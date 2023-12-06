package accounts

import (
	"context"
	"errors"
	"fmt"
	"math/big"

	"github.com/kwilteam/kwil-db/core/types/transactions"
	"github.com/kwilteam/kwil-db/internal/accounts"
	"github.com/kwilteam/kwil-db/internal/modules"
)

var (
	transferPrice = big.NewInt(210_000)
	zeroInt       = big.NewInt(0)
)

// AccountsStore is the datastore required by the AccountsModule.
type AccountsStore interface {
	GetAccount(ctx context.Context, acctID []byte) (*accounts.Account, error)
	Credit(ctx context.Context, acctID []byte, amt *big.Int) error
	Transfer(ctx context.Context, to, from []byte, amt *big.Int) error
	Spend(ctx context.Context, spend *accounts.Spend) error
}

// AccountsModule is the blockchain module dealing with accounts updates and
// value transfers and credits.
type AccountsModule struct {
	store AccountsStore
}

// NewAccountsModule creates a new AccountsModule using the provided datastore.
func NewAccountsModule(store AccountsStore) *AccountsModule {
	return &AccountsModule{
		store: store,
	}
}

// PriceTransfer returns the cost of a transfer transaction.
func (am *AccountsModule) PriceTransfer(ctx context.Context) (*big.Int, error) {
	return transferPrice, nil
}

// TxAcct is the data from a transaction that pertains to the sending account.
type TxAcct struct {
	Sender []byte
	Fee    *big.Int
	Nonce  int64
}

// TransferTx is used to handle a transfer transaction. This involves:
//   - check that transaction fee is adequate
//   - pay for the transaction gas, updating balance and nonce in the DB
//   - transfer value from the sender account to the recipient account
//
// The blockchain application will have decoded the transaction and payload, and
// passed only the required data via TxAcct.
func (am *AccountsModule) TransferTx(ctx context.Context, tx *TxAcct, to []byte, amt *big.Int) error {
	if tx.Fee.Cmp(transferPrice) < 0 {
		return &modules.ABCIModuleError{
			Code:   transactions.CodeInsufficientFee,
			Detail: fmt.Sprintf(`fee %s is less than price %s`, tx.Fee.String(), transferPrice),
		}
		// return fmt.Errorf(`fee %s is less than price %s`, tx.Fee.String(), transferPrice)
	}

	// Pay for gas (the tx fee), and increment nonce.
	spend := &accounts.Spend{
		AccountID: tx.Sender,
		Amount:    transferPrice, // not the Fee???
		Nonce:     tx.Nonce,
	}
	err := am.store.Spend(ctx, spend)
	if err != nil {
		return fmt.Errorf("spend failed: %w", err) // this needs more distinction and coding
	} // STOP - make no state changes if nonce and balance did not update

	// Negative send amounts should be blocked at various levels, so we should
	// never get this, but be extra defensive since we cannot allow thievery.
	if amt.Cmp(zeroInt) < 0 {
		return errors.New("negative transfer amount")
	}

	err = am.store.Transfer(ctx, to, tx.Sender, amt)
	if err != nil {
		return fmt.Errorf("transfer failed: %w", err)
		// yes, they already paid for the transaction's gas and the block space
	}
	return nil
}

// Account returns the account information for the given account identifier.
func (am *AccountsModule) Account(ctx context.Context, acctID []byte) (*accounts.Account, error) {
	return am.store.GetAccount(ctx, acctID)
}

// Credit adds a certain amount to an account's balance, creating the account if
// it does not exist.
func (am *AccountsModule) Credit(ctx context.Context, acctID []byte, amt *big.Int) error {
	return am.store.Credit(ctx, acctID, amt)
}
