package config_testing

import (
	"github.com/kwilteam/kwil-db/internal/chain/config"
	"github.com/kwilteam/kwil-db/internal/common/utils"
)

const configPath = "/test_config.json"

func GetTestConfig() *config.Config {
	dir := utils.GetGoFilePathOfCaller() + configPath
	con, err := config.LoadConfig(dir)
	if err != nil {
		panic(err)
	}
	return con
}
