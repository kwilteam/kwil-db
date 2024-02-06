package accounts

import (
	"context"
	"fmt"
	"math/big"
)

const (
	schemaName      = `kwild_accts`
	sqlCreateSchema = `CREATE SCHEMA IF NOT EXISTS ` + schemaName

	sqlInitTables = `CREATE TABLE IF NOT EXISTS ` + schemaName + `.accounts (
		identifier BYTEA PRIMARY KEY,
		balance TEXT NOT NULL, -- consider: NUMERIC(32) for uint256 and pgx.Numeric will handle it and provide a *big.Int field
		nonce BIGINT NOT NULL -- a.k.a. INT8
	);`

	sqlCreateAccount = `INSERT INTO ` + schemaName + `.accounts (identifier, balance, nonce) VALUES ($1, $2, $3)`

	sqlUpdateAccount = `UPDATE ` + schemaName + `.accounts SET balance = $1, nonce = $2
		WHERE identifier = $3`

	sqlGetAccount = `SELECT balance, nonce FROM ` + schemaName + `.accounts WHERE identifier = $1`
)

func (a *AccountStore) initTables(ctx context.Context) error {
	if _, err := a.db.Execute(ctx, sqlCreateSchema); err != nil {
		return err
	}
	_, err := a.db.Execute(ctx, sqlInitTables)
	if err != nil {
		return fmt.Errorf("failed to initialize tables: %w", err)
	}
	return nil
}

func (a *AccountStore) updateAccount(ctx context.Context, ident []byte, amount *big.Int, nonce int64) error {
	_, err := a.db.Execute(ctx, sqlUpdateAccount, amount.String(), nonce, ident)
	return err
}

// createAccountWithBalance creates an account with the given identifier and
// initial balance.
func (a *AccountStore) createAccountWithBalance(ctx context.Context, ident []byte, amt *big.Int) error {
	_, err := a.db.Execute(ctx, sqlCreateAccount, ident, amt.String(), 0)
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
	results, err := a.db.Query(ctx, sqlGetAccount, ident)
	if err != nil {
		return nil, err
	}

	acc, err := accountFromRecords(ident, results)
	if err == ErrAccountNotFound {
		return emptyAccount(), nil
	}
	return acc, err
}

// getAccountSynchronous gets an account using a read-write transaction, if in
// one. It will show uncommitted changes. It also is different from
// getAccountReadOnly in that a nil account is returned if none exists. This
// should ONLY be used from calls where a write transaction exists (in session).
func (a *AccountStore) getAccountSynchronous(ctx context.Context, ident []byte) (*Account, error) {
	results, err := a.db.Execute(ctx, sqlGetAccount, ident)
	// if errors.Is(err, sql.ErrNoTransaction) { // if this is needed, it's a sign of bad design elsewhere
	// 	results, err = a.db.Query(ctx, sqlGetAccount, ident)
	// }
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
