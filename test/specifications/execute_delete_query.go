package specifications

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func ExecuteDBDeleteSpecification(ctx context.Context, t *testing.T, execute ExecuteQueryDsl) {
	t.Logf("Executing delete action specification")
	// Given a valid database schema
	db := SchemaLoader.Load(t, schemaTestDB)
	dbID := GenerateSchemaId(db.Owner, db.Name)

	actionName := "delete_user"
	actionInput := []any{}

	//// get user id
	receipt, err := execute.ExecuteAction(ctx, dbID, listUsersActionName, nil)
	assert.NoError(t, err)
	assert.NotNil(t, receipt)
	// TODO: get result
	//if len(results) != 1 {
	//	t.Errorf("expected 1 statement result, got %d", len(results))
	//}
	//returnedUser1 := results[0]
	//
	//user1Id, _ := conv.Int32(returnedUser1["id"])

	// When i execute query to database
	_, err = execute.ExecuteAction(ctx, dbID, actionName, actionInput)
	assert.NoError(t, err)

	// Then i expect row to be deleted
	receipt, err = execute.ExecuteAction(ctx, dbID, listUsersActionName, nil)
	assert.NoError(t, err)
	assert.NotNil(t, receipt)

	//if len(results) != 0 {
	//	t.Errorf("expected 0 user statement result, got %d", len(results))
	//}

	////// check foreign key constraint
	//getUserPostsByUserIdActionName := "get_user_posts_by_userid"
	//actionInput = []map[string]any{
	//	{"$id": user1Id},
	//}
	//actionInput = []any{
	//	[]any{user1Id},
	//}
	//receipt, err = execute.ExecuteAction(ctx, dbID, getUserPostsByUserIdActionName, actionInput)
	//assert.NoError(t, err)
	//assert.NotNil(t, receipt)
	//assert.Zerof(t, len(results), "user's posts should be deleted after user is deleted")
}
