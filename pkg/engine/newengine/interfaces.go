package engine

import (
	"context"

	"github.com/kwilteam/kwil-db/pkg/engine/dataset3"
	"github.com/kwilteam/kwil-db/pkg/engine/master"
	"github.com/kwilteam/kwil-db/pkg/engine/types"
)

type Datastore interface {
	Close() error
	Delete() error
	Execute(ctx context.Context, stmt string, args map[string]any) error
	Prepare(stmt string) (Statement, error)
	Query(ctx context.Context, query string, args map[string]any) ([]map[string]any, error)
	Savepoint() (Savepoint, error)
	TableExists(ctx context.Context, table string) (bool, error)
}

type Statement interface {
	Close() error
	Execute(ctx context.Context, args map[string]any) ([]map[string]any, error)
}

type Savepoint interface {
	Commit() error
	Rollback() error
}

type Dataset interface {
	Close() error
	Procedures() []*types.Procedure
	Tables(ctx context.Context) ([]*types.Table, error)
	Delete() error
	Query(ctx context.Context, stmt string, args map[string]any) ([]map[string]any, error)
	Execute(ctx context.Context, procedure string, args []map[string]any, opts *dataset3.TxOpts) ([]map[string]any, error)
}

type MasterDB interface {
	Close() error
	ListDatasets(ctx context.Context) ([]*master.DatasetInfo, error)
	ListDatasetsByOwner(ctx context.Context, owner string) ([]string, error)
	RegisterDataset(ctx context.Context, name, owner string) error
	UnregisterDataset(ctx context.Context, dbid string) error
}
