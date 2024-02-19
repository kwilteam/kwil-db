//go:build pglive

package accounts

import (
	"context"
	"errors"
	"math/big"
	"testing"

	sql "github.com/kwilteam/kwil-db/common/sql"
	"github.com/kwilteam/kwil-db/internal/sql/pg"

	"github.com/stretchr/testify/require"
)

var (
	account1 = []byte("account1")
	account2 = []byte("account2")
)

func Test_Accounts(t *testing.T) {
	cfg := &pg.DBConfig{
		PoolConfig: pg.PoolConfig{
			ConnConfig: pg.ConnConfig{
				Host:   "127.0.0.1",
				Port:   "5432",
				User:   "kwild",
				Pass:   "kwild", // would be ignored if pg_hba.conf set with trust
				DBName: "kwil_test_db",
			},
			MaxConns: 11,
		},
	}

	type testCase struct {
		name string
		fn   func(t *testing.T, db sql.DB)
	}

	// once we have a way to increase balances in accounts, we will have to add tests
	// for spending a valid amount
	testCases := []testCase{
		{
			name: "credit and debit",
			fn: func(t *testing.T, db sql.DB) {
				ctx := context.Background()

				err := Credit(ctx, db, account1, big.NewInt(100))
				require.NoError(t, err)

				err = Credit(ctx, db, account1, big.NewInt(-100))
				require.NoError(t, err)
			},
		},
		{
			name: "debit non-existent account",
			fn: func(t *testing.T, db sql.DB) {
				ctx := context.Background()

				err := Credit(ctx, db, account1, big.NewInt(-100))
				require.ErrorIs(t, err, ErrNegativeBalance)
			},
		},
		{
			name: "credit and over-debit",
			fn: func(t *testing.T, db sql.DB) {
				ctx := context.Background()

				err := Credit(ctx, db, account1, big.NewInt(100))
				require.NoError(t, err)

				err = Credit(ctx, db, account1, big.NewInt(-101))
				require.ErrorIs(t, err, ErrNegativeBalance)
			},
		},
		{
			name: "transfer to nonexistent account",
			fn: func(t *testing.T, db sql.DB) {
				ctx := context.Background()

				err := Credit(ctx, db, account1, big.NewInt(100))
				require.NoError(t, err)

				err = Transfer(ctx, db, account1, account2, big.NewInt(100))
				require.NoError(t, err)

				acc, err := GetAccount(ctx, db, account1)
				require.NoError(t, err)

				require.Equal(t, big.NewInt(0), acc.Balance)

				acc, err = GetAccount(ctx, db, account2)
				require.NoError(t, err)

				require.Equal(t, big.NewInt(100), acc.Balance)
			},
		},
		{
			name: "transfer to existing account",
			fn: func(t *testing.T, db sql.DB) {
				ctx := context.Background()

				err := Credit(ctx, db, account1, big.NewInt(100))
				require.NoError(t, err)

				err = Credit(ctx, db, account2, big.NewInt(100))
				require.NoError(t, err)

				err = Transfer(ctx, db, account1, account2, big.NewInt(50))
				require.NoError(t, err)

				acc, err := GetAccount(ctx, db, account1)
				require.NoError(t, err)

				require.Equal(t, big.NewInt(50), acc.Balance)

				acc, err = GetAccount(ctx, db, account2)
				require.NoError(t, err)

				require.Equal(t, big.NewInt(150), acc.Balance)
			},
		},
		{
			name: "transfer negative amount",
			fn: func(t *testing.T, db sql.DB) {
				ctx := context.Background()

				err := Credit(ctx, db, account1, big.NewInt(100))
				require.NoError(t, err)

				err = Transfer(ctx, db, account1, account2, big.NewInt(-50))
				require.ErrorIs(t, err, ErrNegativeTransfer)
			},
		},
		{
			name: "transfer more than you have",
			fn: func(t *testing.T, db sql.DB) {
				ctx := context.Background()

				err := Credit(ctx, db, account1, big.NewInt(100))
				require.NoError(t, err)

				err = Transfer(ctx, db, account1, account2, big.NewInt(150))
				require.ErrorIs(t, err, ErrInsufficientFunds)
			},
		},
		{
			name: "get non existent account",
			fn: func(t *testing.T, db sql.DB) {
				ctx := context.Background()

				acc, err := GetAccount(ctx, db, account1)
				require.NoError(t, err)

				require.Equal(t, big.NewInt(0), acc.Balance)
				require.Equal(t, int64(0), acc.Nonce)
			},
		},
		{
			name: "spend from non existent account",
			fn: func(t *testing.T, db sql.DB) {
				ctx := context.Background()

				err := Spend(ctx, db, account1, big.NewInt(100), 1)
				require.ErrorIs(t, err, ErrAccountNotFound)
			},
		},
		{
			name: "spend more than you have",
			fn: func(t *testing.T, db sql.DB) {
				ctx := context.Background()

				err := Credit(ctx, db, account1, big.NewInt(100))
				require.NoError(t, err)

				err = Spend(ctx, db, account1, big.NewInt(101), 1)
				require.ErrorIs(t, err, ErrInsufficientFunds)

				acc, err := GetAccount(ctx, db, account1)
				require.NoError(t, err)

				require.Equal(t, big.NewInt(100), acc.Balance)
			},
		},
		{
			name: "spend with invalid nonce",
			fn: func(t *testing.T, db sql.DB) {
				ctx := context.Background()

				err := Credit(ctx, db, account1, big.NewInt(100))
				require.NoError(t, err)

				err = Spend(ctx, db, account1, big.NewInt(50), 2)
				require.ErrorIs(t, err, ErrInvalidNonce)

				acc, err := GetAccount(ctx, db, account1)
				require.NoError(t, err)

				require.Equal(t, big.NewInt(100), acc.Balance)
			},
		},
		{
			name: "valid spend",
			fn: func(t *testing.T, db sql.DB) {
				ctx := context.Background()

				err := Credit(ctx, db, account1, big.NewInt(100))
				require.NoError(t, err)

				err = Spend(ctx, db, account1, big.NewInt(50), 1)
				require.NoError(t, err)

				acc, err := GetAccount(ctx, db, account1)
				require.NoError(t, err)

				require.Equal(t, big.NewInt(50), acc.Balance)
			},
		},
		{
			name: "spend 0 on non-existent account",
			fn: func(t *testing.T, db sql.DB) {
				ctx := context.Background()

				err := Spend(ctx, db, account1, big.NewInt(0), 1)
				require.NoError(t, err)

				acc, err := GetAccount(ctx, db, account1)
				require.NoError(t, err)

				require.Equal(t, big.NewInt(0), acc.Balance)
				require.Equal(t, int64(1), acc.Nonce)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			db, err := pg.NewDB(ctx, cfg)
			require.NoError(t, err)
			defer db.Close()
			tx, err := db.BeginTx(ctx)

			require.NoError(t, err)
			defer tx.Rollback(ctx) // always rollback to avoid cleanup

			defer db.Execute(ctx, `DROP SCHEMA IF EXISTS `+schemaName+` CASCADE;`)

			err = InitializeAccountStore(ctx, tx)
			require.NoError(t, err)

			tc.fn(t, tx)
		})
	}
}

// func newSpend(address string, amount int64, nonce int64) *accounts.Spend {
// 	return &accounts.Spend{
// 		AccountID: []byte(address),
// 		Amount:    big.NewInt(amount),
// 		Nonce:     nonce,
// 	}
// }

// func newAccount(address string, balance int64, nonce int64) *accounts.Account {
// 	return &accounts.Account{
// 		Identifier: []byte(address),
// 		Balance:    big.NewInt(balance),
// 		Nonce:      nonce,
// 	}
// }

func assertErr(t *testing.T, errs []error, target error) {
	t.Helper()
	if target == nil {
		if len(errs) > 0 {
			t.Fatalf("expected no error, got %s", errs)
		}
		return
	}

	contains := false
	for _, err := range errs {
		if errors.Is(err, target) {
			contains = true
		}
	}

	if !contains {
		t.Fatalf("expected error %s, got %s", target, errs)
	}
}
