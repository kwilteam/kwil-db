package specifications

import (
	"context"
	"fmt"
	"kwil/pkg/engine/models"
	"testing"

	"github.com/stretchr/testify/assert"
)

// DatabaseDeployDsl is dsl for database deployment specification
type DatabaseDeployDsl interface {
	DeployDatabase(ctx context.Context, db *models.Dataset) error
	DatabaseShouldExists(ctx context.Context, owner string, dbName string) error
}

func DatabaseDeploySpecification(ctx context.Context, t *testing.T, deploy DatabaseDeployDsl) {
	t.Logf("Executing database deploy specification")
	// Given a valid database schema
	db := SchemaLoader.Load(t)
	dbid := db.ID()
	fmt.Println(dbid)
	// When i deploy the database
	err := deploy.DeployDatabase(ctx, db)

	// Then i expect success
	assert.NoError(t, err)

	// And i expect database should exist
	err = deploy.DatabaseShouldExists(ctx, db.Owner, db.Name)
	assert.NoError(t, err)
}
