package interpreter

import (
	"fmt"

	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/common/sql"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/internal/engine"
	pggenerate "github.com/kwilteam/kwil-db/internal/engine/pg_generate"
	"github.com/kwilteam/kwil-db/internal/engine/planner/logical"
	"github.com/kwilteam/kwil-db/parse"
)

// executionContext is the context of the entire execution.
type executionContext struct {
	// txCtx is the transaction context.
	txCtx *common.TxContext
	// scope is the current scope.
	scope *scopeContext
	// mutatingState is true if the execution is capable of mutating state.
	// If true, it must also be deterministic.
	mutatingState bool
	// namespaces is a map of all namespaces.
	namespaces map[string]*namespace
	// db is the database to execute against.
	db sql.DB
	// accessController holds information about roles and privileges
	accessController *accessController
}

// checkPrivilege checks that the current user has a privilege,
// and returns an error if they do not.
func (e *executionContext) checkPrivilege(priv privilege) error {
	if !e.accessController.HasPrivilege(e.txCtx.Caller, &e.scope.namespace, priv) {
		return fmt.Errorf("%w: %s", ErrDoesNotHavePriv, priv)
	}

	return nil
}

// currentNamespace gets the current namespace.
func (e *executionContext) currentNamespace() *namespace {
	return e.namespaces[e.scope.namespace]
}

// getNamespace gets the specified namespace.
// If the namespace does not exist, it will return an error.
// If the namespace is empty, it will return the current namespace.
func (e *executionContext) getNamespace(namespace string) (*namespace, error) {
	if namespace == "" {
		namespace = e.scope.namespace
	}

	ns, ok := e.namespaces[namespace]
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
func (e *executionContext) query(sql string, fn func(*RecordValue) error) error {
	res, err := parse.Parse(sql)
	if err != nil {
		return err
	}

	if len(res) != 1 {
		// this is an internal bug b/c `query` is only called with a single statement
		// from the interpreter
		return fmt.Errorf("internal bug: expected exactly 1 statement, got %d", len(res))
	}

	sqlStmt, ok := res[0].(*parse.SQLStatement)
	if !ok {
		return fmt.Errorf("internal bug: expected *parse.SQLStatement, got %T", res[0])
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
		e.mutatingState,
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

	return query(e.txCtx.Ctx, e.db, generatedSQL, scanValues, func() error {
		record := newRecordValue()
		for i, field := range analyzed.Plan.Relation().Fields {
			if field.Name == "" {
				continue
			}

			err = record.AddValue(field.Name, scanValues[i])
			if err != nil {
				return err
			}
		}

		return fn(record)
	}, args)
}

// executable is the interface and function to call a built-in Postgres function,
// a user-defined Postgres procedure, or a user-defined Kwil action.
type executable struct {
	// Name is the name of the function.
	Name string
	// ReturnType is a function that takes the arguments for the function and returns the return type.
	// The function can return a nil error AND a nil return type if the function does not return anything.
	ReturnType returnTypeFunc
	// Func is a function that executes the function.
	Func execFunc
}

type execFunc func(exec *executionContext, args []Value, returnFn func([]Value) error) error

// returnTypeFunc is a function that validates the arguments of a function and gives
// the return type of the function based on the arguments.
// The return type can be nil if the function does not return anything.
type returnTypeFunc func([]Value) (*ActionReturn, error)

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
			return &TextValue{Val: e.txCtx.Caller}, true
		case "txid":
			return &TextValue{Val: e.txCtx.TxID}, true
		case "signer":
			return &BlobValue{Val: e.txCtx.Signer}, true
		case "height":
			return &IntValue{Val: e.txCtx.BlockContext.Height}, true
		case "foreign_caller":
			if e.scope.parent != nil {
				return &TextValue{Val: e.scope.parent.namespace}, true
			} else {
				return &TextValue{Val: ""}, true
			}
		case "block_timestamp":
			return &IntValue{Val: e.txCtx.BlockContext.Timestamp}, true
		case "authenticator":
			return &TextValue{Val: e.txCtx.Authenticator}, true
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

	ns := e.namespaces[e.scope.namespace]

	ns.tables = make(map[string]*engine.Table)
	for _, table := range tables {
		ns.tables[table.Name] = table
	}

	return nil
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
