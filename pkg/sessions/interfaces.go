package sessions

import (
	"context"
)

// Committable is an interface for objects that need to be able to commit changes atomically.
// It features two phases: a commit phase, and an apply phase.
// The commit phase is used to write changes to a WAL.
// A committer should write all changes to the WAL in the EndCommit function.
// In the apply phase, the session will read the changes from the WAL and apply them to the
// committable.  It can be called multiple times.
type Committable interface {
	// BeginCommit is called to begin a session.
	BeginCommit(ctx context.Context) error

	// EndCommit is used to ask the committable for any changes it wishes to commit.
	// It provides a function with which to append changes to the session.
	// It returns a commit ID, which is a determinstic hash of the changes.
	// The commit id is used to track the state of the committable.
	// EndCommit can only be called once per session, per committable.
	// When EndCommit is called, the committer should write ALL changes to AppendFunc.
	EndCommit(ctx context.Context, appender func([]byte) error) (err error)

	// BeginApply is used to signal to the committable that changes are about to be applied
	// to its datastore(s).
	// It is called before changes are applied.
	BeginApply(ctx context.Context) error

	// Apply is used to apply a change to the committable.
	// It is called after changes have been committed.
	Apply(ctx context.Context, changes []byte) error

	// EndApply is used to signal to the committable that changes have been applied to its
	// datastore(s).
	// It is called after changes have been applied.
	EndApply(ctx context.Context) error

	// Cancel is used to cancel a session.
	Cancel(ctx context.Context)

	// ID returns a unique ID representative of the state changes that have occurred so far for this committable.
	// It should be deterministic, and should change if and only if the committable has changed.
	ID(ctx context.Context) ([]byte, error)
}

// Wal is an interface for a write-ahead log.
type Wal interface {
	// Append appends a new entry to the WAL
	Append(ctx context.Context, data []byte) error
	// ReadNext reads the next entry from the WAL
	// the ReadNext method will return an io.EOF when it has reached the end of the WAL,
	// or if the WAL is empty, or corrupt.
	ReadNext(ctx context.Context) ([]byte, error)

	// Truncate truncates the WAL, deleting all entries (if any exist)
	// If none exist, it should return nil
	Truncate(ctx context.Context) error
}
