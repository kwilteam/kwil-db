package specifications

import (
	"context"
	"fmt"
	"kwil/x/types/databases"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func ExecuteDBDeleteSpecification(t *testing.T, ctx context.Context, execute ExecuteQueryDsl) {
	t.Logf("Executing delete query specification")
	//Given a valid database schema
	db := SchemaLoader.Load(t)
	dbId := databases.GenerateSchemaName(db.Owner, db.Name)

	userQueryName := "delete_user"
	userTableName := "users"
	userQ := userTable{
		Id:     "1111",
		Name:   "test_user",
		Age:    "33",
		Wallet: strings.ToLower(db.Owner),
		Bool:   "true",
	}
	qualifiedUserTableName := fmt.Sprintf("%s.%s", dbId, userTableName)
	userQueryInput := []any{"name", userQ.Name}

	//When i execute query to database
	err := execute.ExecuteQuery(ctx, db.Owner, db.Name, userQueryName, userQueryInput)
	assert.NoError(t, err)

	rawSql := fmt.Sprintf("SELECT id, name, age, wallet, boolean FROM %s WHERE id = $1", qualifiedUserTableName)

	//Then i expect row to be deleted
	res, err := execute.QueryDatabase(ctx, rawSql, userQ.Id)
	assert.NoError(t, err)

	var user userTable
	err = res.Scan(&user.Id, &user.Name, &user.Age, &user.Wallet, &user.Bool)
	assert.Error(t, err)
}
