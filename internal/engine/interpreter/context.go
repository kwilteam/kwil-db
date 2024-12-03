package interpreter

import (
	"context"
	"fmt"

	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/common/sql"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/internal/engine"
	"github.com/kwilteam/kwil-db/internal/engine/generate"
	"github.com/kwilteam/kwil-db/internal/engine/planner/logical"
	"github.com/kwilteam/kwil-db/parse"
)

// executionContext is the context of the entire execution.
type executionContext struct {
	// txCtx is the transaction context.
	txCtx *common.TxContext
	// scope is the current scope.
	scope *scopeContext
	// availableFunctions is a map of both built-in functions and user-defined PL/pgSQL functions.
	// When the interpreter planner is created, it will be populated with all built-in functions,
	// and then it will be updated with user-defined functions, effectively allowing users to override
	// some function name with their own implementation. This allows Kwil to add new built-in
	// functions without worrying about breaking user schemas.
	// This will not include aggregate and window functions, as those can only be used in SQL.
	// availableFunctions maps local action names to their execution func.
	availableFunctions map[string]*executable
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

// hasPrivilege checks if the current user has a privilege.
func (e *executionContext) hasPrivilege(priv privilege) bool {
	return e.accessController.HasPrivilege(e.txCtx.Caller, &e.scope.namespace, priv)
}

// getTable gets a table from the interpreter.
// It can optionally be given a namespace to search in.
// If the namespace is empty, it will search the current namespace.
func (e *executionContext) getTable(namespace, tableName string) (*engine.Table, bool) {
	if namespace == "" {
		namespace = e.scope.namespace
	}

	ns, ok := e.namespaces[namespace]
	if !ok {
		return nil, false
	}

	table, ok := ns.tables[tableName]
	return table, ok
}

// plan creates a logical plan from the AST.
// It will also rewrite the SQL statement to be deterministic if necessary.
func (e *executionContext) plan(sql string) (*logical.AnalyzedPlan, *parse.SQLStatement, error) {
	res, err := parse.ParseSQL(sql)
	if err != nil {
		return nil, nil, err
	}
	if res.ParseErrs.Err() != nil {
		return nil, nil, res.ParseErrs.Err()
	}

	analyzed, err := logical.CreateLogicalPlan(
		res.AST,
		e.getTable,
		func(varName string) (dataType *types.DataType, found bool) {
			val, found := e.getVariable(varName)
			if !found {
				return nil, false
			}

			// if it is a record, we need to return the record type
			if _, ok := val.(*RecordValue); !ok {
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
	)
	if err != nil {
		return nil, nil, err
	}

	return analyzed, res.AST, nil
}

// query executes a query.
func (e *executionContext) query(sql string, fn func(*RecordValue) error) error {
	res, err := parse.ParseSQL(sql)
	if err != nil {
		return err
	}
	if res.ParseErrs.Err() != nil {
		return res.ParseErrs.Err()
	}

	analyzed, err := logical.CreateLogicalPlan(
		res.AST,
		e.getTable,
		func(varName string) (dataType *types.DataType, found bool) {
			val, found := e.getVariable(varName)
			if !found {
				return nil, false
			}

			// if it is a record, we need to return the record type
			if _, ok := val.(*RecordValue); !ok {
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
	)
	if err != nil {
		return err
	}

	generatedSQL, params, err := generate.WriteSQL(res.AST, true, pgSchema)
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
		record := RecordValue{}
		for i, field := range analyzed.Plan.Relation().Fields {
			if field.Name == "" {
				continue
			}

			err = record.AddValue(field.Name, scanValues[i])
			if err != nil {
				return err
			}
		}

		return fn(&record)
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
	v, _, f := getVarFromScope(name, e.scope)
	return v, f
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

// Cursor is the cursor for the current execution.
// It is used to handle iteration over the results.
type Cursor interface {
	// Next moves the cursor to the next result.
	// It returns the value returned, if the cursor is done, and an error.
	// If the cursor is done, the value returned is not valid.
	Next(context.Context) (RecordValue, bool, error)
	// Close closes the cursor.
	Close() error
}

// returnableCursor is a cursor that can be directly returned to
// by the interpreter. This allows for progressively iterating over
// results.
type returnableCursor struct {
	expectedShape []*types.DataType
	recordChan    chan []Value
	errChan       chan error
}

func newReturnableCursor(expectedShape []*types.DataType) *returnableCursor {
	return &returnableCursor{
		expectedShape: expectedShape,
		recordChan:    make(chan []Value),
		errChan:       make(chan error),
	}
}

func (r *returnableCursor) Record() chan<- []Value {
	return r.recordChan
}

func (r *returnableCursor) Err() chan<- error {
	return r.errChan
}

func (r *returnableCursor) Close() error {
	close(r.recordChan)
	close(r.errChan)
	return nil
}

func (r *returnableCursor) Next(ctx context.Context) (RecordValue, bool, error) {
	select {
	case rec, ok := <-r.recordChan:
		if !ok {
			return RecordValue{}, true, nil
		} else {
			// check if the shape is correct
			if len(r.expectedShape) != len(rec) {
				return RecordValue{}, false, fmt.Errorf("expected %d columns, got %d", len(r.expectedShape), len(rec))
			}

			record := RecordValue{
				Fields: map[string]Value{},
			}

			for i, expected := range r.expectedShape {
				if !expected.EqualsStrict(rec[i].Type()) {
					return RecordValue{}, false, fmt.Errorf("expected type %s, got %s", expected, rec[i].Type())
				}

				record.Fields[expected.Name] = rec[i]
				record.Order = append(record.Order, expected.Name)
			}

			return record, false, nil
		}
	case err := <-r.errChan:
		// if the err is errReturn, we can swallow and just return done.
		if err == errReturn {
			return RecordValue{}, true, nil
		}

		return RecordValue{}, false, err
	case <-ctx.Done():
		return RecordValue{}, false, ctx.Err()
	}
}

// returnChans is a helper interface for returning values to channels.
type returnChans interface {
	Record() chan<- []Value
	Err() chan<- error
}

type returning interface {
	Values([]Value)
	// Err returns the error channel.
	Err(error)
}

func newReturningIntercept() *returningIntercept {
	recordChan := make(chan []Value)
	errChan := make(chan error)

	return &returningIntercept{
		recordChan: recordChan,
		errChan:    errChan,
	}
}

type returningIntercept struct {
	recordChan chan []Value
	errChan    chan error
}

var _ returning = (*returningIntercept)(nil)

func (r *returningIntercept) Values(vals []Value) {
	r.recordChan <- vals
}

func (r *returningIntercept) Err(err error) {
	r.errChan <- err
}

// singleRecordCursor is a cursor that returns a single record.
type singleValueCursor struct {
	rec  RecordValue
	done bool
}

func (s *singleValueCursor) Next(context.Context) (RecordValue, bool, error) {
	if s.done {
		return RecordValue{}, true, nil
	}

	s.done = true
	return s.rec, false, nil
}

func (s *singleValueCursor) Close() error {
	s.done = true
	return nil
}

func newSingleValueCursor(rec RecordValue) *singleValueCursor {
	return &singleValueCursor{
		rec: rec,
	}
}

type result struct {
	Record *RecordValue
	Err    error
}
