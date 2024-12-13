package interpreter

import (
	"context"
	_ "embed"
	"fmt"
	"maps"

	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/common/sql"
	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/internal/engine"
	"github.com/kwilteam/kwil-db/parse"
)

// Interpreter interprets Kwil SQL statements.
type Interpreter struct {
	// TODO: make this thread-safe
	// availableFunctions is a map of function names to their implementations.
	// At this level, it is only scalar SQL functions, and does not include any
	// actions.
	availableFunctions map[string]*executable
	namespaces         map[string]*namespace
	// accessController is used to check if a user has access to a namespace
	accessController *accessController
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
	tables             map[string]*engine.Table
	actions            map[string]*Action

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
func NewInterpreter(ctx context.Context, db sql.DB, log log.Logger) (*Interpreter, error) {
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

	availableFuncs := make(map[string]*executable)

	// we will convert all built-in functions to be executables
	for funcName, impl := range parse.Functions {
		if scalarImpl, ok := impl.(*parse.ScalarFunctionDefinition); ok {
			availableFuncs[funcName] = funcDefToExecutable(funcName, scalarImpl)
		}
	}

	namespaces, err := listNamespaces(ctx, db)
	if err != nil {
		return nil, err
	}

	interpreter := &Interpreter{
		availableFunctions: availableFuncs,
		namespaces:         make(map[string]*namespace),
	}
	for _, ns := range namespaces {
		tables, err := listTablesInNamespace(ctx, db, ns.Name)
		if err != nil {
			return nil, err
		}

		tblMap := make(map[string]*engine.Table)
		for _, tbl := range tables {
			tblMap[tbl.Name] = tbl
		}

		actions, err := listActionsInNamespace(ctx, db, ns.Name)
		if err != nil {
			return nil, err
		}

		// now, we override the built-in functions with the actions
		namespaceFunctions := maps.Clone(availableFuncs)
		actionsMap := make(map[string]*Action)
		for _, action := range actions {
			exec := makeActionToExecutable(action)
			namespaceFunctions[exec.Name] = exec
			actionsMap[action.Name] = action
		}

		interpreter.namespaces[ns.Name] = &namespace{
			tables:             tblMap,
			availableFunctions: namespaceFunctions,
			actions:            actionsMap,
			namespaceType:      ns.Type,
		}
	}

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
		ReturnType: func(v []Value) (*ActionReturn, error) {
			dataTypes := make([]*types.DataType, len(v))
			for i, arg := range v {
				dataTypes[i] = arg.Type()
			}

			retTyp, err := funcDef.ValidateArgsFunc(dataTypes)
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
			iters := 0
			err = query(e.txCtx.Ctx, e.db, pgFormat, []Value{zeroVal}, func() error {
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
		txCtx:            txCtx,
		scope:            newScope(namespace),
		mutatingState:    am.AccessMode() == sql.ReadWrite,
		db:               db,
		namespaces:       i.namespaces,
		accessController: i.accessController,
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
