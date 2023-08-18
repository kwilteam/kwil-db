package validators

import (
	"context"
	"math/big"

	"github.com/kwilteam/kwil-db/pkg/balances"
	"github.com/kwilteam/kwil-db/pkg/transactions"
)

// Join/Leave/Approve required a spend. There is currently no pricing associated
// with the actions, although there probably should be for Join.

func (vm *ValidatorModule) spend(ctx context.Context, acctAddr string,
	amt *big.Int, nonce uint64) error {
	return vm.accts.Spend(ctx, &balances.Spend{
		AccountAddress: acctAddr,
		Amount:         amt,
		Nonce:          int64(nonce),
	})
}

// Join creates a join request for a prospective validator.
func (vm *ValidatorModule) Join(ctx context.Context, joiner []byte, power int64,
	txn *transactions.Transaction) (*transactions.TransactionStatus, error) {
	joinerAddr := vm.addr.Address(joiner)
	// comet-aware way:
	// candidateAddr, _ := pubkeyToAddr(joiner)

	err := vm.spend(ctx, joinerAddr, txn.Body.Fee, txn.Body.Nonce)
	if err != nil {
		return nil, err
	}

	if err = vm.mgr.Join(ctx, joiner, power); err != nil {
		return nil, err
	}

	return &transactions.TransactionStatus{
		Fee: txn.Body.Fee,
	}, nil
}

// Leave creates a leave request for a current validator.
func (vm *ValidatorModule) Leave(ctx context.Context, leaver []byte,
	txn *transactions.Transaction) (*transactions.TransactionStatus, error) {
	leaverAddr := vm.addr.Address(leaver)

	err := vm.spend(ctx, leaverAddr, txn.Body.Fee, txn.Body.Nonce)
	if err != nil {
		return nil, err
	}

	if err = vm.mgr.Leave(ctx, leaver); err != nil {
		return nil, err
	}

	return &transactions.TransactionStatus{
		Fee: txn.Body.Fee,
	}, nil
}

// Approve records an approval transaction from a current validator.
//
// ISSUE: The approver is the tx Sender, with the BIG special case that Sender
// is the base64-encoded pubkey, not an address as with most other Kwil txns.
func (vm *ValidatorModule) Approve(ctx context.Context, joiner []byte,
	txn *transactions.Transaction) (*transactions.TransactionStatus, error) {
	approver := txn.Sender
	approverAddr := vm.addr.Address(approver)

	err := vm.spend(ctx, approverAddr, txn.Body.Fee, txn.Body.Nonce)
	if err != nil {
		return nil, err
	}

	if err = vm.mgr.Approve(ctx, joiner, approver); err != nil {
		return nil, err
	}

	return &transactions.TransactionStatus{
		Fee: txn.Body.Fee,
	}, nil
}
