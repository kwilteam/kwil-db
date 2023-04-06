package specifications

import (
	"context"
	"fmt"
	"kwil/pkg/databases"
	"testing"

	"github.com/stretchr/testify/assert"
)

func ExecuteDBDeleteSpecification(ctx context.Context, t *testing.T, execute ExecuteQueryDsl) {
	return
	t.Logf("Executing delete query specification")
	// Given a valid database schema
	db := SchemaLoader.Load(t)
	dbID := databases.GenerateSchemaId(db.Owner, db.Name)

	userQueryName := "delete_user"
	userTableName := "users"
	userQ := userTable{
		ID:       1111,
		UserName: "test_user",
		Age:      33,
	}
	// qualifiedUserTableName
	_ = fmt.Sprintf("%s_%s", dbID, userTableName)
	userQueryInput := []map[string]any{
		{"where_id": userQ.ID},
	}

	// When i execute query to database
	_, _, err := execute.ExecuteAction(ctx, db.Name, userQueryName, userQueryInput)
	assert.NoError(t, err)

	// Then i expect row to be deleted
	// query := fmt.Sprintf(`query MyQuery { %s (where: {id: {_eq: %d}}) {id username age wallet degen}}`,
	// 	qualifiedUserTableName, userQ.ID)
	// resByte, err := execute.QueryDatabase(ctx, query)
	// assert.NoError(t, err)

	// var resp hasuraResp
	// err = json.Unmarshal(resByte, &resp)
	// assert.NoError(t, err)

	// data := resp["data"]
	// res := data[qualifiedUserTableName]
	// assert.Equal(t, 0, len(res))
}
