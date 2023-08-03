package balances

import (
	"context"
	"fmt"
	"math/big"
	"strings"
)

type preparedStatements struct {
	getAccount PreparedStatement
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
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	address TEXT NOT NULL UNIQUE,
	balance TEXT NOT NULL,
	nonce INTEGER NOT NULL
	);

CREATE TABLE IF NOT EXISTS chains (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	chain_code INTEGER NOT NULL UNIQUE,
	height INTEGER NOT NULL
);`
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

const sqlUpdateAccount = "UPDATE accounts SET balance = $balance, nonce = $nonce WHERE address = $address COLLATE NOCASE"

func (a *AccountStore) updateAccount(ctx context.Context, address string, amount *big.Int, nonce int64) error {
	return a.db.Execute(ctx, sqlUpdateAccount, map[string]interface{}{
		"$address": strings.ToLower(address),
		"$balance": amount.String(),
		"$nonce":   nonce,
	})
}

const sqlCreateAccount = "INSERT INTO accounts (address, balance, nonce) VALUES ($address, $balance, $nonce)"

// createAccount creates an account with the given address.
func (a *AccountStore) createAccount(ctx context.Context, address string) error {
	return a.db.Execute(ctx, sqlCreateAccount, map[string]interface{}{
		"$address": strings.ToLower(address),
		"$balance": big.NewInt(0).String(),
		"$nonce":   0,
	})
}

const sqlGetAccount = "SELECT balance, nonce FROM accounts WHERE address = $address COLLATE NOCASE"

// getAccountReadOnly gets an account using a read-only connection.
// it will not show uncommitted changes.
func (a *AccountStore) getAccountReadOnly(ctx context.Context, address string) (*Account, error) {
	results, err := a.db.Query(ctx, sqlGetAccount, map[string]interface{}{
		"$address": strings.ToLower(address),
	})
	if err != nil {
		return nil, err
	}

	acc, err := accountFromRecords(address, results)
	if err == errAccountNotFound {
		return emptyAccount(), nil
	}
	return acc, err
}

// getAccountSynchronous gets an account using a read-write connection.
// it will show uncommitted changes.
func (a *AccountStore) getAccountSynchronous(ctx context.Context, address string) (*Account, error) {
	results, err := a.stmts.getAccount.Execute(ctx, map[string]interface{}{
		"$address": strings.ToLower(address),
	})
	if err != nil {
		return nil, err
	}

	return accountFromRecords(address, results)
}

// accountFromRecords gets the first account from a list of records.
func accountFromRecords(address string, results []map[string]interface{}) (*Account, error) {
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
		Address: address,
		Balance: balance,
		Nonce:   nonce,
	}, nil
}

// getOrCreateAccount gets an account, creating it if it doesn't exist.
func (a *AccountStore) getOrCreateAccount(ctx context.Context, address string) (*Account, error) {
	account, err := a.getAccountSynchronous(ctx, address)
	if account == nil && err == errAccountNotFound {
		err = a.createAccount(ctx, address)
		if err != nil {
			return nil, fmt.Errorf("failed to create account: %w", err)
		}
		return emptyAccount(), nil
	}
	return account, err
}
