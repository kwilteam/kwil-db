package specifications

import (
	"context"
	"kwil/x/types/databases"
	"testing"

	"github.com/stretchr/testify/assert"
)

// DatabaseDeployDsl is dsl for database deployment specification
type DatabaseDeployDsl interface {
	DeployDatabase(ctx context.Context, db *databases.Database[[]byte]) error
	DatabaseShouldExists(ctx context.Context, owner string, dbName string) error
}

func DatabaseDeploySpecification(t *testing.T, ctx context.Context, deploy DatabaseDeployDsl) {
	//Given a valid database schema
	db := SchemaLoader.Load(t)

	//When i deploy the database
	err := deploy.DeployDatabase(ctx, db)

	//Then i expect success
	assert.NoError(t, err)

	//And i expect database should exist
	err = deploy.DatabaseShouldExists(ctx, db.Owner, db.Name)
	assert.NoError(t, err)
}
