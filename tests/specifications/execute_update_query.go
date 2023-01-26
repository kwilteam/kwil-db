package specifications

import (
	"context"
	"fmt"
	"kwil/x/types/databases"
	"testing"

	"github.com/stretchr/testify/assert"
)

func ExecuteDBUpdateSpecification(t *testing.T, ctx context.Context, execute ExecuteQueryDsl) {
	t.Logf("Executing update query specification")
	//Given a valid database schema
	db := SchemaLoader.Load(t)

	queryName := "update_user"
	tableName := "users"
	inputId := "1111"
	inputName := "test_user"
	inputAge := "33"
	//inputWallet := "guesswhothisis"
	queryInputs := []string{"name", inputName, "age", inputAge, "where_name", inputName}
	dbId := databases.GenerateSchemaName(db.Owner, db.Name)
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
