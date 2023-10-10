package test

import (
	"context"

	"github.com/kwilteam/kwil-db/internal/engine/db"
	sqlTesting "github.com/kwilteam/kwil-db/internal/sql/testing"
)

func OpenTestDB(ctx context.Context) (*db.DB, func() error, error) {
	testDb, closeFunc, err := sqlTesting.OpenTestDB("test")
	if err != nil {
		return nil, nil, err
	}

	datastore, err := db.NewDB(ctx, testDb)
	if err != nil {
		return nil, nil, err
	}

	return datastore, closeFunc, nil
}
