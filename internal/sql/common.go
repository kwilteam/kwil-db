package sql

import (
	"context"
	"fmt"
)

// KVStore is a key-value store.
type KVStore interface {
	// Set sets a key to a value.
	Set(ctx context.Context, key []byte, value []byte) error
	// Get gets a value for a key.
	Get(ctx context.Context, key []byte) ([]byte, error)
	// Delete deletes a key.
	Delete(ctx context.Context, key []byte) error
}

type ResultSetFunc func(ctx context.Context, stmt string, args map[string]any) (*ResultSet, error)

// Connection is a connection to a database.
type Connection interface {
	KVStore
	Execute(ctx context.Context, stmt string, args map[string]any) (Result, error)
	Close() error
	CreateSession() (Session, error)
	Savepoint() (Savepoint, error)
}

// ReturnableConnection is a connection that can be returned to a pool.
type ReturnableConnection interface {
	Connection
	Return()
}

type Savepoint interface {
	Rollback() error
	Commit() error
}

type Session interface {
	Delete() error
	ChangesetID(ctx context.Context) ([]byte, error)
}

type Changeset interface {
	// ID generates a deterministic ID for the changeset.
	ID() ([]byte, error)
}

// Result is the result of a query.
type Result interface {
	Close() error
	// Columns gets the columns of the result.
	Columns() []string
	// Finish finishes any execution that is in progress and closes the result.
	Finish() error

	// Next gets the next row of the result.
	Next() (rowReturned bool, err error)

	// Values gets the values of the current row.
	Values() ([]any, error)
}

type ResultSet struct {
	ReturnedColumns []string
	Rows            [][]any

	i int
}

func (r *ResultSet) Columns() []string {
	v := make([]string, len(r.ReturnedColumns))
	copy(v, r.ReturnedColumns)

	return v
}

func (r *ResultSet) Next() (rowReturned bool, err error) {
	if r.i >= len(r.Rows) {
		return false, nil
	}

	r.i++
	return true, nil
}

func (r *ResultSet) Values() ([]any, error) {
	if r.i >= len(r.Rows) {
		return nil, fmt.Errorf("result set has no more rows")
	}

	v := make([]any, len(r.Rows[r.i]))
	copy(v, r.Rows[r.i])

	return v, nil
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

func (e *EmptyResult) Next() (rowReturned bool, err error) {
	return false, nil
}

func (e *EmptyResult) Values() ([]any, error) {
	return nil, nil
}
