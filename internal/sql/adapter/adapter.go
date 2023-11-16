// package adapater provides an adapter for backwards compatibility
// to other packages in the codebase
package adapter

import (
	"context"

	"github.com/kwilteam/kwil-db/internal/sql"
	"github.com/kwilteam/kwil-db/internal/sql/sqlite"
)

type PoolAdapater struct {
	*sqlite.Pool
}

func (p *PoolAdapater) Execute(ctx context.Context, stmt string, args map[string]any) ([]map[string]any, error) {
	res, err := p.Pool.Execute(ctx, stmt, args)
	if err != nil {
		return nil, err
	}

	return toMaps(res), nil
}

func (p *PoolAdapater) Query(ctx context.Context, stmt string, args map[string]any) ([]map[string]any, error) {
	res, err := p.Pool.Query(ctx, stmt, args)
	if err != nil {
		return nil, err
	}

	return toMaps(res), nil
}

func toMaps(res *sql.ResultSet) []map[string]any {
	if res == nil {
		return nil
	}

	rows := make([]map[string]any, len(res.Rows))
	for i, row := range res.Rows {
		rows[i] = make(map[string]any)
		for j, value := range row {
			rows[i][res.Columns[j]] = value
		}
	}
	return rows
}
