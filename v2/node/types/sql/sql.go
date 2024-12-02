// Package sql defines common type required by SQL database implementations and
// consumers.
package sql

import (
	"context"
	"errors"
)

var (
	ErrNoTransaction = errors.New("no transaction")
	ErrNoRows        = errors.New("no rows in result set")
)

// ResultSet is the result of a query or execution.
// It contains the returned columns and the rows.
type ResultSet struct {
	Columns []string
	Rows    [][]any
	Status  CommandTag
}

// CommandTag is the result of a command execution.
type CommandTag struct {
	// Text is the text of the command tag.
	Text string
	// RowsAffected is the number of rows affected by the command.
	RowsAffected int64
}

// Executor is an interface that can execute queries.
type Executor interface {
	// Execute executes a query or command.
	Execute(ctx context.Context, stmt string, args ...any) (*ResultSet, error)
}

// QueryScanner represents a type that provides the ability to execute an SQL
// statement, where for each row:
//
//  1. result values are scanned into the variables in the scans slice
//  2. the provided function is then called
//
// The function would typically capture the variables in the scans slice,
// allowing it to operator on the values. For instance, append the values to
// slices allocated by the caller, or perform reduction operations like
// sum/mean/min/etc.
//
// NOTE: This method may end up being included in the Tx interface alongside
// Executor since all of the concrete transaction implementations provided by
// this package implement this method.
type QueryScanner interface {
	QueryScanFn(ctx context.Context, stmt string,
		scans []any, fn func() error, args ...any) error
}

// TxMaker is an interface that creates a new transaction. In the context of the
// recursive Tx interface, is creates a nested transaction.
type TxMaker interface {
	BeginTx(ctx context.Context) (Tx, error)
}

// Tx represents a database transaction. It can be nested within other
// transactions, and create new nested transactions. An implementation of Tx may
// also be an AccessModer, but it is not required.
type Tx interface {
	Executor
	TxMaker // recursive interface
	// note: does not embed DB for clear semantics (DB makes a Tx, not the reverse)

	// Rollback rolls back the transaction.
	Rollback(ctx context.Context) error
	// Commit commits the transaction.
	Commit(ctx context.Context) error
}

// DB is a top level database interface, which may directly execute queries or
// create transactions, which may be closed or create additional nested
// transactions.
//
// Some implementations may also be an PreparedTxMaker and/or a ReadTxMaker. Embed
// with those interfaces to compose the minimal interface required.
type DB interface {
	Executor
	TxMaker
}

// ReadTxMaker can make read-only transactions. This is necessarily an outermost
// transaction since nested transactions inherit their access mode from their
// parent. Many read-only transactions can be made at once.
type ReadTxMaker interface {
	BeginReadTx(ctx context.Context) (OuterReadTx, error)
}

// DelayedReadTxMaker is an interface that creates a transaction for reading
// from the database. The transaction won't actually be created until it is used
// for the first time, which is useful for avoiding unnecessary transactions.
type DelayedReadTxMaker interface {
	BeginDelayedReadTx() OuterReadTx
}

// PreparedTx is an outermost database transaction that uses two-phase commit
// with the Precommit method.
//
// NOTE: A PreparedTx may be used where only a Tx or DB is required since those
// interfaces are a subset of the PreparedTx method set.
// It takes a writer to write the full changeset to.
// If the writer is nil, the changeset will not be written.
type PreparedTx interface {
	Tx
	Subscriber
	Precommit(ctx context.Context, changes chan<- any) ([]byte, error)
}

// OuterReadTx is the outermost read-only database transaction.
type OuterReadTx interface {
	Tx
	Subscriber
}

// PreparedTxMaker is the special kind of transaction that creates a transaction
// that has a Precommit method (see OuterTx), which supports obtaining a commit
// ID using a (two-phase) prepared transaction prior to Commit. This is a
// different method name so that an implementation may satisfy both PreparedTxMaker
// and TxMaker.
type PreparedTxMaker interface {
	BeginPreparedTx(ctx context.Context) (PreparedTx, error)
}

// SnapshotTxMaker is an interface that creates a transaction for taking a
// snapshot of the database. This uses serializable isolation level to ensure
// internal consistency.
type SnapshotTxMaker interface {
	BeginSnapshotTx(ctx context.Context) (Tx, string, error)
}

// AccessMode is the type of access to a database.
// It can be read-write or read-only.
type AccessMode uint8

const (
	// ReadWrite is the default access mode.
	// It allows for reading and writing to the database.
	ReadWrite AccessMode = iota
	// ReadOnly allows for reading from the database, but not writing.
	ReadOnly
)

// AccessModer may be satisfied by implementations of Tx and DB, but is not
// universally required for those interfaces (type assert as needed).
type AccessModer interface {
	// AccessMode gets the access mode of the database or transaction.
	AccessMode() AccessMode
}

// Subscriber is a transaction that can be subscribed to.
// When subscribed, the passed channel will receive notifications.
// Only one subscription is allowed per transaction.
type Subscriber interface {
	// Subscribe subscribes to notifications passed using the special `notice()`
	// function. Only `notice()` calls made after the subscription, on this tx,
	// will be received. A done function is returned that should be called to
	// signal that no more statements that emit notices will be executed. The
	// channel will only be closed after all notices have been sent AND the done
	// is called, or if the database shuts down prematurely. In case of an
	// unexpected DB shutdown, an empty string is sent on the channel before it
	// is closed. It is the caller's responsibility to call done.
	Subscribe(ctx context.Context) (ch <-chan string, done func(context.Context) error, err error)
}
