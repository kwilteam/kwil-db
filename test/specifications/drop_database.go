package specifications

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// DatabaseDropDsl is dsl for database drop specification
type DatabaseDropDsl interface {
	TxQueryDsl
	DropDatabase(ctx context.Context, dbName string) (txHash []byte, err error)
	DatabaseShouldExists(ctx context.Context, owner string, dbName string) error
}

func DatabaseDropSpecification(ctx context.Context, t *testing.T, drop DatabaseDropDsl) {
	t.Logf("Executing database drop specification")
	// Given a valid database schema
	db := SchemaLoader.Load(t, schemaTestDB)

	// When i drop the database
	txHash, err := drop.DropDatabase(ctx, db.Name)
	require.NoError(t, err, "failed to send drop database tx")

	// Then i expect success
	expectTxSuccess(t, drop, ctx, txHash, defaultTxQueryTimeout)()

	// And i expect database should not exist
	err = drop.DatabaseShouldExists(ctx, db.Owner, db.Name)
	assert.Error(t, err)

	// Drop again
	txHash, err = drop.DropDatabase(ctx, db.Name)
	require.NoError(t, err, "failed to send drop database tx")

	// Then i expect tx failure
	expectTxFail(t, drop, ctx, txHash, defaultTxQueryTimeout)()
}
