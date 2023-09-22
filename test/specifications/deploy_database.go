package specifications

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

// TODO: a better way to do this would be to retrieve the deployed schema structure and do a deep comparison
func DatabaseDeploySpecification(ctx context.Context, t *testing.T, deploy DatabaseDeployDsl) {
	t.Logf("Executing database deploy specification")

	// Given a valid database schema
	db := SchemaLoader.Load(t, SchemaTestDB)

	// When i deploy the database
	txHash, err := deploy.DeployDatabase(ctx, db)
	require.NoError(t, err, "failed to send deploy database tx")

	// Then i expect success
	expectTxSuccess(t, deploy, ctx, txHash, defaultTxQueryTimeout)()

	// And i expect database should exist
	err = deploy.DatabaseExists(ctx, deploy.DBID(db.Name))
	require.NoError(t, err)
}

// DatabaseDeployInvalidSql1Specification tests invalid SQL1 syntax, Kuneiform parser will fail for SQL1 syntax
func DatabaseDeployInvalidSql1Specification(ctx context.Context, t *testing.T, deploy DatabaseDeployDsl) {
	t.Logf("Executing database deploy invalid SQL1 specification")

	// Given an invalid database schema
	db := SchemaLoader.LoadWithoutValidation(t, schemaInvalidSqlSyntax)

	// When i deploy faulty database
	txHash, err := deploy.DeployDatabase(ctx, db)
	require.NoError(t, err, "failed to send deploy database tx")

	// Then i expect tx failure
	expectTxFail(t, deploy, ctx, txHash, defaultTxQueryTimeout)()

	// read in fixed schema
	db2 := SchemaLoader.Load(t, schemaInvalidSqlSyntaxFixed)
	_, err = deploy.DeployDatabase(ctx, db2)
	require.NoError(t, err, "failed to send deploy database tx")

	expectTxSuccess(t, deploy, ctx, txHash, defaultTxQueryTimeout)()

	err = deploy.DatabaseExists(ctx, deploy.DBID(db.Name))
	require.NoError(t, err)
}

func DatabaseDeployInvalidExtensionSpecification(ctx context.Context, t *testing.T, deploy DatabaseDeployDsl) {
	t.Logf("Executing database deploy invalid Extension init specification")

	db := SchemaLoader.Load(t, schemaInvalidExtensionInit)

	// When i deploy faulty database
	txHash, err := deploy.DeployDatabase(ctx, db)
	require.NoError(t, err, "failed to send deploy database tx")

	// Then i expect tx failure
	expectTxFail(t, deploy, ctx, txHash, defaultTxQueryTimeout)()

	// And i expect database should not exist
	err = deploy.DatabaseExists(ctx, deploy.DBID(db.Name))
	require.Error(t, err)
}

func DatabaseVerifySpecification(ctx context.Context, t *testing.T, deploy DatabaseDeployDsl, exists bool) {
	t.Logf("Executing database verify specification")

	// Given a valid database schema
	db := SchemaLoader.Load(t, SchemaTestDB)

	// And i expect database should exist
	err := deploy.DatabaseExists(ctx, deploy.DBID(db.Name))
	if exists {
		require.NoError(t, err)
	} else {
		require.Error(t, err)
	}
}
