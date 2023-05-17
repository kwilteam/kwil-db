package dataset

import (
	"context"

	"github.com/kwilteam/kwil-db/pkg/engine2/dto"
)

type DB interface {
	// Prepare returns a prepared statement, bound to this connection.
	Prepare(query string) (Statement, error)

	// Query executes a read-only query and returns the result.
	Query(ctx context.Context, query string, args map[string]any) (dto.Result, error)

	// Close closes the database connection.
	Close() error

	// Savepoint creates a new savepoint.
	Savepoint() (Savepoint, error)

	// ListTables returns a list of all tables in the database.
	ListTables() ([]*dto.Table, error)

	// ListActions returns a list of all actions in the database.
	ListActions() ([]*dto.Action, error)

	// CreateTable creates a new table.
	CreateTable(table *dto.Table) error

	// CreateAction persists a new action.
	CreateAction(action *dto.Action) error
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
