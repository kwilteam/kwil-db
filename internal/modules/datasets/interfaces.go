package datasets

import (
	"context"

	coreTypes "github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/internal/accounts"
	"github.com/kwilteam/kwil-db/internal/engine/types"
	"github.com/kwilteam/kwil-db/internal/sql"
)

type AccountStore interface {
	Spend(ctx context.Context, spend *accounts.Spend) error
}

type Engine interface {
	CreateDataset(ctx context.Context, schema *types.Schema, caller []byte) (err error)
	DeleteDataset(ctx context.Context, dbid string, caller []byte) error
	Execute(ctx context.Context, data *types.ExecutionData) (*sql.ResultSet, error)
	GetSchema(ctx context.Context, dbid string) (*types.Schema, error)
	ListDatasets(ctx context.Context, caller []byte) ([]*coreTypes.DatasetInfo, error)
	Query(ctx context.Context, dbid string, query string) (*sql.ResultSet, error)
}
