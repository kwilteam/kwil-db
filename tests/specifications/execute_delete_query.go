package specifications

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"kwil/x/types/databases"
	"testing"
)

func ExecuteDBDeleteSpecification(t *testing.T, ctx context.Context, execute ExecuteQueryDsl) {
	//Given a valid database schema
	db := SchemaLoader.Load(t)

	queryName := "delete_from_table1"
	tableName := "test_table1"
	inputId := "1111"
	queryInputs := []string{queryName, "id", inputId}
	dbId := databases.GenerateSchemaName(db.Owner, db.Name)
	qualifiedTableName := fmt.Sprintf("%s.%s", dbId, tableName)

	//When i execute query to database
	execute.ExecuteQuery(ctx, db.Owner, db.Name, queryName, queryInputs)
	rawSql := fmt.Sprintf("SELECT id, name, age FROM %s WHERE id = $1", qualifiedTableName)

	//Then i expect row to be deleted
	res, err := execute.QueryDatabase(ctx, rawSql, inputId)
	assert.NoError(t, err)
	assert.Error(t, res.Err())
}
