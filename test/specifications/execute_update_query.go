package specifications

import (
	"context"
	"testing"

	"github.com/cstockton/go-conv"

	"github.com/stretchr/testify/assert"
)

func ExecuteDBUpdateSpecification(ctx context.Context, t *testing.T, execute ExecuteQueryDsl) {
	t.Logf("Executing update action specification")
	// Given a valid database schema
	db := SchemaLoader.Load(t)
	dbID := GenerateSchemaId(db.Owner, db.Name)
	actionName := "update_username"
	userQ := userTable{
		ID:       1111,
		UserName: "test_user_update",
		Age:      22,
	}
	actionInput := []map[string]any{
		{"$username": userQ.UserName},
	}

	// When i execute action to database
	_, _, err := execute.ExecuteAction(ctx, dbID, actionName, actionInput)
	assert.NoError(t, err)

	// Then i expect row to be updated
	receipt, results, err := execute.ExecuteAction(ctx, dbID, listUsersActionName, nil)
	assert.NoError(t, err)
	assert.NotNil(t, receipt)

	if len(results) != 1 {
		t.Errorf("expected 1 statement result, got %d", len(results))
	}

	returnedUser1 := results[0]

	user1Id, _ := conv.Int32(returnedUser1["id"])
	user1Username := returnedUser1["username"].(string)
	user1Age, _ := conv.Int32(returnedUser1["age"])

	assert.EqualValues(t, userQ.ID, user1Id)
	assert.EqualValues(t, userQ.UserName, user1Username)
	assert.EqualValues(t, userQ.Age, user1Age)
}
