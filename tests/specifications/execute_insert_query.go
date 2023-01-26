package specifications

import (
	"context"
	"database/sql"
	"fmt"
	"kwil/x/types/databases"
	"testing"

	"github.com/stretchr/testify/assert"
)

type ExecuteQueryDsl interface {
	// ExecuteQuery executes QUERY to a database
	// @yaiba TODO: owner is not needed?? because user can only execute queries using his private key
	ExecuteQuery(ctx context.Context, owner string, dbName string, queryName string, queryInputs []any) error
	QueryDatabase(ctx context.Context, rawSql string, args ...interface{}) (*sql.Row, error)
}

func ExecuteDBInsertSpecification(t *testing.T, ctx context.Context, execute ExecuteQueryDsl) {
	t.Logf("Executing insert query specification")
	//Given a valid database schema
	db := SchemaLoader.Load(t)

	queryName := "create_user"
	tableName := "users"
	inputId := "1111"
	inputName := "test_user"
	inputAge := "22"
	//inputWallet := "guesswhothisis"
	queryInputs := []any{"id", inputId, "name", inputName, "age", inputAge}

	dbId := databases.GenerateSchemaName(db.Owner, db.Name)
	qulifiedTableName := fmt.Sprintf("%s.%s", dbId, tableName)

	//When i execute query to database
	err := execute.ExecuteQuery(ctx, db.Owner, db.Name, queryName, queryInputs)
	assert.NoError(t, err)

	rawSql := fmt.Sprintf("SELECT id, name, age FROM %s WHERE id = $1", qulifiedTableName)

	//Then i expect row to be inserted
	res, err := execute.QueryDatabase(ctx, rawSql, inputId)
	assert.NoError(t, err)

	var rowId, rowName, rowAge string
	err = res.Scan(&rowId, &rowName, &rowAge)
	assert.NoError(t, err)

	assert.Equal(t, inputId, rowId)
	assert.Equal(t, inputName, rowName)
	assert.Equal(t, inputAge, rowAge)
}
