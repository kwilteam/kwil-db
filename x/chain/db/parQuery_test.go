package db_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	types "kwil/pkg/types/db"
	pdb "kwil/x/chain/db"
	ktest "kwil/x/chain/testing"
)

func TestDB_StoreParQuer(t *testing.T) {

	pq := types.ParameterizedQuery{
		Name:  "query_1",
		Query: "SELECT * FROM table_1",
		Parameters: []types.Parameter{
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

	err := pdb.StoreParQuer(&pq, db)
	assert.NoError(t, err)

	retPQ, err := db.GetParQuer("query_1")
	assert.NoError(t, err)
	assert.Equal(t, pq, *retPQ)
}
