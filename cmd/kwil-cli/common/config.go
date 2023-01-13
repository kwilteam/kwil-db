package common

import (
	"fmt"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
)

func LoadConfig() {
	home, err := homedir.Dir()
	if err != nil {
		return
	}
	configFile := filepath.Join(home, ".kwil/config/cli.toml")
	_, err = os.Stat(configFile)
	if err != nil {
		if err := os.MkdirAll(filepath.Dir(configFile), 0755); err != nil {
			fmt.Println(err)
			return
		}

		file, err := os.Create(configFile)
		if err != nil {
			fmt.Println(err)
			return
		}
		file.Close()
	}

	viper.SetConfigFile(configFile)
	_ = viper.ReadInConfig()
}
