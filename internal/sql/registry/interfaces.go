package registry

import (
	"context"

	sql "github.com/kwilteam/kwil-db/internal/sql"
)

// These are the interfaces required by the Registry to manage datasets

type Queryer interface { // dup of as sql.Queryer?
	Query(ctx context.Context, query string, args ...any) (*sql.ResultSet, error)
}

type Executor interface {
	Execute(ctx context.Context, stmt string, args ...any) (*sql.ResultSet, error)
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

// Tx should be returned by the connection pool's BeginTx method. It can do it
// all including writes, unlike Pool methods.
type Tx interface {
	Queryer
	Executor
	sql.Tx // just Commit and Rollback
}

// Pool is a connection pool. Writes must be done by creating a Tx with BeginTx
// and using it's Exec and Set methods. The implementation may deny more than
// one concurrent write transaction from BeginTx.
type PoolTransactor interface {
	Queryer

	// BeginTx starts a transaction. All writes (and uncommitted reads) are done
	// through the methods of the returned Tx. End the transaction with the
	// Commit or Rollback method of the Tx.
	BeginTx(context.Context) (Tx, error)
}
