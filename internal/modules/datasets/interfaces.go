package datasets

import (
	"context"

	"github.com/kwilteam/kwil-db/internal/accounts"
	"github.com/kwilteam/kwil-db/internal/engine"
	engineTypes "github.com/kwilteam/kwil-db/internal/engine/types"
)

type AccountStore interface {
	Spend(ctx context.Context, spend *accounts.Spend) error
}

type Engine interface {
	CreateDataset(ctx context.Context, schema *engineTypes.Schema, caller *engineTypes.User) (dbid string, finalErr error)
	DropDataset(ctx context.Context, dbid string, sender *engineTypes.User) error
	Execute(ctx context.Context, dbid string, procedure string, args [][]any, opts ...engine.ExecutionOpt) ([]map[string]any, error)
	ListDatasets(ctx context.Context, owner []byte) ([]string, error)
	Query(ctx context.Context, dbid string, query string) ([]map[string]any, error)
	GetSchema(ctx context.Context, dbid string) (*engineTypes.Schema, error)
}
