package specifications

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func ExecuteDBInsertSpecification(ctx context.Context, t *testing.T, execute ExecuteQueryDsl) {
	if execute.SupportBatch() {
		ExecuteDBBatchInsertSpecification(ctx, t, execute)
	} else {
		ExecuteDBSingleInsertSpecification(ctx, t, execute)
	}
}

// ExecuteDBSingleInsertSpecification is a specification for database insert, it test
// related table inserts, it will insert 1 user and 1 post
func ExecuteDBSingleInsertSpecification(ctx context.Context, t *testing.T, execute ExecuteQueryDsl) {
	t.Logf("Executing insert action specification")
	// Given a valid database schema
	db := SchemaLoader.Load(t, SchemaTestDB)
	dbID := execute.DBID(db.Name)

	// When i execute action to database
	user1 := userTable{
		ID:       1111,
		UserName: "test_user",
		Age:      22,
	}

	createUserActionInput := []any{user1.ID, user1.UserName, user1.Age}

	txHash, err := execute.Execute(ctx, dbID, createUserActionName, createUserActionInput)
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
	}

	txHash, err = execute.Execute(ctx, dbID, createPostQueryName, post1...)
	assert.NoError(t, err)

	expectTxSuccess(t, execute, ctx, txHash, defaultTxQueryTimeout)()

	records, err = execute.QueryDatabase(ctx, dbID, "SELECT * FROM posts")
	assert.NoError(t, err)
	assert.NotNil(t, records)

	counter := len(records.Export())

	assert.EqualValues(t, 1, counter)

	// TODO: move to a new specification
	//multiStmtActionName := "multi_select"
	//// execute multi statement action
	//txHash, err = execute.Execute(ctx, dbID, multiStmtActionName, nil)
	//assert.NoError(t, err)
	//
	//expectTxSuccess(t, execute, ctx, txHash, defaultTxQueryTimeout)()
}

// ExecuteDBBatchInsertSpecification is a specification for database batch insert,
// it will insert 1 user and 2 posts
func ExecuteDBBatchInsertSpecification(ctx context.Context, t *testing.T, execute ExecuteQueryDsl) {
	t.Logf("Executing batch insert action specification")
	// Given a valid database schema
	db := SchemaLoader.Load(t, SchemaTestDB)
	dbID := execute.DBID(db.Name)

	// When i execute action to database

	user1 := userTable{
		ID:       1111,
		UserName: "test_user",
		Age:      22,
	}

	createUserActionInput := []any{user1.ID, user1.UserName, user1.Age}

	txHash, err := execute.Execute(ctx, dbID, createUserActionName, createUserActionInput)
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

	txHash, err = execute.Execute(ctx, dbID, createPostQueryName, post1...)
	assert.NoError(t, err)

	expectTxSuccess(t, execute, ctx, txHash, defaultTxQueryTimeout)()

	records, err = execute.QueryDatabase(ctx, dbID, "SELECT * FROM posts")
	assert.NoError(t, err)
	require.NotNil(t, records)

	counter := len(records.Export())

	assert.EqualValues(t, 2, counter)
}

func ExecuteDBRecordsVerifySpecification(ctx context.Context, t *testing.T, execute ExecuteQueryDsl, numRecords int) {
	t.Logf("Executing verify db records specification")
	// Given a valid database schema
	db := SchemaLoader.Load(t, SchemaTestDB)
	dbID := execute.DBID(db.Name)

	if execute.SupportBatch() {
		numRecords = numRecords * 2
	}

	records, err := execute.QueryDatabase(ctx, dbID, "SELECT * FROM posts")
	assert.NoError(t, err)
	require.NotNil(t, records)

	counter := 0
	for records.Next() {
		_ = records.Record()
		counter++
	}
	assert.EqualValues(t, numRecords, counter)
}

func ExecuteDBRecordsVerifySpecificationEventually(ctx context.Context, t *testing.T, execute ExecuteQueryDsl, numRecords int) {
	t.Logf("Executing verify db records eventually specification")
	// Given a valid database schema
	db := SchemaLoader.Load(t, SchemaTestDB)
	dbID := execute.DBID(db.Name)

	if execute.SupportBatch() {
		numRecords = numRecords * 2
	}

	require.Eventually(t, func() bool {
		records, err := execute.QueryDatabase(ctx, dbID, "SELECT * FROM posts")
		assert.NoError(t, err)
		require.NotNil(t, records)

		counter := 0
		for records.Next() {
			_ = records.Record()
			counter++
		}
		return counter == numRecords
	}, 1*time.Minute, 500*time.Millisecond)
}
