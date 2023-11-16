package sessions

import (
	"bytes"
	"context"
	"fmt"
	"sync"

	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/internal/utils/order"
	"go.uber.org/zap"
)

// Committable is a datastore that can be included in a commit session.
type Committable interface {
	// Begin signals that a session is starting, and that the datastores
	// should begin allowing writes.
	// Commit should be called to signal that the session is complete.
	Begin(ctx context.Context, idempotencyKey []byte) error
	// BeginRecovery signals that the server is recovering from a crash.
	// It should be called instead of Begin, and Commit should be called
	// to signal that the recovery is complete.
	BeginRecovery(ctx context.Context, idempotencyKey []byte) error
	// Commit signals that the session is complete.
	// It returns a unique identifier for the session that
	// is generated deterministically from the applied changes.
	Commit(ctx context.Context, idempotencyKey []byte) ([]byte, error)
	// Cancel signals that the session is cancelled.
	// If a session is cancelled, it will be reset to the state that
	// it was in before Begin was called.
	Cancel(ctx context.Context) error
}

// MultiCommitter combines multiple committables into one.
// It does not implement Committable itself, but can be used
// to begin and commit multiple committables at once.
// It will persist the idempotency key in the KV store before
// beginning the session, and will delete it after the session.
// If it comes online and still has a session in progress, it will
// check to ensure the idempotency keys match, and if so, will begin recovery.
type MultiCommitter struct {
	committables map[string]Committable // the committables to commit
	inSession    bool                   // whether the committables are in a session
	kv           KV                     // KV tracks state about the committables
	mu           sync.Mutex             // mu protects the committables
	log          log.Logger
}

// NewCommitter creates a new committer.
func NewCommitter(kv KV, committables map[string]Committable, opts ...CommitterOpt) *MultiCommitter {
	c := &MultiCommitter{
		committables: committables,
		kv:           kv,
		log:          log.NewNoOp(),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// Begin begins a session for the committables.
// It will automatically detect if the idempotency key has been used before,
// and if so, will handle the recovery automatically.
func (m *MultiCommitter) Begin(ctx context.Context, idempotencyKey []byte) (err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	defer m.handleErr(ctx, &err)

	if m.inSession {
		return ErrInSession
	}
	m.inSession = true

	lastKey, err := getCurrentKey(m.kv)
	if err != nil {
		return err
	}

	// if lastKey is not nil, we are recovering from a crash
	if lastKey != nil {
		if !bytes.Equal(lastKey, idempotencyKey) {
			return fmt.Errorf("%w on recovery: expected %s, got %s", ErrIdempotencyKeyMismatch, lastKey, idempotencyKey)
		}
		for _, committable := range m.committables {
			err = committable.BeginRecovery(ctx, idempotencyKey)
			if err != nil {
				return err
			}
		}

		return nil
	}

	err = setCurrentKey(m.kv, idempotencyKey)
	if err != nil {
		return err
	}

	for _, committable := range m.committables {
		err = committable.Begin(ctx, idempotencyKey)
		if err != nil {
			return err
		}
	}

	return nil
}

// Commit commits the session for the committables.
// It returns a unique identifier for the session that
// is generated deterministically from the applied changes.
func (m *MultiCommitter) Commit(ctx context.Context, idempotencyKey []byte) (id []byte, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	defer m.handleErr(ctx, &err)

	if !m.inSession {
		return nil, ErrNotInSession
	}
	m.inSession = false

	lastKey, err := getCurrentKey(m.kv)
	if err != nil {
		return nil, err
	}

	if !bytes.Equal(lastKey, idempotencyKey) {
		return nil, fmt.Errorf("%w on commit: expected %s, got %s", ErrIdempotencyKeyMismatch, lastKey, idempotencyKey)
	}

	id = []byte{}

	orderedMap := order.OrderMap(m.committables)
	for _, c := range orderedMap {
		newId, err := c.Value.Commit(ctx, idempotencyKey)
		if err != nil {
			return nil, err
		}

		id = append(id, newId...)
	}

	err = deleteCurrentKey(m.kv)
	return id, err
}

// Register registers a committable with the committer.
func (m *MultiCommitter) Register(name string, committable Committable) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, ok := m.committables[name]
	if ok {
		return fmt.Errorf("committable %s already registered", name)
	}

	m.committables[name] = committable
	return nil
}

// handleErr checks if an error is nil or not.
// If it is not nil, it logs it, and notifies the committables that the session has been cancelled.
// It then sets the session state to not in progress.
func (a *MultiCommitter) handleErr(ctx context.Context, err *error) {
	if *err != nil {
		a.log.Error("error during atomic commit", zap.Error(*err))

		for _, committable := range a.committables {
			err := committable.Cancel(ctx)
			if err != nil {
				a.log.Error("error cancelling committable", zap.Error(err))
			}
		}
		a.inSession = false
	}

}
