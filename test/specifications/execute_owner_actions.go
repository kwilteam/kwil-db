package specifications

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	ownerOnlyActionName = "owner_only"
)

func ExecuteOwnerActionSpecification(ctx context.Context, t *testing.T, execute ExecuteQueryDsl) {
	t.Logf("Executing owner action specification")

	db := SchemaLoader.Load(t, schema_testdb)
	dbID := GenerateSchemaId(db.Owner, db.Name)

	actionInputs := []any{}
	_, err := execute.ExecuteAction(ctx, dbID, ownerOnlyActionName, actionInputs)
	assert.NoError(t, err, "error executing owner action")
}

func ExecuteOwnerActionFailSpecification(ctx context.Context, t *testing.T, execute ExecuteQueryDsl) {
	t.Logf("Executing owner action fail specification")

	db := SchemaLoader.Load(t, schema_testdb)
	dbID := GenerateSchemaId(db.Owner, db.Name)

	actionInputs := []any{}

	_, err := execute.ExecuteAction(ctx, dbID, ownerOnlyActionName, actionInputs)
	assert.Error(t, err, "error executing owner action")
}
