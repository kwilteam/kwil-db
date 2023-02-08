package config

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"kwil/pkg/kclient"
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
var AppConfig *kclient.Config

// BindGlobalFlags binds the global flags to the command.
func BindGlobalFlags(fs *pflag.FlagSet) {
	fs.String("node.endpoint", "", "the endpoint of the Kwil node")

	fs.String("fund.wallet", "", "you wallet private key")
	fs.String("fund.pool_address", "", "the address of the funding pool")
	fs.String("fund.chain_code", "", "the chain code of the funding pool chain")
	fs.String("fund.rpc_url", "", "the provider url of the funding pool chain")
	// cli does not need to set these flags
	//fs.Int64("fund.reconnect_interval", 0, "the reconnect interval of the funding pool")
	//fs.Int64("fund.block_confirmation", 0, "the block confirmation of the funding pool")

	// log flags
	fs.String("log.level", "", "the level of the Kwil log")
	// ignore the log.output_paths flag
	//fs.StringSlice("log.output_paths", []string{}, "the output path of the log (default: ['stdout']), use comma to separate multiple output paths")
}

// BindGlobalEnv binds the global flags to the environment variables.
func BindGlobalEnv(fs *pflag.FlagSet) {
	// node.endpoint maps to PREFIX_NODE_ENDPOINT
	viper.SetEnvPrefix(EnvPrefix)

	viper.BindEnv("node.endpoint")
	viper.BindPFlag("node.endpoint", fs.Lookup("node.endpoint")) //flag override env

	viper.BindEnv("fund.wallet")
	viper.BindPFlag("fund.wallet", fs.Lookup("fund.wallet"))
	viper.BindEnv("fund.pool_address")
	viper.BindPFlag("fund.pool_address", fs.Lookup("fund.pool_address"))
	viper.BindEnv("fund.chain_code")
	viper.BindPFlag("fund.chain_code", fs.Lookup("fund.chain_code"))
	viper.BindEnv("fund.rpc_url")
	viper.BindPFlag("fund.rpc_url", fs.Lookup("fund.rpc_url"))
	//viper.BindEnv("fund.reconnect_interval")
	//viper.BindPFlag("fund.reconnect_interval", fs.Lookup("fund.reconnect_interval"))
	//viper.BindEnv("fund.block_confirmation")
	//viper.BindPFlag("fund.block_confirmation", fs.Lookup("fund.block_confirmation"))

	// log key & env
	viper.BindEnv("log.level")
	viper.BindPFlag("log.level", fs.Lookup("log.level"))
	viper.SetDefault("log.level", "info")
	viper.BindEnv("log.output_paths")
	viper.BindPFlag("log.output_paths", fs.Lookup("log.output_paths"))
	viper.SetDefault("log.output_paths", []string{"stdout"})
}

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
	viper.AutomaticEnv()
	//viper.Debug()
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// cfg file not found; ignore error if desired
			fmt.Fprintln(os.Stdout, "cfg file not found:", viper.ConfigFileUsed())
		} else {
			// cfg file was found but another error was produced
			fmt.Fprintln(os.Stderr, "Error loading config file :", err)
		}
	}

	if err := viper.Unmarshal(&AppConfig, viper.DecodeHook(utils.StringPrivateKeyHookFunc())); err != nil {
		fmt.Fprintln(os.Stderr, "Error unmarshaling config file:", err)
	}
}
