package interpreter

import (
	"context"
	"errors"
	"fmt"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/parse/common"
)

// executionContext is the context of the entire execution.
type executionContext struct {
	// maxCost is the maximum allowable cost of the execution.
	maxCost int64
	// currentCost is the current cost of the execution.
	currentCost int64
	// scope is the current scope.
	scope *scopeContext
	// costTable is the cost table for the execution.
	costTable *CostTable
	// procedures maps local procedure names to their procedure func.
	procedures map[string]procedureCallFunc
}

// newScope creates a new scope.
func newScope() *scopeContext {
	return &scopeContext{
		variables: make(map[string]common.Value),
	}
}

// subScope creates a new sub-scope, which has access to the parent scope.
func (s *scopeContext) subScope() *scopeContext {
	return &scopeContext{
		parent:    s,
		variables: make(map[string]common.Value),
	}
}

// Spend spends a certain amount of cost.
// If the cost exceeds the maximum cost, it returns an error.
func (e *executionContext) Spend(cost int64) error {
	if e.currentCost+cost > e.maxCost {

		e.currentCost = e.maxCost
		return fmt.Errorf("exceeded maximum cost: %d", e.maxCost)
	}
	e.currentCost += cost
	return nil
}

// Notice logs a notice.
func (e *executionContext) Notice(format string) {
	panic("notice not implemented")
}

// setVariable sets a variable in the current scope.
// It will allocate the variable if it does not exist.
func (e *executionContext) setVariable(name string, value common.Value) error {
	err := e.Spend(e.costTable.GetVariableCost)
	if err != nil {
		return err
	}

	_, foundScope, err := getVarFromScope(name, e.scope)
	if err != nil {
		if errors.Is(err, ErrVariableNotFound) {
			return e.allocateVariable(name, value)
		} else {
			return err
		}
	}

	err = e.Spend(e.costTable.SetVariableCost + e.costTable.SizeCostConstant*int64(value.Size()))
	if err != nil {
		return err
	}

	foundScope.variables[name] = value
	return nil
}

// allocateVariable allocates a variable in the current scope.
func (e *executionContext) allocateVariable(name string, value common.Value) error {
	err := e.Spend(e.costTable.AllocateVariableCost + e.costTable.SizeCostConstant*int64(value.Size()))
	if err != nil {
		return err
	}

	_, ok := e.scope.variables[name]
	if ok {
		return fmt.Errorf(`variable "%s" already exists`, name)
	}

	e.scope.variables[name] = value
	return nil
}

// getVariable gets a variable from the current scope.
// It searches the parent scopes if the variable is not found.
func (e *executionContext) getVariable(name string) (common.Value, error) {
	err := e.Spend(e.costTable.GetVariableCost)
	if err != nil {
		return nil, err
	}

	v, _, err := getVarFromScope(name, e.scope)
	return v, err
}

// getVarFromScope recursively searches the scopes for a variable.
// It returns the value, as well as the scope it was found in.
func getVarFromScope(variable string, scope *scopeContext) (common.Value, *scopeContext, error) {
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
	variables map[string]common.Value
}

// Cursor is the cursor for the current execution.
// It is used to handle iteration over the results.
type Cursor interface {
	// Next moves the cursor to the next result.
	// It returns the value returned, if the cursor is done, and an error.
	// If the cursor is done, the value returned is not valid.
	Next(context.Context) (common.RecordValue, bool, error)
	// Close closes the cursor.
	Close() error
}

// returnableCursor is a cursor that can be directly returned to
// by the interpreter. This allows for progressively iterating over
// results.
type returnableCursor struct {
	expectedShape []*types.DataType
	recordChan    chan []common.Value
	errChan       chan error
}

func newReturnableCursor(expectedShape []*types.DataType) *returnableCursor {
	return &returnableCursor{
		expectedShape: expectedShape,
		recordChan:    make(chan []common.Value),
		errChan:       make(chan error),
	}
}

func (r *returnableCursor) Record() chan<- []common.Value {
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

func (r *returnableCursor) Next(ctx context.Context) (common.RecordValue, bool, error) {
	select {
	case rec, ok := <-r.recordChan:
		if !ok {
			return common.RecordValue{}, true, nil
		} else {
			// check if the shape is correct
			if len(r.expectedShape) != len(rec) {
				return common.RecordValue{}, false, fmt.Errorf("expected %d columns, got %d", len(r.expectedShape), len(rec))
			}

			record := common.RecordValue{
				Fields: map[string]common.Value{},
			}

			for i, expected := range r.expectedShape {
				if !expected.EqualsStrict(rec[i].Type()) {
					return common.RecordValue{}, false, fmt.Errorf("expected type %s, got %s", expected, rec[i].Type())
				}

				record.Fields[expected.Name] = rec[i]
				record.Order = append(record.Order, expected.Name)
			}

			return record, false, nil
		}
	case err := <-r.errChan:
		if err == errReturn {
			return common.RecordValue{}, true, nil
		}

		return common.RecordValue{}, false, err
	case <-ctx.Done():
		return common.RecordValue{}, false, ctx.Err()
	}
}

// returnChans is a helper interface for returning values to channels.
type returnChans interface {
	Record() chan<- []common.Value
	Err() chan<- error
}
