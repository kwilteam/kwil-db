package accounts

import (
	"context"
	"errors"
	"fmt"
	"math/big"

	sql "github.com/kwilteam/kwil-db/common/sql"
	"github.com/kwilteam/kwil-db/core/types"
)

// InitializeAccountStore initializes the account store schema and tables.
func InitializeAccountStore(ctx context.Context, db sql.DB) error {
	sp, err := db.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer sp.Rollback(ctx)

	err = initTables(ctx, sp)
	if err != nil {
		return fmt.Errorf("failed to initialize tables: %w", err)
	}

	return sp.Commit(ctx)
}

// Credit credits an account with the given amount. If the account does not exist, it will be created.
// A negative amount will be treated as a debit. Accounts cannot have negative balances, and will
// return an error if the amount would cause the balance to go negative.
func Credit(ctx context.Context, tx sql.DB, account []byte, amt *big.Int) error {
	acct, err := getAccount(ctx, tx, account)
	if err != nil {
		if errors.Is(err, ErrAccountNotFound) {
			// if account does not exist, we should create it with a balance,
			// as long as the credit amount is not negative
			if amt.Sign() < 0 {
				return ErrNegativeBalance
			}

			return createAccount(ctx, tx, account, amt)
		}
		return err
	}

	newBal := new(big.Int).Add(acct.Balance, amt)

	// if the new balance is negative (which is possible with a debit), we should fail
	if newBal.Sign() < 0 {
		return ErrNegativeBalance
	}

	return updateAccount(ctx, tx, account, newBal, acct.Nonce)
}

// GetAccount retrieves the account with the given identifier. If the account does not exist, it will
// return an account with a balance of 0 and a nonce of 0.
func GetAccount(ctx context.Context, tx sql.DB, account []byte) (*types.Account, error) {
	acct, err := getAccount(ctx, tx, account)
	if err == ErrAccountNotFound {
		return &types.Account{
			Identifier: account,
			Balance:    big.NewInt(0),
			Nonce:      0,
		}, nil
	}
	return acct, err
}

// Spend spends an amount from an account and records nonces. It blocks until the spend is written to the database.
// The nonce passed must be exactly one greater than the account's nonce. If the nonce is not valid, the spend will fail.
// If the account does not have enough funds to spend the amount, an ErrInsufficientFunds error will be returned.
func Spend(ctx context.Context, tx sql.DB, account []byte, amount *big.Int, nonce int64) error {
	acct, err := getAccount(ctx, tx, account)
	if err != nil {
		// if amount is 0 and account not found, create the account
		// we check that nonce is 1, since this is the first tx
		if errors.Is(err, ErrAccountNotFound) && amount.Sign() == 0 && nonce == 1 {
			return createAccountWithNonce(ctx, tx, account, amount, nonce)
		}

		return err
	}

	if nonce != acct.Nonce+1 {
		return fmt.Errorf("%w: expected %d, got %d", ErrInvalidNonce, acct.Nonce+1, nonce)
	}

	newBal := new(big.Int).Sub(acct.Balance, amount)
	// if negative, spend the entire balance and increment the nonce
	if newBal.Sign() < 0 {
		return errInsufficientFunds(account, amount, acct.Balance)
	}

	return updateAccount(ctx, tx, account, newBal, nonce)
}

// Transfer transfers an amount from one account to another. If the from account does not have enough funds to transfer the amount,
// it will fail. If the to account does not exist, it will be created. The amount must be greater than 0.
func Transfer(ctx context.Context, db sql.DB, from, to []byte, amt *big.Int) error {
	if amt.Sign() < 0 {
		return ErrNegativeTransfer
	}

	tx, err := db.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Ensure that the from account balance is sufficient.
	account, err := getAccount(ctx, tx, from)
	if err != nil {
		return err
	}

	newFromBal := new(big.Int).Sub(account.Balance, amt)
	if newFromBal.Sign() < 0 {
		return errInsufficientFunds(from, amt, account.Balance)
	}

	// add the balance to the to new account
	receiver, err := getAccount(ctx, tx, to)
	if err != nil {
		if errors.Is(err, ErrAccountNotFound) {
			err2 := createAccount(ctx, tx, to, amt)
			if err2 != nil {
				return err2
			}
		} else {
			return err
		}
	} else {
		newToBal := new(big.Int).Add(receiver.Balance, amt)
		err = updateAccount(ctx, tx, to, newToBal, receiver.Nonce)
		if err != nil {
			return err
		}
	}

	// decrement the balance from the from account
	err = updateAccount(ctx, tx, from, newFromBal, account.Nonce)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
