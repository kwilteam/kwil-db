package datasets

import (
	"fmt"
	"math/big"

	"github.com/kwilteam/kwil-db/internal/entity"
	"github.com/kwilteam/kwil-db/pkg/engine/dto"
	"github.com/kwilteam/kwil-db/pkg/tx"
)

func (u *DatasetUseCase) Execute(action *entity.ExecuteAction) (*tx.Receipt, error) {
	price, err := u.PriceExecute(action)
	if err != nil {
		return nil, err
	}

	ds, err := u.engine.GetDataset(action.ExecutionBody.DBID)
	if err != nil {
		return nil, err
	}

	err = u.compareAndSpend(action.Tx.Sender, action.Tx.Fee, action.Tx.Nonce, price)
	if err != nil {
		return nil, err
	}

	res, err := ds.Execute(&dto.TxContext{
		Caller:  action.Tx.Sender,
		Action:  action.ExecutionBody.Action,
		Dataset: action.ExecutionBody.DBID,
	}, action.ExecutionBody.Params)
	if err != nil {
		return nil, err
	}
	fmt.Println("cherry exec res", res)
	bts, err := readQueryResult(res)
	if err != nil {
		return nil, err
	}

	return &tx.Receipt{
		TxHash: action.Tx.Hash,
		Fee:    price.String(),
		Body:   bts,
	}, nil
}

func (u *DatasetUseCase) PriceExecute(action *entity.ExecuteAction) (*big.Int, error) {
	return actionPrice, nil
}
