package specifications

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// DatabaseDropDsl is dsl for database drop specification
type DatabaseDropDsl interface {
	DropDatabase(ctx context.Context, dbName string) error
	DatabaseShouldExists(ctx context.Context, owner string, dbName string) error
}

func DatabaseDropSpecification(ctx context.Context, t *testing.T, drop DatabaseDropDsl) {
	t.Logf("Executing database drop specification")
	// Given a valid database schema
	db := SchemaLoader.Load(t)

	// When i drop the database, it should success
	err := drop.DropDatabase(ctx, db.Name)
	assert.NoError(t, err)

	// Drop again should fail
	err = drop.DropDatabase(ctx, db.Name)
	assert.Error(t, err)

	// And i expect database should not exist
	err = drop.DatabaseShouldExists(ctx, db.Owner, db.Name)
	assert.Error(t, err)
}
