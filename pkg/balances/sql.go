package balances

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/kwilteam/kwil-db/pkg/sql"
)

type preparedStatements struct {
	getAccount sql.Statement
}

func (p *preparedStatements) Close() error {
	return p.getAccount.Close()
}

func (a *AccountStore) prepareStatements() error {
	if a.stmts == nil {
		a.stmts = &preparedStatements{}
	}

	stmt, err := a.db.Prepare(sqlGetAccount)
	if err != nil {
		return fmt.Errorf("failed to prepare get account statement: %w", err)
	}

	a.stmts.getAccount = stmt

	return nil
}

const (
	sqlInitTables = `
CREATE TABLE IF NOT EXISTS accounts (
	public_key BLOB PRIMARY KEY,
	balance TEXT NOT NULL,
	nonce INTEGER NOT NULL
	) WITHOUT ROWID, STRICT;
	`
)

func getTableInits() []string {
	inits := strings.Split(sqlInitTables, ";")
	return inits[:len(inits)-1]
}

func (ar *AccountStore) initTables(ctx context.Context) error {
	inits := getTableInits()

	for _, init := range inits {
		err := ar.db.Execute(ctx, init, nil)
		if err != nil {
			return fmt.Errorf("failed to initialize tables: %w", err)
		}
	}

	return nil
}

const sqlUpdateAccount = "UPDATE accounts SET balance = $balance, nonce = $nonce WHERE public_key = $public_key COLLATE NOCASE"

func (a *AccountStore) updateAccount(ctx context.Context, pubKey []byte, amount *big.Int, nonce int64) error {
	return a.db.Execute(ctx, sqlUpdateAccount, map[string]interface{}{
		"$public_key": pubKey,
		"$balance":    amount.String(),
		"$nonce":      nonce,
	})
}

const sqlCreateAccount = "INSERT INTO accounts (public_key, balance, nonce) VALUES ($public_key, $balance, $nonce)"

// createAccount creates an account with the given public_key.
func (a *AccountStore) createAccount(ctx context.Context, pubKey []byte) error {
	return a.db.Execute(ctx, sqlCreateAccount, map[string]interface{}{
		"$public_key": pubKey,
		"$balance":    big.NewInt(0).String(),
		"$nonce":      0,
	})
}

const sqlGetAccount = "SELECT balance, nonce FROM accounts WHERE public_key = $public_key COLLATE NOCASE"

// getAccountReadOnly gets an account using a read-only connection.
// it will not show uncommitted changes.
func (a *AccountStore) getAccountReadOnly(ctx context.Context, pubKey []byte) (*Account, error) {
	results, err := a.db.Query(ctx, sqlGetAccount, map[string]interface{}{
		"$public_key": pubKey,
	})
	if err != nil {
		return nil, err
	}

	acc, err := accountFromRecords(pubKey, results)
	if err == errAccountNotFound {
		return emptyAccount(), nil
	}
	return acc, err
}

// getAccountSynchronous gets an account using a read-write connection.
// it will show uncommitted changes.
func (a *AccountStore) getAccountSynchronous(ctx context.Context, pubKey []byte) (*Account, error) {
	results, err := a.stmts.getAccount.Execute(ctx, map[string]interface{}{
		"$public_key": pubKey,
	})
	if err != nil {
		return nil, err
	}

	return accountFromRecords(pubKey, results)
}

// accountFromRecords gets the first account from a list of records.
func accountFromRecords(publicKey []byte, results []map[string]interface{}) (*Account, error) {
	if len(results) == 0 {
		return nil, errAccountNotFound
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
		PublicKey: publicKey,
		Balance:   balance,
		Nonce:     nonce,
	}, nil
}

// getOrCreateAccount gets an account, creating it if it doesn't exist.
func (a *AccountStore) getOrCreateAccount(ctx context.Context, pubKey []byte) (*Account, error) {
	account, err := a.getAccountSynchronous(ctx, pubKey)
	if account == nil && err == errAccountNotFound {
		err = a.createAccount(ctx, pubKey)
		if err != nil {
			return nil, fmt.Errorf("failed to create account: %w", err)
		}
		return emptyAccount(), nil
	}
	return account, err
}
