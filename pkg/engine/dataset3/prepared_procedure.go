package dataset2

import (
	"fmt"
	"strings"

	"github.com/kwilteam/kwil-db/pkg/engine/dto"
)

// a procedureContext is the context for executing a procedure.
// it contains the executionContext, as well as the values scoped to the procedure.
type procedureContext struct {
	*executionContext
	values map[string]any
}

// addVariables adds variables to the procedure context.
// if there are any conflicts, the new values will override the old ones.
func (p *procedureContext) addVariables(variables map[string]any) {
	if variables == nil {
		return
	}

	if p.values == nil {
		p.values = make(map[string]any)
	}

	for k, v := range variables {
		p.values[k] = v
	}
}

// An StoredProcedure is a set of statements that takes a predefined set of inputs,
// and can be executed atomically.  It is the primary way to interact with
// a dataset.
type StoredProcedure struct {
	*dto.Action
	operations []operation
	dataset    *Dataset
}

// Execute executes the procedure.
// It takes context.
// It returns an error.
// If any dml statements are executed, it will return the result of the last one.
func (a *StoredProcedure) Execute(ctx *executionContext, args []any) error {
	err := a.checkAccessControl(ctx)
	if err != nil {
		return fmt.Errorf("procedure.execute: failed access control: %w", err)
	}

	values, err := a.getValues(ctx, args)
	if err != nil {
		return fmt.Errorf("procedure.execute: failed to get values: %w", err)
	}

	procCtx := &procedureContext{
		executionContext: ctx,
		values:           values,
	}

	savepoint, err := a.dataset.db.Savepoint()
	if err != nil {
		return fmt.Errorf("procedure.execute: failed to create savepoint: %w", err)
	}
	defer savepoint.Rollback()

	err = a.evaluateOperations(procCtx)
	if err != nil {
		return fmt.Errorf("procedure.execute: failed to evaluate operations: %w", err)
	}

	err = savepoint.Commit()
	if err != nil {
		return fmt.Errorf("procedure.execute: failed to commit savepoint: %w", err)
	}

	return nil
}

// evaluateOperations executes the operations in the procedure.
// It takes context and a map of values. The values will be perpetually updated as the operations are executed.
func (a *StoredProcedure) evaluateOperations(ctx *procedureContext) error {
	for _, op := range a.operations {
		requiredArgs := op.requiredVariables()
		args := make([]any, len(requiredArgs))

		for i, requiredArg := range requiredArgs {
			argValue, ok := ctx.values[requiredArg]
			if !ok {
				return fmt.Errorf("missing argument '%s'", requiredArg)
			}

			args[i] = argValue
		}

		operationResults, err := op.evaluate(ctx, args...)
		if err != nil {
			return fmt.Errorf("failed to execute operation: %w", err)
		}

		if operationResults != nil {
			ctx.addVariables(operationResults)
		}
	}

	return nil
}

// getValues returns a map of values containing the necessary inputs for the procedure.
// if there are missing values, it will return an error.
func (a *StoredProcedure) getValues(ctx *executionContext, args []any) (map[string]any, error) {
	if len(args) != len(a.Inputs) {
		return nil, fmt.Errorf("function block '%s' requires %d inputs, but %d were provided", a.Name, len(a.Inputs), len(args))
	}

	values := ctx.contextualVariables()
	for i, arg := range args {
		values[a.Inputs[i]] = arg
	}

	return values, nil
}

// close closes the action.
func (a *StoredProcedure) close() error {
	var errs []string

	for _, stmt := range a.operations {
		err := stmt.close()
		if err != nil {
			errs = append(errs, err.Error())
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to close procedure '%s': %s", a.Name, strings.Join(errs, "; "))
	}

	return nil
}

func (a *StoredProcedure) checkAccessControl(opts *executionContext) error {
	if a.Public {
		return nil
	}

	if opts == nil {
		return fmt.Errorf("failed to execute private action '%s': could not authenticate caller", a.Name)
	}

	if opts.caller != a.dataset.owner {
		return fmt.Errorf("failed to execute private action '%s': caller is not the owner", a.Name)
	}

	return nil
}
