package repository

import (
	"context"
	"kwil/kwil/repository/gen"
	"kwil/x/types/databases"
)

// wrapping db retriever

func (q *queries) GetDatabase(ctx context.Context, id *databases.DatabaseIdentifier) (*databases.Database, error) {
	return q.dbRetriever.GetDatabase(ctx, id)
}

func (q *queries) ListDatabases(ctx context.Context) ([]*databases.DatabaseIdentifier, error) {
	return q.dbRetriever.ListDatabases(ctx)
}

func (q *queries) ListTables(ctx context.Context, dbid int32) ([]*gen.ListTablesRow, error) {
	return q.dbRetriever.ListTables(ctx, dbid)
}
