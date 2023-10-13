package accounts

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"hash"
	"math/big"
	"sync"

	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/internal/sessions"
	sqlSessions "github.com/kwilteam/kwil-db/internal/sessions/sql-session"
)

type AccountStore struct {
	db            Datastore
	log           log.Logger
	rw            sync.RWMutex
	gasEnabled    bool
	noncesEnabled bool
	stmts         *preparedStatements

	blockHasher hash.Hash
}

// Wrapper around the sql session committable that allows us to
// define module specific logic on how to capture the state of the
// account store independent of the SQL specifics.
type Committable struct {
	sessions.Committable

	dbHash      func() ([]byte, error)
	resetDBHash func()
}

// ID overrides the base Committable. We do this so that we can have the
// persistence of the state be part of the 2pc process, but have the ID reflect
// the actual state free from SQL specifics.
func (ac *Committable) ID(ctx context.Context) ([]byte, error) {
	return ac.dbHash()
}

// Wrapper around the Cancel method on the base Committable.
// This reset the updates recorded in the account store within a commit session.
func (ac *Committable) Cancel(ctx context.Context) error {
	ac.resetDBHash()
	return ac.Committable.Cancel(ctx)
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

	ar.blockHasher = sha256.New()
	return ar, nil
}

func (a *AccountStore) Close() error {
	// Close prepared statements
	if a.stmts != nil {
		if err := a.stmts.Close(); err != nil {
			return fmt.Errorf("failed to close prepared statements: %w", err)
		}
	}
	return nil
}

func (a *AccountStore) GetAccount(ctx context.Context, pubKey []byte) (*Account, error) {
	a.rw.RLock()
	defer a.rw.RUnlock()

	return a.getAccountReadOnly(ctx, pubKey)
}

// WrapCommittable wraps a sql session committable with the account store specific logic
func (a *AccountStore) WrapCommittable(committable *sqlSessions.SqlCommitable) *Committable {
	return &Committable{
		Committable: committable,
		dbHash:      a.dbHash,
		resetDBHash: a.resetBlockHasher,
	}
}

// dbHash captures the hash of the updates received by the account store
// within a commit session.
// Each update includes: [accountPubKey, amount, nonce]
func (a *AccountStore) dbHash() ([]byte, error) {
	hash := a.blockHasher.Sum(nil)
	a.blockHasher.Reset()
	return hash, nil
}

// resetBlockHasher resets the block hasher to its initial state.
// This is called when a commit session is cancelled.
func (a *AccountStore) resetBlockHasher() {
	a.log.Debug("resetting accounts block hasher")
	a.blockHasher.Reset()
}

type Spend struct {
	AccountPubKey []byte
	Amount        *big.Int
	Nonce         int64
}

// Spend spends an amount from an account. It blocks until the spend is written to the database.
func (a *AccountStore) Spend(ctx context.Context, spend *Spend) error {
	a.rw.Lock()
	defer a.rw.Unlock()

	balance, nonce, err := a.checkSpend(ctx, spend)
	if err != nil {
		return fmt.Errorf("failed to check spend: %w", err)
	}

	err = a.updateAccount(ctx, spend.AccountPubKey, balance, nonce)
	if err != nil {
		return fmt.Errorf("failed to update account: %w", err)
	}

	a.registerSpend(ctx, spend.AccountPubKey, balance, nonce)
	return nil
}

// AccountStore can get really large with many accounts.
// Therefore, computing the apphash of the entire account store can be expensive,
// and would involve a full scan of the db on write connection to get updated account information.
// Instead, we compute the apphash based on the updates to the account store per block.
// registerSpend: Hashes the user account info [pubkey, balance, nonce] after spending.
func (a *AccountStore) registerSpend(ctx context.Context, pubKey []byte, balance *big.Int, nonce int64) {
	a.blockHasher.Write(pubKey)
	a.blockHasher.Write(balance.Bytes())
	binary.Write(a.blockHasher, binary.LittleEndian, nonce)
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
