package dbretriever

import (
	"context"
	"fmt"
	"kwil/pkg/databases"
)

func (q *dbRetriever) GetIndexes(ctx context.Context, dbid int32) ([]*databases.Index, error) {
	tableList, err := q.gen.ListTables(ctx, dbid)
	if err != nil {
		return nil, fmt.Errorf(`error getting tables for dbid %d: %w`, dbid, err)
	}

	indexes := make([]*databases.Index, 0)
	for _, table := range tableList {
		indexList, err := q.gen.GetIndexes(ctx, table.ID)
		if err != nil {
			return nil, fmt.Errorf(`error getting indexes for table %s: %w`, table.TableName, err)
		}

		for _, index := range indexList {
			indexes = append(indexes, &databases.Index{
				Name:    index.IndexName,
				Table:   table.TableName,
				Columns: index.Columns,
				Using:   databases.IndexType(index.IndexType),
			})
		}
	}

	return indexes, nil
}
