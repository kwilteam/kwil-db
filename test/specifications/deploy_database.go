package specifications

import (
	"context"
	"github.com/kwilteam/kwil-db/pkg/transactions"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// DatabaseDeployDsl is dsl for database deployment specification
type DatabaseDeployDsl interface {
	DeployDatabase(ctx context.Context, db *transactions.Schema) error
	// TODO: verify more than just existence, check schema structure
	DatabaseShouldExists(ctx context.Context, owner string, dbName string) error
}

func DatabaseDeploySpecification(ctx context.Context, t *testing.T, deploy DatabaseDeployDsl) {
	t.Logf("Executing database deploy specification")

	// Given a valid database schema
	db := SchemaLoader.Load(t, schema_testdb)

	// When i deploy the database
	err := deploy.DeployDatabase(ctx, db)

	// Then i expect success
	assert.NoError(t, err, "failed to deploy database")

	// And i expect database should exist
	err = deploy.DatabaseShouldExists(ctx, db.Owner, db.Name)
	assert.NoError(t, err)

}

func DatabaseDeployInvalidSqlSpecification(ctx context.Context, t *testing.T, deploy DatabaseDeployDsl) {
	t.Logf("Executing database deploy invalid SQL specification")

	db := SchemaLoader.LoadWithoutValidation(t, schema_invalidSQLSyntax)

	// Deploy faulty database and expect error
	deploy.DeployDatabase(ctx, db)
	// assert.Error(t, err)
	time.Sleep(5 * time.Second)
	// And i expect database should not exist
	err := deploy.DatabaseShouldExists(ctx, db.Owner, db.Name)
	assert.Error(t, err)

	// read in fixed schema
	db = SchemaLoader.Load(t, schema_invalidSQLSyntaxFixed)

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

	db := SchemaLoader.Load(t, schema_invalidExtensionInit)

	// Deploy faulty database and expect error
	err := deploy.DeployDatabase(ctx, db)
	assert.Error(t, err)

	// And i expect database should not exist
	err = deploy.DatabaseShouldExists(ctx, db.Owner, db.Name)
	assert.Error(t, err)
}

func DatabaseVerifySpecification(ctx context.Context, t *testing.T, deploy DatabaseDeployDsl, exisits bool) {
	t.Logf("Executing database verify specification")
	// Given a valid database schema
	db := SchemaLoader.Load(t, schema_testdb)

	// And i expect database should exist
	err := deploy.DatabaseShouldExists(ctx, db.Owner, db.Name)
	if exisits {
		assert.NoError(t, err)
	} else {
		assert.Error(t, err)
	}
}
