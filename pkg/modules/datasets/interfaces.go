package datasets

import (
	"context"

	"github.com/kwilteam/kwil-db/pkg/balances"
	"github.com/kwilteam/kwil-db/pkg/engine"
	engineTypes "github.com/kwilteam/kwil-db/pkg/engine/types"
)

type AccountStore interface {
	Spend(ctx context.Context, spend *balances.Spend) error
}

type Engine interface {
	CreateDataset(ctx context.Context, schema *engineTypes.Schema) (dbid string, finalErr error)
	DropDataset(ctx context.Context, sender, dbid string) error
	Execute(ctx context.Context, dbid string, procedure string, args []map[string]any, opts ...engine.ExecutionOpt) ([]map[string]any, error)
	ListDatasets(ctx context.Context, owner string) ([]string, error)
	Query(ctx context.Context, dbid string, query string) ([]map[string]any, error)
	GetSchema(ctx context.Context, dbid string) (*engineTypes.Schema, error)
}
