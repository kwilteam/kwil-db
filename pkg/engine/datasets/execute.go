package datasets

import (
	"encoding/json"
	"fmt"
	"kwil/pkg/engine/models"
	"kwil/pkg/sql/driver"
	"math/big"
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

//func (d *Dataset) executeSingle

// ExecuteAction executes a predefined database action
func (d *Dataset) ExecuteAction(exec *models.ActionExecution, opts *ExecOpts) (ActionResult, error) {
	sp, err := d.conn.Savepoint()
	if err != nil {
		return nil, fmt.Errorf("failed to create savepoint: %w", err)
	}
	defer sp.Rollback()

	buildExecOpts(opts)

	ac, ok := d.actions[exec.Action]
	if !ok {
		return nil, fmt.Errorf("action %s does not exist", exec.Action)
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
func (d *Dataset) execAction(action *PreparedAction, inputs []map[string]any) (ActionResult, error) {
	if len(inputs) == 0 {
		return d.execActionWithNoInputs(action)
	}

	result := make(ActionResult, 0)
	for _, record := range inputs {
		for _, statement := range action.Statements {
			res, err := d.execStatement(statement.Stmt, record)
			if err != nil {
				return nil, fmt.Errorf("error executing statement: %w", err)
			}

			result = append(result, res)
		}
	}

	return result, nil
}

// execActionWithNoInputs executes an action with no inputs.
// the action will execute once
func (d *Dataset) execActionWithNoInputs(action *PreparedAction) (ActionResult, error) {
	result := make(ActionResult, 0)
	for _, statement := range action.Statements {
		res, err := d.execStatement(statement.Stmt, nil)
		if err != nil {
			return nil, fmt.Errorf("error executing statement: %w", err)
		}

		result = append(result, res)
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

var actionPrice = big.NewInt(2000000000000000)

func (d *Dataset) GetActionPrice(action string, execOpts *ExecOpts) (res *big.Int, err error) {
	return actionPrice, nil
}

type RecordSet []driver.Record

type ActionResult []RecordSet

func (r *RecordSet) Bytes() ([]byte, error) {
	bts, err := json.Marshal(r)
	if err != nil {
		return nil, fmt.Errorf("error marshaling response: %w", err)
	}

	return bts, nil
}

func (a *ActionResult) Bytes() ([]byte, error) {
	bts, err := json.Marshal(a)
	if err != nil {
		return nil, fmt.Errorf("error marshaling response: %w", err)
	}

	return bts, nil
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
