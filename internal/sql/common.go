package sql

import (
	"context"
	"errors"
)

var (
	ErrNoTransaction = errors.New("no transaction")
	ErrNoRows        = errors.New("no rows in result set")
)

type Queryer interface {
	Query(ctx context.Context, stmt string, args ...any) (*ResultSet, error)
}

type Executor interface {
	Execute(ctx context.Context, stmt string, args ...any) (*ResultSet, error)
}

type KVGetter interface {
	// Get gets a value for a key.
	Get(ctx context.Context, key []byte, sync bool) ([]byte, error)
}

type KV interface {
	KVGetter

	// Set sets a key to a value.
	Set(ctx context.Context, key []byte, value []byte) error
}

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

// Tx is like a transaction, but it can be nested.
type Tx interface {
	TxCloser
	DB
}

// DB is an interface that can execute queries and make
// nested transactions (savepoints).
type DB interface {
	Executor
	TxMaker
	// AccessMode gets the access mode of the database.
	// It can be either read-write or read-only.
	AccessMode() AccessMode
}

type ExecResult interface {
	RowsAffected() (int64, error)
}

// ResultSet is the result of a query or execution.
// It contains the returned columns and the rows.
type ResultSet struct {
	Columns []string
	Rows    [][]any

	Status CommandTag
}

// Map returns the result set as a slice of maps.
// For example, if the result set has two columns, "name" and "age",
// and two rows, "alice" and 30, and "bob" and 40, the result would be:
//
//	[]map[string]any{
//		{"name": "alice", "age": 30},
//		{"name": "bob", "age": 40},
//	}
//
// It is not recommended to use this method for large result sets.
func (r *ResultSet) Map() []map[string]any {
	m := make([]map[string]any, len(r.Rows))
	for i, row := range r.Rows {
		m2 := make(map[string]any)
		for j, col := range row {
			m2[r.Columns[j]] = col
		}

		m[i] = m2
	}

	return m
}

type CommandTag struct {
	Text         string
	RowsAffected int64
}

func (ct *CommandTag) String() string {
	return ct.Text // tip: prefix will be select, insert, etc.
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
