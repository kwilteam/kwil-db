package config_testing

import (
	"github.com/kwilteam/kwil-db/internal/config"
	"github.com/kwilteam/kwil-db/internal/utils/files"
<<<<<<< HEAD
=======
	"github.com/kwilteam/kwil-db/pkg/types"
	"github.com/spf13/viper"
	"path/filepath"
	"strings"
>>>>>>> b64dc94cf02f1f9d814336627f167ff5d29bb7d5
)

const configPath = "/test_config.json"

<<<<<<< HEAD
func GetTestConfig() *config.Config {
	dir := files.GetCurrentPath() + configPath
	con, err := config.LoadConfig(dir)
=======
func GetTestConfig() *types.Config {
	dir := files.GetCurrentPath() + configPath
	con, err := getConfig(dir)
>>>>>>> b64dc94cf02f1f9d814336627f167ff5d29bb7d5
	if err != nil {
		panic(err)
	}
	return con
}
<<<<<<< HEAD
=======

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
>>>>>>> b64dc94cf02f1f9d814336627f167ff5d29bb7d5
