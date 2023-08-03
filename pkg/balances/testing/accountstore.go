package testing

import (
	"context"

	"github.com/kwilteam/kwil-db/pkg/balances"
	sqlTesting "github.com/kwilteam/kwil-db/pkg/sql/testing"
)

func NewTestAccountStore(ctx context.Context, opts ...balances.AccountStoreOpts) (*balances.AccountStore, func() error, error) {
	ds, td, err := sqlTesting.OpenTestDB("test_account_store")
	if err != nil {
		return nil, nil, err
	}

	accStore, err := balances.NewAccountStore(ctx,
		append(opts, balances.WithDatastore(&dbAdapter{ds}))...,
	)
	if err != nil {
		return nil, nil, err
	}

	return accStore, td, nil
}

// TODO: adapter should get deleted once we merge in main

type dbAdapter struct {
	sqlTesting.TestSqliteClient
}

func (s *dbAdapter) Savepoint() (balances.Savepoint, error) {
	return s.TestSqliteClient.Savepoint()
}

func (s *dbAdapter) Prepare(stmt string) (balances.PreparedStatement, error) {
	return s.TestSqliteClient.Prepare(stmt)
}

func (s *dbAdapter) CreateSession() (balances.Session, error) {
	return s.TestSqliteClient.CreateSession()
}
