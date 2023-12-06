package accounts

import (
	"context"
	"fmt"
	"math/big"
)

const (
	sqlInitTables = `CREATE TABLE IF NOT EXISTS accounts (
		identifier BLOB PRIMARY KEY,
		balance TEXT NOT NULL,
		nonce INTEGER NOT NULL
	) WITHOUT ROWID, STRICT;`

	sqlCreateAccount = `INSERT INTO accounts (identifier, balance, nonce) VALUES ($identifier, $balance, $nonce)`

	sqlUpdateAccount = `UPDATE accounts SET balance = $balance,
						nonce = $nonce WHERE identifier = $identifier COLLATE NOCASE`

	sqlGetAccount = `SELECT balance, nonce FROM accounts WHERE identifier = $identifier`
)

func (ar *AccountStore) initTables(ctx context.Context) error {
	_, err := ar.db.Execute(ctx, sqlInitTables, nil)
	if err != nil {
		return fmt.Errorf("failed to initialize tables: %w", err)
	}
	// insert genesisAllocs (IF NOT EXISTS)?
	return nil
}

func (a *AccountStore) updateAccount(ctx context.Context, ident []byte, amount *big.Int, nonce int64) error {
	_, err := a.db.Execute(ctx, sqlUpdateAccount, map[string]interface{}{
		"$identifier": ident,
		"$balance":    amount.String(),
		"$nonce":      nonce,
	})
	return err
}

// createAccountWithBalance creates an account with the given identifier and
// initial balance.
func (a *AccountStore) createAccountWithBalance(ctx context.Context, ident []byte, amt *big.Int) error {
	_, err := a.db.Execute(ctx, sqlCreateAccount, map[string]interface{}{
		"$identifier": ident,
		"$balance":    amt.String(),
		"$nonce":      0,
	})
	return err
}

// createAccount creates an account with the given identifier.
func (a *AccountStore) createAccount(ctx context.Context, ident []byte) error {
	return a.createAccountWithBalance(ctx, ident, big.NewInt(0))
}

// getAccountReadOnly gets an account using a read-only connection. it will not
// show uncommitted changes. If the account does not exist, no error is
// returned, but an account with a nil identifier is returned.
func (a *AccountStore) getAccountReadOnly(ctx context.Context, ident []byte) (*Account, error) {
	results, err := a.db.Query(ctx, sqlGetAccount, map[string]interface{}{
		"$identifier": ident,
	})
	if err != nil {
		return nil, err
	}

	acc, err := accountFromRecords(ident, results)
	if err == ErrAccountNotFound {
		return emptyAccount(), nil
	}
	return acc, err
}

// getAccountSynchronous gets an account using a read-write connection.
// it will show uncommitted changes.
func (a *AccountStore) getAccountSynchronous(ctx context.Context, ident []byte) (*Account, error) {
	results, err := a.db.Execute(ctx, sqlGetAccount, map[string]interface{}{
		"$identifier": ident,
	})
	if err != nil {
		return nil, err
	}

	return accountFromRecords(ident, results)
}

// accountFromRecords gets the first account from a list of records.
func accountFromRecords(identifier []byte, results []map[string]interface{}) (*Account, error) {
	if len(results) == 0 {
		return nil, ErrAccountNotFound
	}

	stringBal, ok := results[0]["balance"].(string)
	if !ok {
		return nil, fmt.Errorf("failed to convert stored string balance to big int")
	}

	balance, ok := new(big.Int).SetString(stringBal, 10)
	if !ok {
		return nil, ErrConvertToBigInt
	}

	nonce, ok := results[0]["nonce"].(int64)
	if !ok {
		return nil, fmt.Errorf("failed to convert stored nonce to int64")
	}

	return &Account{
		Identifier: identifier,
		Balance:    balance,
		Nonce:      nonce,
	}, nil
}

// getOrCreateAccount gets an account, creating it if it doesn't exist.
func (a *AccountStore) getOrCreateAccount(ctx context.Context, ident []byte) (*Account, error) {
	account, err := a.getAccountSynchronous(ctx, ident)
	if account == nil && err == ErrAccountNotFound {
		err = a.createAccount(ctx, ident)
		if err != nil {
			return nil, fmt.Errorf("failed to create account: %w", err)
		}
		return emptyAccount(), nil
	}
	return account, err
}
