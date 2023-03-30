package balances

import (
	"fmt"
	"kwil/pkg/log"
	"kwil/pkg/sql/driver"
	"math/big"
	"sync"
)

type AccountStore struct {
	path string
	db   *driver.Connection
	log  log.Logger
	mu   *sync.Mutex
	wipe bool
}

func NewAccountStore(opts ...balancesOpts) (*AccountStore, error) {
	ar := &AccountStore{
		path: DefaultPath,
		log:  log.NewNoOp(),
		mu:   &sync.Mutex{},
		wipe: false,
	}

	for _, opt := range opts {
		opt(ar)
	}

	db, err := driver.OpenConn(accountDBName, driver.WithPath(ar.path))
	if err != nil {
		return nil, fmt.Errorf("failed to open account connection: %w", err)
	}
	err = db.AcquireLock()
	if err != nil {
		return nil, fmt.Errorf("failed to acquire lock: %w", err)
	}

	ar.db = db

	err = ar.initTables()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize tables: %w", err)
	}

	return ar, nil
}

func (a *AccountStore) GetAccount(address string) (*Account, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.getAccount(address)
}

type Spend struct {
	AccountAddress string
	Amount         *big.Int
	Nonce          int64
}

func (a *AccountStore) Spend(spend *Spend) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.spend(spend)
}

func (a *AccountStore) BatchSpend(spendList []*Spend) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	sp, err := a.db.Savepoint()
	if err != nil {
		return fmt.Errorf("failed to create savepoint: %w", err)
	}
	defer sp.Rollback()

	for _, spend := range spendList {
		err := a.spend(spend)
		if err != nil {
			return fmt.Errorf("failed to spend: %w", err)
		}
	}

	return sp.Commit()
}

func (a *AccountStore) spend(spend *Spend) error {
	account, err := a.getAccount(spend.AccountAddress)
	if err != nil {
		return fmt.Errorf("failed to get account: %w", err)
	}

	if account.Nonce != spend.Nonce {
		return ErrInvalidNonce
	}

	newBal := new(big.Int).Sub(account.Balance, spend.Amount)
	if newBal.Cmp(big.NewInt(0)) < 0 {
		return ErrInsufficientFunds
	}

	err = a.setBalance(spend.AccountAddress, newBal)
	if err != nil {
		return fmt.Errorf("failed to set balance: %w", err)
	}

	err = a.setNonce(spend.AccountAddress, spend.Nonce+1)
	if err != nil {
		return fmt.Errorf("failed to set nonce: %w", err)
	}

	return nil
}

type Credit struct {
	AccountAddress string
	Amount         *big.Int
}

func (a *AccountStore) Credit(credit *Credit) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.credit(credit)
}

func (a *AccountStore) BatchCredit(creditList []*Credit) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	sp, err := a.db.Savepoint()
	if err != nil {
		return fmt.Errorf("failed to create savepoint: %w", err)
	}
	defer sp.Rollback()

	for _, credit := range creditList {
		err := a.credit(credit)
		if err != nil {
			return fmt.Errorf("failed to credit: %w", err)
		}
	}

	return sp.Commit()
}

func (a *AccountStore) credit(credit *Credit) error {
	account, err := a.getAccount(credit.AccountAddress)
	if err != nil {
		if err == ErrAccountNotFound {
			err = a.createAccount(credit.AccountAddress)
			account = &Account{
				Address: credit.AccountAddress,
				Balance: big.NewInt(0),
				Nonce:   0,
			}

			if err != nil {
				return fmt.Errorf("failed to create account: %w", err)
			}
		} else {
			return fmt.Errorf("failed to get account: %w", err)
		}
	}

	newBal := new(big.Int).Add(account.Balance, credit.Amount)
	return a.setBalance(credit.AccountAddress, newBal)
}

func (a *AccountStore) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.db.Close()
}
