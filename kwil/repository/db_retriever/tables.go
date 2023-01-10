package dbretriever

import (
	"context"
	"kwil/kwil/repository/gen"
	"kwil/x/types/databases"
)

func (q *dbRetriever) ListTables(ctx context.Context, dbid int32) ([]*gen.ListTablesRow, error) {
	return q.gen.ListTables(ctx, dbid)
}

func (q *dbRetriever) GetTables(ctx context.Context, dbid int32) ([]*databases.Table, error) {
	tableList, err := q.ListTables(ctx, dbid)
	if err != nil {
		return nil, err
	}

	tables := make([]*databases.Table, len(tableList))
	for i, table := range tableList {
		columns, err := q.GetColumns(ctx, table.ID)
		if err != nil {
			return nil, err
		}

		tables[i] = &databases.Table{
			Name:    table.TableName,
			Columns: columns,
		}
	}

	return tables, nil
}
