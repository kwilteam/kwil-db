// Package committable provides an easy to use interface for creating committables.
package committable

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"hash"
	"sync"

	"github.com/kwilteam/kwil-db/internal/sql"
)

type CommittableStore interface {
	Set(ctx context.Context, key []byte, value []byte) error
	Get(ctx context.Context, key []byte, sync bool) ([]byte, error)
	Savepoint() (sql.Savepoint, error)
}

// SavepointCommittable is a committable that can be used to commit a savepoint.
type SavepointCommittable struct {
	db        CommittableStore
	mu        sync.Mutex
	savepoint sql.Savepoint
	// hash creates a perpetual hash of the committable.
	hash hash.Hash
	skip bool

	// idFn is an alternative to hash, that allows the caller
	// to pass a function to generate an id.
	// This is useful for when the caller wants to define its own logic for generating IDs
	idFn func() ([]byte, error)

	writable bool
}

func New(db CommittableStore) *SavepointCommittable {
	return &SavepointCommittable{
		db:   db,
		hash: sha256.New(),
	}
}

// Begin begins a session for the account store.
// It will check the stored idempotency key and return an error if it is the same as the one provided.
func (a *SavepointCommittable) Begin(ctx context.Context, idempotencyKey []byte) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.savepoint != nil {
		return fmt.Errorf("old session savepoint still exists")
	}

	if a.writable {
		return fmt.Errorf("old session still exists")
	}
	a.writable = true

	lastUsedKey, err := a.db.Get(ctx, IdempotencyKeyKey, true)
	if err != nil {
		return err
	}

	if bytes.Equal(lastUsedKey, idempotencyKey) {
		return fmt.Errorf("received duplicate idempotency key during non-recovery: %s", idempotencyKey)
	}

	sp, err := a.db.Savepoint()
	if err != nil {
		return err
	}
	a.savepoint = sp

	a.hash = sha256.New()

	return a.db.Set(ctx, IdempotencyKeyKey, idempotencyKey)
}

// BeginRecovery begins a session for the account store in recovery mode.
// If the idempotency key is the same as the last used one, it will skip all incoming spends.
func (a *SavepointCommittable) BeginRecovery(ctx context.Context, idempotencyKey []byte) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.writable {
		return fmt.Errorf("old session still exists")
	}
	a.writable = true

	lastUsedKey, err := a.db.Get(ctx, IdempotencyKeyKey, true)
	if err != nil {
		return err
	}

	if bytes.Equal(lastUsedKey, idempotencyKey) {
		a.skip = true
		return nil
	}

	sp, err := a.db.Savepoint()
	if err != nil {
		return err
	}
	a.savepoint = sp

	a.hash = sha256.New()

	return a.db.Set(ctx, IdempotencyKeyKey, idempotencyKey)
}

// Cancel cancels the session for the committables.
func (a *SavepointCommittable) Cancel(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.hash.Reset()
	a.skip = false
	a.writable = false

	if a.savepoint != nil {
		err := a.savepoint.Rollback()
		if err != nil {
			return err
		}
		a.savepoint = nil
	}

	return nil
}

// Commit commits the session for the account store.
func (a *SavepointCommittable) Commit(ctx context.Context, idempotencyKey []byte) ([]byte, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	defer a.hash.Reset()

	if !a.writable {
		return nil, fmt.Errorf("no session exists")
	}
	a.writable = false

	if a.skip {
		a.skip = false
		return a.db.Get(ctx, ApphashKey, true)
	}

	if a.savepoint == nil {
		return nil, fmt.Errorf("no session exists")
	}

	var appHash []byte
	if a.idFn != nil {
		var err error
		appHash, err = a.idFn()
		if err != nil {
			return nil, err
		}
	} else {
		appHash = a.hash.Sum(nil)
	}

	err := a.db.Set(ctx, ApphashKey, appHash)
	if err != nil {
		return nil, err
	}

	err = a.db.Set(ctx, IdempotencyKeyKey, idempotencyKey)
	if err != nil {
		return nil, err
	}

	err = a.savepoint.Commit()
	if err != nil {
		return nil, err
	}
	a.savepoint = nil

	return appHash, nil
}

// Register registers a value to be used in the commit hash.
func (a *SavepointCommittable) Register(value []byte) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.writable {
		return fmt.Errorf("cannot register value when not in a session")
	}

	if a.idFn != nil {
		return fmt.Errorf("cannot register value when using idFn")
	}

	if a.skip {
		return nil
	}

	a.hash.Write(value)

	return nil
}

// SetIDFunc sets the id function.
func (a *SavepointCommittable) SetIDFunc(idFn func() ([]byte, error)) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.idFn = idFn
}

// Skip returns whether or not the committable is in skip mode.
func (a *SavepointCommittable) Skip() bool {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.skip
}

var (
	// used to identify the last used idempotency key in the store
	IdempotencyKeyKey = []byte("idempotencyKey")
	// used to identify the last used apphash in the store
	ApphashKey = []byte("apphash")
)
