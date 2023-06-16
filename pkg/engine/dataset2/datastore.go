package dataset2

import (
	"context"

	"github.com/kwilteam/kwil-db/pkg/engine/dto"
)

// Datastore is an interface for a datastore, usually a sqlite DB.
type Datastore interface {
	Prepare(query string) (PreparedStatement, error)
	StoreExtension(ctx context.Context, ext *dto.ExtensionInitialization) error
	GetExtensions(ctx context.Context) ([]*dto.ExtensionInitialization, error)
	CreateTable(ctx context.Context, table *dto.Table) error
	ListTables(ctx context.Context) ([]*dto.Table, error)
	StoreProcedure(ctx context.Context, action *dto.Action) error
	ListProcedures(ctx context.Context) ([]*dto.Action, error)
	Savepoint() (Savepoint, error)
	Close() error
	Delete() error
	Query(context.Context, string, map[string]any) (dto.Result, error)
}

type PreparedStatement interface {
	// Execute executes a prepared statement with the given arguments.
	Execute(args map[string]any) (dto.Result, error)

	// Close closes the statement.
	Close() error
}

type Savepoint interface {
	Rollback() error
	Commit() error
}
