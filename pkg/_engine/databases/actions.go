package databases

import (
	"fmt"

	"github.com/kwilteam/kwil-db/pkg/sql/sqlite"
)

// An Action is a single action that can be executed on a database.
type Action struct {
	Name           string
	Public         bool
	RequiredInputs []string
	stmts          []*sqlite.Statement
	dataset        *Dataset
}

const (
	defaultCallerAddress = "0x0000000000000000000000000000000000000000"
	callerVarName        = "@caller"
)

// ExecOpts are options for executing an action.
// Things like caller, block height, etc. are included here.
type ExecOpts struct {
	// Caller is the wallet address of the caller.
	Caller string
}

// fillDefaults fills in default values for the options.
func (e *ExecOpts) fillDefaults() *ExecOpts {
	if e == nil {
		e = &ExecOpts{}
	}

	if e.Caller == "" {
		e.Caller = defaultCallerAddress
	}

	return e
}

// fillInputs adds the ExecOpts values to the inputs map.
func (e *ExecOpts) fillInputs(inputs map[string]any) map[string]any {
	e = e.fillDefaults()
	if inputs == nil {
		inputs = make(map[string]any)
	}

	inputs[callerVarName] = e.Caller

	return inputs
}

// Execute executes the action.
// It takes in a map of inputs and options.
// It returns a result set and an error.
func (a *Action) Execute(userInputs map[string]any, opts *ExecOpts) (results *sqlite.ResultSet, err error) {
	inputs := opts.fillInputs(userInputs)

	err = a.checkAccessControl(opts)
	if err != nil {
		return nil, fmt.Errorf("Action.Execute: failed access control: %w", err)
	}

	savepoint, err := a.dataset.conn.Savepoint()
	if err != nil {
		return nil, err
	}
	defer savepoint.Rollback()

	for _, stmt := range a.stmts {
		err = stmt.Execute(
			sqlite.WithNamedArgs(inputs),
			sqlite.WithResultSet(results),
		)
		if err != nil {
			return nil, fmt.Errorf("Action.Execute: failed to execute statement: %w", err)
		}
	}

	err = savepoint.Commit()
	if err != nil {
		return nil, fmt.Errorf("Action.Execute: failed to commit savepoint: %w", err)
	}

	return nil, nil
}

// BatchExecute executes the action multiple times.
// It takes in a map of inputs and options.
// It returns a result set and an error.
func (a *Action) BatchExecute(userInputs []map[string]any, opts *ExecOpts) (results *sqlite.ResultSet, err error) {
	savepoint, err := a.dataset.conn.Savepoint()
	if err != nil {
		return nil, fmt.Errorf("Action.BatchExecute: failed to create savepoint: %w", err)
	}
	defer savepoint.Rollback()

	for _, inputs := range userInputs {
		results, err = a.Execute(inputs, opts)
		if err != nil {
			return nil, fmt.Errorf("Action.BatchExecute: failed to execute statement: %w", err)
		}
	}

	err = savepoint.Commit()
	if err != nil {
		return nil, fmt.Errorf("Action.BatchExecute: failed to commit savepoint: %w", err)
	}

	return nil, nil
}

// Close closes the action.
func (a *Action) Close() error {
	for _, stmt := range a.stmts {
		err := stmt.Finalize()
		if err != nil {
			return fmt.Errorf("Action.Close: failed to finalize statement: %w", err)
		}
	}
	return nil
}

func (a *Action) checkAccessControl(opts *ExecOpts) error {
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
