package accounts

import (
	"context"
	"encoding/binary"
	"fmt"
	"math/big"
	"sync"

	"github.com/kwilteam/kwil-db/core/log"
)

// CommitRegister is an interface for registering a commit.
type CommitRegister interface {
	// Skip returns true if the commit should be skipped.
	// This isgnals that the account store should not be updated,
	// and simply return nil.
	Skip() bool

	// Register registers a commit.
	// This should be called when data is written to the database.
	Register(value []byte) error
}

type Datastore interface {
	Execute(ctx context.Context, stmt string, args map[string]any) ([]map[string]any, error)
	Query(ctx context.Context, query string, args map[string]any) ([]map[string]any, error)
}

type AccountStore struct {
	db            Datastore
	log           log.Logger
	rw            sync.RWMutex
	gasEnabled    bool
	noncesEnabled bool

	committable CommitRegister
}

func NewAccountStore(ctx context.Context, datastore Datastore, committable CommitRegister, opts ...AccountStoreOpts) (*AccountStore, error) {
	ar := &AccountStore{
		log:         log.NewNoOp(),
		db:          datastore,
		committable: committable,
	}

	for _, opt := range opts {
		opt(ar)
	}

	err := ar.initTables(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize tables: %w", err)
	}

	return ar, nil
}

func (a *AccountStore) Close() error {
	return nil
}

func (a *AccountStore) GetAccount(ctx context.Context, pubKey []byte) (*Account, error) {
	a.rw.RLock()
	defer a.rw.RUnlock()

	return a.getAccountReadOnly(ctx, pubKey)
}

type Spend struct {
	AccountPubKey []byte
	Amount        *big.Int
	Nonce         int64
}

func (s *Spend) bytes() []byte {
	bts := s.AccountPubKey
	bts = append(bts, s.Amount.Bytes()...)

	binary.LittleEndian.AppendUint64(bts, uint64(s.Nonce))

	return bts
}

// Spend spends an amount from an account. It blocks until the spend is written to the database.
func (a *AccountStore) Spend(ctx context.Context, spend *Spend) error {
	a.rw.Lock()
	defer a.rw.Unlock()

	if a.committable.Skip() {
		return nil
	}

	balance, nonce, err := a.checkSpend(ctx, spend)
	if err != nil {
		return fmt.Errorf("failed to check spend: %w", err)
	}

	err = a.updateAccount(ctx, spend.AccountPubKey, balance, nonce)
	if err != nil {
		return fmt.Errorf("failed to update account: %w", err)
	}

	return a.committable.Register(spend.bytes())
}

// checkSpend checks that a spend is valid.  If gas costs are enabled, it checks that the account has enough gas to pay for the spend.
// If nonces are enabled, it checks that the nonce is correct.  It returns the new balance and nonce if the spend is valid. It returns an
// error if the spend is invalid.
func (a *AccountStore) checkSpend(ctx context.Context, spend *Spend) (*big.Int, int64, error) {
	account, err := a.getOrCreateAccount(ctx, spend.AccountPubKey)
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
