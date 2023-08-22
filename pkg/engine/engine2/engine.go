package engine2

import (
	"context"

	"github.com/kwilteam/kwil-db/pkg/engine"
	engineTypes "github.com/kwilteam/kwil-db/pkg/engine/types"
)

// TODO: delete this
type IEngine interface {
	EngineStateMachine
	ListDatasets(ctx context.Context, owner string) ([]string, error)
	Query(ctx context.Context, dbid string, query string) ([]map[string]any, error)
	GetSchema(ctx context.Context, dbid string) (*engineTypes.Schema, error)
}

type EngineStateMachine interface {
	CreateDataset(ctx context.Context, schema *engineTypes.Schema) (dbid string, finalErr error)
	DropDataset(ctx context.Context, sender, dbid string) error
	Execute(ctx context.Context, dbid string, procedure string, args [][]any, opts ...engine.ExecutionOpt) ([]map[string]any, error)
}

type Engine struct{}
