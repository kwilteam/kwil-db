package datasets

import (
	"context"
	"fmt"
	"math/big"

	"github.com/kwilteam/kwil-db/pkg/balances"
	"github.com/kwilteam/kwil-db/pkg/engine"
	engineTypes "github.com/kwilteam/kwil-db/pkg/engine/types"
	"github.com/kwilteam/kwil-db/pkg/transactions"
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
func (u *DatasetModule) Deploy(ctx context.Context, schema *engineTypes.Schema, tx *transactions.Transaction) (*transactions.TransactionStatus, error) {
	price, err := u.PriceDeploy(ctx, schema)
	if err != nil {
		if price == nil {
			price = big.NewInt(0)
		}
		return failure(tx, price, err)
	}

	err = u.compareAndSpend(ctx, price, tx)
	if err != nil {
		return failure(tx, price, err)
	}

	_, err = u.engine.CreateDataset(ctx, schema)
	if err != nil {
		return failure(tx, price, fmt.Errorf("failed to create dataset: %w", err))
	}

	return success(tx, price)
}

// Drop drops a database.
func (u *DatasetModule) Drop(ctx context.Context, dbid string, tx *transactions.Transaction) (*transactions.TransactionStatus, error) {
	price, err := u.PriceDrop(ctx, dbid)
	if err != nil {
		if price == nil {
			price = big.NewInt(0)
		}
		return failure(tx, price, err)
	}

	err = u.compareAndSpend(ctx, price, tx)
	if err != nil {
		return failure(tx, price, err)
	}

	err = u.engine.DropDataset(ctx, tx.Sender.Address().String(), dbid)
	if err != nil {
		return failure(tx, price, fmt.Errorf("failed to drop dataset: %w", err))
	}

	return success(tx, price)
}

// Execute executes an action against a database.
// TODO: I think args should be [][]any, not [][]string
func (u *DatasetModule) Execute(ctx context.Context, dbid string, action string, args [][]any, tx *transactions.Transaction) (*transactions.TransactionStatus, error) {
	price, err := u.PriceExecute(ctx, dbid, action, args)
	if err != nil {
		if price == nil {
			price = big.NewInt(0)
		}
		return failure(tx, price, err)
	}

	err = u.compareAndSpend(ctx, price, tx)
	if err != nil {
		return failure(tx, price, err)
	}

	_, err = u.engine.Execute(ctx, dbid, action, args,
		engine.WithCaller(tx.Sender.Address().String()),
	)
	if err != nil {
		return failure(tx, price, fmt.Errorf("failed to execute action: %w", err))
	}

	return success(tx, price)
}

// compareAndSpend compares the calculated price to the transaction's fee, and spends the price if the fee is sufficient.
func (u *DatasetModule) compareAndSpend(ctx context.Context, price *big.Int, tx *transactions.Transaction) error {

	if tx.Body.Fee.Cmp(price) < 0 {
		return fmt.Errorf(`%w: fee %s is less than price %s`, ErrInsufficientFee, tx.Body.Fee.String(), price.String())
	}

	return u.accountStore.Spend(ctx, &balances.Spend{
		AccountAddress: tx.Sender.Address().String(),
		Amount:         price,
		Nonce:          int64(tx.Body.Nonce),
	})
}

func success(tx *transactions.Transaction, fee *big.Int) (*transactions.TransactionStatus, error) {
	txid, err := tx.GetHash()
	if err != nil {
		return nil, fmt.Errorf("failed to get tx hash: %w", err)
	}

	return &transactions.TransactionStatus{
		ID:     txid,
		Fee:    fee,
		Status: transactions.StatusSuccess,
	}, nil
}

func failure(tx *transactions.Transaction, fee *big.Int, errs ...error) (*transactions.TransactionStatus, error) {
	txid, err := tx.GetHash()
	if err != nil {
		return nil, fmt.Errorf("failed to get tx hash: %w", err)
	}

	var errStrings []string
	for _, err := range errs {
		errStrings = append(errStrings, err.Error())
	}

	return &transactions.TransactionStatus{
		ID:     txid,
		Fee:    fee,
		Status: transactions.StatusFailed,
		Errors: errStrings,
	}, nil
}
