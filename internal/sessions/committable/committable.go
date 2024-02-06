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

// Aggregator is a dummy committable that aggregates data passed to the Register
// method, only to return the hash from Commit. No actual DB transaction is
// created. This is an adapter for MultiCommitter when the underlying database
// is shared with another Committable that actually creates/commits/rollsback
// the transaction.
type Aggregator struct {
	hash hash.Hash
}

func (a *Aggregator) Begin(context.Context, []byte) error {
	a.hash = sha256.New()
	return nil
}

// Precommit gets the commit ID for the (end of the) session.
func (a *Aggregator) Precommit(context.Context) ([]byte, error) {
	defer a.hash.Reset()
	return a.hash.Sum(nil), nil
}

func (a *Aggregator) Commit(context.Context, []byte) error {
	return nil
}

func (a *Aggregator) Cancel(context.Context) error {
	if a.hash == nil {
		return nil
	}
	a.hash.Reset()
	return nil
}

func (a *Aggregator) Register(value []byte) error {
	_, err := a.hash.Write(value)
	return err
}

// Dummy is a dummy committable that simply returns the value from the ID
// function from Commit. No actual DB transaction is created. This is an adapter
// for MultiCommitter when the underlying database is shared with another
// Committable that actually creates/commits/rollsback the transaction.
type Dummy struct {
	ID func() ([]byte, error)
}

func (a *Dummy) Begin(context.Context, []byte) error {
	return nil
}

// Precommit gets the commit ID for the (end of the) session.
func (a *Dummy) Precommit(context.Context) ([]byte, error) {
	return a.ID()
}

// Commit commits the session ...
func (a *Dummy) Commit(context.Context, []byte) error {
	return nil
}

// Cancel cancels the session for the committables.
func (a *Dummy) Cancel(context.Context) error {
	return nil
}

type CommittableStore interface {
	Set(ctx context.Context, key []byte, value []byte) error
	Get(ctx context.Context, key []byte, sync bool) ([]byte, error) // sync should indicate if write conn to be used
	// BeginTx should create a transaction on the connection pool's one *write* conn
	// BeginTx(ctx context.Context) (sql.Tx, error) // need wrapper for pgx.TxOptions in and pgx.Tx out
	sql.TxBeginner
}

// Committable manages a database transaction lifetime. Unlike the
// sql/registry.Registry type that also implements internal/sessions.Committable
// interface, the commit ID derives from data passed to the Register method.
//
// WARNING: this type should NOT be used for the same underlying database
// connection since it will also attempt to begin a transaction. If sharing the
// connection with e.g. a Registry, one of Aggregator or Dummy type should
// instead be used for the purposes of just customizing the ID returned from
// Commit. This Committable type is only useful on a separate connection, and
// even in that case the replication monitor should only be sensitive to tables
// and schemas pertaining to that the other type.
type Committable struct { // this would be a connection Pool
	db CommittableStore

	mu sync.Mutex
	tx sql.TxCloser
	// hash creates a perpetual hash of the committable.
	hash hash.Hash
	skip bool // informs users of a corresponding data store to not execute

	// idFn is an alternative to hash, that allows the caller
	// to pass a function to generate an id.
	// This is useful for when the caller wants to define its own logic for generating IDs
	idFn func() ([]byte, error)
}

func New(db CommittableStore) *Committable {
	return &Committable{
		db:   db,
		hash: sha256.New(),
	}
}

// Begin begins a session for ...
// It will check the stored idempotency key and return an error if it is the same as the one provided.
func (a *Committable) Begin(ctx context.Context, idempotencyKey []byte) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.tx != nil {
		return fmt.Errorf("old session transaction still exists")
	}

	lastUsedKey, err := a.db.Get(ctx, IdempotencyKeyKey, true)
	if err != nil {
		return err
	}

	if bytes.Equal(lastUsedKey, idempotencyKey) {
		return fmt.Errorf("received duplicate idempotency key, no recovery possible")
	}

	tx, err := a.db.Begin(ctx)
	if err != nil {
		return err
	}
	a.tx = tx

	a.hash = sha256.New()

	return nil
}

// Cancel cancels the session for the committables.
func (a *Committable) Cancel(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.hash.Reset()
	a.skip = false

	if a.tx != nil {
		err := a.tx.Rollback(ctx)
		if err != nil {
			return err
		}
		a.tx = nil
	}

	return nil
}

// Commit commits the session ...
func (a *Committable) Commit(ctx context.Context, idempotencyKey []byte) ([]byte, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	defer a.hash.Reset()

	if a.tx == nil {
		return nil, fmt.Errorf("no session exists")
	}

	// NOTE: with dataset Registry using the same pg database, the WAL data
	// collected the Registry's Commit => CommitID includes ALL table changes
	// (across all postgresql schema) including the ones covered by this
	// Committable! There must be only one type using the replication monitor
	// for the commit hash, and it should not include changes pertaining to this
	// Committable's tables (schema).
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

	err := a.db.Set(ctx, ApphashKey, appHash) // just for recovery when re-commit still needs to return the apphash
	if err != nil {
		return nil, err
	}

	err = a.db.Set(ctx, IdempotencyKeyKey, idempotencyKey)
	if err != nil {
		return nil, err
	}

	err = a.tx.Commit(ctx)
	if err != nil {
		return nil, err
	}
	a.tx = nil
	// NOTE: a WAL based approach would use the DB's CommitID method here

	return appHash, nil
}

// Register registers a value to be used in the commit hash. NOTE: for accounts
// store, where vstore uses an idFunc, and engine does NOT USE
// Committable! (engine implements a Registry with the Commit method returning apphash)
func (a *Committable) Register(value []byte) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.tx == nil {
		return fmt.Errorf("cannot register value when not in a session")
	}

	if a.idFn != nil {
		return fmt.Errorf("cannot register value when using idFn")
	}

	a.hash.Write(value)

	return nil
}

// SetIDFunc sets the id function.
func (a *Committable) SetIDFunc(idFn func() ([]byte, error)) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.idFn = idFn
}

var (
	// used to identify the last used idempotency key in the store
	IdempotencyKeyKey = []byte("idempotencyKey")
	// used to identify the last used apphash in the store
	ApphashKey = []byte("apphash")
)
