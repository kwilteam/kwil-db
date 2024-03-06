//go:build pglive

package accounts

import (
	"context"
	"testing"

	"github.com/kwilteam/kwil-db/internal/sql/pg"

	"github.com/stretchr/testify/require"
)

func Test_AccountsLive(t *testing.T) {
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

	for _, tc := range acctsTestCases {
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
