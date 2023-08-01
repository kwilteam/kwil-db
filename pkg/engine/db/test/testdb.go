package test

import (
	"context"
	"io"

	"github.com/kwilteam/kwil-db/pkg/engine/db"
	sqlTesting "github.com/kwilteam/kwil-db/pkg/sql/testing"
)

func OpenTestDB(ctx context.Context) (*db.DB, func() error, error) {
	testDb, closeFunc, err := sqlTesting.OpenTestDB("test")
	if err != nil {
		return nil, nil, err
	}

	datastore, err := db.NewDB(ctx, databaseAdapter{testDb})
	if err != nil {
		return nil, nil, err
	}

	return datastore, closeFunc, nil
}

type databaseAdapter struct {
	sqlTesting.TestSqliteClient
}

func (d databaseAdapter) Prepare(query string) (db.Statement, error) {
	return d.TestSqliteClient.Prepare(query)
}

func (d databaseAdapter) Savepoint() (db.Savepoint, error) {
	return d.TestSqliteClient.Savepoint()
}

func (d databaseAdapter) CreateSession() (db.Session, error) {
	return d.TestSqliteClient.CreateSession()
}

func (d databaseAdapter) ApplyChangeset(changeset io.Reader) error {
	return d.TestSqliteClient.ApplyChangeset(changeset)
}
