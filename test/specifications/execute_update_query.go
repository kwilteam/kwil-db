package specifications

import (
	"context"
	"encoding/json"
	"fmt"
	"kwil/pkg/databases"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func ExecuteDBUpdateSpecification(ctx context.Context, t *testing.T, execute ExecuteQueryDsl) {
	t.Logf("Executing update query specification")
	// Given a valid database schema
	db := SchemaLoader.Load(t)
	dbID := databases.GenerateSchemaId(db.Owner, db.Name)

	userQueryName := "update_user"
	userTableName := "users"
	userQ := userTable{
		ID:       1111,
		UserName: "test_user",
		Age:      33,
		Wallet:   strings.ToLower(db.Owner),
		Degen:    true,
	}
	qualifiedUserTableName := fmt.Sprintf("%s_%s", dbID, userTableName)
	userQueryInput := []any{"username", userQ.UserName, "age", userQ.Age, "where_username", userQ.UserName}

	// When i execute query to database
	err := execute.ExecuteQuery(ctx, db.Name, userQueryName, userQueryInput)
	assert.NoError(t, err)

	// Then i expect row to be updated
	query := fmt.Sprintf(`query MyQuery { %s (where: {id: {_eq: %d}}) {id username age wallet degen}}`,
		qualifiedUserTableName, userQ.ID)
	resByte, err := execute.QueryDatabase(ctx, query)
	assert.NoError(t, err)

	var resp hasuraResp
	err = json.Unmarshal(resByte, &resp)
	assert.NoError(t, err)

	data := resp["data"]
	res := data[qualifiedUserTableName][0]
	assert.EqualValues(t, userQ, res)
}
