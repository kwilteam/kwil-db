package accounts

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"kwil/node/types/sql"
	"kwil/node/versioning"
	"kwil/types"
	"math/big"
	"sync"
)

// Accounts represents an in-memory cache of accounts stored in a PostgreSQL database.
// This is primarily used to optimize data reads.
type Accounts struct {
	mtx sync.RWMutex
	// records is an in-memory cache of account records.
	records map[string]*types.Account
	// updates is a map of account identifiers (hex-encoded) to updated record values.
	// These updates are applied to the records at the end of the block.
	updates map[string]*types.Account

	// TODO: use lru cache of a capacity
	// lru "github.com/hashicorp/golang-lru/v2"
	// cache   *lru.Cache[string, *types.Account]

}

func InitializeAccountStore(ctx context.Context, db sql.DB) (*Accounts, error) {
	upgradeFns := map[int64]versioning.UpgradeFunc{
		0: initTables,
	}

	err := versioning.Upgrade(ctx, db, schemaName, upgradeFns, accountStoreVersion)
	if err != nil {
		return nil, err
	}

	return &Accounts{
		records: make(map[string]*types.Account),
		updates: make(map[string]*types.Account),
	}, nil
}

// GetAccount retrieves the account with the given identifier. If the account does not exist,
// it will return an account with a balance of 0 and a nonce of 0.
func (a *Accounts) GetAccount(ctx context.Context, tx sql.Executor, account []byte) (*types.Account, error) {
	acct, err := a.getAccount(ctx, tx, account, false)
	if err != nil {
		if err == ErrAccountNotFound {
			return &types.Account{
				Identifier: account,
				Balance:    big.NewInt(0),
				Nonce:      0,
			}, nil
		}
		return nil, err
	}
	return acct, nil
}

// getAccount retrieves the account with the given identifier.
// If the account does not exist, it will return an error.
// If uncommitted is true, it will check the in-memory cache for the account.
func (a *Accounts) getAccount(ctx context.Context, tx sql.Executor, account []byte, uncommitted bool) (acct *types.Account, err error) {
	a.mtx.RLock()
	defer a.mtx.RUnlock()

	var ok bool
	if uncommitted {
		acct, ok = a.updates[hex.EncodeToString(account)]
		if ok {
			return acct, nil
		}
	}

	acct, ok = a.records[hex.EncodeToString(account)]
	if ok {
		return acct, nil
	}

	acct, err = getAccount(ctx, tx, account)
	if err != nil {
		return nil, err
	}

	// Add the account to the in-memory cache
	a.records[hex.EncodeToString(account)] = acct
	return acct, nil
}

// Credit credits an account with the given amount. If the account does not exist, it will be created.
// A negative amount will be treated as a debit. Accounts cannot have negative balances, and will
// return an error if the amount would cause the balance to go negative.
// It also adds a record to the in-memory cache.
func (a *Accounts) Credit(ctx context.Context, tx sql.Executor, account []byte, amt *big.Int) error {
	acct, err := a.getAccount(ctx, tx, account, true)
	if err != nil {
		if errors.Is(err, ErrAccountNotFound) {
			if amt.Sign() < 0 {
				return ErrNegativeBalance
			}

			return a.createAccount(ctx, tx, account, amt, 0)
		}
		return err
	}

	newBal := new(big.Int).Add(acct.Balance, amt)

	// If the new balance is negative (which is possible with a debit), we should fail
	if newBal.Sign() < 0 {
		return ErrNegativeBalance
	}

	return a.updateAccount(ctx, tx, account, newBal, acct.Nonce)
}

// Spend spends an amount from an account and records nonces. It blocks until the spend is written to the database.
// The nonce passed must be exactly one greater than the account's nonce. If the nonce is not valid, the spend will fail.
// If the account does not have enough funds to spend the amount, an ErrInsufficientFunds error will be returned.
func (a *Accounts) Spend(ctx context.Context, tx sql.Executor, account []byte, amount *big.Int, nonce int64) error {
	acct, err := a.getAccount(ctx, tx, account, true)
	if err != nil {
		// If amount is 0 and account does not exist, create the account
		// Ensure that the nonce is 1, as this is the first tx spend on this account
		if errors.Is(err, ErrAccountNotFound) && amount.Sign() == 0 && nonce == 1 {
			return a.createAccount(ctx, tx, account, amount, nonce)
		}

		return err
	}

	// Ensure that the nonce is exactly one greater than the account's nonce
	if acct.Nonce+1 != nonce {
		return fmt.Errorf("%w: expected nonce %d, got %d", ErrInvalidNonce, acct.Nonce+1, nonce)
	}

	// Ensure that the balance is sufficient
	newBal := new(big.Int).Sub(acct.Balance, amount)
	if newBal.Sign() < 0 {
		return errInsufficientFunds(account, amount, acct.Balance)
	}

	return a.updateAccount(ctx, tx, account, newBal, nonce)
}

// ApplySpend spends an amount from an account. It blocks until the spend is written to the database.
// This is used by the new nodes during migration to replicate spends from the old network to the new network.
// If the account does not have enough funds to spend the amount, spend the entire balance.
// Nonces on the new network take precedence over the old network, so the nonces are not checked.
func (a *Accounts) ApplySpend(ctx context.Context, tx sql.Executor, account []byte, amount *big.Int, nonce int64) error {
	acct, err := a.getAccount(ctx, tx, account, true)
	if err != nil {
		// Spends should not occur on accounts that do not exist during migration as credits are disabled.
		return err
	}

	// If the balance is insufficient, spend the entire balance, else spend the amount
	newBal := new(big.Int).Sub(acct.Balance, amount)
	if newBal.Sign() < 0 {
		newBal = big.NewInt(0)
	}

	return a.updateAccount(ctx, tx, account, newBal, acct.Nonce)
}

// Transfer transfers an amount from one account to another. If the from account does not have enough funds to transfer the amount,
// it will fail. If the to account does not exist, it will be created. The amount must be greater than 0.
func (a *Accounts) Transfer(ctx context.Context, db sql.TxMaker, from, to []byte, amt *big.Int) error {
	if amt.Sign() < 0 {
		return ErrNegativeTransfer
	}

	tx, err := db.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Ensure that the from account balance is sufficient.
	account, err := a.getAccount(ctx, tx, from, true)
	if err != nil {
		return err
	}

	newFromBal := new(big.Int).Sub(account.Balance, amt)
	if newFromBal.Sign() < 0 {
		return errInsufficientFunds(from, amt, account.Balance)
	}

	// add the balance to the to new account
	receiver, err := a.getAccount(ctx, tx, to, true)
	if err != nil {
		if errors.Is(err, ErrAccountNotFound) {
			err2 := a.createAccount(ctx, tx, to, amt, 0)
			if err2 != nil {
				return err2
			}
		} else {
			return err
		}
	} else {
		newToBal := new(big.Int).Add(receiver.Balance, amt)
		err = a.updateAccount(ctx, tx, to, newToBal, receiver.Nonce)
		if err != nil {
			return err
		}
	}

	// decrement the balance from the from account
	err = a.updateAccount(ctx, tx, from, newFromBal, account.Nonce)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// Commit applies all the updates to the in-memory cache.
// This is called after the updates are written to the pg database.
func (a *Accounts) Commit() error {
	a.mtx.Lock()
	defer a.mtx.Unlock()

	for k, v := range a.updates {
		a.records[k] = v
	}

	a.updates = make(map[string]*types.Account)
	return nil
}

func (a *Accounts) createAccount(ctx context.Context, tx sql.Executor, account []byte, amt *big.Int, nonce int64) error {
	if err := createAccount(ctx, tx, account, amt, nonce); err != nil {
		return err
	}

	// Record the account creation in the updates
	a.mtx.Lock()
	defer a.mtx.Unlock()
	a.updates[hex.EncodeToString(account)] = &types.Account{
		Identifier: account,
		Balance:    amt,
		Nonce:      nonce,
	}

	return nil
}

func (a *Accounts) updateAccount(ctx context.Context, tx sql.Executor, account []byte, amount *big.Int, nonce int64) error {
	if err := updateAccount(ctx, tx, account, amount, nonce); err != nil {
		return err
	}

	// Record the account update in the updates
	a.mtx.Lock()
	defer a.mtx.Unlock()

	acct, ok := a.updates[hex.EncodeToString(account)]
	if ok {
		acct.Balance = amount
		acct.Nonce = nonce
	}

	return nil
}
