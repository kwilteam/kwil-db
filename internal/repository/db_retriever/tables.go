package dbretriever

import (
	"context"
	"kwil/internal/repository/gen"
	"kwil/pkg/types/data_types/any_type"
	"kwil/pkg/types/databases"
)

func (q *dbRetriever) ListTables(ctx context.Context, dbid int32) ([]*gen.ListTablesRow, error) {
	return q.gen.ListTables(ctx, dbid)
}

func (q *dbRetriever) GetTables(ctx context.Context, dbid int32) ([]*databases.Table[anytype.KwilAny], error) {
	tableList, err := q.ListTables(ctx, dbid)
	if err != nil {
		return nil, err
	}

	tables := make([]*databases.Table[anytype.KwilAny], len(tableList))
	for i, table := range tableList {
		columns, err := q.GetColumns(ctx, table.ID)
		if err != nil {
			return nil, err
		}

		tables[i] = &databases.Table[anytype.KwilAny]{
			Name:    table.TableName,
			Columns: columns,
		}
	}

	return tables, nil
}
