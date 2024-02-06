package types

import (
	"context"

	"github.com/kwilteam/kwil-db/internal/sql"
)

type ResultSetFunc func(ctx context.Context, stmt string, args map[string]any) (*sql.ResultSet, error)

// ExecFunc allows a simpler implementation for certain functions.
type ExecFunc func(ctx context.Context, stmt string, args map[string]any) error
