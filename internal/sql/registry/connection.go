package registry

import (
	"context"

	sql "github.com/kwilteam/kwil-db/internal/sql"
)

// PoolOpener is a function that opens a connection pool.
// If create is true, then it will create the database if it does not exist.
type PoolOpener func(ctx context.Context, dbid string, create bool) (Pool, error)

// Pool is a connection pool.
type Pool interface {
	KV
	// Close closes the pool.
	Close() error

	Execute(ctx context.Context, stmt string, args map[string]any) (*sql.ResultSet, error)
	Query(ctx context.Context, query string, args map[string]any) (*sql.ResultSet, error)

	// Savepoint creates a savepoint.
	Savepoint() (sql.Savepoint, error)
	// CreateSession creates a session.
	CreateSession() (sql.Session, error)
}

type KV interface {
	// Set sets a key to a value.
	Set(ctx context.Context, key []byte, value []byte) error
	// Get gets a value for a key.
	Get(ctx context.Context, key []byte, sync bool) ([]byte, error)
}
