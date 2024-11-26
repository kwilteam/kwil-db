package interpreter

import (
	"context"
	"fmt"
	"maps"

	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/common/sql"
	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/internal/engine"
	"github.com/kwilteam/kwil-db/parse"
)

const defaultNamespace = "main"

// TODO: move this into the SQL package once pg supports this
type DB interface {
	sql.TxMaker
	Execute(ctx context.Context, stmt string, args ...any) (Rows, error)
}

type Rows interface {
	Next() bool
	Scan(dest ...interface{}) error
	Close() error
	Err() error
}

// TODO: delete this
// simply using this for planning
type IInterpeter interface {
	// ExecuteRaw executes a raw Postgres SQL statement against the database.
	ExecuteRaw(ctx context.Context, db sql.DB, statement string) (Rows, error)
	// Execute executes Kwil SQL statements against the database.
	ExecuteAdmin(ctx context.Context, db DB, statement string) (Rows, error)
	// Call executes an action against the database.
	Call(ctx *common.TxContext, db DB, action string, args []any) (Rows, error)
}

// Interpreter interprets Kwil SQL statements.
type Interpreter struct {
	// TODO: make this thread-safe
	// availableFunctions is a map of function names to their implementations.
	// At this level, it is only scalar SQL functions, and does not include any
	// actions.
	availableFunctions map[string]*executable
	namespaces         map[string]*namespace
}

// a namespace is a collection of tables and actions.
// It is conceptually equivalent to Postgres's schema, but is given a
// different name to avoid confusion.
type namespace struct {
	// owner is the public key of the owner of the namespace
	owner              []byte
	availableFunctions map[string]*executable
	tables             map[string]*engine.Table
	actions            map[string]*Action
}

// NewInterpreter creates a new interpreter.
// It reads currently stored namespaces and loads them into memory.
func NewInterpreter(ctx context.Context, db sql.DB, log log.Logger) (*Interpreter, error) {
	availableFuncs := make(map[string]*executable)

	// we will convert all built-in functions to be executables
	for funcName, impl := range parse.Functions {
		if scalarImpl, ok := impl.(*parse.ScalarFunctionDefinition); ok {
			funcName := funcName // avoid shadowing
			exec := &executable{
				Name: funcName,
				ReturnType: func(v []Value) (*ActionReturn, error) {
					dataTypes := make([]*types.DataType, len(v))
					for i, arg := range v {
						dataTypes[i] = arg.Type()
					}

					retTyp, err := scalarImpl.ValidateArgsFunc(dataTypes)
					if err != nil {
						return nil, err
					}

					return &ActionReturn{
						IsTable: false,
						Fields:  []*NamedType{{Name: funcName, Type: retTyp}},
					}, nil
				},
				Func: func(e *executionContext, args []Value, fn func([]Value) error) error {
					//convert args to any
					params := make([]string, len(args))
					argTypes := make([]*types.DataType, len(args))
					for i, arg := range args {
						params[i] = fmt.Sprintf("$%d", i+1)
						argTypes[i] = arg.Type()
					}

					// get the expected return type
					retTyp, err := scalarImpl.ValidateArgsFunc(argTypes)
					if err != nil {
						return err
					}

					zeroVal, err := NewZeroValue(retTyp)
					if err != nil {
						return err
					}

					// format the function
					pgFormat, err := scalarImpl.PGFormatFunc(params)
					if err != nil {
						return err
					}

					// execute the query
					iters := 0
					err = query(ctx, db, pgFormat, []Value{zeroVal}, func() error {
						iters++
						return nil
					}, args)
					if err != nil {
						return err
					}
					if iters != 1 {
						return fmt.Errorf("expected 1 row, got %d", iters)
					}

					return fn([]Value{zeroVal})
				},
			}

			availableFuncs[funcName] = exec
		}
	}

	namespaces, err := listNamespaceTables(ctx, db)
	if err != nil {
		return nil, err
	}
	actions, err := loadActions(ctx, db)
	if err != nil {
		return nil, err
	}
	namespaceOwners, err := getNamespaceOwners(ctx, db)
	if err != nil {
		return nil, err
	}

	interpreter := &Interpreter{
		availableFunctions: availableFuncs,
		namespaces:         make(map[string]*namespace),
	}
	for name, tables := range namespaces {
		actions, ok := actions[name]
		if !ok {
			// can be empty if no actions are defined for the namespace
			actions = make(map[string]*Action)
		}

		owner, ok := namespaceOwners[name]
		if !ok {
			owner = []byte{}
		}

		// now, we override the built-in functions with the actions
		namespaceFunctions := maps.Clone(availableFuncs)
		for _, action := range actions {
			exec := makeActionToExecutable(owner, action)
			namespaceFunctions[exec.Name] = exec
		}

		interpreter.namespaces[name] = &namespace{
			owner:              owner,
			tables:             tables,
			availableFunctions: namespaceFunctions,
			actions:            actions,
		}
	}

	return interpreter, nil
}

// Call executes an action against the database.
// The resultFn is called with the result of the action, if any.
func (i *Interpreter) Call(ctx *common.TxContext, db sql.DB, namespace, action string, args []any, resultFn func([]Value) error) error {
	accessModer, ok := db.(sql.AccessModer)
	if !ok {
		return fmt.Errorf("could not determine access mode")
	}

	if namespace == "" {
		namespace = defaultNamespace
	}

	ns, ok := i.namespaces[namespace]
	if !ok {
		return fmt.Errorf(`namespace "%s" does not exist`, namespace)
	}

	// ensure that the user isnt calling a built-in function
	act, ok := ns.actions[action]
	if !ok {
		return fmt.Errorf(`action "%s" does not exist in namespace "%s"`, action, namespace)
	}

	if !act.IsView() && accessModer.AccessMode() == sql.ReadOnly {
		return fmt.Errorf(`cannot call mutable action "%s" in read-only mode`, action)
	}

	// now we can call the executable. The executable checks that the caller is allowed to call the action
	// (e.g. in case of a private action or owner action)
	exec, ok := ns.availableFunctions[action]
	if !ok {
		// this should never happen
		return fmt.Errorf(`internal bug: action "%s" does not exist in namespace "%s"`, action, namespace)
	}

	argVals := make([]Value, len(args))
	for i, arg := range args {
		val, err := NewValue(arg)
		if err != nil {
			return err
		}

		argVals[i] = val
	}

	return exec.Func(&executionContext{
		txCtx:              ctx,
		scope:              newScope(namespace),
		availableFunctions: ns.availableFunctions,
		// if we can write, then we are in execution, and should be deterministic
		mutatingState: accessModer.AccessMode() == sql.ReadWrite,
		db:            db,
		interpreter:   i,
	}, argVals, resultFn)
}

// Initialize initializes the interpreter.
// It takes a single slice of bytes, which should be a public key.
// This public key will be made the initial owner of the "main" namespace.
func (i *Interpreter) Initialize(ctx context.Context, db sql.DB, owner []byte) error {
	if len(owner) == 0 {
		return fmt.Errorf("owner must not be empty for initialization")
	}

	_, ok := i.namespaces[defaultNamespace]
	if ok {
		return fmt.Errorf("default namespace has already been initialized")
	}

	// create the namespace
	i.namespaces[defaultNamespace] = &namespace{}

	return i.CreateNamespace(ctx, db, defaultNamespace, owner)
}

// CreateNamespace creates a new namespace, owned by the given public key.
// It is only meant to be called from within extensions, and not from the SQL layer.
// TODO: this should probably be replaced with a general SQL interface, where they can
// do "CREATE NAMESPACE" and "CREATE TABLE" statements.
func (i *Interpreter) CreateNamespace(ctx context.Context, db sql.DB, name string, owner []byte) error {
	if len(owner) == 0 {
		return fmt.Errorf("owner must not be empty for namespace creation")
	}

	_, ok := i.namespaces[name]
	if ok {
		return fmt.Errorf(`namespace "%s" already exists`, name)
	}

	// insert the namespace into the database
	err := createUserNamespace(ctx, db, name, owner)
	if err != nil {
		return err
	}

	// create the namespace
	i.namespaces[name] = &namespace{
		owner:              owner,
		availableFunctions: maps.Clone(i.availableFunctions),
		tables:             make(map[string]*engine.Table),
		actions:            make(map[string]*Action),
	}

	return nil
}
