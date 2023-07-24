package balances

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/kwilteam/kwil-db/pkg/log"
	"github.com/kwilteam/kwil-db/pkg/sql/driver"

	"go.uber.org/zap"
)

type AccountStore struct {
	path           string
	db             *driver.Connection
	log            log.Logger
	mu             *sync.Mutex
	wipe           bool
	gas_enabled    bool
	nonces_enabled bool
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
	ar.log.Named("account_store")

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

func (a *AccountStore) UpdateGasCosts(gas_enabled bool) {
	a.gas_enabled = gas_enabled
}

func (a *AccountStore) GasEnabled() bool {
	return a.gas_enabled
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

type ChainConfig struct {
	ChainCode int32
	Height    int64
}

// Spend spends an amount from an account.
func (a *AccountStore) Spend(spend *Spend) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.spend(spend)
}

// BatchSpend spends a list of spends in a single transaction.  It can optionally
// update the chain height, however nil can be passed to skip this.
func (a *AccountStore) BatchSpend(spendList []*Spend, chain *ChainConfig) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	sp, err := a.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to create Begin: %w", err)
	}
	defer sp.Rollback()

	for _, spend := range spendList {
		err := a.spend(spend)
		if err != nil {
			return fmt.Errorf("failed to spend: %w", err)
		}
	}

	if chain != nil {
		err := a.setChainHeight(chain.ChainCode, chain.Height)
		if err != nil {
			return fmt.Errorf("failed to set chain height: %w", err)
		}
	}

	return sp.Commit()
}

func (a *AccountStore) spend(spend *Spend) error {
	account, err := a.getOrCreateAccount(spend.AccountAddress)
	if err != nil {
		return fmt.Errorf("failed to get account: %w", err)
	}

	if a.nonces_enabled && account.Nonce+1 != spend.Nonce {
		a.log.Debug("tx nonce incorrect", zap.String("address", spend.AccountAddress), zap.Int64("expected", account.Nonce), zap.Int64("actual", spend.Nonce))
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

	if a.nonces_enabled {
		err = a.setNonce(spend.AccountAddress, spend.Nonce)
		if err != nil {
			return fmt.Errorf("failed to set nonce: %w", err)
		}
	}

	return nil
}

type Credit struct {
	AccountAddress string
	Amount         *big.Int
}

// Credit credits an account.
func (a *AccountStore) Credit(credit *Credit) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.credit(credit)
}

// BatchCredit credits a list of credits in a single transaction.  It can optionally
// update the chain height, however nil can be passed to skip this.
func (a *AccountStore) BatchCredit(creditList []*Credit, chain *ChainConfig) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	sp, err := a.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to create Begin: %w", err)
	}
	defer sp.Rollback()

	for _, credit := range creditList {
		err := a.credit(credit)
		if err != nil {
			return fmt.Errorf("failed to credit: %w", err)
		}
	}

	if chain != nil {
		err := a.setChainHeight(chain.ChainCode, chain.Height)
		if err != nil {
			return fmt.Errorf("failed to set chain height: %w", err)
		}
	}

	return sp.Commit()
}

func (a *AccountStore) credit(credit *Credit) error {
	//TODO:  use getOrCreateAccount here?
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

	a.log.Info("crediting account", zap.String("address", account.Address), zap.String("amount", credit.Amount.String()))
	newBal := new(big.Int).Add(account.Balance, credit.Amount)
	return a.setBalance(credit.AccountAddress, newBal)
}

func (a *AccountStore) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.db.Close()
}
