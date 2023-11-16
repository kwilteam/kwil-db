package datasets

import (
	"context"

	"github.com/kwilteam/kwil-db/internal/accounts"
	"github.com/kwilteam/kwil-db/internal/engine/types"
	"github.com/kwilteam/kwil-db/internal/sql"
)

type AccountStore interface {
	Spend(ctx context.Context, spend *accounts.Spend) error
}

type Engine interface {
	// TODO: signer and caller will become the same thing after we update auth
	Call(ctx context.Context, dbid string, procedure string, caller []byte, signer []byte, args []any) (*sql.ResultSet, error)
	CreateDataset(ctx context.Context, schema *types.Schema, caller []byte) (err error)
	DeleteDataset(ctx context.Context, dbid string, caller []byte) error
	// TODO: signer and caller will become the same thing after we update auth
	Execute(ctx context.Context, dbid string, procedure string, caller []byte, signer []byte, args []any) error
	GetSchema(ctx context.Context, dbid string) (*types.Schema, error)
	ListDatasets(ctx context.Context, caller []byte) ([]string, error)
	Query(ctx context.Context, dbid string, query string) (*sql.ResultSet, error)
}
