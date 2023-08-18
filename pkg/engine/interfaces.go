package engine

import (
	"context"
	"io"

	"github.com/kwilteam/kwil-db/pkg/engine/dataset"
	"github.com/kwilteam/kwil-db/pkg/engine/types"
	"github.com/kwilteam/kwil-db/pkg/sql"
)

type Dataset interface {
	Close() error
	ListProcedures() []*types.Procedure
	ListExtensions(ctx context.Context) ([]*types.Extension, error)
	ListTables(ctx context.Context) ([]*types.Table, error)
	Metadata() (name, owner string)
	Delete() error
	Query(ctx context.Context, stmt string, args map[string]any) ([]map[string]any, error)
	Execute(ctx context.Context, procedure string, args [][]any, opts *dataset.TxOpts) ([]map[string]any, error)
	ApplyChangeset(changeset io.Reader) error
	Savepoint() (sql.Savepoint, error)
	CreateSession() (sql.Session, error)
	Call(ctx context.Context, procedure string, args []any, opts *dataset.TxOpts) ([]map[string]any, error)
}

type MasterDB interface {
	Close() error
	ListDatasets(ctx context.Context) ([]*types.DatasetInfo, error)
	ListDatasetsByOwner(ctx context.Context, owner string) ([]string, error)
	RegisterDataset(ctx context.Context, name, owner string) error
	UnregisterDataset(ctx context.Context, dbid string) error
}

// CommitRegister is an interface for registering atomically committable data stores
// Any database registered to this will be atomically synced in a 2pc transaction
type CommitRegister interface {
	// Register registers a database to the commit register
	Register(ctx context.Context, name string, db sql.Database) error
	// Unregister unregisters a database from the commit register
	Unregister(ctx context.Context, name string) error
}
