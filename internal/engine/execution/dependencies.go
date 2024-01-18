package execution

import (
	"context"

	sql "github.com/kwilteam/kwil-db/internal/sql"
)

// Registry is the interface for a database registry.
// The database registry should handle connections, locking, persistence, and transaction atomicity.
type Registry interface {
	Databases

	// Create creates a database.
	Create(ctx context.Context, dbid string) error

	// Delete deletes a database.
	Delete(ctx context.Context, dbid string) error

	// List lists the databases that are available.
	List(ctx context.Context) ([]string, error)
}

// Databases is an interface for interacting with databases.
type Databases interface {
	// Query executes a query against a reader connection
	// It will not read uncommitted data, and cannot be used to write data.
	Query(ctx context.Context, dbid string, stmt string, params map[string]any) (*sql.ResultSet, error)

	// Execute executes a statement against the database.
	// The statement can mutate state, and will read uncommitted data.
	Execute(ctx context.Context, dbid string, stmt string, params map[string]any) (*sql.ResultSet, error)

	// Set sets a key to a value.
	Set(ctx context.Context, dbid string, key []byte, value []byte) error

	// Get gets a value for a key.
	// it contains a sync flag, which indicates whether it should read uncommitted data.
	Get(ctx context.Context, dbid string, key []byte, sync bool) ([]byte, error)
}
