package dataset

import (
	"fmt"

	"github.com/kwilteam/kwil-db/pkg/engine2/dto"
	"github.com/kwilteam/kwil-db/pkg/engine2/sqldb"
)

// An preparedAction is a set of statements that takes a predefined set of inputs,
// and can be executed atomically.  It is the primary way to interact with
// a dataset.
type preparedAction struct {
	*dto.Action
	stmts   []sqldb.Statement
	dataset *Dataset
}

// Execute executes the action.
// It takes in a map of inputs and options.
// It returns a result set and an error.
func (a *preparedAction) Execute(txCtx *dto.TxContext, userInputs map[string]any) (dto.Result, error) {
	err := a.checkAccessControl(txCtx)
	if err != nil {
		return nil, fmt.Errorf("Action.Execute: failed access control: %w", err)
	}

	savepoint, err := a.dataset.db.Savepoint()
	if err != nil {
		return nil, err
	}
	defer savepoint.Rollback()

	inputs := txCtx.FillInputs(userInputs)

	var res dto.Result
	for _, stmt := range a.stmts {
		res, err = stmt.Execute(inputs)
		if err != nil {
			return nil, fmt.Errorf("Action.Execute: failed to execute statement: %w", err)
		}
	}

	err = savepoint.Commit()
	if err != nil {
		return nil, fmt.Errorf("Action.Execute: failed to commit savepoint: %w", err)
	}

	return res, nil
}

// BatchExecute executes the action multiple times.
// It takes in a map of inputs and options.
// It returns a result set and an error.
func (a *preparedAction) BatchExecute(txCtx *dto.TxContext, userInputs []map[string]any) (dto.Result, error) {
	savepoint, err := a.dataset.db.Savepoint()
	if err != nil {
		return nil, fmt.Errorf("Action.BatchExecute: failed to create savepoint: %w", err)
	}
	defer savepoint.Rollback()

	var results dto.Result
	for _, inputs := range userInputs {
		results, err = a.Execute(txCtx, inputs)
		if err != nil {
			return nil, fmt.Errorf("Action.BatchExecute: failed to execute statement: %w", err)
		}
	}

	err = savepoint.Commit()
	if err != nil {
		return nil, fmt.Errorf("Action.BatchExecute: failed to commit savepoint: %w", err)
	}

	return results, nil
}

// Close closes the action.
func (a *preparedAction) Close() error {
	for _, stmt := range a.stmts {
		err := stmt.Close()
		if err != nil {
			return fmt.Errorf("Action.Close: failed to close statement: %w", err)
		}
	}
	return nil
}

func (a *preparedAction) checkAccessControl(opts *dto.TxContext) error {
	if a.Public {
		return nil
	}

	if opts == nil {
		return fmt.Errorf("failed to execute private action '%s': could not authenticate caller", a.Name)
	}

	if opts.Caller == "" {
		return fmt.Errorf("failed to execute private action '%s': caller not provided", a.Name)
	}

	if opts.Caller != a.dataset.Ctx.Owner {
		return fmt.Errorf("failed to execute private action '%s': caller is not the owner", a.Name)
	}

	return nil
}
