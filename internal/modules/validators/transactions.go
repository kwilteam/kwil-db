package validators

import (
	"context"
	"fmt"
	"math/big"

	"github.com/kwilteam/kwil-db/core/types/transactions"
	"github.com/kwilteam/kwil-db/internal/accounts"
)

type ExecutionResponse struct {
	// Fee is the amount of tokens spent on the execution
	Fee     *big.Int
	GasUsed int64
}

func resp(fee *big.Int) *ExecutionResponse {
	return &ExecutionResponse{
		Fee:     fee,
		GasUsed: 0,
	}
}

// Join/Leave/Approve required a spend. There is currently no pricing associated
// with the actions, although there probably should be for Join.

func (vm *ValidatorModule) spend(ctx context.Context, acctPubKey []byte,
	amt *big.Int, nonce uint64) error {
	return vm.accts.Spend(ctx, &accounts.Spend{
		AccountPubKey: acctPubKey,
		Amount:        amt,
		Nonce:         int64(nonce),
	})
}

// Join creates a join request for a prospective validator.
func (vm *ValidatorModule) Join(ctx context.Context, joiner []byte, power int64,
	txn *transactions.Transaction) (*ExecutionResponse, error) {
	price, err := vm.PriceJoin(ctx)
	if err != nil {
		return nil, err
	}

	if txn.Body.Fee.Cmp(price) < 0 {
		return nil, fmt.Errorf("insufficient fee: %d < %d", txn.Body.Fee, price)
	}

	err = vm.spend(ctx, joiner, txn.Body.Fee, txn.Body.Nonce)
	if err != nil {
		return nil, err
	}

	if err = vm.mgr.Join(ctx, joiner, power); err != nil {
		return nil, err
	}

	return resp(txn.Body.Fee), nil
}

// Leave creates a leave request for a current validator.
func (vm *ValidatorModule) Leave(ctx context.Context, leaver []byte,
	txn *transactions.Transaction) (*ExecutionResponse, error) {
	price, err := vm.PriceLeave(ctx)
	if err != nil {
		return nil, err
	}

	if txn.Body.Fee.Cmp(price) < 0 {
		return nil, fmt.Errorf("insufficient fee: %d < %d", txn.Body.Fee, price)
	}

	err = vm.spend(ctx, leaver, txn.Body.Fee, txn.Body.Nonce)
	if err != nil {
		return nil, err
	}

	if err = vm.mgr.Leave(ctx, leaver); err != nil {
		return nil, err
	}

	return resp(txn.Body.Fee), nil
}

// Approve records an approval transaction from a current validator..
func (vm *ValidatorModule) Approve(ctx context.Context, joiner []byte,
	txn *transactions.Transaction) (*ExecutionResponse, error) {
	approver := txn.Sender
	price, err := vm.PriceApprove(ctx)
	if err != nil {
		return nil, err
	}

	if txn.Body.Fee.Cmp(price) < 0 {
		return nil, fmt.Errorf("insufficient fee: %d < %d", txn.Body.Fee, price)
	}

	err = vm.spend(ctx, approver, txn.Body.Fee, txn.Body.Nonce)
	if err != nil {
		return nil, err
	}

	if err = vm.mgr.Approve(ctx, joiner, approver); err != nil {
		return nil, err
	}

	return &ExecutionResponse{
		Fee:     txn.Body.Fee,
		GasUsed: 0,
	}, nil
}
