package dba_test

import (
	"github.com/kwilteam/kwil-db/internal/dba"
	ktest "github.com/kwilteam/kwil-db/internal/testing"
	types "github.com/kwilteam/kwil-db/pkg/types/dba"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDB_StoreParQuer(t *testing.T) {

	pq := types.ParameterizedQuery{
		Name:  "query_1",
		Query: "SELECT * FROM table_1",
		Parameters: []types.Paramater{
			{
				Name: "param_1",
				Type: "string",
			},
			{
				Name: "param_2",
				Type: "int",
			},
		},
	}

	db := ktest.GetEmptyTestDB(t)
	defer db.Close()

	err := dba.StoreParQuer(&pq, db)
	assert.NoError(t, err)

	retPQ, err := db.GetParQuer("query_1")
	assert.NoError(t, err)
	assert.Equal(t, pq, *retPQ)
}
