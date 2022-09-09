package utils

import (
	"github.com/spf13/viper"
	"os"
)

func LoadConfig() error {
	// getting home dir
	d, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	// the config will be stored at $HOME/.kwil_cli.json
	viper.AddConfigPath(d + "/.kwil/config")
	viper.SetConfigName("cli")
	viper.SetConfigType("toml")
	err = viper.ReadInConfig()
	if err != nil {
		// if error, try to create the file
		f, err := os.Create(d + "/.kwil/config/cli.toml")
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
