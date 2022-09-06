package config_testing

import (
	"github.com/kwilteam/kwil-db/internal/config"
	"github.com/kwilteam/kwil-db/internal/utils/files"
	"github.com/kwilteam/kwil-db/pkg/types"
	"github.com/spf13/viper"
	"path/filepath"
	"strings"
)

const configPath = "/test_config.json"

func GetTestConfig() *types.Config {
	dir := files.GetCurrentPath() + configPath
	con, err := getConfig(dir)
	if err != nil {
		panic(err)
	}
	return con
}

// Currently only used for testing
func getConfig(path string) (*types.Config, error) {
	var dbConfig types.Config

	dir, file := filepath.Split(path)
	strs := strings.Split(file, ".")

	viper.AddConfigPath(dir)
	viper.SetConfigName(file)
	viper.SetConfigType(strs[len(strs)-1])

	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	err = viper.Unmarshal(&dbConfig)
	if err != nil {
		return nil, err
	}

	config.Init(&dbConfig)

	return &dbConfig, nil
}
