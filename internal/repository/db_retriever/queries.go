package dbretriever

import (
	"context"
	"fmt"
	"kwil/pkg/databases"
	"kwil/pkg/databases/convert"
	"kwil/pkg/databases/spec"
)

func (q *dbRetriever) GetQueries(ctx context.Context, dbid int32) ([]*databases.SQLQuery[*spec.KwilAny], error) {
	queryList, err := q.gen.GetQueries(ctx, dbid)
	if err != nil {
		return nil, fmt.Errorf(`error getting queries for dbid %d: %w`, dbid, err)
	}

	queries := make([]*databases.SQLQuery[*spec.KwilAny], len(queryList))
	for i, query := range queryList {
		var q databases.SQLQuery[[]byte]
		err = q.DecodeGOB(query.Query)
		if err != nil {
			return nil, fmt.Errorf(`error decoding query %d: %w`, query.ID, err)
		}

		// convert bytes to anytype.KwilAny
		qry, err := convert.Bytes.SQLQueryToKwilAny(&q)
		if err != nil {
			return nil, fmt.Errorf(`error converting query %d: %w`, query.ID, err)
		}

		queries[i] = qry
	}

	return queries, nil
}
