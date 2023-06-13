package sqldb

import (
	"context"

	"github.com/kwilteam/kwil-db/pkg/engine/dto"
)

// Datastore is an interface for interacting with a database.
// It is used by both the main application as well as extensions.
type Datastore interface {
	// Prepare returns a prepared statement, bound to this connection.
	Prepare(query string) (Statement, error)

	// Query executes a read-only query and returns the result.
	Query(ctx context.Context, query string, args map[string]any) (dto.Result, error)

	// Close closes the database connection.
	Close() error

	// Delete deletes the database.
	Delete() error

	// Savepoint creates a new savepoint.
	Savepoint() (Savepoint, error)

	// ListTables returns a list of all tables in the database.
	ListTables(ctx context.Context) ([]*dto.Table, error)

	// CreateTable creates a new table.
	CreateTable(ctx context.Context, table *dto.Table) error
}

// DB is an interface used by deployed datasets to interact with the underlying database.
type DB interface {
	Datastore

	// ListActions returns a list of all actions in the database.
	ListActions(ctx context.Context) ([]*dto.Action, error)

	// StoreAction persists a new action.
	StoreAction(ctx context.Context, action *dto.Action) error
}

type Statement interface {
	// Execute executes a prepared statement with the given arguments.
	Execute(args map[string]any) (dto.Result, error)

	// Close closes the statement.
	Close() error
}

type Savepoint interface {
	// Rollback rolls back the savepoint.
	Rollback() error

	// Commit commits the savepoint.
	Commit() error
}
