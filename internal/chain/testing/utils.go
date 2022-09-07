package testing

import (
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

func loadConfig(path string) (config, error) {
	var dbConfig config

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
	return dbConfig, nil
}
