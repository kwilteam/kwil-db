package specifications

import (
	"context"
	"github.com/kwilteam/kwil-db/pkg/engine/models"
	"github.com/stretchr/testify/assert"
	"testing"
)

func ExecuteDBDeleteSpecification(ctx context.Context, t *testing.T, execute ExecuteQueryDsl) {
	t.Logf("Executing delete action specification")
	// Given a valid database schema
	db := SchemaLoader.Load(t)
	dbID := models.GenerateSchemaId(db.Owner, db.Name)

	actionName := "delete_user"
	actionInput := []map[string]any{}

	// When i execute query to database
	_, _, err := execute.ExecuteAction(ctx, dbID, actionName, actionInput)
	assert.NoError(t, err)

	// Then i expect row to be deleted
	receipt, results, err := execute.ExecuteAction(ctx, dbID, listUsersActionName, nil)
	assert.NoError(t, err)
	assert.NotNil(t, receipt)

	if len(results) != 0 {
		t.Errorf("expected 0 statement result, got %d", len(results))
	}
}
