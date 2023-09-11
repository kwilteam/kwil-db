package specifications

import (
	"context"
	"testing"

	"github.com/kwilteam/kwil-db/pkg/client"
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
	DatabaseIdentifier
	TxQueryDsl
	// ExecuteAction executes QUERY to a database
	ExecuteAction(ctx context.Context, dbid string, actionName string, actionInputs ...[]any) ([]byte, error)
	QueryDatabase(ctx context.Context, dbid, query string) (*client.Records, error)
}

func ExecuteDBInsertSpecification(ctx context.Context, t *testing.T, execute ExecuteQueryDsl) {
	t.Logf("Executing insert action specification")
	// Given a valid database schema
	db := SchemaLoader.Load(t, schemaTestDB)
	dbID := execute.DBID(db.Name)

	// When i execute action to database

	user1 := userTable{
		ID:       1111,
		UserName: "test_user",
		Age:      22,
	}

	createUserActionInput := []any{user1.ID, user1.UserName, user1.Age}

	txHash, err := execute.ExecuteAction(ctx, dbID, createUserActionName, createUserActionInput)
	assert.NoError(t, err)

	expectTxSuccess(t, execute, ctx, txHash, defaultTxQueryTimeout)()

	// testing query database
	records, err := execute.QueryDatabase(ctx, dbID, "SELECT * FROM users")
	assert.NoError(t, err)
	assert.NotNil(t, records)

	// create post
	const createPostQueryName = "create_post"
	post1 := [][]any{
		{1111, "test_post", "test_body"},
		{2222, "test_post2", "test_body2"},
	}

	txHash, err = execute.ExecuteAction(ctx, dbID, createPostQueryName, post1...)
	assert.NoError(t, err)

	expectTxSuccess(t, execute, ctx, txHash, defaultTxQueryTimeout)()

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
	post2 := [][]any{
		{3333, "test_post3", "test_body3"},
		{4444, "test_post4", "test_body4"},
	}

	txHash, err = execute.ExecuteAction(ctx, dbID, createPostQueryName, post2...)
	assert.NoError(t, err)

	expectTxSuccess(t, execute, ctx, txHash, defaultTxQueryTimeout)()

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
	txHash, err = execute.ExecuteAction(ctx, dbID, multiStmtActionName, nil)
	assert.NoError(t, err)

	expectTxSuccess(t, execute, ctx, txHash, defaultTxQueryTimeout)()
}

func ExecuteDBRecordsVerifySpecification(ctx context.Context, t *testing.T, execute ExecuteQueryDsl, numRecords int) {
	t.Logf("Executing verify db records specification")
	// Given a valid database schema
	db := SchemaLoader.Load(t, schemaTestDB)
	dbID := execute.DBID(db.Name)

	records, err := execute.QueryDatabase(ctx, dbID, "SELECT * FROM posts")
	assert.NoError(t, err)
	assert.NotNil(t, records)

	counter := 0
	for records.Next() {
		_ = records.Record()
		counter++
	}
	assert.EqualValues(t, numRecords, counter)
}
