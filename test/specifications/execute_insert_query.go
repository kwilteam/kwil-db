package specifications

import (
	"context"
	"kwil/pkg/databases"
	"testing"

	kTx "kwil/pkg/tx"

	"github.com/cstockton/go-conv"
	"github.com/stretchr/testify/assert"
)

const listUsersActionName = "list_users"

type userTable struct {
	ID       int32  `json:"id"`
	UserName string `json:"username"`
	Age      int32  `json:"age"`
}

type ExecuteQueryDsl interface {
	// ExecuteAction executes QUERY to a database
	// @yaiba TODO: owner is not needed?? because user can only execute queries using his private key
	ExecuteAction(ctx context.Context, dbid string, queryName string, queryInputs []map[string]any) (*kTx.Receipt, [][]map[string]any, error)
	QueryDatabase(ctx context.Context, dbid, query string) ([]map[string]any, error)
}

func ExecuteDBInsertSpecification(ctx context.Context, t *testing.T, execute ExecuteQueryDsl) {
	t.Logf("Executing insert query specification")
	// Given a valid database schema
	db := SchemaLoader.Load(t)
	dbID := databases.GenerateSchemaId(db.Owner, db.Name)

	createUserQueryName := "create_user"
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
	_, _, err := execute.ExecuteAction(ctx, dbID, createUserQueryName, userQueryInput)
	assert.NoError(t, err)

	receipt, results, err := execute.ExecuteAction(ctx, dbID, listUsersActionName, nil)
	assert.NoError(t, err)
	assert.NotNil(t, receipt)

	if len(results) != 1 {
		t.Errorf("expected 1 statement result, got %d", len(results))
	}

	stmt1Results := results[0]

	if len(stmt1Results) != 2 {
		t.Errorf("expected 2 rows, got %d", len(stmt1Results))
	}

	returnedUser1 := stmt1Results[0]

	user1Id, _ := conv.Int32(returnedUser1["id"])
	user1Username := returnedUser1["username"].(string)
	user1Age, _ := conv.Int32(returnedUser1["age"])

	assert.EqualValues(t, user1.ID, user1Id)
	assert.EqualValues(t, user1.UserName, user1Username)
	assert.EqualValues(t, user1.Age, user1Age)

	returnedUser2 := stmt1Results[1]
	user2Id, _ := conv.Int32(returnedUser2["id"])
	user2Username := returnedUser2["username"].(string)
	user2Age, _ := conv.Int32(returnedUser2["age"])

	assert.EqualValues(t, user2.ID, user2Id)
	assert.EqualValues(t, user2.UserName, user2Username)
	assert.EqualValues(t, user2.Age, user2Age)

}
