package accounts

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"sync"

	"github.com/kwilteam/kwil-db/core/log"
)

type Datastore interface {
	Execute(ctx context.Context, stmt string, args ...any) ([]map[string]any, error)
	Query(ctx context.Context, query string, args ...any) ([]map[string]any, error)
}

type AccountStore struct {
	db            Datastore
	log           log.Logger
	rw            sync.RWMutex
	gasEnabled    bool
	noncesEnabled bool
}

func NewAccountStore(ctx context.Context, datastore Datastore, opts ...AccountStoreOpts) (*AccountStore, error) {
	ar := &AccountStore{
		log: log.NewNoOp(),
		db:  datastore,
	}

	for _, opt := range opts {
		opt(ar)
	}

	err := ar.initTables(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize tables: %w", err)
	}

	return ar, nil
}

func (a *AccountStore) GetAccount(ctx context.Context, ident []byte) (*Account, error) {
	a.rw.RLock()
	defer a.rw.RUnlock()

	return a.getAccountReadOnly(ctx, ident)
}

// Transfer sends an amount from the sender's balance to another account. The
// amount sent is given by the amount. This does not affect the sending
// account's nonce; a Spend should precede this to pay for required transaction
// gas and validate/advance the nonce.
func (a *AccountStore) Transfer(ctx context.Context, to, from []byte, amt *big.Int) error {
	// Ensure that the from account balance is sufficient.
	account, err := a.getAccountSynchronous(ctx, from)
	if err != nil {
		return err
	}
	newBal, err := account.validateSpend(amt)
	if err != nil {
		return err
	}
	// Update or create the to account with the transferred amount.
	toAcct, err := a.getOrCreateAccount(ctx, to)
	if err != nil {
		return err
	}
	// Decrement the from account balance first.
	err = a.updateAccount(ctx, from, newBal, account.Nonce)
	if err != nil {
		return err
	}
	toBal := big.NewInt(0).Add(toAcct.Balance, amt)
	return a.updateAccount(ctx, to, toBal, toAcct.Nonce)
}

// Spend specifies a the fee and nonce of a transaction for an account. The
// amount has historically been associated with the transaction's fee (to pay
// for gas) i.e. the price of a certain transaction type.
type Spend struct {
	AccountID []byte
	Amount    *big.Int
	Nonce     int64
}

// Spend spends an amount from an account and records nonces. It blocks until the spend is written to the database.
// The following scenarios are possible when spending from an account:
// InvalidNonce:
//
//	If the nonce validation fails, no updates are made to the account and transaction is aborted.
//
// InsufficientFunds:
//
//	If GasCosts are enabled and the account doesn't have enough balance to pay for the transaction,
//	the entire balance is spent and records the nonce for the account and transaction is aborted.
//
// ValidSpend:
//
//	If account has enough funds, the amount is spent and the nonce is updated.
//	If GasCosts are disabled, only the nonces are updated for the account.
func (a *AccountStore) Spend(ctx context.Context, spend *Spend) error {
	a.rw.Lock()
	defer a.rw.Unlock()

	var account *Account
	var err error
	if a.gasEnabled && spend.Amount.Cmp(big.NewInt(0)) > 0 { // don't automatically create accounts when gas is required
		account, err = a.getAccountSynchronous(ctx, spend.AccountID)
	} else { // with no gas or a free transaction, we'll create the account if it doesn't exist
		account, err = a.getOrCreateAccount(ctx, spend.AccountID)
	}
	if err != nil {
		return fmt.Errorf("Spend: failed to get account: %w", err)
	}

	// Invalid Nonce: No updates to the account
	err = account.validateNonce(spend.Nonce)
	if err != nil {
		return fmt.Errorf("Spend: failed to validate nonce: %w", err)
	}

	// Spend only if the GasCosts are enabled.
	if a.gasEnabled {
		_, err = account.validateSpend(spend.Amount)
		if err != nil {
			// Insufficient Funds: spend the entire balance and update the nonce
			spend.Amount = account.Balance
			return errors.Join(err, a.spend(ctx, spend, account))
		}
	} else {
		spend.Amount = big.NewInt(0)
	}

	// Valid spend: update account balance and nonce
	return a.spend(ctx, spend, account)
}

func (a *AccountStore) spend(ctx context.Context, spend *Spend, account *Account) error {
	newBal := new(big.Int).Sub(account.Balance, spend.Amount)
	err := a.updateAccount(ctx, spend.AccountID, newBal, spend.Nonce)
	if err != nil {
		return fmt.Errorf("failed to update account: %w", err)
	}

	return nil
}

// Credit credits an account. If the account does not exist, it will be created.
func (a *AccountStore) Credit(ctx context.Context, acctID []byte, amt *big.Int) error {
	a.rw.Lock()
	defer a.rw.Unlock()

	// If exists, add to balance; if not, insert this balance and zero nonce.
	account, err := a.getAccountSynchronous(ctx, acctID)
	if err != nil {
		if !errors.Is(err, ErrAccountNotFound) {
			return err
		}
		return a.createAccountWithBalance(ctx, acctID, amt)
	}

	bal := new(big.Int).Add(account.Balance, amt)
	err = a.updateAccount(ctx, account.Identifier, bal, account.Nonce) // same nonce
	if err != nil {
		return fmt.Errorf("failed to update account: %w", err)
	}

	return nil
}
