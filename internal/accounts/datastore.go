package accounts

import (
	"context"

	"github.com/kwilteam/kwil-db/internal/sql"
)

type Datastore interface {
	Execute(ctx context.Context, stmt string, args map[string]any) error
	Query(ctx context.Context, query string, args map[string]any) ([]map[string]any, error)
	Prepare(stmt string) (sql.Statement, error)
}
