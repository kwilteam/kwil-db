package specifications

import (
	"context"
	"fmt"
	"kwil/x/types/databases"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func ExecuteDBUpdateSpecification(ctx context.Context, t *testing.T, execute ExecuteQueryDsl) {
	t.Logf("Executing update query specification")
	// Given a valid database schema
	db := SchemaLoader.Load(t)
	dbID := databases.GenerateSchemaName(db.Owner, db.Name)

	userQueryName := "update_user"
	userTableName := "users"
	userQ := userTable{
		ID:     "1111",
		Name:   "test_user",
		Age:    "33",
		Wallet: strings.ToLower(db.Owner),
		Bool:   "true",
	}
	qualifiedUserTableName := fmt.Sprintf("%s.%s", dbID, userTableName)
	userQueryInput := []any{"name", userQ.Name, "age", userQ.Age, "where_name", userQ.Name}

	// When i execute query to database
	err := execute.ExecuteQuery(ctx, db.Name, userQueryName, userQueryInput)
	assert.NoError(t, err)

	rawSQL := fmt.Sprintf("SELECT id, name, age, wallet, boolean FROM %s WHERE id = $1", qualifiedUserTableName)

	// Then i expect row to be updated
	res, err := execute.QueryDatabase(ctx, rawSQL, userQ.ID)
	assert.NoError(t, err)

	var user userTable
	err = res.Scan(&user.ID, &user.Name, &user.Age, &user.Wallet, &user.Bool)
	assert.NoError(t, err)

	assert.EqualValues(t, userQ, user)
}
