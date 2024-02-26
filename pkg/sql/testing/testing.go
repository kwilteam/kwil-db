package client

import (
	"context"
	"errors"

	"github.com/kwilteam/kwil-db/pkg/log"
	"github.com/kwilteam/kwil-db/pkg/sql"
	"github.com/kwilteam/kwil-db/pkg/sql/client"
	"github.com/kwilteam/kwil-db/pkg/sql/sqlite"
)

// OpenTestDB opens a test database for use in unit tests.
// It returns a SqliteClient, a cleanup function, and an error.
func OpenTestDB(name string) (connection TestSqliteClient, teardown func() error, err error) {
	db, closeFunc, err := sqlite.OpenDbWithTearDown(name)
	if err != nil {
		return nil, nil, err
	}

	clnt, err := client.WrapSqliteConn(db, log.NewNoOp())
	if err != nil {
		return nil, nil, errors.Join(closeFunc(), err)
	}

	return &wrappedSqliteClient{clnt}, closeFunc, nil
}

type wrappedSqliteClient struct {
	*client.SqliteClient
}

func (w *wrappedSqliteClient) Prepare(query string) (sql.Statement, error) {
	return w.SqliteClient.Prepare(query)
}

func (w *wrappedSqliteClient) Savepoint() (sql.Savepoint, error) {
	return w.SqliteClient.Savepoint()
}

// we need to get rid of close and delete since the teardown function will handle that
func (w *wrappedSqliteClient) Close() error {
	return nil
}

func (w *wrappedSqliteClient) Delete() error {
	return nil
}

type TestSqliteClient interface {
	Close() error
	Delete() error
	Execute(context.Context, string, map[string]any) error
	Prepare(string) (sql.Statement, error)
	Query(context.Context, string, map[string]any) ([]map[string]any, error)
	Savepoint() (sql.Savepoint, error)
	TableExists(context.Context, string) (bool, error)
}
