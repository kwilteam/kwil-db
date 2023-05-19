package engine2

import (
	"context"

	"github.com/kwilteam/kwil-db/pkg/engine2/dto"
	"github.com/kwilteam/kwil-db/pkg/engine2/sqldb"
)

type Dataset interface {
	Savepoint() (sqldb.Savepoint, error)

	ListActions() []*dto.Action

	ListTables() []*dto.Table

	CreateTable(ctx context.Context, table *dto.Table) error

	CreateAction(ctx context.Context, action *dto.Action) error

	Execute(txCtx *dto.TxContext, inputs []map[string]any) (dto.Result, error)

	Close() error
}
