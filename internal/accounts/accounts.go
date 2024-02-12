package accounts

import (
	"context"
	"errors"
	"fmt"
	"math/big"

	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/internal/sql"
)

type AccountStore struct {
	log        log.Logger
	gasEnabled bool
}

func NewAccountStore(ctx context.Context, db sql.DB, opts ...AccountStoreOpts) (*AccountStore, error) {
	ar := &AccountStore{
		log: log.NewNoOp(),
	}

	sp, err := db.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer sp.Rollback(ctx)

	for _, opt := range opts {
		opt(ar)
	}

	err = initTables(ctx, sp)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize tables: %w", err)
	}

	return ar, sp.Commit(ctx)
}

func (a *AccountStore) GetAccount(ctx context.Context, tx sql.DB, ident []byte) (*Account, error) {
	acct, err := getAccount(ctx, tx, ident)
	if err == ErrAccountNotFound {
		return &Account{
			Identifier: ident,
			Balance:    big.NewInt(0),
			Nonce:      0,
		}, nil
	}
	return acct, err
}

// Transfer sends an amount from the sender's balance to another account. The
// amount sent is given by the amount. This does not affect the sending
// account's nonce; a Spend should precede this to pay for required transaction
// gas and validate/advance the nonce.
func (a *AccountStore) Transfer(ctx context.Context, tx sql.DB, to, from []byte, amt *big.Int) error {
	sp, err := tx.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer sp.Rollback(ctx)

	// Ensure that the from account balance is sufficient.
	account, err := getAccount(ctx, sp, from)
	if err != nil {
		return err
	}
	newBal, err := account.validateSpend(amt)
	if err != nil {
		return err
	}
	// Update or create the to account with the transferred amount.
	toAcct, err := getOrCreateAccount(ctx, sp, to)
	if err != nil {
		return err
	}
	// Decrement the from account balance first.
	err = updateAccount(ctx, sp, from, newBal, account.Nonce)
	if err != nil {
		return err
	}
	toBal := big.NewInt(0).Add(toAcct.Balance, amt)
	err = updateAccount(ctx, sp, to, toBal, toAcct.Nonce)
	if err != nil {
		return err
	}

	return sp.Commit(ctx)
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
func (a *AccountStore) Spend(ctx context.Context, tx sql.DB, spend *Spend) error {
	sp, err := tx.BeginTx(ctx) // using a tx in case we make an account but spend fails for some reason
	if err != nil {
		return fmt.Errorf("Spend: failed to begin transaction: %w", err)
	}
	defer sp.Rollback(ctx)

	var account *Account
	if a.gasEnabled && spend.Amount.Cmp(big.NewInt(0)) > 0 { // don't automatically create accounts when gas is required
		account, err = getAccount(ctx, sp, spend.AccountID)
	} else { // with no gas or a free transaction, we'll create the account if it doesn't exist
		account, err = getOrCreateAccount(ctx, sp, spend.AccountID)
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
			// Insufficient Funds: spend the entire balance in the account and increment the nonce
			err2 := updateAccount(ctx, sp, spend.AccountID, big.NewInt(0), spend.Nonce)
			if err2 != nil {
				return errors.Join(err, fmt.Errorf("Spend: failed to update account: %w", err2))
			}
			err2 = sp.Commit(ctx)
			if err2 != nil {
				return errors.Join(err, fmt.Errorf("Spend: failed to commit transaction: %w", err2))
			}

			return fmt.Errorf("Spend: failed to spend: %w", err)
		}
	} else {
		spend.Amount = big.NewInt(0)
	}

	newBal := new(big.Int).Sub(account.Balance, spend.Amount)
	err = updateAccount(ctx, sp, spend.AccountID, newBal, spend.Nonce)
	if err != nil {
		return fmt.Errorf("Spend: failed to update account: %w", err)
	}

	return sp.Commit(ctx)
}

// Credit credits an account. If the account does not exist, it will be created.
func (a *AccountStore) Credit(ctx context.Context, tx sql.DB, acctID []byte, amt *big.Int) error {
	// If exists, add to balance; if not, insert this balance and zero nonce.
	account, err := getAccount(ctx, tx, acctID)
	if err != nil {
		if !errors.Is(err, ErrAccountNotFound) {
			return err
		}
		return createAccount(ctx, tx, acctID, amt)
	}

	bal := new(big.Int).Add(account.Balance, amt)
	err = updateAccount(ctx, tx, account.Identifier, bal, account.Nonce) // same nonce
	if err != nil {
		return fmt.Errorf("failed to update account: %w", err)
	}

	return nil
}
