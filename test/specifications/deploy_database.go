package specifications

import (
	"context"
	"encoding/hex"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/kwilteam/kwil-db/pkg/transactions"

	"github.com/stretchr/testify/assert"
)

// DatabaseDeployDsl is dsl for database deployment specification
type DatabaseDeployDsl interface {
	DeployDatabase(ctx context.Context, db *transactions.Schema) (txHash []byte, err error)
	// TODO: verify more than just existence, check schema structure
	DatabaseShouldExists(ctx context.Context, owner string, dbName string) error
	TxSuccess(ctx context.Context, txHash []byte) error
}

func DatabaseDeploySpecification(ctx context.Context, t *testing.T, deploy DatabaseDeployDsl) {
	t.Logf("Executing database deploy specification")

	// Given a valid database schema
	db := SchemaLoader.Load(t, schemaTestDB)

	// When i deploy the database
	//err := deploy.DeployDatabase(ctx, db)

	txHash, err := deploy.DeployDatabase(ctx, db)
	t.Logf("txHash: %s", hex.EncodeToString(txHash))

	var status strings.Builder
	require.Eventually(t, func() bool {
		// prevent appending to the prior invocation(s)
		status.Reset()
		if err := deploy.TxSuccess(ctx, txHash); err == nil {
			return true
		} else {
			status.WriteString(err.Error())
			return false
		}
	}, time.Second*15, time.Second*2, "deploy database failed: %s", status.String())

	//// Then i expect success
	require.NoError(t, err, "failed to deploy database")

	// And i expect database should exist
	err = deploy.DatabaseShouldExists(ctx, db.Owner, db.Name)
	require.NoError(t, err)
}

func DatabaseDeployInvalidSqlSpecification(ctx context.Context, t *testing.T, deploy DatabaseDeployDsl) {
	t.Logf("Executing database deploy invalid SQL specification")

	db := SchemaLoader.LoadWithoutValidation(t, schemaInvalidSqlSyntax)

	// Deploy faulty database and expect error
	deploy.DeployDatabase(ctx, db)
	// assert.Error(t, err)
	time.Sleep(5 * time.Second)
	// And i expect database should not exist
	err := deploy.DatabaseShouldExists(ctx, db.Owner, db.Name)
	assert.Error(t, err)

	// read in fixed schema
	db = SchemaLoader.Load(t, schemaInvalidSqlSyntaxFixed)

	// Deploy fault database and expect error
	deploy.DeployDatabase(ctx, db)
	// assert.NoError(t, err)
	time.Sleep(5 * time.Second)
	// And i expect database should exist

	err = deploy.DatabaseShouldExists(ctx, db.Owner, db.Name)
	assert.NoError(t, err)
}

func DatabaseDeployInvalidExtensionSpecification(ctx context.Context, t *testing.T, deploy DatabaseDeployDsl) {
	t.Logf("Executing database deploy invalid Extension init specification")

	db := SchemaLoader.Load(t, schemaInvalidExtensionInit)

	// Deploy faulty database and expect error
	txHash, err := deploy.DeployDatabase(ctx, db)

	var status strings.Builder
	require.Eventually(t, func() bool {
		// prevent appending to the prior invocation(s)
		status.Reset()
		if err := deploy.TxSuccess(ctx, txHash); err == nil {
			return true
		} else {
			status.WriteString(err.Error())
			return false
		}
	}, time.Second*5, time.Millisecond*100, "deploy database failed: %s", status.String())

	// And i expect database should not exist
	err = deploy.DatabaseShouldExists(ctx, db.Owner, db.Name)
	assert.Error(t, err)
}

func DatabaseVerifySpecification(ctx context.Context, t *testing.T, deploy DatabaseDeployDsl, exisits bool) {
	t.Logf("Executing database verify specification")
	// Given a valid database schema
	db := SchemaLoader.Load(t, schemaTestDB)

	// And i expect database should exist
	err := deploy.DatabaseShouldExists(ctx, db.Owner, db.Name)
	if exisits {
		assert.NoError(t, err)
	} else {
		assert.Error(t, err)
	}
}
