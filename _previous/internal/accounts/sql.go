package accounts

import (
	"context"
	"fmt"
	"math/big"

	sql "github.com/kwilteam/kwil-db/common/sql"
	"github.com/kwilteam/kwil-db/core/types"
)

const (
	schemaName = `kwild_accts`

	accountStoreVersion = 0

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

func initTables(ctx context.Context, tx sql.DB) error {
	_, err := tx.Execute(ctx, sqlInitTables)
	if err != nil {
		return fmt.Errorf("failed to initialize tables: %w", err)
	}

	return nil
}

// updateAccount updates the balance and nonce of an account.
func updateAccount(ctx context.Context, db sql.Executor, ident []byte, amount *big.Int, nonce int64) error {
	_, err := db.Execute(ctx, sqlUpdateAccount, amount.String(), nonce, ident)
	return err
}

// createAccount creates an account with the given identifier and
// initial balance. The nonce will be set to 0.
func createAccount(ctx context.Context, db sql.Executor, ident []byte, amt *big.Int) error {
	_, err := db.Execute(ctx, sqlCreateAccount, ident, amt.String(), int64(0))
	return err
}

func createAccountWithNonce(ctx context.Context, db sql.Executor, ident []byte, amt *big.Int, nonce int64) error {
	_, err := db.Execute(ctx, sqlCreateAccount, ident, amt.String(), nonce)
	return err
}

// getAccount retrieves an account from the database.
// if the account is not found, it returns nil, ErrAccountNotFound.
func getAccount(ctx context.Context, db sql.Executor, ident []byte) (*types.Account, error) {
	results, err := db.Execute(ctx, sqlGetAccount, ident)
	if err != nil {
		return nil, err
	}

	if len(results.Rows) == 0 {
		return nil, ErrAccountNotFound
	}
	if len(results.Rows) > 1 {
		return nil, fmt.Errorf("expected 1 row, got %d", len(results.Rows))
	}

	// rows[0][0] == balance
	// rows[0][1] == nonce

	stringBal, ok := results.Rows[0][0].(string)
	if !ok {
		return nil, fmt.Errorf("failed to convert stored string balance to big int")
	}

	balance, ok := new(big.Int).SetString(stringBal, 10)
	if !ok {
		return nil, ErrConvertToBigInt
	}

	nonce, ok := results.Rows[0][1].(int64)
	if !ok {
		return nil, fmt.Errorf("failed to convert stored nonce to int64")
	}

	return &types.Account{
		Identifier: ident,
		Balance:    balance,
		Nonce:      nonce,
	}, nil
}
