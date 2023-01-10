package dbretriever

import (
	"context"
	"fmt"
	"kwil/x/types/databases"
)

func (q *dbRetriever) GetQueries(ctx context.Context, dbid int32) ([]*databases.SQLQuery, error) {
	queryList, err := q.gen.GetQueries(ctx, dbid)
	if err != nil {
		return nil, fmt.Errorf(`error getting queries for dbid %d: %w`, dbid, err)
	}

	queries := make([]*databases.SQLQuery, len(queryList))
	for i, query := range queryList {
		var q databases.SQLQuery
		err = q.DecodeGOB(query.Query)
		if err != nil {
			return nil, fmt.Errorf(`error decoding query %d: %w`, query.ID, err)
		}

		queries[i] = &q
	}

	return queries, nil
}
