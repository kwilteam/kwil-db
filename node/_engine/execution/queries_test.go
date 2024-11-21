//go:build pglive

package execution

import (
	"context"
	"strings"
	"testing"

	"github.com/kwilteam/kwil-db/core/types/testdata"
	"github.com/kwilteam/kwil-db/internal/sql/pg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_StoringSchemas(t *testing.T) {
	ctx := context.Background()
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
		SchemaFilter: func(s string) bool {
			return strings.Contains(s, pg.DefaultSchemaFilterPrefix)
		},
	}

	db, err := pg.NewDB(ctx, cfg)
	require.NoError(t, err)

	tx, err := db.BeginTx(ctx)
	require.NoError(t, err)
	defer tx.Rollback(ctx) // we always want to rollback, never commit

	err = InitializeEngine(ctx, tx)
	require.NoError(t, err)

	err = createSchemasTableIfNotExists(ctx, tx)
	require.NoError(t, err)

	err = createSchema(ctx, tx, testdata.TestSchema, "txid")
	require.NoError(t, err)

	defer func() {
		err := deleteSchema(ctx, tx, testdata.TestSchema.DBID())
		require.NoError(t, err)
	}()

	schemas, err := getSchemas(ctx, tx, nil)
	require.NoError(t, err)

	require.Len(t, schemas, 1)

	assert.EqualValuesf(t, testdata.TestSchema, schemas[0], "expected: %v, got: %v", testdata.TestSchema, schemas[0])
}
