package interpreter

import (
	"fmt"

	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/extensions/precompiles"
	"github.com/kwilteam/kwil-db/node/engine"
	"github.com/kwilteam/kwil-db/node/engine/parse"
	pggenerate "github.com/kwilteam/kwil-db/node/engine/pg_generate"
	"github.com/kwilteam/kwil-db/node/engine/planner/logical"
	"github.com/kwilteam/kwil-db/node/types/sql"
)

// executionContext is the context of the entire execution.
type executionContext struct {
	// txCtx is the transaction context.
	txCtx *common.TxContext
	// scope is the current scope.
	scope *scopeContext
	// canMutateState is true if the execution is capable of mutating state.
	// If true, it must also be deterministic.
	canMutateState bool
	// db is the database to execute against.
	db sql.DB
	// interpreter is the interpreter that created this execution context.
	interpreter *BaseInterpreter
	// logs are the logs that have been generated.
	logs []string
}

// checkPrivilege checks that the current user has a privilege,
// and returns an error if they do not.
func (e *executionContext) checkPrivilege(priv privilege) error {
	if !e.interpreter.accessController.HasPrivilege(e.txCtx.Caller, &e.scope.namespace, priv) {
		return fmt.Errorf("%w: %s", ErrDoesNotHavePriv, priv)
	}

	return nil
}

// getNamespace gets the specified namespace.
// If the namespace does not exist, it will return an error.
// If the namespace is empty, it will return the current namespace.
func (e *executionContext) getNamespace(namespace string) (*namespace, error) {
	if namespace == "" {
		namespace = e.scope.namespace
	}

	ns, ok := e.interpreter.namespaces[namespace]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrNamespaceNotFound, namespace)
	}

	return ns, nil
}

// getTable gets a table from the interpreter.
// It can optionally be given a namespace to search in.
// If the namespace is empty, it will search the current namespace.
func (e *executionContext) getTable(namespace, tableName string) (*engine.Table, bool) {
	ns, err := e.getNamespace(namespace)
	if err != nil {
		panic(err) // we should never hit an error here
	}

	table, ok := ns.tables[tableName]
	return table, ok
}

// query executes a query.
// It will parse the SQL, create a logical plan, and execute the query.
func (e *executionContext) query(sql string, fn func(*row) error) error {
	res, err := parse.Parse(sql)
	if err != nil {
		return err
	}

	if len(res) != 1 {
		// this is an node bug b/c `query` is only called with a single statement
		// from the interpreter
		return fmt.Errorf("node bug: expected exactly 1 statement, got %d", len(res))
	}

	sqlStmt, ok := res[0].(*parse.SQLStatement)
	if !ok {
		return fmt.Errorf("node bug: expected *parse.SQLStatement, got %T", res[0])
	}

	// create a logical plan. This will make the query deterministic (if necessary),
	// as well as tell us what the return types will be.
	analyzed, err := logical.CreateLogicalPlan(
		sqlStmt,
		e.getTable,
		func(varName string) (dataType *types.DataType, found bool) {
			val, found := e.getVariable(varName)
			if !found {
				return nil, false
			}

			// if it is a record, then return nil
			if _, ok := val.(*RecordValue); ok {
				return nil, false
			}

			return val.Type(), true
		},
		func(objName string) (obj map[string]*types.DataType, found bool) {
			val, found := e.getVariable(objName)
			if !found {
				return nil, false
			}

			if rec, ok := val.(*RecordValue); ok {
				dt := make(map[string]*types.DataType)
				for _, field := range rec.Order {
					dt[field] = rec.Fields[field].Type()
				}

				return dt, true
			}

			return nil, false
		},
		e.canMutateState,
		e.scope.namespace,
	)
	if err != nil {
		return err
	}

	generatedSQL, params, err := pggenerate.GenerateSQL(sqlStmt, e.scope.namespace)
	if err != nil {
		return err
	}

	// get the params we will pass
	var args []Value
	for _, param := range params {
		val, found := e.getVariable(param)
		if !found {
			return fmt.Errorf("%w: %s", ErrVariableNotFound, param)
		}

		args = append(args, val)
	}

	// get the scan values as well:
	var scanValues []Value
	for _, field := range analyzed.Plan.Relation().Fields {
		scalar, err := field.Scalar()
		if err != nil {
			return err
		}

		zVal, err := NewZeroValue(scalar)
		if err != nil {
			return err
		}

		scanValues = append(scanValues, zVal)
	}

	cols := make([]string, len(analyzed.Plan.Relation().Fields))
	for i, field := range analyzed.Plan.Relation().Fields {
		cols[i] = field.Name
	}

	return query(e.txCtx.Ctx, e.db, generatedSQL, scanValues, func() error {
		if len(scanValues) != len(cols) {
			// should never happen, but just in case
			return fmt.Errorf("node bug: scan values and columns are not the same length")
		}

		return fn(&row{
			columns: cols,
			Values:  scanValues,
		})
	}, args)
}

// executable is the interface and function to call a built-in Postgres function,
// a user-defined Postgres procedure, or a user-defined Kwil action.
type executable struct {
	// Name is the name of the function.
	Name string
	// Func is a function that executes the function.
	Func execFunc
	// Type is the type of the executable.
	Type executableType
}

type executableType string

const (
	// executableTypeFunction is a built-in Postgres function.
	executableTypeFunction executableType = "function"
	// executableTypeAction is a user-defined Kwil action.
	executableTypeAction executableType = "action"
	// executableTypePrecompile is a precompiled extension.
	executableTypePrecompile executableType = "precompile"
)

type execFunc func(exec *executionContext, args []Value, returnFn resultFunc) error

// newScope creates a new scope.
func newScope(namespace string) *scopeContext {
	return &scopeContext{
		variables: make(map[string]Value),
		namespace: namespace,
	}
}

// subScope creates a new sub-scope, which has access to the parent scope.
func (s *scopeContext) subScope() *scopeContext {
	return &scopeContext{
		parent:    s,
		variables: make(map[string]Value),
	}
}

// setVariable sets a variable in the current scope.
// It will allocate the variable if it does not exist.
// if we are setting a variable that was defined in an outer scope,
// it will overwrite the variable in the outer scope.
func (e *executionContext) setVariable(name string, value Value) error {
	_, foundScope, found := getVarFromScope(name, e.scope)
	if !found {
		return e.allocateVariable(name, value)
	}

	foundScope.variables[name] = value
	return nil
}

// allocateVariable allocates a variable in the current scope.
func (e *executionContext) allocateVariable(name string, value Value) error {
	_, ok := e.scope.variables[name]
	if ok {
		return fmt.Errorf(`variable "%s" already exists`, name)
	}

	e.scope.variables[name] = value
	return nil
}

// getVariable gets a variable from the current scope.
// It searches the parent scopes if the variable is not found.
// It returns the value and a boolean indicating if the variable was found.
func (e *executionContext) getVariable(name string) (Value, bool) {
	if len(name) == 0 {
		return nil, false
	}

	switch name[0] {
	case '$':
		v, _, f := getVarFromScope(name, e.scope)
		return v, f
	case '@':
		switch name[1:] {
		case "caller":
			return newText(e.txCtx.Caller), true
		case "txid":
			return newText(e.txCtx.TxID), true
		case "signer":
			return newBlob(e.txCtx.Signer), true
		case "height":
			return newInt(e.txCtx.BlockContext.Height), true
		case "foreign_caller":
			if e.scope.parent != nil {
				return newText(e.scope.parent.namespace), true
			} else {
				return newText(""), true
			}
		case "block_timestamp":
			return newInt(e.txCtx.BlockContext.Timestamp), true
		case "authenticator":
			return newText(e.txCtx.Authenticator), true
		}
	}

	return nil, false
}

// reloadTables reloads the cached tables from the database for the current namespace.
func (e *executionContext) reloadTables() error {
	tables, err := listTablesInNamespace(e.txCtx.Ctx, e.db, e.scope.namespace)
	if err != nil {
		return err
	}

	ns := e.interpreter.namespaces[e.scope.namespace]

	ns.tables = make(map[string]*engine.Table)
	for _, table := range tables {
		ns.tables[table.Name] = table
	}

	return nil
}

// canExecute checks if the context can execute the action.
// It returns an error if it cannot.
func (e *executionContext) canExecute(namespace string, name string, modifiers precompiles.Modifiers) error {
	// if the ctx cannot mutate state and the action is not a view (and thus might try to mutate state),
	// then return an error
	if !modifiers.Has(precompiles.VIEW) && !e.canMutateState {
		return fmt.Errorf("%w: cannot execute action %s in a read-only transaction", ErrActionMutatesState, name)
	}

	// if the action is private, then the calling namespace must be the same as the action's namespace
	if modifiers.Has(precompiles.PRIVATE) && e.scope.namespace != namespace {
		return fmt.Errorf("%w: action %s is private", ErrActionPrivate, name)
	}

	// if it is system-only, then this must not be called without an outer scope
	if modifiers.Has(precompiles.SYSTEM) && e.scope.parent == nil {
		return fmt.Errorf("%w: action %s is system-only", ErrSystemOnly, name)
	}

	// if the action is owner only, then check if the user is the owner
	if modifiers.Has(precompiles.OWNER) && !e.interpreter.accessController.IsOwner(e.txCtx.Caller) {
		return fmt.Errorf("%w: action %s can only be executed by the owner", ErrActionOwnerOnly, name)
	}

	return nil
}

func (e *executionContext) app() *common.App {
	// we need to wait until we make changes to the engine interface for extensions before we can implement this
	return &common.App{
		Service: e.interpreter.service,
		DB:      e.db,
		Engine:  e.interpreter,
	}
}

// getVarFromScope recursively searches the scopes for a variable.
// It returns the value, as well as the scope it was found in.
func getVarFromScope(variable string, scope *scopeContext) (Value, *scopeContext, bool) {
	if v, ok := scope.variables[variable]; ok {
		return v, scope, true
	}
	if scope.parent == nil {
		return nil, nil, false
	}
	return getVarFromScope(variable, scope.parent)
}

// scopeContext is the context for the current block of code.
type scopeContext struct {
	// parent is the parent scope.
	// if the parent is nil, this is the root
	parent *scopeContext
	// variables are the variables stored in memory.
	variables map[string]Value
	// namespace is the current namespace.
	namespace string
}
