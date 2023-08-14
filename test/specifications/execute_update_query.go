package specifications

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func ExecuteDBUpdateSpecification(ctx context.Context, t *testing.T, execute ExecuteQueryDsl) {
	t.Logf("Executing update action specification")
	// Given a valid database schema
	db := SchemaLoader.Load(t, schema_testdb)
	dbID := GenerateSchemaId(db.Owner, db.Name)
	actionName := "update_user"
	userQ := userTable{
		ID:       2222,
		UserName: "test_user_update",
		Age:      22,
	}

	actionInput := []any{
		[]any{userQ.ID, userQ.UserName, userQ.Age},
	}

	// When i execute action to database
	_, err := execute.ExecuteAction(ctx, dbID, actionName, actionInput)
	assert.NoError(t, err)

	// Then i expect row to be updated
	receipt, err := execute.ExecuteAction(ctx, dbID, listUsersActionName, nil)
	assert.NoError(t, err)
	assert.NotNil(t, receipt)

	// TODO: get result
	//if len(results) != 1 {
	//	t.Errorf("expected 1 statement result, got %d", len(results))
	//}
	//fmt.Println(results)
	//returnedUser1 := results[0]
	//
	//user1Id, _ := conv.Int32(returnedUser1["id"])
	//user1Username := returnedUser1["username"].(string)
	//user1Age, _ := conv.Int32(returnedUser1["age"])
	//
	//assert.EqualValues(t, userQ.ID, user1Id)
	//assert.EqualValues(t, userQ.UserName, user1Username)
	//assert.EqualValues(t, userQ.Age, user1Age)

	// TODO: delete
	records, err := execute.QueryDatabase(ctx, dbID, "SELECT * FROM posts")
	assert.NoError(t, err)
	assert.NotZero(t, len(records.Export()), "should get user's posts before user_id updated")
	// TODO: undelete

	//// check foreign key constraint
	getUserPostsByUserIdActionName := "get_user_posts_by_userid"
	actionInput = []any{
		[]any{userQ.ID},
	}
	receipt, err = execute.ExecuteAction(ctx, dbID, getUserPostsByUserIdActionName, actionInput)
	assert.NoError(t, err)
	assert.NotNil(t, receipt)

	// TODO: get result
	//assert.NotZero(t, len(results), "should get user's posts after user_id updated")

	//getUserPostsActionName := "get_user_posts"
	//actionInput = []map[string]any{
	//	{"$username": userQ.UserName},
	//}
	//receipt, results, err = execute.ExecuteAction(ctx, dbID, getUserPostsActionName, actionInput)
	//assert.NoError(t, err)
	//assert.NotNil(t, receipt)
	//assert.NotZero(t, len(results), "should get user's posts after user_id updated")
}
