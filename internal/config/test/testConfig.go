package config

import (
	"github.com/kwilteam/kwil-db/internal/utils"
	"github.com/kwilteam/kwil-db/pkg/types"
	"testing"
)

const configPath = "/test_config.json"

func GetTestConfig(t *testing.T) *types.Config {
	dir := utils.GetCurrentPath() + configPath
	con, err := loadConfig(dir)
	if err != nil {
		t.Fatal(err)
	}
	return con
}
