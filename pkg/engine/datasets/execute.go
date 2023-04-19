package datasets

import (
	"encoding/json"
	"fmt"
	"kwil/pkg/engine/models"
	"kwil/pkg/sql/driver"
	"math/big"
	"strings"
)

const (
	callerVar = "@caller"
)

type ExecOpts struct {
	// Caller is the user that is calling the action.
	Caller string
}

// buildExecOpts ensures that the execOpts are initialized to avoid panics
func buildExecOpts(eo *ExecOpts) {
	if eo == nil {
		eo = &ExecOpts{}
	}
	if eo.Caller == "" {
		eo.Caller = "0x0000000000000000000000000000000000000000"
	}
}

// ExecuteAction executes a predefined database action
func (d *Dataset) ExecuteAction(exec *models.ActionExecution, opts *ExecOpts) (RecordSet, error) {
	sp, err := d.conn.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to create savepoint: %w", err)
	}
	defer sp.Rollback()

	buildExecOpts(opts)

	ac, ok := d.actions[exec.Action]
	if !ok {
		return nil, fmt.Errorf("action %s does not exist", exec.Action)
	}

	canExecute, err := d.CanExecute(exec.Action, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to check if action access control: %w", err)
	}
	if !canExecute {
		return nil, fmt.Errorf("action %s is not allowed for caller %s", exec.Action, opts.Caller)
	}

	inputs, err := ac.Prepare(exec, opts)
	if err != nil {
		return nil, fmt.Errorf("error preparing action %s: %w", exec.Action, err)
	}

	result, err := d.execAction(ac, inputs)
	if err != nil {
		return nil, fmt.Errorf("error executing action %s: %w", exec.Action, err)
	}

	err = sp.Commit()
	if err != nil {
		return nil, fmt.Errorf("failed to commit savepoint: %w", err)
	}

	return result, nil
}

// execAction executes an action for as many inputs as exists
func (d *Dataset) execAction(action *PreparedAction, inputs []map[string]any) (result RecordSet, err error) {
	if len(inputs) == 0 {
		return d.execActionWithNoInputs(action)
	}

	for _, record := range inputs {
		for _, statement := range action.Statements {
			result, err = d.execStatement(statement.Stmt, record)
			if err != nil {
				return nil, fmt.Errorf("error executing statement '%s': %w", statement.Stmt, err)
			}
		}
	}

	return result, nil
}

// execActionWithNoInputs executes an action with no inputs.
// the action will execute once
func (d *Dataset) execActionWithNoInputs(action *PreparedAction) (result RecordSet, err error) {
	for _, statement := range action.Statements {
		result, err = d.execStatement(statement.Stmt, nil)
		if err != nil {
			return nil, fmt.Errorf("error executing statement '%s': %w", statement.Stmt, err)
		}
	}

	return result, nil
}

// execStatement executes a single statement
func (d *Dataset) execStatement(stmt string, input map[string]any) (RecordSet, error) {
	recordSet := make(RecordSet, 0)
	err := d.conn.ExecuteNamed(stmt, input, func(stmt *driver.Statement) error {
		row, err := stmt.GetRecord()
		if err != nil {
			return err
		}

		recordSet = append(recordSet, row)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error executing statement: %w", err)
	}

	return recordSet, nil
}

func (d *Dataset) CanExecute(action string, execOpts *ExecOpts) (bool, error) {
	acc, ok := d.actions[action]
	if !ok {
		return false, fmt.Errorf("action %s does not exist", action)
	}

	if acc.Public {
		return true, nil
	}

	buildExecOpts(execOpts)

	if strings.EqualFold(execOpts.Caller, d.Owner) {
		return true, nil
	}

	return false, nil
}

var actionPrice = big.NewInt(2000000000000000)

func (d *Dataset) GetActionPrice(action string, execOpts *ExecOpts) (res *big.Int, err error) {
	return actionPrice, nil
}

type RecordSet []driver.Record

func (r *RecordSet) Bytes() ([]byte, error) {
	bts, err := json.Marshal(r)
	if err != nil {
		return nil, fmt.Errorf("error marshaling response: %w", err)
	}

	return bts, nil
}

func (r *RecordSet) ForEach(fn func(row map[string]any) error) error {
	for _, row := range *r {
		err := fn(row)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *Dataset) Query(stmt string) (res RecordSet, err error) {
	err = d.readOnlyConn.Query(stmt, func(stmt *driver.Statement) error {
		row, err := stmt.GetRecord()
		if err != nil {
			return err
		}

		res = append(res, row)
		return nil
	})

	return res, err
}
