package interpreter

import (
	"context"
	"errors"
	"fmt"

	"github.com/kwilteam/kwil-db/core/types"
)

// executionContext is the context of the entire execution.
type executionContext struct {
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
	Func functionCall
}

// returnTypeFunc is a function that validates the arguments of a function and gives
// the return type of the function based on the arguments.
// The return type can be nil if the function does not return anything.
type returnTypeFunc func([]Value) (*types.ProcedureReturn, error)

// newScope creates a new scope.
func newScope() *scopeContext {
	return &scopeContext{
		variables: make(map[string]Value),
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
	_, foundScope, err := getVarFromScope(name, e.scope)
	if err != nil {
		if errors.Is(err, ErrVariableNotFound) {
			return e.allocateVariable(name, value)
		} else {
			return err
		}
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
func (e *executionContext) getVariable(name string) (Value, error) {
	v, _, err := getVarFromScope(name, e.scope)
	return v, err
}

// getVarFromScope recursively searches the scopes for a variable.
// It returns the value, as well as the scope it was found in.
func getVarFromScope(variable string, scope *scopeContext) (Value, *scopeContext, error) {
	if v, ok := scope.variables[variable]; ok {
		return v, scope, nil
	}
	if scope.parent == nil {
		return nil, nil, fmt.Errorf(`%w: "%s"`, ErrVariableNotFound, variable)
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
