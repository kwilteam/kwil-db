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
	"sort"
	"sync"

	"github.com/kwilteam/kwil-db/pkg/log"
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

	// State to indicate whether the session is in progress or not
	inProgress bool
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
func NewAtomicCommitter(ctx context.Context, committables map[string]Committable, wal Wal, opts ...CommiterOpt) (*AtomicCommitter, error) {
	committablesMap := make(map[CommittableId]Committable)
	for id, committable := range committables {
		committablesMap[CommittableId(id)] = committable
	}

	a := &AtomicCommitter{
		committables: committablesMap,
		wal:          &sessionWal{wal},
		log:          log.NewNoOp(),
	}

	for _, opt := range opts {
		opt(a)
	}

	err := a.flushWal(ctx)
	if err != nil {
		return nil, err
	}

	return a, nil
}

func (a *AtomicCommitter) Begin(ctx context.Context) (err error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	defer a.handleErr(ctx, &err)

	if a.inProgress {
		return ErrSessionInProgress
	}
	a.inProgress = true

	return a.beginCommit(ctx)
}

// Commit commits the atomic session.
// It aggregates all commit ids from the committables and returns them as a single Sha256 hash.
// It can be given a callback function to handle any errors that occur during the apply phase (which procedes asynchronously) after this function returns.
func (a *AtomicCommitter) Commit(ctx context.Context, applyCallback func(error)) (commitId []byte, err error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	defer a.handleErr(ctx, &err)

	if !a.inProgress {
		return nil, ErrNoSessionInProgress
	}

	err = a.wal.Begin(ctx)
	if err != nil {
		return nil, err
	}

	commitId, err = a.endCommit(ctx)
	if err != nil {
		return nil, err
	}

	err = a.wal.Commit(ctx)
	if err != nil {
		return nil, err
	}

	go func() {
		err2 := a.apply(ctx)
		applyCallback(err2)
	}()

	return commitId, nil
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

// handleErr checks if an error is nil or not.  If it is not nil, it logs it, and notifies the committables that the session has been cancelled.
func (a *AtomicCommitter) handleErr(ctx context.Context, err *error) {
	if *err != nil {
		a.log.Error("error during atomic commit", zap.Error(*err))
		a.cancel(ctx)
	}
}

// flushWal will try to apply all changes in the WAL to the committables.
// If the wal does not contain a commit record, it will delete all changes in the WAL.
// If the wal contains a commit record, it will apply all changes in the WAL to the committables.
// If the wal contains a commit record, but the commit fails, it will return an error.
func (a *AtomicCommitter) flushWal(ctx context.Context) (err error) {
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

// beginCommit calls BeginCommit on all committables.
func (a *AtomicCommitter) beginCommit(ctx context.Context) error {
	return a.callAll(ErrBeginCommit, func(c Committable) error {
		return c.BeginCommit(ctx)
	})
}

// endCommit calls EndCommit on all committables.
// it orders the committables alphabetically by their unique identifier, to ensure that the commit id is deterministic.
func (a *AtomicCommitter) endCommit(ctx context.Context) ([]byte, error) {
	orderedCommittables := orderAlphabetically(a.committables)

	hash := sha256.New()
	for _, c := range orderedCommittables {
		commitId, err := c.committable.EndCommit(ctx, func(b []byte) error {
			return a.wal.WriteChangeset(ctx, c.id, b)
		})
		if err != nil {
			return nil, wrapError(ErrEndCommit, err)
		}

		fmt.Println("commitId", commitId)

		_, err = hash.Write(commitId)
		if err != nil {
			return nil, wrapError(ErrEndCommit, err)
		}
	}

	return hash.Sum(nil), nil
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

// orderAlphabetically orders the committables alphabetically by their unique identifier.
func orderAlphabetically(commitableMap map[CommittableId]Committable) []*struct {
	id          CommittableId
	committable Committable
} {
	// Extracting keys
	keys := make([]string, 0, len(commitableMap))
	for k := range commitableMap {
		keys = append(keys, k.String())
	}

	// Sorting keys
	sort.Strings(keys)

	// Creating output slice
	datasets := make([]*struct {
		id          CommittableId
		committable Committable
	}, 0)

	for _, k := range keys {
		datasets = append(datasets, &struct {
			id          CommittableId
			committable Committable
		}{
			id:          CommittableId(k),
			committable: commitableMap[CommittableId(k)],
		})
	}
	return datasets
}
