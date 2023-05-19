package engine2

import (
	"context"

	"github.com/kwilteam/kwil-db/pkg/engine2/dto"
	"github.com/kwilteam/kwil-db/pkg/engine2/sqldb"
)

type datastore interface {
	// PrepareRaw prepares a sql statement without parsing it.
	// This should never be exposed to the user.
	PrepareRaw(query string) (sqldb.Statement, error)

	// CreateTable creates a new table.
	CreateTable(ctx context.Context, table *dto.Table) error

	// Close closes the database connection.
	Close() error

	// Query executes a read-only query and returns the result.
	Query(ctx context.Context, query string, args map[string]any) (dto.Result, error)

	// Delete deletes the database.

	// Execute executes a write query
	Execute(query string, args map[string]any) error
}
