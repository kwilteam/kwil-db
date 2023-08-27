/*
Package sessions provides a session abstraction for atomic commits across various datastores.

It implements a basic two-phase commit protocol, where the first phase is to commit
idempotent changes to a WAL, and the second phase is to apply those changes to the
datastores.

The session writes changes from any committable to a WAL.  All changes are appended with the
unique identifier for the committable.  When the session is committed, the changes are read
from the WAL and applied to the datastores.
*/
package sessions

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/kwilteam/kwil-db/pkg/log"
	"github.com/kwilteam/kwil-db/pkg/utils/order"
	"go.uber.org/zap"
)

// AtomicCommitter handles atomic commits across various datastores.
// It uses a WAL to store idempotent changes, and then applies them to the datastores.
type AtomicCommitter struct {
	// committables are the objects that need to be committed atomically.
	// They are indexed by their unique identifier.
	committables map[CommittableId]Committable

	// wal is the write-ahead log for the session.
	wal *sessionWal

	// log is self-explanatory.
	log log.Logger

	// Mutex to ensure thread safety when checking session state
	mu sync.Mutex

	// phase is the current phase of the session.
	phase CommitPhase

	// closed is set to true when the committer is closed.
	// this is to protect against AtomicCommitter consumers
	// from calling methods on a closed committer, which could
	// lead to an inconsistent state.
	closed bool

	// timeout is the amount of time to wait on shutdown
	timeout time.Duration
}

// CommittableId is the unique identifier for a committable.
type CommittableId string

func (id CommittableId) String() string {
	return string(id)
}

func (id CommittableId) Bytes() []byte {
	return []byte(id)
}

// NewAtomicCommitter creates a new atomic session.
func NewAtomicCommitter(ctx context.Context, wal Wal, opts ...CommiterOpt) *AtomicCommitter {
	a := &AtomicCommitter{
		committables: make(map[CommittableId]Committable),
		wal:          &sessionWal{wal},
		log:          log.NewNoOp(),
		timeout:      5 * time.Second,
		phase:        CommitPhaseNone,
	}

	for _, opt := range opts {
		opt(a)
	}

	return a
}

// ClearWal clears the wal.  This method will check if there are values in the WAL.
// If so, it will attempt to commit them.  If no commit record is found, then it will
// discard the WAL.
// This should be called on startup or recovery.
func (a *AtomicCommitter) ClearWal(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.closed {
		return ErrClosed
	}

	if a.phase.InSession() {
		return ErrSessionInProgress
	}

	return a.applyWal(ctx)
}

// Begin begins a new session.
// It will signal to all committables that a session has begun.
func (a *AtomicCommitter) Begin(ctx context.Context) (err error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.closed {
		return ErrClosed
	}

	if a.phase.InSession() {
		return ErrSessionInProgress
	}
	a.phase = CommitPhaseCommit

	// we handle errors after checking to see if a session
	// is in progress because, if begin is called while a session
	// is in progress, it will cancel the session.
	defer a.handleErr(ctx, &err)

	return a.beginCommit(ctx)
}

// Commit commits the atomic session.
// It can be given a callback function to handle any errors that occur during the apply phase (which proceeds asynchronously) after this function returns.
func (a *AtomicCommitter) Commit(ctx context.Context, applyCallback func(error)) (err error) {
	a.mu.Lock()

	// if no session in progress, then return without cancelling
	if a.phase != CommitPhaseCommit {
		a.mu.Unlock()
		return phaseError("Commit", CommitPhaseCommit, a.phase)
	}
	a.phase = CommitPhaseApply

	// if error, cancel the session
	defer a.handleErr(ctx, &err)

	// if session is in progress but the committer is closed, then cancel the session
	if a.closed {
		a.mu.Unlock()
		return ErrClosed
	}

	// begin the commit in the WAL
	err = a.wal.WriteBegin(ctx)
	if err != nil {
		a.mu.Unlock()
		return err
	}

	// tell all committables to finish phase 1, and submit
	// any changes to the WAL
	err = a.endCommit(ctx)
	if err != nil {
		a.mu.Unlock()
		return err
	}

	// commit the WAL
	err = a.wal.WriteCommit(ctx)
	if err != nil {
		a.mu.Unlock()
		return err
	}

	go func() {
		err2 := a.apply(ctx)
		applyCallback(err2)
		a.phase = CommitPhaseNone
		a.mu.Unlock()
	}()

	return nil
}

// ID returns a deterministic identifier representative of all state changes that have occurred in the session.
// It can only be called in between Begin and Commit.
func (a *AtomicCommitter) ID(ctx context.Context) (id []byte, err error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.phase.InSession() {
		return nil, ErrNoSessionInProgress
	}

	// this comes after checking if a session is in progress because
	// we do not want to call cancel if a session is not in progress.
	defer a.handleErr(ctx, &err)

	if a.closed {
		return nil, ErrClosed
	}

	return a.id(ctx)
}

// Apply applies the atomic session.
// It will notify all committables that changes are about to be applied, and then apply the changes.
// It will then notify all committables that it has read all changes.
func (a *AtomicCommitter) apply(ctx context.Context) (err error) {
	defer a.handleErr(ctx, &err)

	beginRecord, err := a.wal.ReadNext(ctx)
	if err != nil {
		return err
	}

	if beginRecord.Type != WalRecordTypeBegin {
		a.log.Error("missing begin record in WAL")
		return ErrMissingBegin
	}

	err = a.beginApply(ctx)
	if err != nil {
		return err
	}

	applyErrs := []error{}
	for {
		record, err := a.wal.ReadNext(ctx)
		if err != nil {
			return err
		}
		if record.Type == WalRecordTypeCommit {
			break
		}

		committable := a.committables[record.CommittableId]
		err = committable.Apply(ctx, record.Data)
		if err != nil {
			applyErrs = append(applyErrs, err)
		}
	}
	err = errors.Join(applyErrs...)
	if err != nil {
		return wrapError(ErrApply, err)
	}

	err = a.endApply(ctx)
	if err != nil {
		return err
	}

	return a.wal.Truncate(ctx)
}

func (a *AtomicCommitter) cancel(ctx context.Context) {
	for _, committable := range a.committables {
		committable.Cancel(ctx)
	}
}

// handleErr checks if an error is nil or not.
// If it is not nil, it logs it, and notifies the committables that the session has been cancelled.
// It then sets the session state to not in progress.
func (a *AtomicCommitter) handleErr(ctx context.Context, err *error) {
	if *err != nil {
		a.log.Error("error during atomic commit", zap.Error(*err))
		a.cancel(ctx)
		a.phase = CommitPhaseNone
	}
}

// applyWal will try to apply all changes in the WAL to the committables.
// If the wal does not contain a commit record, it will delete all changes in the WAL.
// If the wal contains a commit record, it will apply all changes in the WAL to the committables.
// If the wal contains a commit record, but the commit fails, it will return an error.
func (a *AtomicCommitter) applyWal(ctx context.Context) (err error) {
	beginRecord, err := a.wal.ReadNext(ctx)
	if err == io.EOF {
		return a.wal.Truncate(ctx)
	}
	if err != nil {
		return err
	}

	if beginRecord.Type != WalRecordTypeBegin {
		a.log.Error("missing begin record in WAL")
		return a.wal.Truncate(ctx)
	}

	// if we reach here, then it means the wal contains data and we are starting apply phase
	defer a.handleErr(ctx, &err)

	err = a.beginApply(ctx)
	if err != nil {
		return err
	}

	applyErrs := []error{}
	for {
		record, err := a.wal.ReadNext(ctx)
		if err == io.EOF {
			// if we have reached io.EOF, we want to truncate the wal and cancel the session
			// we do not want to return an error, but we do want to tell all committables to cancel
			a.log.Error("missing commit record in WAL, truncating")
			truncErr := a.wal.Truncate(ctx)
			// if there is a truncate error, we want to return that
			// this will trigger the deferred handleErr function, which calls cancel
			if truncErr != nil {
				return truncErr
			}
			a.cancel(ctx)
			return nil
		}
		if err != nil {
			return err
		}
		if record.Type == WalRecordTypeCommit {
			break
		}

		committable, ok := a.committables[record.CommittableId]
		if !ok {
			a.log.Error("cannot find target data store for wal record", zap.String("committable_id", record.CommittableId.String()))
		}
		err = committable.Apply(ctx, record.Data)
		if err != nil {
			applyErrs = append(applyErrs, err)
		}
	}
	err = errors.Join(applyErrs...)
	if err != nil {
		return wrapError(ErrApply, err)
	}

	err = a.endApply(ctx)
	if err != nil {
		return err
	}

	return a.wal.Truncate(ctx)
}

// beginCommit calls BeginCommit on all committables.
func (a *AtomicCommitter) beginCommit(ctx context.Context) error {
	return a.callAll(ErrBeginCommit, func(c Committable) error {
		return c.BeginCommit(ctx)
	})
}

// endCommit calls EndCommit on all committables.
// it orders the committables alphabetically by their unique identifier, to ensure that the commit id is deterministic.
func (a *AtomicCommitter) endCommit(ctx context.Context) error {
	for id, c := range a.committables {
		err := c.EndCommit(ctx, func(b []byte) error {
			return a.wal.WriteChangeset(ctx, id, b)
		})
		if err != nil {
			return wrapError(ErrEndCommit, err)
		}
	}

	return nil
}

// beginApply calls BeginApply on all committables.
func (a *AtomicCommitter) beginApply(ctx context.Context) error {
	return a.callAll(ErrBeginApply, func(c Committable) error {
		return c.BeginApply(ctx)
	})
}

// endApply calls EndApply on all committables.
func (a *AtomicCommitter) endApply(ctx context.Context) error {
	return a.callAll(ErrEndApply, func(c Committable) error {
		return c.EndApply(ctx)
	})
}

// id calls ID on all committables.
// it orders the committables alphabetically by their unique identifier, to ensure that the commit id is deterministic.
func (a *AtomicCommitter) id(ctx context.Context) (id []byte, err error) {
	hash := sha256.New()

	for _, c := range order.OrderMapLexicographically[CommittableId, Committable](a.committables) {
		commitId, err := c.Value.ID(ctx)
		if err != nil {
			return nil, wrapError(ErrID, err)
		}

		_, err = hash.Write(commitId)
		if err != nil {
			return nil, wrapError(ErrID, err)
		}
	}

	return hash.Sum(nil), nil
}

func (a *AtomicCommitter) callAll(errType error, f func(Committable) error) error {
	errs := []error{}
	for _, committable := range a.committables {
		err := f(committable)
		if err != nil {
			errs = append(errs, err)
		}
	}

	err := errors.Join(errs...)
	if err != nil {
		return wrapError(errType, err)
	}

	return nil
}

// TODO: we need to test register and unregister being used in different phases

// Register registers a committable with the atomic committer.
// If a session is already in progress, the newly registered committer will be added to the session,
// and BeginCommit will immediately be called on the committable.
// If BeginCommit fails, the entire session will be cancelled.
func (a *AtomicCommitter) Register(ctx context.Context, id string, committable Committable) (err error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	_, ok := a.committables[CommittableId(id)]
	if ok {
		return ErrAlreadyRegistered
	}
	a.committables[CommittableId(id)] = committable

	// we need to call commit if in phase 1, and apply if in phase 2
	// right now we only call apply
	defer a.handleErr(ctx, &err)
	switch a.phase {
	default:
		panic(fmt.Sprintf("unknown phase: %d", a.phase))
	case CommitPhaseNone:
		return nil
	case CommitPhaseCommit:
		err = committable.BeginCommit(ctx)
		if err != nil {
			return wrapError(ErrBeginCommit, err)
		}
	case CommitPhaseApply:
		err = committable.BeginApply(ctx)
		if err != nil {
			return wrapError(ErrBeginApply, err)
		}
	}

	return nil
}

// Unregister unregisters a committable from the atomic committer.
// If a session is already in progress, Cancel will immediately be called on the committable.
func (a *AtomicCommitter) Unregister(ctx context.Context, id string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	committable, ok := a.committables[CommittableId(id)]
	if !ok {
		return wrapError(ErrUnknownCommittable, fmt.Errorf("committable id: %s", id))
	}
	delete(a.committables, CommittableId(id))

	if a.phase.InSession() {
		committable.Cancel(ctx)
	}

	return nil
}

// Close closes the atomic committer.
// If a session is in progress, it will cancel the session.
// If a session is not in progress, it will close immediately.
// We only have to worry about this being called outside of a session,
// or in the middle of phase 1.  Since the committer mutex is locked
// from the end of phase1 to the end of phase2, we know that this
// function will not be called in the middle of phase 2.
func (a *AtomicCommitter) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.log.Info("closing atomic committer")

	if !a.phase.InSession() {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), a.timeout)
	defer cancel()

	a.cancel(ctx)
	return nil
}

type CommitPhase uint8

const (
	// Signals that no session is in progress.
	CommitPhaseNone CommitPhase = iota
	// Signals that the session is in between Begin and Commit.
	CommitPhaseCommit
	// Signals that the session is in between Commit and BeginApply.
	CommitPhaseApply
)

// InSession returns true if the committer is in a session.
func (c CommitPhase) InSession() bool {
	return c != CommitPhaseNone
}

func (c CommitPhase) String() string {
	switch c {
	case CommitPhaseNone:
		return "none"
	case CommitPhaseCommit:
		return "commit"
	case CommitPhaseApply:
		return "apply"
	default:
		return fmt.Sprintf("unknown phase: %d", c)
	}
}

// phaseError returns an error indicating that a method was called in the wrong phase.
func phaseError(method string, desiredPhase CommitPhase, actual CommitPhase) error {
	return fmt.Errorf("%w, method '%s' can only be called in '%s' phase, but was called in '%s'", ErrCommitPhase, method, desiredPhase, actual)
}
