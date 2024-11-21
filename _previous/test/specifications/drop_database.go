package specifications

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func DatabaseDropSpecification(ctx context.Context, t *testing.T, drop DatabaseDropDsl) {
	t.Logf("Executing database drop specification")
	// Given a valid database schema
	db := SchemaLoader.Load(t, SchemaTestDB)

	// When i drop the database
	txHash, err := drop.DropDatabase(ctx, db.Name)
	require.NoError(t, err, "failed to send drop database tx")

	// Then i expect success
	expectTxSuccess(t, drop, ctx, txHash, defaultTxQueryTimeout)()

	// And i expect database should not exist
	err = drop.DatabaseExists(ctx, drop.DBID(db.Name))
	assert.Error(t, err)

	// Drop again
	txHash, err = drop.DropDatabase(ctx, db.Name)
	require.NoError(t, err, "failed to send drop database tx")

	// Then i expect tx failure
	expectTxFail(t, drop, ctx, txHash, defaultTxQueryTimeout)()
}
