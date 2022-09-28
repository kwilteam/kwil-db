package config_testing

import (
	"kwil/x/chain/config"
	"kwil/x/utils"
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
