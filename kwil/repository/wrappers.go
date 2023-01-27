package repository

import (
	"context"
	"kwil/kwil/repository/gen"
	anytype "kwil/x/types/data_types/any_type"
	"kwil/x/types/databases"
)

// wrapping db retriever

func (q *queries) GetDatabase(ctx context.Context, id *databases.DatabaseIdentifier) (*databases.Database[anytype.KwilAny], error) {
	return q.dbRetriever.GetDatabase(ctx, id)
}

func (q *queries) ListDatabases(ctx context.Context) ([]*databases.DatabaseIdentifier, error) {
	return q.dbRetriever.ListDatabases(ctx)
}

func (q *queries) ListTables(ctx context.Context, dbid int32) ([]*gen.ListTablesRow, error) {
	return q.dbRetriever.ListTables(ctx, dbid)
}

func (q *queries) ListDatabasesByOwner(ctx context.Context, owner string) ([]string, error) {
	return q.dbRetriever.ListDatabasesByOwner(ctx, owner)
}

// wrapping schema manager
func (q *queries) CreateSchema(ctx context.Context, name string) error {
	return q.schema.CreateSchema(ctx, name)
}

func (q *queries) SchemaExists(ctx context.Context, name string) (bool, error) {
	return q.schema.SchemaExists(ctx, name)
}

func (q *queries) DropSchema(ctx context.Context, name string) error {
	return q.schema.DropSchema(ctx, name)
}
