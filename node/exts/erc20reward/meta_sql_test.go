package erc20reward

import (
	"context"
	"crypto/sha256"
	"math/big"
	"testing"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/node/engine/interpreter"
	"github.com/kwilteam/kwil-db/node/pg"
	"github.com/kwilteam/kwil-db/node/types/sql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// deterministic address generator
func genEthAddress(t string) ethcommon.Address {
	hash := sha256.Sum256([]byte(t))

	return ethcommon.BytesToAddress(hash[:20])
}

func Test_RewardStore(t *testing.T) {
	type testcase struct {
		name    string
		initial userProvidedData
		synced  syncedRewardData
		balance int64
	}

	tests := []testcase{
		{
			name: "simple",
			initial: userProvidedData{
				ID:                 types.MustParseUUID("fc2717ab-e5dd-4f42-bd70-8eac96d0d4c9"),
				ChainID:            "1",
				EscrowAddress:      genEthAddress("escrow"),
				DistributionPeriod: 1000,
			},
			synced: syncedRewardData{
				Erc20Address:  genEthAddress("erc20"),
				Erc20Decimals: 18,
			},
			balance: 100,
		},
	}

	db, err := newTestDB()
	require.NoError(t, err)
	defer db.Close()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			tx, err := db.BeginTx(ctx)
			require.NoError(t, err)
			defer tx.Rollback(ctx) // always rollback to reset the state

			app := newTestApp(t, tx)

			err = createSchema(ctx, app)
			require.NoError(t, err)

			err = createNewRewardInstance(ctx, app, &test.initial)
			require.NoError(t, err)

			// get it and ensure it is not pending
			rewards, err := getStoredRewards(ctx, app)
			require.NoError(t, err)
			require.Len(t, rewards, 1)
			rew := rewards[0]

			// ensure initial data is right, and that synced data is empty
			assert.EqualValues(t, test.initial, rew.userProvidedData)
			assert.False(t, rew.synced)
			assert.EqualValues(t, 0, rew.syncedAt)
			assert.EqualValues(t, syncedRewardData{}, rew.syncedRewardData)

			// store synced data
			err = setRewardSynced(ctx, app, test.initial.ID, []byte("blockhash"), 100, &test.synced)
			require.NoError(t, err)

			dec, err := types.NewDecimalFromBigInt(big.NewInt(test.balance), 0)
			require.NoError(t, err)
			err = dec.SetPrecisionAndScale(78, 0)
			require.NoError(t, err)

			// set balance
			err = addBalanceToReward(ctx, app, test.initial.ID, dec)
			require.NoError(t, err)

			// read it back
			rewards, err = getStoredRewards(ctx, app)
			require.NoError(t, err)
			require.Len(t, rewards, 1)
			rew = rewards[0]

			assert.EqualValues(t, test.initial, rew.userProvidedData)
			assert.EqualValues(t, test.synced, rew.syncedRewardData)
			assert.EqualValues(t, dec, rew.ownedBalance)

			assert.True(t, rew.synced)
			assert.EqualValues(t, 100, rew.syncedAt)
		})
	}
}

func newTestDB() (*pg.DB, error) {
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

	return pg.NewDB(ctx, cfg)
}

const defaultCaller = "owner"

func newTestApp(t *testing.T, tx sql.DB) *common.App {
	interp, err := interpreter.NewInterpreter(context.Background(), tx, &common.Service{}, nil, nil, nil)
	require.NoError(t, err)

	err = interp.ExecuteWithoutEngineCtx(context.Background(), tx, "TRANSFER OWNERSHIP TO $user", map[string]any{
		"user": defaultCaller,
	}, nil)
	require.NoError(t, err)

	return &common.App{
		DB:     tx,
		Engine: interp,
		Service: &common.Service{
			Logger: log.New(),
		},
	}
}
