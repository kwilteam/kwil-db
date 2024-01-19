package sql

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

var (
	ErrNoTransaction = errors.New("no transaction")
	ErrNoRows        = errors.New("no rows in result set")
)

func CleanStmt(stmt string) string {
	trimmed := strings.TrimSpace(stmt)
	// Ensure it ends with a semicolon.
	if !strings.HasSuffix(trimmed, ";") {
		return trimmed + ";"
	}
	return trimmed
}

type Queryer interface {
	Query(ctx context.Context, stmt string, args ...any) (*ResultSet, error)
}

type PendingQueryer interface {
	QueryPending(ctx context.Context, query string, args ...any) (*ResultSet, error)
}

type Executor interface {
	Execute(ctx context.Context, stmt string, args ...any) error
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

type TxBeginner interface {
	Begin(ctx context.Context) (TxCloser, error)
}

// type Tx = TxCloser // refactoring

// TxMaker is the special kind of transaction beginner that can make nested
// transactions, and that explicitly scopes Query/Execute to the tx.
type TxMaker interface {
	BeginTx(ctx context.Context) (Tx, error)
}

// Tx can do it all, including start nested transactions.
//
// TODO: after refactoring this should become Tx and anything using implicit
// tx/session management should use TxCloser.
type Tx interface {
	Queryer
	Executor
	TxCloser // just Commit and Rollback
	TxMaker  // for nested transactions to isolate failures
}

// KVStore is a key-value store.
// type KVStore interface {
// 	// Set sets a key to a value.
// 	Set(ctx context.Context, key []byte, value []byte) error
// 	// Get gets a value for a key.
// 	Get(ctx context.Context, key []byte) ([]byte, error)
// 	// Delete deletes a key.
// 	// Delete(ctx context.Context, key []byte) error
// }

type ExecResult interface {
	RowsAffected() (int64, error)
}

// Result is the result of a query. TODO: this is detail, not needed outside of sqlite impl
type Result interface {
	Close() error
	// Columns gets the columns of the result.
	Columns() []string
	// Finish finishes any execution that is in progress and closes the result.
	Finish() error

	// Next gets the next row of the result.
	Next() (rowReturned bool)

	// Err gets any error that occurred during iteration.
	Err() error

	// Values gets the values of the current row.
	Values() ([]any, error)

	// A ResultSet() (*ResultSet, error) method is a red flag for me. Something
	// should not be returning a sql.Result (or the caller should assert to a
	// *sql.ResultSet). If an interface has a method to return the concrete type
	// that's actually implementing the interface, then the interface is
	// pointless.
}

type ResultSet struct {
	ReturnedColumns []string
	Rows            [][]any

	i int // starts at 0
}

var _ Result = (*ResultSet)(nil)

func (r *ResultSet) Columns() []string {
	v := make([]string, len(r.ReturnedColumns))
	copy(v, r.ReturnedColumns)

	return v
}

func (r *ResultSet) Next() (rowReturned bool) {
	if r.i >= len(r.Rows) {
		return false
	}

	r.i++
	return true
}

func (r *ResultSet) Values() ([]any, error) {
	if r.i > len(r.Rows) {
		return nil, fmt.Errorf("result set has no more rows")
	}

	v := make([]any, len(r.Rows[r.i-1]))
	copy(v, r.Rows[r.i-1])

	return v, nil
}

func (r *ResultSet) Map() []map[string]any {
	m := make([]map[string]any, len(r.Rows))
	for i, row := range r.Rows {
		m2 := make(map[string]any)
		for j, col := range row {
			m2[r.ReturnedColumns[j]] = col
		}

		m[i] = m2
	}

	return m
}

// implements Result
func (r *ResultSet) Close() error {
	return nil
}

// implements Result
func (r *ResultSet) Err() error {
	return nil
}

// implements Result
func (r *ResultSet) Finish() error {
	return nil
}

type ConnectionFlag int

const (
	// OpenNone indicates that the connection should be read-write and not created if it does not exist.
	OpenNone ConnectionFlag = 1 << iota
	// OpenReadOnly indicates that the connection should be read-only.
	OpenReadOnly
	// OpenCreate indicates that the connection should be created if it does not exist.
	OpenCreate
	// OpenMemory indicates that the connection should be in-memory.
	OpenMemory
)

// EmptyResult is a result that has no rows.
type EmptyResult struct{}

var _ Result = (*EmptyResult)(nil)

func (e *EmptyResult) Close() error {
	return nil
}

func (e *EmptyResult) Columns() []string {
	return nil
}

func (e *EmptyResult) Finish() error {
	return nil
}

func (e *EmptyResult) Next() (rowReturned bool) {
	return false
}

func (e *EmptyResult) Err() error {
	return nil
}

func (e *EmptyResult) Values() ([]any, error) {
	return nil, nil
}
