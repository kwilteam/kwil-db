package testing

import (
	"context"

	"github.com/kwilteam/kwil-db/pkg/balances"
	sqlTesting "github.com/kwilteam/kwil-db/pkg/sql/testing"
)

// NewTestAccountStore creates a new account store for testing.
// It returns an account store, a function to tear down the database, and an error.
func NewTestAccountStore(ctx context.Context, opts ...balances.AccountStoreOpts) (*balances.AccountStore, func() error, error) {
	ds, td, err := sqlTesting.OpenTestDB("test_account_store")
	if err != nil {
		return nil, nil, err
	}

	accStore, err := balances.NewAccountStore(ctx,
		ds,
		opts...,
	)
	if err != nil {
		return nil, nil, err
	}

	return accStore, td, nil
}
