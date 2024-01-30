package plan

import (
	"context"

	"github.com/kwilteam/kwil-db/internal/engine/types"
)

type SchemaGetter interface {
	//SchemaByID(dbid string) *types.Schema
	//TableByName(schemaName, tableName string) *types.Table
	GetSchema(ctx context.Context, dbid string) (*types.Schema, error)
	TableByName(ctx context.Context, schema *types.Schema, tableName string) (*types.Table, error)
}
