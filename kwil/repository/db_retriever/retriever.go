package dbretriever

import (
	"context"
	"database/sql"
	"kwil/kwil/repository/gen"
	"kwil/x/types/databases"
)

type DatabaseRetriever interface {
	// GetDatabase returns a database by name and owner
	GetDatabase(ctx context.Context, database *databases.DatabaseIdentifier) (*databases.Database, error)

	// ListDatabases returns a list of all databases
	ListDatabases(ctx context.Context) ([]*databases.DatabaseIdentifier, error)

	// ListTables returns a list of all table ids in a database
	ListTables(ctx context.Context, dbid int32) ([]*gen.ListTablesRow, error)
}

type dbRetriever struct {
	gen *gen.Queries
}

func New(gen *gen.Queries) DatabaseRetriever {
	return &dbRetriever{
		gen: gen,
	}
}

func (q *dbRetriever) WithTx(tx *sql.Tx) DatabaseRetriever {
	return &dbRetriever{
		gen: q.gen.WithTx(tx),
	}
}
