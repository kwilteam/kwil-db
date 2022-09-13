package utils

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

func LoadConfig() error {
	// getting home dir
	d, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	cfgPath := filepath.Join(d, ".kwil/config")
	viper.AddConfigPath(cfgPath)

	viper.SetConfigName("cli")
	viper.SetConfigType("toml")
	err = viper.ReadInConfig()
	if err != nil {
		// if error, try to create the file
		err = os.MkdirAll(cfgPath, 0750)
		if err != nil {
			return err
		}

		f, err := os.Create(filepath.Join(cfgPath, "cli.toml"))
		if err != nil {
			return err
		}
		f.Close()

		// try to read the file again
		err = viper.ReadInConfig()
		if err != nil {
			return err
		}
		return err
	}
	return nil

}
