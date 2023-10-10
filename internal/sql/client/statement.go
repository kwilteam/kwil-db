package client

import (
	"context"

	"github.com/kwilteam/kwil-db/internal/sql/sqlite"
)

type Statement struct {
	stmt *sqlite.Statement
}

func (s *Statement) Execute(ctx context.Context, args map[string]any) ([]map[string]any, error) {
	results, err := s.stmt.Start(ctx,
		sqlite.WithNamedArgs(args),
	)
	if err != nil {
		return nil, err
	}

	return NewCursor(results).Export()
}

func (s *Statement) Close() error {
	return s.stmt.Finalize()
}
