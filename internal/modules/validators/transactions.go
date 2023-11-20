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

func (vm *ValidatorModule) spend(ctx context.Context, acctID []byte,
	amt *big.Int, nonce uint64) error {
	return vm.accts.Spend(ctx, &accounts.Spend{
		AccountID: acctID,
		Amount:    amt,
		Nonce:     int64(nonce),
	})
}

// Join creates a join request for a prospective validator.
func (vm *ValidatorModule) Join(ctx context.Context, power int64,
	txn *transactions.Transaction) (*ExecutionResponse, error) {
	price, err := vm.PriceJoin(ctx)
	if err != nil {
		return nil, err
	}

	if txn.Body.Fee.Cmp(price) < 0 {
		return nil, fmt.Errorf("insufficient fee: %d < %d", txn.Body.Fee, price)
	}

	err = vm.spend(ctx, txn.Sender, txn.Body.Fee, txn.Body.Nonce)
	if err != nil {
		return nil, err
	}

	if err = vm.mgr.Join(ctx, txn.Sender, power); err != nil {
		return nil, err
	}

	return resp(txn.Body.Fee), nil
}

// Leave creates a leave request for a current validator.
func (vm *ValidatorModule) Leave(ctx context.Context, txn *transactions.Transaction) (*ExecutionResponse, error) {
	price, err := vm.PriceLeave(ctx)
	if err != nil {
		return nil, err
	}

	if txn.Body.Fee.Cmp(price) < 0 {
		return nil, fmt.Errorf("insufficient fee: %d < %d", txn.Body.Fee, price)
	}

	err = vm.spend(ctx, txn.Sender, txn.Body.Fee, txn.Body.Nonce)
	if err != nil {
		return nil, err
	}

	if err = vm.mgr.Leave(ctx, txn.Sender); err != nil {
		return nil, err
	}

	return resp(txn.Body.Fee), nil
}

// Approve records an approval transaction from a current validator.
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

// Remove records a removal transaction targeting the given validator pubkey.
// The transaction sender must be a current validator. The caller ensures that
// the transaction signature is verified.
func (vm *ValidatorModule) Remove(ctx context.Context, validator []byte,
	txn *transactions.Transaction) (*ExecutionResponse, error) {
	remover := txn.Sender
	price, err := vm.PriceRemove(ctx)
	if err != nil {
		return nil, err
	}

	if txn.Body.Fee.Cmp(price) < 0 {
		return nil, fmt.Errorf("insufficient fee: %d < %d", txn.Body.Fee, price)
	}

	err = vm.spend(ctx, remover, txn.Body.Fee, txn.Body.Nonce)
	if err != nil {
		return nil, err
	}

	if err = vm.mgr.Remove(ctx, validator, remover); err != nil {
		return nil, err
	}

	return &ExecutionResponse{
		Fee:     txn.Body.Fee,
		GasUsed: 0,
	}, nil
}
