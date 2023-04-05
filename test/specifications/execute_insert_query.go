package specifications

import (
	"context"
	"encoding/json"
	"kwil/pkg/databases"
	"testing"

	kTx "kwil/pkg/tx"

	"github.com/stretchr/testify/assert"
)

const listUsersActionName = "list_users"

type userTable struct {
	ID       int32  `json:"id"`
	UserName string `json:"username"`
	Age      int32  `json:"age"`
}

type hasuraTable map[string][]userTable
type hasuraResp map[string]hasuraTable

type ExecuteQueryDsl interface {
	// ExecuteAction executes QUERY to a database
	// @yaiba TODO: owner is not needed?? because user can only execute queries using his private key
	ExecuteAction(ctx context.Context, dbid string, queryName string, queryInputs []map[string]any) (*kTx.Receipt, error)
	QueryDatabase(ctx context.Context, query string) ([]byte, error)
}

func ExecuteDBInsertSpecification(ctx context.Context, t *testing.T, execute ExecuteQueryDsl) {
	t.Logf("Executing insert query specification")
	// Given a valid database schema
	db := SchemaLoader.Load(t)
	dbID := databases.GenerateSchemaId(db.Owner, db.Name)

	userQueryName := "create_user"
	user1 := userTable{
		ID:       1111,
		UserName: "test_user",
		Age:      22,
	}

	user2 := userTable{
		ID:       2222,
		UserName: "test_user2",
		Age:      33,
	}

	/*
			{"id": userQ.ID},
		{"username": userQ.UserName},
		{"age": userQ.Age},
		{"address": userQ.Wallet},
	*/

	userQueryInput := []map[string]any{
		{
			"$id":       user1.ID,
			"$username": user1.UserName,
			"$age":      user1.Age,
		},
		{
			"$id":       user2.ID,
			"$username": user2.UserName,
			"$age":      user2.Age,
		},
	}

	// TODO test insert post table
	// When i execute query to database
	_, err := execute.ExecuteAction(ctx, dbID, userQueryName, userQueryInput)
	assert.NoError(t, err)

	res, err := execute.ExecuteAction(ctx, dbID, listUsersActionName, nil)
	assert.NoError(t, err)

	var results []map[string]any
	err = json.Unmarshal(res.Body, &results)
	assert.NoError(t, err)

	returnedUser1 := results[0]
	user1Id := returnedUser1["id"].(int32)
	user1Username := returnedUser1["username"].(string)
	user1Age := returnedUser1["age"].(int32)

	assert.EqualValues(t, user1.ID, user1Id)
	assert.EqualValues(t, user1.UserName, user1Username)
	assert.EqualValues(t, user1.Age, user1Age)

	returnedUser2 := results[1]
	user2Id := returnedUser2["id"].(int32)
	user2Username := returnedUser2["username"].(string)
	user2Age := returnedUser2["age"].(int32)

	assert.EqualValues(t, user2.ID, user2Id)
	assert.EqualValues(t, user2.UserName, user2Username)
	assert.EqualValues(t, user2.Age, user2Age)

}
