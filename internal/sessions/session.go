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

	// Precommit prepares to commit the changes, returning a hash of the
	// updates or current state.
	Precommit(context.Context) ([]byte, error)

	// Commit signals that the session is complete.
	// It returns a unique identifier for the session that
	// is generated deterministically from the applied changes.
	Commit(ctx context.Context, idempotencyKey []byte) error

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
	sessionKey   []byte                 // whether the committables are in a session
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

	if len(m.sessionKey) > 0 {
		return ErrInSession
	}
	m.sessionKey = idempotencyKey

	lastKey, err := getCurrentKey(m.kv)
	if err != nil {
		return err
	}

	// if the last key is the same as the current key, we are recovering
	// from a crash, therefore begin recovery mode.
	if bytes.Equal(lastKey, idempotencyKey) {
		// At this point we don't know if any of the committables had gotten to
		// the Commit loop previously. Since the previous recover/Skip
		// mechanisms did not properly support DB reads at the prior state, we
		// can only warn and see if the individual committables fail to Begin
		// based on their own recorded idempotencyKeys.
		//
		// Maybe we remove MultiCommitter and replace it with another mechanisms
		// just for aggregating app hashes from each committer in a
		// deterministic order.
		m.log.Warn("trying to apply same transaction again, individual committables may error")
		// return errors.New("trying to apply same transaction again, no recovery possible")
	} else {
		err = setCurrentKey(m.kv, idempotencyKey)
		if err != nil {
			return err
		}
	}

	for _, committable := range m.committables {
		err = committable.Begin(ctx, idempotencyKey)
		if err != nil {
			return err
		}
	}

	return nil
}

// Precommit prepares to commit the session. No more updates may be performed
// after this. This should precede Commit or Cancel. The unique identifier is
// deterministically generated from the commit IDs of the individual committers.
func (m *MultiCommitter) Precommit(ctx context.Context) (id []byte, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	defer m.handleErr(ctx, &err)

	if len(m.sessionKey) == 0 {
		return nil, ErrNotInSession
	}

	orderedMap := order.OrderMap(m.committables)
	for _, c := range orderedMap {
		newID, err := c.Value.Precommit(ctx)
		if err != nil {
			return nil, err
		}

		id = append(id, newID...)
	}

	return
}

// Commit commits the session for the committables.
func (m *MultiCommitter) Commit(ctx context.Context) (err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	defer m.handleErr(ctx, &err)

	if len(m.sessionKey) == 0 {
		return ErrNotInSession
	}

	for _, c := range m.committables {
		if err := c.Commit(ctx, m.sessionKey); err != nil {
			return err
		}
	}

	err = setCurrentKey(m.kv, m.sessionKey)
	if err != nil {
		return err
	}
	m.sessionKey = nil

	return err
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
func (m *MultiCommitter) handleErr(ctx context.Context, err *error) {
	if *err != nil {
		m.log.Error("error during atomic commit", zap.Error(*err))

		for _, committable := range m.committables {
			err := committable.Cancel(ctx)
			if err != nil {
				m.log.Error("error cancelling committable", zap.Error(err))
			}
		}
		m.sessionKey = nil
	}

}
