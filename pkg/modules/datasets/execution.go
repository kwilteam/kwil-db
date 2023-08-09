package datasets

import (
	"context"
	"fmt"
	"math/big"

	"github.com/kwilteam/kwil-db/pkg/balances"
	"github.com/kwilteam/kwil-db/pkg/engine"
	engineTypes "github.com/kwilteam/kwil-db/pkg/engine/types"
	transaction "github.com/kwilteam/kwil-db/pkg/tx"
)

/*
	This files contains logic for executing state changes against the database.
*/

// ExecutionResponse is the response from any interaction that modifies state.
type ExecutionResponse struct {
	// Fee is the amount of tokens spent on the execution
	Fee *big.Int
}

// Deploy deploys a database.
func (u *DatasetModule) Deploy(ctx context.Context, schema *engineTypes.Schema, tx *transaction.Transaction) (*transaction.ExecutionResponse, error) {
	price, err := u.PriceDeploy(ctx, schema)
	if err != nil {
		return nil, err
	}

	err = u.compareAndSpend(ctx, price, tx)
	if err != nil {
		return nil, err
	}

	_, err = u.engine.CreateDataset(ctx, schema)
	if err != nil {
		return nil, fmt.Errorf("failed to create dataset: %w", err)
	}

	return &transaction.ExecutionResponse{
		Fee: price,
	}, nil
}

// Drop drops a database.
func (u *DatasetModule) Drop(ctx context.Context, dbid string, tx *transaction.Transaction) (*transaction.ExecutionResponse, error) {
	price, err := u.PriceDrop(ctx, dbid)
	if err != nil {
		return nil, err
	}

	err = u.compareAndSpend(ctx, price, tx)
	if err != nil {
		return nil, err
	}

	err = u.engine.DropDataset(ctx, tx.Sender, dbid)
	if err != nil {
		return nil, fmt.Errorf("failed to drop dataset: %w", err)
	}

	return &transaction.ExecutionResponse{
		Fee: price,
	}, nil
}

// Execute executes an action against a database.
func (u *DatasetModule) Execute(ctx context.Context, dbid string, action string, params []map[string]any, tx *transaction.Transaction) (*transaction.ExecutionResponse, error) {
	price, err := u.PriceExecute(ctx, dbid, action, params)
	if err != nil {
		return nil, err
	}

	err = u.compareAndSpend(ctx, price, tx)
	if err != nil {
		return nil, err
	}

	_, err = u.engine.Execute(ctx, dbid, action, params,
		engine.WithCaller(tx.Sender),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to execute action: %w", err)
	}

	return &transaction.ExecutionResponse{
		Fee: price,
	}, nil
}

// compareAndSpend compares the calculated price to the transaction's fee, and spends the price if the fee is sufficient.
func (u *DatasetModule) compareAndSpend(ctx context.Context, price *big.Int, tx *transaction.Transaction) error {
	bigFee := new(big.Int)
	_, ok := bigFee.SetString(tx.Fee, 10)
	if !ok {
		return fmt.Errorf("failed to parse fee %s", tx.Fee)
	}

	if bigFee.Cmp(price) < 0 {
		return fmt.Errorf(`%w: fee %s is less than price %s`, ErrInsufficientFee, tx.Fee, price.String())
	}

	return u.accountStore.Spend(ctx, &balances.Spend{
		AccountAddress: tx.Sender,
		Amount:         price,
		Nonce:          tx.Nonce,
	})
}
