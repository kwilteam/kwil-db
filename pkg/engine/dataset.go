package engine

import (
	"context"

	"github.com/kwilteam/kwil-db/pkg/engine/dto"
	"github.com/kwilteam/kwil-db/pkg/engine/sqldb"
)

// A Dataset is a stored database instance.
// It is conceptually similar to a Postgres "schema".
type Dataset interface {
	// Savepoint creates a new savepoint.
	Savepoint() (sqldb.Savepoint, error)

	// ListActions returns a list of all actions in the database.
	ListActions() []*dto.Action

	// ListTables returns a list of all tables in the database.
	ListTables() []*dto.Table

	// CreateTable creates a new table.
	CreateTable(ctx context.Context, table *dto.Table) error

	// CreateAction creates a new action.
	CreateAction(ctx context.Context, action *dto.Action) error

	// Execute executes an action and returns the result.
	Execute(txCtx *dto.TxContext, inputs []map[string]any) (dto.Result, error)

	// Query executes a read-only query and returns the result.
	Query(ctx context.Context, stmt string, args map[string]any) (dto.Result, error)

	// Id returns the id of the dataset.
	Id() string

	// Owner returns the owner of the dataset.
	Owner() string

	// Name returns the name of the dataset.
	Name() string

	Delete(txCtx *dto.TxContext) error
}

// internalDataset exposes more than the public Dataset interface.
type internalDataset interface {
	Dataset

	// Delete deletes the dataset.
	Delete(txCtx *dto.TxContext) error

	// Close closes the dataset.
	Close() error

	// Owner returns the owner of the dataset.
	Owner() string
}
