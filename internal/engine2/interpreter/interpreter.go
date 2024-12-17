package interpreter

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/common/sql"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/internal/engine2"
	"github.com/kwilteam/kwil-db/internal/engine2/parse"
)

// Interpreter interprets Kwil SQL statements.
type Interpreter struct {
	// TODO: make this thread-safe
	namespaces map[string]*namespace
	// accessController is used to check if a user has access to a namespace
	accessController *accessController
	// service is the base application
	service *common.Service
}

// a namespace is a collection of tables and actions.
// It is conceptually equivalent to Postgres's schema, but is given a
// different name to avoid confusion.
type namespace struct {
	// availableFunctions is a map of both built-in functions and user-defined PL/pgSQL functions.
	// When the interpreter planner is created, it will be populated with all built-in functions,
	// and then it will be updated with user-defined functions, effectively allowing users to override
	// some function name with their own implementation. This allows Kwil to add new built-in
	// functions without worrying about breaking user schemas.
	// This will not include aggregate and window functions, as those can only be used in SQL.
	// availableFunctions maps local action names to their execution func.
	availableFunctions map[string]*executable
	tables             map[string]*engine2.Table

	// namespaceType is the type of namespace.
	// It can be user-created, built-in, or extension.
	namespaceType namespaceType
}

type namespaceType string

const (
	namespaceTypeUser      namespaceType = "USER"
	namespaceTypeSystem    namespaceType = "SYSTEM"
	namespaceTypeExtension namespaceType = "EXTENSION"
)

func (n namespaceType) valid() bool {
	switch n {
	case namespaceTypeUser, namespaceTypeSystem, namespaceTypeExtension:
		return true
	default:
		return false
	}
}

// NewInterpreter creates a new interpreter.
// It reads currently stored namespaces and loads them into memory.
func NewInterpreter(ctx context.Context, db sql.DB, service *common.Service) (*Interpreter, error) {
	// we need to check if it is initialized. We will do this by checking if the schema kwild_engine exists
	res, err := db.Execute(ctx, "SELECT EXISTS (SELECT 1 FROM information_schema.schemata WHERE schema_name = 'kwild_engine')")
	if err != nil {
		return nil, err
	}

	switch len(res.Rows) {
	case 0:
		return nil, fmt.Errorf("could not determine if the database is initialized")
	case 1:
		if !res.Rows[0][0].(bool) {
			err = initSQL(ctx, db)
			if err != nil {
				return nil, err
			}
		}
	default:
		return nil, fmt.Errorf("unexpected number of rows returned")
	}

	namespaces, err := listNamespaces(ctx, db)
	if err != nil {
		return nil, err
	}

	interpreter := &Interpreter{
		namespaces: make(map[string]*namespace),
		service:    service,
	}
	for _, ns := range namespaces {
		tables, err := listTablesInNamespace(ctx, db, ns.Name)
		if err != nil {
			return nil, err
		}

		tblMap := make(map[string]*engine2.Table)
		for _, tbl := range tables {
			tblMap[tbl.Name] = tbl
		}

		actions, err := listActionsInNamespace(ctx, db, ns.Name)
		if err != nil {
			return nil, err
		}

		// now, we override the built-in functions with the actions
		namespaceFunctions := copyBuiltinExecutables()
		for _, action := range actions {
			exec := makeActionToExecutable(ns.Name, action)
			namespaceFunctions[exec.Name] = exec
		}

		interpreter.namespaces[ns.Name] = &namespace{
			tables:             tblMap,
			availableFunctions: namespaceFunctions,
			namespaceType:      ns.Type,
		}
	}

	// TODO: set up all extensions

	accessController, err := newAccessController(ctx, db)
	if err != nil {
		return nil, err
	}
	interpreter.accessController = accessController

	return interpreter, nil
}

// funcDefToExecutable converts a function definition to an executable.
func funcDefToExecutable(funcName string, funcDef *parse.ScalarFunctionDefinition) *executable {
	return &executable{
		Name: funcName,
		Func: func(e *executionContext, args []Value, fn func([]Value) error) error {
			//convert args to any
			params := make([]string, len(args))
			argTypes := make([]*types.DataType, len(args))
			for i, arg := range args {
				params[i] = fmt.Sprintf("$%d", i+1)
				argTypes[i] = arg.Type()
			}

			// get the expected return type
			retTyp, err := funcDef.ValidateArgsFunc(argTypes)
			if err != nil {
				return err
			}

			zeroVal, err := NewZeroValue(retTyp)
			if err != nil {
				return err
			}

			// format the function
			pgFormat, err := funcDef.PGFormatFunc(params)
			if err != nil {
				return err
			}

			// execute the query
			// We could avoid a roundtrip here by having go implementating of the function.
			// Since for now we are more concerned about expanding functionality than scalability,
			// we will use the roundtrip.
			iters := 0
			err = query(e.txCtx.Ctx, e.db, "SELECT "+pgFormat+";", []Value{zeroVal}, func() error {
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
		Type: executableTypeFunction,
	}
}

// Execute executes a statement against the database.
func (i *Interpreter) Execute(ctx *common.TxContext, db sql.DB, statement string, fn func([]Value) error) error {
	if fn == nil {
		fn = func([]Value) error { return nil }
	}

	// parse the statement
	ast, err := parse.Parse(statement)
	if err != nil {
		return err
	}

	if len(ast) == 0 {
		return fmt.Errorf("no valid statements provided: %s", statement)
	}

	execCtx, err := i.newExecCtx(ctx, db, defaultNamespace)
	if err != nil {
		return err
	}

	interpPlanner := interpreterPlanner{}

	for _, stmt := range ast {
		err = stmt.Accept(&interpPlanner).(stmtFunc)(execCtx, fn)
		if err != nil {
			return err
		}
	}

	return nil
}

// Call executes an action against the database.
// The resultFn is called with the result of the action, if any.
func (i *Interpreter) Call(ctx *common.TxContext, db sql.DB, namespace, action string, args []any, resultFn func([]Value) error) error {
	if namespace == "" {
		namespace = defaultNamespace
	}

	ns, ok := i.namespaces[namespace]
	if !ok {
		return fmt.Errorf(`namespace "%s" does not exist`, namespace)
	}

	// now we can call the executable. The executable checks that the caller is allowed to call the action
	// (e.g. in case of a private action or owner action)
	exec, ok := ns.availableFunctions[action]
	if !ok {
		// this should never happen
		return fmt.Errorf(`internal bug: action "%s" does not exist in namespace "%s"`, action, namespace)
	}

	switch exec.Type {
	case executableTypeFunction:
		return fmt.Errorf(`%w: action "%s" is a built-in function and cannot be called directly`, ErrCannotCall, action)
	case executableTypeAction, executableTypePrecompile:
		// do nothing, this is what we want
	default:
		return fmt.Errorf(`internal bug: unknown executable type "%s"`, exec.Type)
	}

	argVals := make([]Value, len(args))
	for i, arg := range args {
		val, err := NewValue(arg)
		if err != nil {
			return err
		}

		argVals[i] = val
	}

	execCtx, err := i.newExecCtx(ctx, db, namespace)
	if err != nil {
		return err
	}

	return exec.Func(execCtx, argVals, resultFn)
}

// newExecCtx creates a new execution context.
func (i *Interpreter) newExecCtx(txCtx *common.TxContext, db sql.DB, namespace string) (*executionContext, error) {
	am, ok := db.(sql.AccessModer)
	if !ok {
		return nil, fmt.Errorf("database does not implement AccessModer")
	}

	return &executionContext{
		txCtx:          txCtx,
		scope:          newScope(namespace),
		canMutateState: am.AccessMode() == sql.ReadWrite,
		db:             db,
		interpreter:    i,
	}, nil
}

// SetOwner initializes the interpreter's database by setting the owner.
// It will overwrite the owner if it is already set.
func (i *Interpreter) SetOwner(ctx context.Context, db sql.DB, owner string) error {
	err := i.accessController.SetOwnership(ctx, db, string(owner))
	if err != nil {
		return err
	}
	return nil
}

const (
	defaultNamespace = "main"
)

var builtInExecutables = func() map[string]*executable {
	execs := make(map[string]*executable)
	for funcName, impl := range parse.Functions {
		if scalarImpl, ok := impl.(*parse.ScalarFunctionDefinition); ok {
			execs[funcName] = funcDefToExecutable(funcName, scalarImpl)
		}
	}

	return execs
}()

// copyBuiltinExecutables returns a map of built-in functions to their executables.
func copyBuiltinExecutables() map[string]*executable {
	b := make(map[string]*executable)
	for k, v := range builtInExecutables {
		b[k] = v
	}

	return b
}
