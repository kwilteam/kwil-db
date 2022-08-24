package dba_test

import (
	"github.com/kwilteam/kwil-db/internal/dba"
	ktest "github.com/kwilteam/kwil-db/internal/testing"
	types "github.com/kwilteam/kwil-db/pkg/types/dba"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDB_StoreAllNoTx(t *testing.T) {

	db := ktest.GetTestDB(t)
	defer db.Close()
	sqlConf := ktest.GetTestSQLConfig(t)

	// Going to set my own name, owner, dbType, and defaultRole

	// Testing for non-transactiom
	db.StoreAll(false)

	retName, err := db.Get([]byte("name"))
	assert.NoError(t, err)
	expectedName := *sqlConf.GetName()
	assert.Equal(t, expectedName, string(retName))

	retOwner, err := db.Get([]byte("owner"))
	assert.NoError(t, err)
	expectedOwner := *sqlConf.GetOwner()
	assert.Equal(t, expectedOwner, string(retOwner))

	retDBType, err := db.Get([]byte("dbType"))
	assert.NoError(t, err)
	expectedDBType := *sqlConf.GetDBType()
	assert.Equal(t, expectedDBType, string(retDBType))

	retDefaultRole, err := db.Get([]byte("defRole"))
	assert.NoError(t, err)
	expectedDefaultRole := *sqlConf.GetDefaultRole()
	assert.Equal(t, expectedDefaultRole, string(retDefaultRole))
}

func TestDB_StoreAllTx(t *testing.T) {

	db := ktest.GetTestDB(t)
	defer db.Close()
	sqlConf := ktest.GetTestSQLConfig(t)

	// Going to set my own name, owner, dbType, and defaultRole

	// Testing for transactiom
	db.StoreAll(true)

	retName, err := db.Get([]byte("name"))
	assert.NoError(t, err)
	expectedName := *sqlConf.GetName()
	assert.Equal(t, expectedName, string(retName))

	retOwner, err := db.Get([]byte("owner"))
	assert.NoError(t, err)
	expectedOwner := *sqlConf.GetOwner()
	assert.Equal(t, expectedOwner, string(retOwner))

	retDBType, err := db.Get([]byte("dbType"))
	assert.NoError(t, err)
	expectedDBType := *sqlConf.GetDBType()
	assert.Equal(t, expectedDBType, string(retDBType))

	retDefaultRole, err := db.Get([]byte("defRole"))
	assert.NoError(t, err)
	expectedDefaultRole := *sqlConf.GetDefaultRole()
	assert.Equal(t, expectedDefaultRole, string(retDefaultRole))
}

func TestDB_StoreAndGetRole(t *testing.T) {

	testRole := types.Role{
		Name: "kwiller",
		Permissions: types.Permissions{
			DDL:                  true,
			ParamaterizedQueries: []string{"query_1", "query_2"},
		},
	}

	db := ktest.GetEmptyTestDB(t)
	defer db.Close()

	dba.StoreRole(&testRole, db)
	retRole, err := db.GetRole("kwiller")
	assert.NoError(t, err)
	assert.Equal(t, testRole, *retRole)
}
