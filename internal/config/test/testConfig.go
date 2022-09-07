package config_testing

import (
	"github.com/kwilteam/kwil-db/internal/config"
	"github.com/kwilteam/kwil-db/internal/utils/files"
)

const configPath = "/test_config.json"

func GetTestConfig() *config.Config {
	dir := files.GetCurrentPath() + configPath
	con, err := config.LoadConfig(dir)
	if err != nil {
		panic(err)
	}
	return con
}
