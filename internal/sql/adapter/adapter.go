// package adapter provides an adapter for backwards compatibility
// to other packages in the codebase
package adapter

import (
	"context"

	"github.com/kwilteam/kwil-db/internal/sql"
)

type queryFunc func(ctx context.Context, stmt string, args ...any) (*sql.ResultSet, error)

func query(ctx context.Context, q queryFunc, stmt string, args ...any) ([]map[string]any, error) {
	res, err := q(ctx, stmt, args...)
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
			rows[i][res.ReturnedColumns[j]] = value
		}
	}
	return rows
}

type Datastore interface {
	sql.Queryer
	sql.Executor
}

type DB struct {
	Datastore
	// *pg.DB
}

func (p *DB) Execute(ctx context.Context, stmt string, args ...any) ([]map[string]any, error) {
	return query(ctx, p.Datastore.Execute, stmt, args...)
}

func (p *DB) Query(ctx context.Context, stmt string, args ...any) ([]map[string]any, error) {
	return query(ctx, p.Datastore.Query, stmt, args...)
}
