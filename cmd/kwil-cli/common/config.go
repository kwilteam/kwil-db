package common

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

func LoadConfig() {
	configFile := GetConfigFile()
	_, err := os.Stat(configFile)
	if err != nil {
		// TODO: create init function
		log.Fatal(err)
	}

	LoadConfigFromPath(configFile)
}

func LoadConfigFromPath(path string) {
	_, err := os.Stat(path)
	if err != nil {
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			fmt.Println(err)
			return
		}

		file, err := os.Create(path)
		if err != nil {
			fmt.Println(err)
			return
		}
		file.Close()
	}

	viper.SetConfigFile(path)

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		fmt.Println(err)
	}
}

func GetConfigFile() string {
	home, err := homedir.Dir()
	if err != nil {
		return ""
	}
	configFile := filepath.Join(home, ".kwil/config/cli.toml")
	return configFile
}
