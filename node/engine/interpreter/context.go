package interpreter

import (
	"fmt"
	"strings"

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
	// engineCtx is the transaction context.
	engineCtx *common.EngineContext
	// scope is the current scope.
	scope *scopeContext
	// canMutateState is true if the execution is capable of mutating state.
	// If true, it must also be deterministic.
	canMutateState bool
	// db is the database to execute against.
	db sql.DB
	// interpreter is the interpreter that created this execution context.
	interpreter *baseInterpreter
	// logs are the logs that have been generated.
	// it is a pointer to a slice to allow for child scopes to allocate
	// space for more logs on the parent.
	logs *[]string
	// queryActive is true if a query is currently active.
	// This is used to prevent nested queries, which can cause
	// a deadlock or unexpected behavior.
	queryActive bool
}

// subscope creates a new subscope execution context.
// A subscope allows for a new context to exist without
// modifying the original. Unlike a child, a subscope does not
// inherit the parent's variables.
// It is used for when an action calls another action / extension method.
func (e *executionContext) subscope(namespace string) *executionContext {
	return &executionContext{
		engineCtx:      e.engineCtx,
		scope:          newScope(namespace),
		canMutateState: e.canMutateState,
		db:             e.db,
		interpreter:    e.interpreter,
		logs:           e.logs,
	}
}

// checkPrivilege checks that the current user has a privilege,
// and returns an error if they do not.
func (e *executionContext) checkPrivilege(priv privilege) error {
	if e.engineCtx.OverrideAuthz {
		return nil
	}

	if !e.interpreter.accessController.HasPrivilege(e.engineCtx.TxContext.Caller, &e.scope.namespace, priv) {
		return fmt.Errorf(`%w %s on namespace "%s"`, engine.ErrDoesNotHavePrivilege, priv, e.scope.namespace)
	}

	return nil
}

// isOwner checks if the current user is the owner of the namespace.
func (e *executionContext) isOwner() bool {
	return e.interpreter.accessController.IsOwner(e.engineCtx.TxContext.Caller)
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
		return nil, fmt.Errorf(`%w: "%s"`, engine.ErrNamespaceNotFound, namespace)
	}

	return ns, nil
}

// getTable gets a table from the interpreter.
// It can optionally be given a namespace to search in.
// If the namespace is empty, it will search the current namespace.
func (e *executionContext) getTable(namespace, tableName string) (*engine.Table, error) {
	ns, err := e.getNamespace(namespace)
	if err != nil {
		return nil, err
	}

	table, ok := ns.tables[tableName]
	if !ok {
		return nil, fmt.Errorf(`%w: table "%s" not found in namespace "%s"`, engine.ErrUnknownTable, tableName, namespace)
	}

	return table, nil
}

// checkNamespaceMutatbility checks if the current namespace is mutable.
// It allows extensions to be overridden, but not the main namespace.
// It does not check for drops; these should be handled separately.
// These rules are not handled in the accessController because they are always
// enforced, regardless of the roles and privileges of the caller.
func (e *executionContext) checkNamespaceMutatbility() error {
	if e.scope.namespace == infoNamespace {
		return engine.ErrCannotMutateInfoNamespace
	}

	ns2, err := e.getNamespace(e.scope.namespace)
	if err != nil {
		return err
	}

	if ns2.namespaceType == namespaceTypeExtension && !e.engineCtx.OverrideAuthz {
		return fmt.Errorf(`%w: "%s"`, engine.ErrCannotMutateExtension, e.scope.namespace)
	}

	return nil
}

// query executes a query.
// It will parse the SQL, create a logical plan, and execute the query.
func (e *executionContext) query(sql string, fn func(*row) error) error {
	if e.queryActive {
		return engine.ErrQueryActive
	}
	e.queryActive = true
	defer func() { e.queryActive = false }()

	res, err := parse.Parse(sql)
	if err != nil {
		return fmt.Errorf("%w: %w", engine.ErrParse, err)
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
		func(varName string) (dataType *types.DataType, err2 error) {
			val, err := e.getVariable(varName)
			if err != nil {
				return nil, err
			}

			// if it is a record, then return nil
			if _, ok := val.(*precompiles.RecordValue); ok {
				return nil, engine.ErrUnknownVariable
			}

			return val.Type(), nil
		},
		func(objName string) (obj map[string]*types.DataType, err2 error) {
			val, err := e.getVariable(objName)
			if err != nil {
				return nil, err
			}

			if rec, ok := val.(*precompiles.RecordValue); ok {
				dt := make(map[string]*types.DataType)
				for _, field := range rec.Order {
					dt[field] = rec.Fields[field].Type()
				}

				return dt, nil
			}

			return nil, engine.ErrUnknownVariable
		},
		e.canMutateState,
		e.scope.namespace,
	)
	if err != nil {
		return fmt.Errorf("%w: %w", engine.ErrQueryPlanner, err)
	}

	generatedSQL, params, err := pggenerate.GenerateSQL(sqlStmt, e.scope.namespace)
	if err != nil {
		return fmt.Errorf("%w: %w", engine.ErrPGGen, err)
	}

	// get the params we will pass
	var args []precompiles.Value
	for _, param := range params {
		val, err := e.getVariable(param)
		if err != nil {
			return err
		}

		args = append(args, val)
	}

	// get the scan values as well:
	var scanValues []precompiles.Value
	for _, field := range analyzed.Plan.Relation().Fields {
		scalar, err := field.Scalar()
		if err != nil {
			return err
		}

		zVal, err := precompiles.NewZeroValue(scalar)
		if err != nil {
			return err
		}

		scanValues = append(scanValues, zVal)
	}

	cols := make([]string, len(analyzed.Plan.Relation().Fields))
	for i, field := range analyzed.Plan.Relation().Fields {
		cols[i] = field.Name
	}

	return query(e.engineCtx.TxContext.Ctx, e.db, generatedSQL, scanValues, func() error {
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
// a user-defined Kwil action, or a precompile method.
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

type execFunc func(exec *executionContext, args []precompiles.Value, returnFn resultFunc) error

// setVariable sets a variable in the current scope.
// It will allocate the variable if it does not exist.
// if we are setting a variable that was defined in an outer scope,
// it will overwrite the variable in the outer scope.
func (e *executionContext) setVariable(name string, value precompiles.Value) error {
	if strings.HasPrefix(name, "@") {
		return fmt.Errorf("%w: cannot set system variable %s", engine.ErrInvalidVariable, name)
	}

	oldVal, foundScope, found := getVarFromScope(name, e.scope)
	if !found {
		return e.allocateVariable(name, value)
	}

	// if the new variable is null, we should set the old variable to null
	if value.Type().EqualsStrict(types.UnknownType) && value.Null() {
		// set it to null
		newVal, err := precompiles.MakeNull(oldVal.Type())
		if err != nil {
			return err
		}
		foundScope.variables[name] = newVal
		return nil
	}

	if !oldVal.Type().EqualsStrict(value.Type()) {
		return fmt.Errorf("%w: cannot assign variable of type %s to existing variable of type %s", engine.ErrType, value.Type(), oldVal.Type())
	}

	foundScope.variables[name] = value
	return nil
}

// allocateVariable allocates a variable in the current scope.
func (e *executionContext) allocateVariable(name string, value precompiles.Value) error {
	if value.Type().EqualsStrict(types.UnknownType) && value.Null() {
		return fmt.Errorf("%w: cannot allocate untyped null value", engine.ErrInvalidNull)
	}

	_, ok := e.scope.variables[name]
	if ok {
		return fmt.Errorf(`variable "%s" already exists`, name)
	}

	e.scope.variables[name] = value
	return nil
}

// allocateNullVariable allocates a null variable in the current scope.
// It requires a valid type to allocate the variable.
func (e *executionContext) allocateNullVariable(name string, dataType *types.DataType) error {
	nv, err := precompiles.MakeNull(dataType)
	if err != nil {
		return err
	}

	return e.allocateVariable(name, nv)
}

// getVariable gets a variable from the current scope.
// It searches the parent scopes if the variable is not found.
// It returns the value and a boolean indicating if the variable was found.
func (e *executionContext) getVariable(name string) (precompiles.Value, error) {
	if len(name) == 0 {
		return nil, fmt.Errorf("%w: variable name is empty", engine.ErrInvalidVariable)
	}

	switch name[0] {
	case '$':
		v, _, f := getVarFromScope(name, e.scope)
		if !f {
			return nil, fmt.Errorf("%w: %s", engine.ErrUnknownVariable, name)
		}
		return v, nil
	case '@':
		switch name[1:] {
		case "caller":
			if e.engineCtx.InvalidTxCtx {
				return nil, engine.ErrInvalidTxCtx
			}
			return precompiles.MakeText(e.engineCtx.TxContext.Caller), nil
		case "txid":
			if e.engineCtx.InvalidTxCtx {
				return nil, engine.ErrInvalidTxCtx
			}
			return precompiles.MakeText(e.engineCtx.TxContext.TxID), nil
		case "signer":
			if e.engineCtx.InvalidTxCtx {
				return nil, engine.ErrInvalidTxCtx
			}
			return precompiles.MakeBlob(e.engineCtx.TxContext.Signer), nil
		case "height":
			if e.engineCtx.InvalidTxCtx {
				return nil, engine.ErrInvalidTxCtx
			}
			return precompiles.MakeInt8(e.engineCtx.TxContext.BlockContext.Height), nil
		case "foreign_caller":
			if e.scope.parent != nil {
				return precompiles.MakeText(e.scope.parent.namespace), nil
			} else {
				return precompiles.MakeText(""), nil
			}
		case "block_timestamp":
			if e.engineCtx.InvalidTxCtx {
				return nil, engine.ErrInvalidTxCtx
			}
			return precompiles.MakeInt8(e.engineCtx.TxContext.BlockContext.Timestamp), nil
		case "authenticator":
			if e.engineCtx.InvalidTxCtx {
				return nil, engine.ErrInvalidTxCtx
			}
			return precompiles.MakeText(e.engineCtx.TxContext.Authenticator), nil
		default:
			return nil, fmt.Errorf("%w: %s", engine.ErrInvalidVariable, name)
		}
	default:
		return nil, fmt.Errorf("%w: %s", engine.ErrInvalidVariable, name)
	}
}

// reloadTables reloads the cached tables from the database for the current namespace.
func (e *executionContext) reloadTables() error {
	tables, err := listTablesInNamespace(e.engineCtx.TxContext.Ctx, e.db, e.scope.namespace)
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
// It should always be called BEFORE you are in the new action's scope.
func (e *executionContext) canExecute(newNamespace, actionName string, modifiers precompiles.Modifiers) error {
	// if the ctx cannot mutate state and the action is not a view (and thus might try to mutate state),
	// then return an error
	if !modifiers.Has(precompiles.VIEW) && !e.canMutateState {
		return fmt.Errorf(`%w: action "%s" requires a writer connection`, engine.ErrCannotMutateState, actionName)
	}

	// the VIEW check protects against state being modified outside of consensus. This is critical to protect
	// against consensus errors. Every other check enforces user-defined rules, and thus can be overridden by
	// extensions.
	// We only pass other checks if this is the top-level call, since we still want typical checks like private
	// and system to apply. We simply want the override to be able to directly call private and system actions.
	if e.engineCtx.OverrideAuthz && e.scope.isTopLevel {
		return nil
	}

	// if the action is private and either:
	// - the calling namespace is not the same as the new namespace
	// - the action is top level
	// then return an error
	if modifiers.Has(precompiles.PRIVATE) && (e.scope.namespace != newNamespace || e.scope.isTopLevel) {
		return fmt.Errorf("%w: action %s is private", engine.ErrActionPrivate, actionName)
	}

	// if it is system-only, then it must be within a subscope
	if modifiers.Has(precompiles.SYSTEM) && e.scope.isTopLevel {
		return fmt.Errorf("%w: action %s is system-only", engine.ErrActionSystemOnly, actionName)
	}

	// if the action is owner only, then check if the user is the owner
	if modifiers.Has(precompiles.OWNER) && !e.interpreter.accessController.IsOwner(e.engineCtx.TxContext.Caller) {
		return fmt.Errorf("%w: action %s can only be executed by the owner", engine.ErrActionOwnerOnly, actionName)
	}

	return nil
}

func (e *executionContext) app() *common.App {
	// we need to wait until we make changes to the engine interface for extensions before we can implement this
	return &common.App{
		Service: e.interpreter.service,
		DB:      e.db,
		Engine: &recursiveInterpreter{
			i:    e.interpreter,
			logs: e.logs,
		},
		Accounts:   e.interpreter.accounts,
		Validators: e.interpreter.validators,
	}
}

// getVarFromScope recursively searches the scopes for a variable.
// It returns the value, as well as the scope it was found in.
func getVarFromScope(variable string, scope *scopeContext) (precompiles.Value, *scopeContext, bool) {
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
	variables map[string]precompiles.Value
	// namespace is the current namespace.
	namespace string
	// isTopLevel is true if this is the top level scope.
	// A scope can not be top level and also not have a parent.
	isTopLevel bool
}

// newScope creates a new scope.
func newScope(namespace string) *scopeContext {
	return &scopeContext{
		variables: make(map[string]precompiles.Value),
		namespace: namespace,
	}
}

// child creates a new sub-scope, which has access to the parent scope.
// It is used for if blocks and for loops, which can access the outer
// blocks variables and modify them, but new variables created are not
// accessible outside of the block.
func (s *scopeContext) child() {
	s.parent = &scopeContext{
		parent:    s.parent,
		variables: s.variables,
		namespace: s.namespace,
	}
	s.variables = make(map[string]precompiles.Value)
	s.namespace = s.parent.namespace
}

// popScope pops the current scope, returning the parent scope.
func (s *scopeContext) popScope() {
	if s.parent == nil {
		panic("cannot pop root scope")
	}

	*s = *s.parent
}
