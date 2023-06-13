package db

import (
	"context"
)

type SqlDB interface {
	// Execute executes a statement.
	Execute(ctx context.Context, stmt string, args map[string]any) error

	// Query executes a query and returns the result.
	Query(ctx context.Context, query string, args map[string]any) ([]map[string]any, error)

	// Prepare prepares a statement for execution, and returns a Statement.
	Prepare(stmt string) (Statement, error)

	// TableExists checks if a table exists.
	TableExists(ctx context.Context, table string) (bool, error)

	// Close closes the connection to the database.
	Close() error

	// Delete deletes the database.
	Delete() error

	// Savepoint creates a savepoint.
	Savepoint() (Savepoint, error)
}

type Savepoint interface {
	Rollback() error
	Commit() error
}

type Statement interface {
	Execute(ctx context.Context, args map[string]any) ([]map[string]any, error)
	Close() error
}
