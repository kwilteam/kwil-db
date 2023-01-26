package specifications

import (
	"context"
	"fmt"
	"kwil/x/types/databases"
	"testing"

	"github.com/stretchr/testify/assert"
)

func ExecuteDBDeleteSpecification(t *testing.T, ctx context.Context, execute ExecuteQueryDsl) {
	t.Logf("Executing delete query specification")
	//Given a valid database schema
	db := SchemaLoader.Load(t)

	queryName := "delete_user"
	tableName := "users"
	inputId := "1111"
	inputName := "test_user"
	//inputAge := "33"
	//inputWallet := "guesswhothisis"
	queryInputs := []any{"name", inputName}

	dbId := databases.GenerateSchemaName(db.Owner, db.Name)
	qualifiedTableName := fmt.Sprintf("%s.%s", dbId, tableName)

	//When i execute query to database
	execute.ExecuteQuery(ctx, db.Owner, db.Name, queryName, queryInputs)
	rawSql := fmt.Sprintf("SELECT id, name, age FROM %s WHERE id = $1", qualifiedTableName)

	//Then i expect row to be deleted
	res, err := execute.QueryDatabase(ctx, rawSql, inputId)
	assert.NoError(t, err)
	var rw string
	err = res.Scan(&rw)
	assert.Error(t, err)
}
