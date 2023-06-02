package specifications

import (
	"context"
	"testing"

	"github.com/kwilteam/kwil-db/pkg/kuneiform/schema"

	"github.com/stretchr/testify/assert"
)

// DatabaseDeployDsl is dsl for database deployment specification
type DatabaseDeployDsl interface {
	DeployDatabase(ctx context.Context, db *schema.Schema) error
	DatabaseShouldExists(ctx context.Context, owner string, dbName string) error
}

func DatabaseDeploySpecification(ctx context.Context, t *testing.T, deploy DatabaseDeployDsl) {
	t.Logf("Executing database deploy specification")
	testDeployFailure(ctx, t, deploy)

	// Given a valid database schema
	db := SchemaLoader.Load(t, schema_testdb)

	// When i deploy the database
	err := deploy.DeployDatabase(ctx, db)

	// Then i expect success
	assert.NoError(t, err)

	// And i expect database should exist
	err = deploy.DatabaseShouldExists(ctx, db.Owner, db.Name)
	assert.NoError(t, err)

}

func testDeployFailure(ctx context.Context, t *testing.T, deploy DatabaseDeployDsl) {
	db := SchemaLoader.LoadWithoutValidation(t, schema_invalidSQLSyntax)

	// Deploy faulty database and expect error
	err := deploy.DeployDatabase(ctx, db)
	assert.Error(t, err)

	// And i expect database should not exist
	err = deploy.DatabaseShouldExists(ctx, db.Owner, db.Name)
	assert.Error(t, err)

	// read in fixed schema
	db = SchemaLoader.Load(t, schema_invalidSQLSyntaxFixed)

	// Deploy fault database and expect error
	err = deploy.DeployDatabase(ctx, db)
	assert.NoError(t, err)

	// And i expect database should exist
	err = deploy.DatabaseShouldExists(ctx, db.Owner, db.Name)
	assert.NoError(t, err)
}
