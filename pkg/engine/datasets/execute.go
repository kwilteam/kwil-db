package datasets

import (
	"encoding/json"
	"fmt"
	"kwil/pkg/engine/models"
	"kwil/pkg/engine/types"
	"kwil/pkg/sql/driver"
	"math/big"
)

const (
	callerVar = "@caller"
)

type ExecOpts struct {
	// Params are the parameters to pass to the action.
	Params []map[string][]byte

	// Caller is the user that is calling the action.
	Caller string
}

func (d *Dataset) ExecuteAction(exec *models.ActionExecution, execOpts *ExecOpts) (res RecordSet, err error) {
	sp, err := d.conn.Savepoint()
	if err != nil {
		return nil, fmt.Errorf("failed to create savepoint: %w", err)
	}
	defer sp.Rollback()

	opts := &ExecOpts{
		Params: []map[string][]byte{{}},
		Caller: "0x0000000000000000000000000000000000000000",
	}
	if exec.Params != nil {
		opts.Params = exec.Params
	}
	if execOpts.Caller != "" {
		opts.Caller = execOpts.Caller
	}

	ac, ok := d.actions[exec.Action]
	if !ok {
		return nil, fmt.Errorf("action %s does not exist", exec.Action)
	}

	for i, record := range opts.Params {
		rec := make(map[string]any)
		for _, input := range ac.Inputs {
			val, ok := record[input]
			if !ok {
				return nil, fmt.Errorf(`missing input "%s"`, input)
			}

			concrete, err := types.NewFromSerial(val)
			if err != nil {
				return nil, fmt.Errorf("error converting serialized input %s: %w", input, err)
			}

			rec[input] = concrete
		}

		rec[callerVar] = opts.Caller
		for _, prepared := range ac.Statements {
			err := d.conn.ExecuteNamed(prepared.Stmt, rec, func(stmt *driver.Statement) error {
				if i == len(opts.Params)-1 {
					row, err := stmt.GetRecord()
					if err != nil {
						return err
					}

					res = append(res, row)
				}
				return nil
			})
			if err != nil {
				return nil, fmt.Errorf("error executing statement: %w", err)
			}
		}
	}

	err = sp.Commit()
	if err != nil {
		return nil, fmt.Errorf("failed to commit savepoint: %w", err)
	}

	return res, nil
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
