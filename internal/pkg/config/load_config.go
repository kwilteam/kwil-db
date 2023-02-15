package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func LoadConfig(defaultConfig map[string]interface{}, configFile, envPrefix, defaultConfigDir, defaultConfigName, defaultConfigType string) {
	for k, v := range defaultConfig {
		viper.SetDefault(k, v)
	}

	if configFile != "" {
		viper.SetConfigFile(configFile)
		fmt.Fprintln(os.Stdout, "Using config file:", viper.ConfigFileUsed())
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(filepath.Join(home, defaultConfigDir))
		viper.SetConfigName(defaultConfigName)
		viper.SetConfigType(defaultConfigType)

		viper.SafeWriteConfig()
	}

	// PREFIX_A_B will be mapped to a.b
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)
	viper.SetEnvPrefix(envPrefix)
	viper.AutomaticEnv()
	// viper.Debug()
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// cfg file not found; ignore error if desired
			fmt.Fprintln(os.Stdout, "cfg file not found:", viper.ConfigFileUsed())
		} else {
			// cfg file was found but another error was produced
			fmt.Fprintln(os.Stderr, "Error loading config file :", err)
		}
	}
	//
	//if err := viper.Unmarshal(&AppConfig, viper.DecodeHook(config.StringPrivateKeyHookFunc())); err != nil {
	//	fmt.Fprintln(os.Stderr, "Error unmarshaling config file:", err)
	//}
}
