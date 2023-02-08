package dbretriever

import (
	"context"
	"database/sql"
	gen2 "kwil/internal/repository/gen"
	"kwil/pkg/databases"
	"kwil/pkg/databases/spec"
)

type DatabaseRetriever interface {
	// GetDatabase returns a database by name and owner
	GetDatabase(ctx context.Context, database *databases.DatabaseIdentifier) (*databases.Database[*spec.KwilAny], error)

	// ListDatabases returns a list of all databases
	ListDatabases(ctx context.Context) ([]*databases.DatabaseIdentifier, error)

	// ListDatabasesByOwner returns a list of all databases owned by a user
	ListDatabasesByOwner(ctx context.Context, owner string) ([]string, error)

	// ListTables returns a list of all table ids in a database
	ListTables(ctx context.Context, dbid int32) ([]*gen2.ListTablesRow, error)
}

type DatabaseRetrieverTxer interface {
	DatabaseRetriever
	WithTx(tx *sql.Tx) DatabaseRetrieverTxer
}

type dbRetriever struct {
	gen *gen2.Queries
}

func New(gen *gen2.Queries) DatabaseRetrieverTxer {
	return &dbRetriever{
		gen: gen,
	}
}

func (q *dbRetriever) WithTx(tx *sql.Tx) DatabaseRetrieverTxer {
	return &dbRetriever{
		gen: q.gen.WithTx(tx),
	}
}
