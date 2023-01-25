package specifications

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"kwil/x/types/databases"
	"testing"
)

func ExecuteDBUpdateSpecification(t *testing.T, ctx context.Context, execute ExecuteQueryDsl) {
	//Given a valid database schema
	db := SchemaLoader.Load(t)

	queryName := "update_table1"
	tableName := "test_table1"
	inputId := "1111"
	inputName := "name33"
	inputAge := "33"
	dbId := databases.GenerateSchemaName(db.Owner, db.Name)
	queryInputs := []string{queryName, "name", inputName, "age", inputAge, "id", inputId}
	qualifiedTableName := fmt.Sprintf("%s.%s", dbId, tableName)

	//When i execute query to database
	err := execute.ExecuteQuery(ctx, db.Owner, db.Name, queryName, queryInputs)
	assert.NoError(t, err)

	rawSql := fmt.Sprintf("SELECT id, name, age FROM %s WHERE id = $1", qualifiedTableName)

	//Then i expect row to be updated
	res, err := execute.QueryDatabase(ctx, rawSql, inputId)
	assert.NoError(t, err)

	var rowId, rowName, rowAge string
	err = res.Scan(&rowId, &rowName, &rowAge)
	assert.NoError(t, err)

	assert.Equal(t, inputId, rowId)
	assert.Equal(t, inputName, rowName)
	assert.Equal(t, inputAge, rowAge)
}
