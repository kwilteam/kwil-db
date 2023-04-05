package balances

import (
	"fmt"
	"kwil/pkg/sql/driver"
	"math/big"
	"strings"

	"go.uber.org/zap"
)

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

func (ar *AccountStore) initTables() error {
	inits := getTableInits()

	if ar.wipe {
		err := ar.dropAllTables()
		if err != nil {
			return fmt.Errorf("failed to wipe balances: %w", err)
		}
	}

	for _, init := range inits {
		err := ar.db.Execute(init)
		if err != nil {
			return fmt.Errorf("failed to initialize tables: %w", err)
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
			a.log.Warn("failed to convert stored string balance to big int", zap.String("address", address), zap.String("balance", stringBal))
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
		a.log.Debug("account not found", zap.String("address", address))
		return nil, ErrAccountNotFound
	}

	return &Account{
		Address: address,
		Balance: balance,
		Nonce:   nonce,
	}, nil
}

const sqlDropAccounts = "DROP TABLE accounts"
const sqlDropChains = "DROP TABLE chains"

func (a *AccountStore) dropAllTables() error {
	exists, err := a.db.TableExists("accounts")
	if err != nil {
		return fmt.Errorf("failed to check if accounts table exists: %w", err)
	}

	if exists {
		err = a.db.Execute(sqlDropAccounts)
		if err != nil {
			return fmt.Errorf("failed to drop accounts table: %w", err)
		}
	}

	exists, err = a.db.TableExists("chains")
	if err != nil {
		return fmt.Errorf("failed to check if chains table exists: %w", err)
	}

	if exists {
		err = a.db.Execute(sqlDropChains)
		if err != nil {
			return fmt.Errorf("failed to drop chains table: %w", err)
		}
	}

	return nil
}

const sqlGetChainHeight = "SELECT height FROM chains WHERE chain_code = $chain_code"

func (a *AccountStore) getChainHeight(chainCode int32) (int64, error) {
	var height int64
	exists := false
	err := a.db.QueryNamed(sqlGetChainHeight, func(stmt *driver.Statement) error {
		exists = true
		height = stmt.GetInt64("height")
		return nil
	}, map[string]interface{}{
		"$chain_code": chainCode,
	})
	if err != nil {
		return 0, err
	}
	if !exists {
		return 0, nil
	}

	return height, nil
}

const sqlSetChainHeight = "UPDATE chains SET height = $height WHERE chain_code = $chain_code"

func (a *AccountStore) setChainHeight(chainCode int32, height int64) error {
	return a.db.ExecuteNamed(sqlSetChainHeight, map[string]interface{}{
		"$chain_code": chainCode,
		"$height":     height,
	})
}

const sqlCreateChain = "INSERT INTO chains (chain_code, height) VALUES ($chain_code, $height)"

func (a *AccountStore) createChain(chainCode int32, height int64) error {
	return a.db.ExecuteNamed(sqlCreateChain, map[string]interface{}{
		"$chain_code": chainCode,
		"$height":     height,
	})
}

const sqlChainExists = "SELECT EXISTS(SELECT 1 FROM chains WHERE chain_code = $chain_code)"

func (a *AccountStore) chainExists(chainCode int32) (bool, error) {
	exists := false
	err := a.db.QueryNamed(sqlChainExists, func(stmt *driver.Statement) error {
		exists = stmt.GetBool("EXISTS(SELECT 1 FROM chains WHERE chain_code = $chain_code)")
		return nil
	}, map[string]interface{}{
		"$chain_code": chainCode,
	})
	if err != nil {
		return false, err
	}

	return exists, nil
}
