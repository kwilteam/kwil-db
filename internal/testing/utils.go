package testing

import (
	"github.com/kwilteam/kwil-db/pkg/types"
	"github.com/spf13/viper"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

func getCurrentPath() string {
	_, filename, _, _ := runtime.Caller(1)

	return path.Dir(filename)
}

func loadConfig(path string) (*types.Config, error) {
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
	return &dbConfig, nil
}
