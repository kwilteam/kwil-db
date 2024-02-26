package specifications

import (
	"context"
	"testing"

	"github.com/kwilteam/kwil-db/pkg/client"

	kTx "github.com/kwilteam/kwil-db/pkg/tx"

	"github.com/cstockton/go-conv"
	"github.com/stretchr/testify/assert"
)

const (
	createUserActionName = "create_user"
	listUsersActionName  = "list_users"
)

type userTable struct {
	ID       int32  `json:"id"`
	UserName string `json:"username"`
	Age      int32  `json:"age"`
}

type ExecuteQueryDsl interface {
	// ExecuteAction executes QUERY to a database
	ExecuteAction(ctx context.Context, dbid string, actionName string, actionInputs []map[string]any) (*kTx.Receipt, []map[string]any, error)
	QueryDatabase(ctx context.Context, dbid, query string) (*client.Records, error)
}

func ExecuteDBInsertSpecification(ctx context.Context, t *testing.T, execute ExecuteQueryDsl) {
	t.Logf("Executing insert action specification")
	// Given a valid database schema
	db := SchemaLoader.Load(t, schema_testdb)
	dbID := GenerateSchemaId(db.Owner, db.Name)

	// When i execute action to database

	user1 := userTable{
		ID:       1111,
		UserName: "test_user",
		Age:      22,
	}

	createUserActionInput := []map[string]any{
		{
			"$id":       user1.ID,
			"$username": user1.UserName,
			"$age":      user1.Age,
		},
	}
	_, _, err := execute.ExecuteAction(ctx, dbID, createUserActionName, createUserActionInput)
	assert.NoError(t, err)

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

	assert.EqualValues(t, user1.ID, user1Id)
	assert.EqualValues(t, user1.UserName, user1Username)
	assert.EqualValues(t, user1.Age, user1Age)

	// testing query database
	records, err := execute.QueryDatabase(ctx, dbID, "SELECT * FROM users")
	assert.NoError(t, err)
	assert.NotNil(t, records)

	// create post
	const createPostQueryName = "create_post"
	post1 := []map[string]any{
		{
			"$id":      1111,
			"$title":   "test_post",
			"$content": "test_body",
		},
		{
			"$id":      2222,
			"$title":   "test_post2",
			"$content": "test_body2",
		},
	}

	_, _, err = execute.ExecuteAction(ctx, dbID, createPostQueryName, post1)
	assert.NoError(t, err)

	records, err = execute.QueryDatabase(ctx, dbID, "SELECT * FROM posts")
	assert.NoError(t, err)
	assert.NotNil(t, records)

	counter := 0
	for records.Next() {
		_ = records.Record()
		counter++
	}

	assert.EqualValues(t, 2, counter)

	// insert more
	post2 := []map[string]any{
		{
			"$id":      3333,
			"$title":   "test_post3",
			"$content": "test_body3",
		},
		{
			"$id":      4444,
			"$title":   "test_post4",
			"$content": "test_body4",
		},
	}

	_, _, err = execute.ExecuteAction(ctx, dbID, createPostQueryName, post2)
	assert.NoError(t, err)

	records, err = execute.QueryDatabase(ctx, dbID, "SELECT * FROM posts")
	assert.NoError(t, err)
	assert.NotNil(t, records)

	counter = 0
	for records.Next() {
		_ = records.Record()
		counter++
	}

	assert.EqualValues(t, 4, counter)

	multiStmtActionName := "multi_select"
	// execute multi statement action
	_, res, err := execute.ExecuteAction(ctx, dbID, multiStmtActionName, nil)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	userRow1 := res[0]
	// users has age, posts does not, but has content
	_, ok := userRow1["age"]
	assert.True(t, ok)

	_, ok = userRow1["content"]
	assert.False(t, ok)
}
