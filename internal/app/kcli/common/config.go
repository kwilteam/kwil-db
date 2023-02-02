package common

import (
	"fmt"
	"github.com/spf13/cobra"
	"kwil/pkg/kwil-client"
	"kwil/pkg/utils"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

const (
	EnvPrefix         = "KCLI"
	DefaultConfigName = "config"
	DefaultConfigDir  = ".kwil_cli"
	DefaultConfigType = "toml"
)

var ConfigFile string
var AppConfig *kwil_client.Config

func LoadConfig() {
	if ConfigFile != "" {
		viper.SetConfigFile(ConfigFile)
		fmt.Fprintln(os.Stdout, "Using config file:", viper.ConfigFileUsed())
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(filepath.Join(home, DefaultConfigDir))
		viper.SetConfigName(DefaultConfigName)
		viper.SetConfigType(DefaultConfigType)
		viper.SafeWriteConfig()
	}

	// PREFIX_A_B will be mapped to a.b
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)
	viper.SetEnvPrefix(EnvPrefix)

	//viper.AllowEmptyEnv(true)
	viper.AutomaticEnv()
	//viper.Debug()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			fmt.Fprintln(os.Stdout, "Config file not found:", viper.ConfigFileUsed())

		} else {
			// Config file was found but another error was produced
			fmt.Fprintln(os.Stderr, "Error loading config file :", err)
		}
	}

	if err := viper.Unmarshal(&AppConfig, viper.DecodeHook(utils.StringPrivateKeyHookFunc())); err != nil {
		fmt.Fprintln(os.Stderr, "Error unmarshaling config file:", err)
	}
}
