package repository

import (
	"context"
	"kwil/internal/repository/gen"
)

type DbInfo interface {
	GetTableSize(ctx context.Context, schema, table string) (int64, error)
	GetIndexedColumnCount(ctx context.Context, schema, table string) (int64, error)
}

func (q *queries) GetTableSize(ctx context.Context, schema, table string) (int64, error) {
	return q.gen.GetTableSize(ctx, &gen.GetTableSizeParams{
		Schemaname: schema,
		Relname:    table,
	})
}

func (q *queries) GetIndexedColumnCount(ctx context.Context, schema, table string) (int64, error) {
	return q.gen.GetIndexedColumnCount(ctx, &gen.GetIndexedColumnCountParams{
		Schemaname: schema,
		Tablename:  table,
	})
}
