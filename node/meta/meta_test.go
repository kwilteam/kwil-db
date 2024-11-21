package meta_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/node/meta"
	"github.com/kwilteam/kwil-db/node/pg"
)

func Test_NetworkParams(t *testing.T) {
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

	ctx := context.Background()

	db, err := pg.NewDB(ctx, cfg)
	require.NoError(t, err)
	defer db.Close()

	tx, err := db.BeginTx(ctx)
	require.NoError(t, err)
	defer tx.Rollback(ctx) // always rollback to reset the test

	err = meta.InitializeMetaStore(ctx, tx)
	require.NoError(t, err)

	// getting params without any having been stored returns
	// ErrParamsNotFound

	_, err = meta.LoadParams(ctx, tx)
	require.Equal(t, meta.ErrParamsNotFound, err)

	param := &common.NetworkParameters{
		MaxBlockSize:     1000,
		JoinExpiry:       100,
		VoteExpiry:       100,
		DisabledGasCosts: true,
		MaxVotesPerTx:    100,
	}

	err = meta.StoreParams(ctx, tx, param)
	require.NoError(t, err)

	param2, err := meta.LoadParams(ctx, tx)
	require.NoError(t, err)

	require.EqualValues(t, param, param2)

	// update some params and perform a diff
	param2.MaxBlockSize = 2000
	param2.JoinExpiry = 200
	param2.DisabledGasCosts = false
	param2.MigrationStatus = types.NoActiveMigration

	err = meta.StoreDiff(ctx, tx, param, param2)
	require.NoError(t, err)

	param3, err := meta.LoadParams(ctx, tx)
	require.NoError(t, err)

	require.EqualValues(t, param2, param3)
}
