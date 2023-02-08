package dbretriever

import (
	"context"
	"kwil/internal/repository/gen"
	"kwil/pkg/databases"
	"kwil/pkg/databases/spec"
)

func (q *dbRetriever) ListTables(ctx context.Context, dbid int32) ([]*gen.ListTablesRow, error) {
	return q.gen.ListTables(ctx, dbid)
}

func (q *dbRetriever) GetTables(ctx context.Context, dbid int32) ([]*databases.Table[*spec.KwilAny], error) {
	tableList, err := q.ListTables(ctx, dbid)
	if err != nil {
		return nil, err
	}

	tables := make([]*databases.Table[*spec.KwilAny], len(tableList))
	for i, table := range tableList {
		columns, err := q.GetColumns(ctx, table.ID)
		if err != nil {
			return nil, err
		}

		tables[i] = &databases.Table[*spec.KwilAny]{
			Name:    table.TableName,
			Columns: columns,
		}
	}

	return tables, nil
}
