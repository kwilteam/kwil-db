package datasets

import (
	"fmt"
	"kwil/internal/entity"
	"kwil/pkg/engine/datasets"
	"kwil/pkg/engine/models"
	"kwil/pkg/tx"
	"math/big"
)

func (u *DatasetUseCase) Execute(action *entity.ExecuteAction) (*tx.Receipt, error) {
	price, err := u.PriceExecute(action)
	if err != nil {
		return nil, err
	}

	ds, ok := u.engine.Datasets[action.ExecutionBody.DBID]
	if !ok {
		return nil, fmt.Errorf("dataset not found '%s'", action.ExecutionBody.DBID)
	}

	err = u.compareAndSpend(action.Tx.Sender, action.Tx.Fee, action.Tx.Nonce, price)
	if err != nil {
		return nil, err
	}

	res, err := ds.ExecuteAction(&models.ActionExecution{
		Action: action.ExecutionBody.Action,
		Params: action.ExecutionBody.Params,
		DBID:   action.ExecutionBody.DBID,
	}, &datasets.ExecOpts{
		Caller: action.Tx.Sender,
	})
	if err != nil {
		return nil, err
	}

	bts, err := res.Bytes()
	if err != nil {
		return nil, err
	}

	return &tx.Receipt{
		Hash: action.Tx.Hash,
		Fee:  price.String(),
		Body: bts,
	}, nil
}

func (u *DatasetUseCase) PriceExecute(action *entity.ExecuteAction) (*big.Int, error) {
	ds, ok := u.engine.Datasets[action.ExecutionBody.DBID]
	if !ok {
		return nil, fmt.Errorf("dataset not found '%s'", action.ExecutionBody.DBID)
	}

	execOpts := datasets.ExecOpts{
		Caller: action.Tx.Sender,
		Params: action.ExecutionBody.Params,
	}

	return ds.GetActionPrice(action.ExecutionBody.Action, &execOpts)
}
