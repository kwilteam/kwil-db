package datasets

import (
	"context"
	"math/big"

	"github.com/kwilteam/kwil-db/internal/entity"
	"github.com/kwilteam/kwil-db/pkg/engine/dataset"
	"github.com/kwilteam/kwil-db/pkg/tx"
)

func (u *DatasetUseCase) Execute(ctx context.Context, action *entity.ExecuteAction) (*tx.Receipt, error) {
	price, err := u.PriceExecute(action)
	if err != nil {
		return nil, err
	}

	ds, err := u.engine.GetDataset(ctx, action.ExecutionBody.DBID)
	if err != nil {
		return nil, err
	}

	err = u.compareAndSpend(action.Tx.Sender, action.Tx.Fee, action.Tx.Nonce, price)
	if err != nil {
		return nil, err
	}

	res, err := ds.Execute(ctx, action.ExecutionBody.Action, action.ExecutionBody.Params, &dataset.TxOpts{
		Caller: action.Tx.Sender,
	})
	if err != nil {
		return nil, err
	}

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
