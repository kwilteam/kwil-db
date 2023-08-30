package validators

import (
	"context"
	"math/big"

	"github.com/kwilteam/kwil-db/pkg/balances"
	"github.com/kwilteam/kwil-db/pkg/transactions"
)

type ExecutionResponse struct {
	// Fee is the amount of tokens spent on the execution
	Fee     *big.Int
	GasUsed int64
}

// Join/Leave/Approve required a spend. There is currently no pricing associated
// with the actions, although there probably should be for Join.

func (vm *ValidatorModule) spend(ctx context.Context, acctPubKey []byte,
	amt *big.Int, nonce uint64) error {
	return vm.accts.Spend(ctx, &balances.Spend{
		AccountPubKey: acctPubKey,
		Amount:        amt,
		Nonce:         int64(nonce),
	})
}

// Join creates a join request for a prospective validator.
func (vm *ValidatorModule) Join(ctx context.Context, joiner []byte, power int64,
	txn *transactions.Transaction) (*ExecutionResponse, error) {

	err := vm.spend(ctx, joiner, txn.Body.Fee, txn.Body.Nonce)
	if err != nil {
		return nil, err
	}

	if err = vm.mgr.Join(ctx, joiner, power); err != nil {
		return nil, err
	}

	return &ExecutionResponse{
		Fee:     txn.Body.Fee,
		GasUsed: 0,
	}, nil
}

// Leave creates a leave request for a current validator.
func (vm *ValidatorModule) Leave(ctx context.Context, leaver []byte,
	txn *transactions.Transaction) (*ExecutionResponse, error) {

	err := vm.spend(ctx, leaver, txn.Body.Fee, txn.Body.Nonce)
	if err != nil {
		return nil, err
	}

	if err = vm.mgr.Leave(ctx, leaver); err != nil {
		return nil, err
	}

	return &ExecutionResponse{
		Fee:     txn.Body.Fee,
		GasUsed: 0,
	}, nil
}

// Approve records an approval transaction from a current validator..
func (vm *ValidatorModule) Approve(ctx context.Context, joiner []byte,
	txn *transactions.Transaction) (*ExecutionResponse, error) {
	approver := txn.Sender

	err := vm.spend(ctx, approver, txn.Body.Fee, txn.Body.Nonce)
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
