//// go:build pglive

package accounts

import (
	"context"
	"kwil/node/pg"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

var testConfig = &pg.DBConfig{
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

func Test_AccountsLive(t *testing.T) {
	for _, tc := range acctsTestCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			db, err := pg.NewDB(ctx, testConfig)
			require.NoError(t, err)
			defer db.Close()
			tx, err := db.BeginTx(ctx)

			require.NoError(t, err)
			defer tx.Rollback(ctx) // always rollback to avoid cleanup

			defer db.Execute(ctx, `DROP SCHEMA IF EXISTS `+schemaName+` CASCADE;`)

			accounts, err := InitializeAccountStore(ctx, tx)
			require.NoError(t, err)

			tc.fn(t, tx, accounts, nil, true)
		})
	}
}

func TestGetAccount(t *testing.T) {
	ctx := context.Background()
	db, err := pg.NewDB(ctx, testConfig)
	require.NoError(t, err)
	defer db.Close()

	tx1, err := db.BeginTx(ctx)
	require.NoError(t, err)
	defer tx1.Rollback(ctx)

	defer db.Execute(ctx, `DROP SCHEMA IF EXISTS `+schemaName+` CASCADE;`)

	accounts, err := InitializeAccountStore(ctx, tx1)
	require.NoError(t, err)
	tx1.Commit(ctx)

	// Credit an account
	tx2, err := db.BeginPreparedTx(ctx)
	require.NoError(t, err)
	err = accounts.Credit(ctx, tx2, account1, big.NewInt(100))
	require.NoError(t, err)

	// Get the account (non-consensus tx)
	readTx, err := db.BeginReadTx(ctx)
	require.NoError(t, err)
	defer readTx.Rollback(ctx)
	acc, err := accounts.GetAccount(ctx, readTx, account1)
	require.NoError(t, err)
	require.Equal(t, int64(0), acc.Balance.Int64())

	// Generally this should be called after the tx is committed, for testing purposes
	accounts.Commit()
	account, err := accounts.GetAccount(ctx, readTx, account1)
	require.NoError(t, err)
	require.Equal(t, big.NewInt(100), account.Balance)
}
