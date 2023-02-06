package specifications

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

// DatabaseDropDsl is dsl for database drop specification
type DatabaseDropDsl interface {
	DropDatabase(ctx context.Context, dbName string) error
	DatabaseShouldExists(ctx context.Context, owner string, dbName string) error
}

func DatabaseDropSpecification(t *testing.T, ctx context.Context, drop DatabaseDropDsl) {
	t.Logf("Executing database drop specification")
	//Given a valid database schema
	db := SchemaLoader.Load(t)

	//When i drop the database
	err := drop.DropDatabase(ctx, db.Name)

	//Then i expect success
	assert.NoError(t, err)

	//And i expect database should exist
	err = drop.DatabaseShouldExists(ctx, db.Owner, db.Name)
	assert.Error(t, err)
}
