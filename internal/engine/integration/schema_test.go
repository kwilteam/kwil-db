package integration_test

import (
	"context"
	"os"
	"testing"

	"github.com/kwilteam/kwil-db/common/sql"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/internal/engine/execution"
	"github.com/kwilteam/kwil-db/kuneiform"
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

				err = global.CreateDataset(ctx, db, schema, owner)
				require.NoError(t, err)
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

	db, err := kuneiform.Parse(string(d))
	if err != nil {
		return nil, err
	}

	return db, nil
}
