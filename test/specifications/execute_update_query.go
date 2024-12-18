package specifications

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ExecuteDBUpdateSpecification(ctx context.Context, t *testing.T, execute ExecuteQueryDsl) {
	t.Logf("Executing update action specification")
	// Given a valid database schema
	db := SchemaLoader.Load(t, SchemaTestDB)
	dbID := execute.DBID(db.Name)
	actionName := "update_user"
	userQ := userTable{
		ID:       2222,
		UserName: "test_user_update",
		Age:      22,
	}

	actionInput := [][]any{
		{userQ.ID, userQ.UserName, userQ.Age},
	}

	// When i execute action to database
	txHash, err := execute.Execute(ctx, dbID, actionName, actionInput...)
	assert.NoError(t, err)

	expectTxSuccess(t, execute, ctx, txHash, defaultTxQueryTimeout)()

	// Then i expect row to be updated
	receipt, err := execute.QueryDatabase(ctx, dbID, "SELECT * FROM users WHERE id = 2222")
	assert.NoError(t, err)
	assert.NotNil(t, receipt)

	results := receipt.Export()
	if len(results) != 1 {
		t.Errorf("expected 1 statement result, got %d", len(results))
	}

	returnedUser1 := results[0]

	user1Id, ok := returnedUser1["id"].(int64)
	require.Truef(t, ok, "expected a int64, got a %T", returnedUser1["id"])

	user1Username, ok := returnedUser1["username"].(string)
	require.Truef(t, ok, "expected a string, got a %T", returnedUser1["username"])

	user1Age, ok := returnedUser1["age"].(int64)
	require.Truef(t, ok, "expected a int64, got a %T", returnedUser1["age"])

	assert.EqualValues(t, userQ.ID, user1Id)
	assert.EqualValues(t, userQ.UserName, user1Username)
	assert.EqualValues(t, userQ.Age, user1Age)

	// check foreign key was updated properly from the previous UPDATE
	receipt, err = execute.QueryDatabase(ctx, dbID, "SELECT title, content FROM posts WHERE user_id = 2222;")
	assert.NoError(t, err)
	assert.NotNil(t, receipt)

	results = receipt.Export()
	length1 := len(results)
	assert.NotZero(t, length1, "user should have more than 0 posts")

	receipt, err = execute.QueryDatabase(ctx, dbID, `SELECT title, content
    FROM posts
    WHERE user_id = (
        SELECT id
        FROM users
        WHERE username = 'test_user_update'
    );`)
	assert.NoError(t, err)
	assert.NotNil(t, receipt)

	results = receipt.Export()

	assert.NotZero(t, len(results), "user should have more than 0 posts")
	assert.Equal(t, length1, len(results), "user should have same number of posts after username update")

	// TODO: get result
	//assert.NotZero(t, len(results), "should get user's posts after user_id updated")

	//getUserPostsActionName := "get_user_posts"
	//actionInput = []map[string]any{
	//	{"$username": userQ.UserName},
	//}
	//receipt, results, err = execute.Execute(ctx, dbID, getUserPostsActionName, actionInput)
	//assert.NoError(t, err)
	//assert.NotNil(t, receipt)
	//assert.NotZero(t, len(results), "should get user's posts after user_id updated")
}
