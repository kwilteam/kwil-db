package testing

import (
	"testing"

	"github.com/kwilteam/kwil-db/internal/chain/db"
	"github.com/kwilteam/kwil-db/internal/chain/store"
	u "github.com/kwilteam/kwil-db/internal/common/utils"
	tdba "github.com/kwilteam/kwil-db/pkg/types/db"
)

const configPath = "/configs/test_config.json"
const sqlConfigPath = "/configs/test_sql_config.json"
const emptyDBPath = "/configs/test_empty_db_config.json"

func GetTestConfig(t *testing.T) config {
	dir := u.GetCallerPath() + configPath
	con, err := loadConfig(dir)
	if err != nil {
		t.Fatal(err)
	}
	return con
}

type config interface {
	GetChainID() int
	GetKVPath() string
}

func GetTestStore(t *testing.T) *store.BadgerDB {
	conf := GetTestConfig(t)
	st, err := store.New(conf)
	if err != nil {
		t.Fatal(err)
	}
	return st
}

func GetTestLoader(t *testing.T) *db.DBLoader {
	conf := GetTestConfig(t)

	kv := GetTestStore(t)

	l, err := db.NewLoader(conf, kv)
	if err != nil {
		t.Fatal(err)
	}
	return l
}

func GetTestSQLConfig(t *testing.T) *tdba.SqlDatabaseConfig {
	curDir := u.GetCallerPath()
	conf, err := db.LoadSQLConfig(curDir + sqlConfigPath)
	if err != nil {
		t.Fatal(err)
	}

	return conf
}

func GetEmptySQLConfig(t *testing.T) *tdba.SqlDatabaseConfig {
	curDir := u.GetCallerPath()
	conf, err := db.LoadSQLConfig(curDir + emptyDBPath)
	if err != nil {
		t.Fatal(err)
	}

	return conf
}

func GetTestDB(t *testing.T) *db.DB {
	conf := GetTestSQLConfig(t)
	l := GetTestLoader(t)
	db := db.NewDB(conf, l)
	return db
}

func GetEmptyTestDB(t *testing.T) *db.DB {
	conf := GetEmptySQLConfig(t)
	l := GetTestLoader(t)
	db := db.NewDB(conf, l)
	return db
}
