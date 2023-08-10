package balances

import (
	"context"
	"fmt"
	"math/big"
	"sync"

	"github.com/kwilteam/kwil-db/pkg/log"
)

type AccountStore struct {
	db            Datastore
	log           log.Logger
	rw            sync.RWMutex
	gasEnabled    bool
	noncesEnabled bool
	stmts         *preparedStatements
}

func NewAccountStore(ctx context.Context, datastore Datastore, opts ...AccountStoreOpts) (*AccountStore, error) {
	ar := &AccountStore{
		log: log.NewNoOp(),
		db:  datastore,
	}

	for _, opt := range opts {
		opt(ar)
	}

	err := ar.initTables(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize tables: %w", err)
	}

	err = ar.prepareStatements()
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statements: %w", err)
	}

	return ar, nil
}
func (a *AccountStore) GetAccount(ctx context.Context, address string) (*Account, error) {
	a.rw.RLock()
	defer a.rw.RUnlock()

	return a.getAccountReadOnly(ctx, address)
}

type Spend struct {
	AccountAddress string
	Amount         *big.Int
	Nonce          int64
}

// Spend spends an amount from an account. It blocks until the spend is written to the database.
func (a *AccountStore) Spend(ctx context.Context, spend *Spend) error {
	a.rw.Lock()
	defer a.rw.Unlock()

	balance, nonce, err := a.checkSpend(ctx, spend)
	if err != nil {
		return fmt.Errorf("failed to check spend: %w", err)
	}

	err = a.updateAccount(ctx, spend.AccountAddress, balance, nonce)
	if err != nil {
		return fmt.Errorf("failed to update account: %w", err)
	}

	return nil
}

// checkSpend checks that a spend is valid.  If gas costs are enabled, it checks that the account has enough gas to pay for the spend.
// If nonces are enabled, it checks that the nonce is correct.  It returns the new balance and nonce if the spend is valid. It returns an
// error if the spend is invalid.
func (a *AccountStore) checkSpend(ctx context.Context, spend *Spend) (*big.Int, int64, error) {
	account, err := a.getOrCreateAccount(ctx, spend.AccountAddress)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get account: %w", err)
	}

	nonce := account.Nonce
	if a.noncesEnabled {
		err = account.validateNonce(spend.Nonce)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to validate nonce: %w", err)
		}

		nonce = spend.Nonce
	}

	balance := account.Balance
	if a.gasEnabled {
		balance, err = account.validateSpend(spend.Amount)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to subtract gas: %w", err)
		}
	}

	return balance, nonce, nil
}
