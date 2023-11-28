package accounts

import (
	"context"
	"fmt"
	"math/big"
)

const (
	// both public_key and address will just become an identifier
	sqlInitAccountsTable = `CREATE TABLE IF NOT EXISTS accounts (
		public_key BLOB PRIMARY KEY,
		address TEXT,
		balance TEXT NOT NULL,
		nonce INTEGER NOT NULL
		) WITHOUT ROWID, STRICT;`
	sqlInitNewAccountsTable = `CREATE TABLE IF NOT EXISTS new_accounts (
		address TEXT PRIMARY KEY,
		balance TEXT NOT NULL
		) WITHOUT ROWID, STRICT;`

	sqlCreateAccount = `INSERT INTO accounts (public_key, address, balance, nonce) VALUES ($public_key, $address, $balance, $nonce)`

	sqlUpdateAccount = `UPDATE accounts SET balance = $balance, nonce = $nonce
		WHERE public_key = $public_key COLLATE NOCASE`
	sqlUpdateAccountBalance = `UPDATE accounts SET balance = $balance
		WHERE public_key = $public_key COLLATE NOCASE`

	sqlGetPendingAccount    = `SELECT address, balance FROM new_accounts WHERE address = $address`
	sqlDelPendingAccount    = `DELETE FROM new_accounts WHERE address = $address`
	sqlInsertPendingAccount = `INSERT INTO new_accounts (address, balance) VALUES ($address, $balance)`
	// ON CONFLICT (address) DO UPDATE SET balance = balance + $balance`
	sqlUpdatePendingAccountBalance = `UPDATE new_accounts SET balance = $balance
		WHERE address = $address COLLATE NOCASE`

	sqlGetAccount       = `SELECT balance, nonce FROM accounts WHERE public_key = $public_key`
	sqlGetAccountByAddr = `SELECT balance, nonce, public_key FROM accounts WHERE address = $address`
)

func (ar *AccountStore) initTables(ctx context.Context) error {
	_, err := ar.db.Execute(ctx, sqlInitAccountsTable, nil)
	if err != nil {
		return fmt.Errorf("failed to initialize accounts table: %w", err)
	}
	_, err = ar.db.Execute(ctx, sqlInitNewAccountsTable, nil)
	if err != nil {
		return fmt.Errorf("failed to initialize new_accounts table: %w", err)
	}
	return nil
}

func (a *AccountStore) updateAccount(ctx context.Context, pubKey []byte, amount *big.Int, nonce int64) error {
	_, err := a.db.Execute(ctx, sqlUpdateAccount, map[string]interface{}{
		"$public_key": pubKey,
		"$balance":    amount.String(),
		"$nonce":      nonce,
	})
	return err
}

func (a *AccountStore) updateAccountBalance(ctx context.Context, pubKey []byte, amount *big.Int) error {
	_, err := a.db.Execute(ctx, sqlUpdateAccountBalance, map[string]interface{}{
		"$public_key": pubKey,
		"$balance":    amount.String(),
	})
	return err
}

// createAccount creates an account with the given public_key.
func (a *AccountStore) createAccount(ctx context.Context, addr string, pubKey []byte, bal *big.Int, nonce int64) error {
	fmt.Println("Create Account with balance: ", bal.String())
	_, err := a.db.Execute(ctx, sqlCreateAccount, map[string]interface{}{
		"$public_key": pubKey,
		"$address":    addr,
		"$balance":    bal.String(),
		"$nonce":      nonce,
	})
	return err
}

// getAccountReadOnly gets an account using a read-only connection. it will not
// show uncommitted changes. If the account does not exist, no error is
// returned, but an account with a nil pubkey is returned.
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
	results, err := a.db.Execute(ctx, sqlGetAccount, map[string]interface{}{
		"$public_key": pubKey,
	})
	if err != nil {
		return nil, err
	}

	return accountFromRecords(pubKey, results)
}

func (a *AccountStore) getAccountByAddress(ctx context.Context, addr string) (*Account, error) {
	results, err := a.db.Execute(ctx, sqlGetAccountByAddr, map[string]interface{}{
		"$address": addr,
	})
	if err != nil {
		return nil, err
	}

	return accountFromRecords(nil, results)
}

func (a *AccountStore) getPendingAccountBalance(ctx context.Context, addr string) (*big.Int, error) {
	results, err := a.db.Execute(ctx, sqlGetPendingAccount, map[string]interface{}{
		"$address": addr,
	})
	if err != nil {
		return nil, err
	}

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

	return balance, nil
}

func (a *AccountStore) createPendingAccount(ctx context.Context, addr string, bal *big.Int) error {
	_, err := a.db.Execute(ctx, sqlInsertPendingAccount, map[string]interface{}{
		"$address": addr,
		"$balance": bal.String(),
	})
	return err
}

func (a *AccountStore) updatePendingAccount(ctx context.Context, addr string, bal *big.Int) error {
	_, err := a.db.Execute(ctx, sqlUpdatePendingAccountBalance, map[string]interface{}{
		"$address": addr,
		"$balance": bal.String(),
	})
	return err
}

func (a *AccountStore) deletePendingAccount(ctx context.Context, addr string) error {
	_, err := a.db.Execute(ctx, sqlDelPendingAccount, map[string]interface{}{
		"$address": addr,
	})
	return err
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
	fmt.Println("Account balance: ", stringBal)

	balance := big.NewInt(0)
	if stringBal != "<nil>" { // TODO: Remove this once identifier changes are merged
		balance, ok = new(big.Int).SetString(stringBal, 10)
		if !ok {
			return nil, ErrConvertToBigInt
		}
	}

	nonce, ok := results[0]["nonce"].(int64)
	if !ok {
		return nil, fmt.Errorf("failed to convert stored nonce to int64")
	}

	if publicKey == nil { // we're using sqlGetAccountByAddr
		publicKey, ok = results[0]["public_key"].([]byte)
		if !ok {
			return nil, fmt.Errorf("failed to convert stored public key to []byte")
		}
	}

	return &Account{
		PublicKey: publicKey,
		Balance:   balance,
		Nonce:     nonce,
	}, nil
}
