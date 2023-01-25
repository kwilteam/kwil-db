package specifications

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/stretchr/testify/assert"
	"kwil/x/types/databases"
	"testing"
)

type ExecuteQueryDsl interface {
	// ExecuteQuery executes QUERY to a database
	// @yaiba TODO: owner is not needed?? because user can only execute queries using his private key
	ExecuteQuery(ctx context.Context, owner string, dbName string, queryName string, queryInputs []string) error
	QueryDatabase(ctx context.Context, rawSql string, args ...interface{}) (*sql.Row, error)
}

func ExecuteDBInsertSpecification(t *testing.T, ctx context.Context, execute ExecuteQueryDsl) {
	//Given a valid database schema
	db := SchemaLoader.Load(t)

	queryName := "insert_into_table1"
	tableName := "test_table1"
	inputId := "1111"
	inputName := "name22"
	inputAge := "22"
	queryInputs := []string{queryName, "id", inputId, "name", inputName, "age", inputAge, "authenticate_user", "true"}
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
