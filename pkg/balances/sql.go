package balances

import (
	"fmt"
	"kwil/pkg/log"
	"kwil/pkg/sql/driver"
	"math/big"
)

const (
	sqlInitTables = `
CREATE TABLE IF NOT EXISTS accounts (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	address TEXT NOT NULL UNIQUE,
	balance TEXT NOT NULL,
	nonce INTEGER NOT NULL
	);`
)

func (ar *AccountStore) initTables() error {
	err := ar.db.Execute(sqlInitTables)
	if err != nil {
		return fmt.Errorf("failed to initialize tables: %w", err)
	}

	if ar.wipe {
		err = ar.wipeBalances()
		if err != nil {
			return fmt.Errorf("failed to wipe balances: %w", err)
		}
	}

	return nil
}

const sqlSetBalance = "UPDATE accounts SET balance = $balance WHERE address = $address"

func (a *AccountStore) setBalance(address string, amount *big.Int) error {
	return a.db.ExecuteNamed(sqlSetBalance, map[string]interface{}{
		"$address": address,
		"$balance": amount.String(),
	})
}

const sqlSetNonce = "UPDATE accounts SET nonce = $nonce WHERE address = $address"

func (a *AccountStore) setNonce(address string, nonce int64) error {
	return a.db.ExecuteNamed(sqlSetNonce, map[string]interface{}{
		"$address": address,
		"$nonce":   nonce,
	})
}

const sqlCreateAccount = "INSERT INTO accounts (address, balance, nonce) VALUES ($address, $balance, $nonce)"

func (a *AccountStore) createAccount(address string) error {
	return a.db.ExecuteNamed(sqlCreateAccount, map[string]interface{}{
		"$address": address,
		"$balance": big.NewInt(0).String(),
		"$nonce":   0,
	})
}

const sqlGetAccount = "SELECT balance, nonce FROM accounts WHERE address = $address"

func (a *AccountStore) getAccount(address string) (*Account, error) {
	var balance *big.Int
	var nonce int64
	exists := false
	err := a.db.QueryNamed(sqlGetAccount, func(stmt *driver.Statement) error {
		exists = true
		var ok bool

		stringBal := stmt.GetText("balance")
		balance, ok = new(big.Int).SetString(stringBal, 10)
		if !ok {
			a.log.Warn("failed to convert stored string balance to big int", log.Field{Key: "balance", String: stringBal}, log.Field{Key: "address", String: address})
			return ErrConvertToBigInt
		}

		nonce = stmt.GetInt64("nonce")

		return nil
	}, map[string]interface{}{
		"$address": address,
	})
	if err != nil {
		return nil, err
	}
	if !exists {
		a.log.Debug("account not found", log.Field{Key: "address", String: address})
		return nil, ErrAccountNotFound
	}

	return &Account{
		Address: address,
		Balance: balance,
		Nonce:   nonce,
	}, nil
}

const sqlWipeBalances = "DELETE FROM accounts"

func (a *AccountStore) wipeBalances() error {
	return a.db.Execute(sqlWipeBalances)
}
