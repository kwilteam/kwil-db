package config

import (
	"github.com/kwilteam/kwil-db/pkg/types"
	"testing"
)

const configPath = "/test_config.json"

func GetTestConfig(t *testing.T) *types.Config {
	dir := getCurrentPath() + configPath
	con, err := loadConfig(dir)
	if err != nil {
		t.Fatal(err)
	}
	return con
}
