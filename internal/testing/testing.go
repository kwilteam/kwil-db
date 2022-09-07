package testing

import (
	"testing"

	"github.com/kwilteam/kwil-db/internal/dba"
	"github.com/kwilteam/kwil-db/internal/store"
	"github.com/kwilteam/kwil-db/pkg/types"
	tdba "github.com/kwilteam/kwil-db/pkg/types/dba"
)

const configPath = "/configs/test_config.json"
const sqlConfigPath = "/configs/test_sql_config.json"
const emptyDBPath = "/configs/test_empty_db_config.json"

func GetTestConfig(t *testing.T) *types.Config {
	dir := getCurrentPath() + configPath
	con, err := loadConfig(dir)
	if err != nil {
		t.Fatal(err)
	}
	return con
}

func GetTestStore(t *testing.T) *store.BadgerDB {
	conf := GetTestConfig(t)
	st, err := store.New(conf)
	if err != nil {
		t.Fatal(err)
	}
	return st
}

func GetTestLoader(t *testing.T) *dba.DBLoader {
	conf := GetTestConfig(t)

	kv := GetTestStore(t)

	l, err := dba.NewLoader(conf, kv)
	if err != nil {
		t.Fatal(err)
	}
	return l
}

func GetTestSQLConfig(t *testing.T) *tdba.SqlDatabaseConfig {
	curDir := getCurrentPath()
	conf, err := dba.LoadSQLConfig(curDir + sqlConfigPath)
	if err != nil {
		t.Fatal(err)
	}

	return conf
}

func GetEmptySQLConfig(t *testing.T) *tdba.SqlDatabaseConfig {
	curDir := getCurrentPath()
	conf, err := dba.LoadSQLConfig(curDir + emptyDBPath)
	if err != nil {
		t.Fatal(err)
	}

	return conf
}

func GetTestDB(t *testing.T) *dba.DB {
	conf := GetTestSQLConfig(t)
	l := GetTestLoader(t)
	db := dba.NewDB(conf, l)
	return db
}

func GetEmptyTestDB(t *testing.T) *dba.DB {
	conf := GetEmptySQLConfig(t)
	l := GetTestLoader(t)
	db := dba.NewDB(conf, l)
	return db
}
