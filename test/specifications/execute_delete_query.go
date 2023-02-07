package specifications

import (
	"context"
	"fmt"
	"kwil/pkg/databases"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func ExecuteDBDeleteSpecification(ctx context.Context, t *testing.T, execute ExecuteQueryDsl) {
	t.Logf("Executing delete query specification")
	// Given a valid database schema
	db := SchemaLoader.Load(t)
	dbID := databases.GenerateSchemaName(db.Owner, db.Name)

	userQueryName := "delete_user"
	userTableName := "users"
	userQ := userTable{
		ID:     "1111",
		Name:   "test_user",
		Age:    "33",
		Wallet: strings.ToLower(db.Owner),
		Bool:   "true",
	}
	qualifiedUserTableName := fmt.Sprintf("%s.%s", dbID, userTableName)
	userQueryInput := []any{"name", userQ.Name}

	// When i execute query to database
	err := execute.ExecuteQuery(ctx, db.Name, userQueryName, userQueryInput)
	assert.NoError(t, err)

	rawSQL := fmt.Sprintf("SELECT id, name, age, wallet, boolean FROM %s WHERE id = $1", qualifiedUserTableName)

	// Then i expect row to be deleted
	res, err := execute.QueryDatabase(ctx, rawSQL, userQ.ID)
	assert.NoError(t, err)

	var user userTable
	err = res.Scan(&user.ID, &user.Name, &user.Age, &user.Wallet, &user.Bool)
	assert.Error(t, err)
}
