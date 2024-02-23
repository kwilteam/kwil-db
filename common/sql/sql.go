package sql

import (
	"context"
	"errors"
)

var (
	ErrNoTransaction = errors.New("no transaction")
	ErrNoRows        = errors.New("no rows in result set")
)

// DB is a connection to a Postgres database.
// It has root user access, and can execute any Postgres command.
type DB interface {
	Executor
	// BeginTx starts a new transaction.
	BeginTx(ctx context.Context) (Tx, error)
	// AccessMode gets the access mode of the database.
	// It can be either read-write or read-only.
	AccessMode() AccessMode
}

// Executor is an interface that can execute queries.
type Executor interface {
	// Execute executes a query or command.
	// The stmt should be a valid Postgres statement.
	Execute(ctx context.Context, stmt string, args ...any) (*ResultSet, error)
}

// Tx is a database transaction. It can be nested within other transactions.
type Tx interface {
	DB
	// Rollback rolls back the transaction.
	Rollback(ctx context.Context) error
	// Commit commits the transaction.
	Commit(ctx context.Context) error
}

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

// TxCloser terminates a transaction by committing or rolling it back. A method
// that returns this alone would keep the tx under the hood of the parent type,
// directing queries internally through the scope of a transaction/session
// started with BeginTx.
type TxCloser interface {
	Rollback(ctx context.Context) error
	Commit(ctx context.Context) error
}

// TxPrecommitter is the special kind of transaction that can prepare a
// transaction for commit.
// It is only available on the outermost transaction.
type TxPrecommitter interface {
	Precommit(ctx context.Context) ([]byte, error)
}

type TxBeginner interface {
	Begin(ctx context.Context) (TxCloser, error)
}

// OuterTxMaker is the special kind of transaction beginner that can make nested
// transactions, and that explicitly scopes Query/Execute to the tx.
type OuterTxMaker interface {
	BeginTx(ctx context.Context) (OuterTx, error)
}

// ReadTxMaker can make read-only transactions.
// Many read-only transactions can be made at once.
type ReadTxMaker interface {
	BeginReadTx(ctx context.Context) (Tx, error)
}

// TxMaker is the special kind of transaction beginner that can make nested
// transactions, and that explicitly scopes Query/Execute to the tx.
type TxMaker interface {
	BeginTx(ctx context.Context) (Tx, error)
}

// OuterTx is a database transaction. It is the outermost transaction type.
// "nested transactions" are called savepoints, and can be created with
// BeginSavepoint. Savepoints can be nested, and are rolled back to the
// innermost savepoint on Rollback.
//
// Anything using implicit tx/session management should use TxCloser.
type OuterTx interface {
	Tx
	TxPrecommitter
}
