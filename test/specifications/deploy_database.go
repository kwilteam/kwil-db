package specifications

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kwilteam/kwil-db/pkg/transactions"

	"github.com/stretchr/testify/assert"
)

// DatabaseDeployDsl is dsl for database deployment specification
type DatabaseDeployDsl interface {
	DatabaseIdentifier
	DatabaseExister
	DeployDatabase(ctx context.Context, db *transactions.Schema) (txHash []byte, err error)
	TxSuccess(ctx context.Context, txHash []byte) error
}

// TODO: a better way to do this would be to retrieve the deployed schema structure and do a deep comparison
func DatabaseDeploySpecification(ctx context.Context, t *testing.T, deploy DatabaseDeployDsl) {
	t.Logf("Executing database deploy specification")

	// Given a valid database schema
	db := SchemaLoader.Load(t, schemaTestDB)

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
	// When i deploy faulty database
	_, err = deploy.DeployDatabase(ctx, db2)
	require.NoError(t, err, "failed to send deploy database tx")

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
	assert.Error(t, err)
}

func DatabaseVerifySpecification(ctx context.Context, t *testing.T, deploy DatabaseDeployDsl, exisits bool) {
	t.Logf("Executing database verify specification")

	// Given a valid database schema
	db := SchemaLoader.Load(t, schemaTestDB)

	// And i expect database should exist
	err := deploy.DatabaseExists(ctx, deploy.DBID(db.Name))
	if exisits {
		assert.NoError(t, err)
	} else {
		assert.Error(t, err)
	}
}
