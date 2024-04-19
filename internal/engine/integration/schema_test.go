// go:build pglive

package integration_test

import (
	"context"
	"os"
	"testing"

	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/common/sql"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/internal/engine/execution"
	"github.com/kwilteam/kwil-db/parse"
	"github.com/stretchr/testify/require"
)

var (
	owner = []byte("test_owner")
)

// Test_Schemas is made to test full kuneiform schemas against the engine.
// The intent of this is to test the full engine, with expected error messages,
// without having to write a full integration test.
func Test_Schemas(t *testing.T) {
	type testCase struct {
		name string
		// fn is the test function
		// the passed db will be in a transaction
		fn func(t *testing.T, global *execution.GlobalContext, db sql.DB, readonly sql.DB)
	}

	testCases := []testCase{
		{
			name: "simple schema",
			fn: func(t *testing.T, global *execution.GlobalContext, db sql.DB, readonly sql.DB) {
				ctx := context.Background()
				schema, err := loadSchema("users.kf")
				require.NoError(t, err)

				err = global.CreateDataset(ctx, db, schema, &common.TransactionData{
					Signer: owner,
					Caller: string(owner),
					TxID:   "test",
				})
				require.NoError(t, err)
				datasets, err := global.ListDatasets(owner)
				require.NoError(t, err)
				require.Len(t, datasets, 1)

				// create user
				_, err = global.Procedure(ctx, db, &common.ExecutionData{
					Dataset:   datasets[0].DBID,
					Procedure: "create_user",
					Args:      []any{"satoshi", "42"},
					TransactionData: common.TransactionData{
						Signer: owner,
						Caller: string(owner),
						TxID:   "1",
					},
				})
				require.NoError(t, err)

				// make a post
				_, err = global.Procedure(ctx, db, &common.ExecutionData{
					Dataset:   datasets[0].DBID,
					Procedure: "create_post",
					Args:      []any{"hello world"},
					TransactionData: common.TransactionData{
						Signer: owner,
						Caller: string(owner),
						TxID:   "2",
					},
				})

				res, err := global.Procedure(ctx, db, &common.ExecutionData{
					Dataset:   datasets[0].DBID,
					Procedure: "get_user",
					Args:      []any{"satoshi"},
					TransactionData: common.TransactionData{
						Signer: owner,
						Caller: string(owner),
						TxID:   "1",
					},
				})
				require.NoError(t, err)

				require.Len(t, res.Rows, 1)

				// check the columns
				require.Len(t, res.Columns, 4)
				// should be id, age, address, post_count
				require.Equal(t, "id", res.Columns[0])
				require.Equal(t, "age", res.Columns[1])
				require.Equal(t, "address", res.Columns[2])
				require.Equal(t, "post_count", res.Columns[3])

				// check the values
				// we will simply check row 0 is some uuid
				_, ok := res.Rows[0][0].([16]byte)
				require.True(t, ok)
				require.Equal(t, int64(42), res.Rows[0][1])
				require.Equal(t, string(owner), res.Rows[0][2])
				require.Equal(t, int64(1), res.Rows[0][3])
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			global, db, err := setup(t)
			if err != nil {
				t.Fatal(err)
			}
			defer cleanup(t, db)

			ctx := context.Background()

			tx, err := db.BeginOuterTx(ctx)
			require.NoError(t, err)
			defer tx.Rollback(ctx)

			readonly, err := db.BeginReadTx(ctx)
			require.NoError(t, err)
			defer readonly.Rollback(ctx)

			tc.fn(t, global, tx, readonly)
		})
	}
}

// loadSchema loads a schema from the schemas directory.
func loadSchema(file string) (*types.Schema, error) {
	d, err := os.ReadFile("./schemas/" + file)
	if err != nil {
		return nil, err
	}

	db, err := parse.ParseKuneiform(string(d))
	if err != nil {
		return nil, err
	}

	return db, nil
}
