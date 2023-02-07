package specifications

import (
	"context"
	"database/sql"
	"fmt"
	"kwil/pkg/types/databases"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type userTable struct {
	ID     string
	Name   string
	Age    string
	Wallet string
	Bool   string
}

type ExecuteQueryDsl interface {
	// ExecuteQuery executes QUERY to a database
	// @yaiba TODO: owner is not needed?? because user can only execute queries using his private key
	ExecuteQuery(ctx context.Context, dbName string, queryName string, queryInputs []any) error
	QueryDatabase(ctx context.Context, rawSQL string, args ...interface{}) (*sql.Row, error)
}

func ExecuteDBInsertSpecification(ctx context.Context, t *testing.T, execute ExecuteQueryDsl) {
	t.Logf("Executing insert query specification")
	// Given a valid database schema
	db := SchemaLoader.Load(t)
	dbID := databases.GenerateSchemaName(db.Owner, db.Name)

	userQueryName := "create_user"
	userTableName := "users"
	userQ := userTable{
		ID:     "1111",
		Name:   "test_user",
		Age:    "22",
		Wallet: strings.ToLower(db.Owner),
		Bool:   "true",
	}
	qualifiedUserTableName := fmt.Sprintf("%s.%s", dbID, userTableName)
	userQueryInput := []any{"id", userQ.ID, "name", userQ.Name, "age", userQ.Age, "boolean", userQ.Bool}

	// TODO test insert post table
	// When i execute query to database
	err := execute.ExecuteQuery(ctx, db.Name, userQueryName, userQueryInput)
	assert.NoError(t, err)

	rawSQL := fmt.Sprintf("SELECT id, name, age, wallet, boolean FROM %s WHERE id = $1", qualifiedUserTableName)

	// Then i expect row to be inserted
	res, err := execute.QueryDatabase(ctx, rawSQL, userQ.ID)
	assert.NoError(t, err)

	var user userTable
	err = res.Scan(&user.ID, &user.Name, &user.Age, &user.Wallet, &user.Bool)
	assert.NoError(t, err)

	assert.Equal(t, userQ.ID, user.ID)
	assert.Equal(t, userQ.Name, user.Name)
	assert.Equal(t, userQ.Age, user.Age)
	assert.Equal(t, userQ.Wallet, user.Wallet)
	assert.Equal(t, userQ.Bool, user.Bool)
	assert.EqualValues(t, userQ, user)
}
