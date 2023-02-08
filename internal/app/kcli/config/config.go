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

// viper keys
const (
	NodeEndpointKey = "node.endpoint"

	FundWalletKey            = "fund.wallet"
	FundPoolAddressKey       = "fund.pool_address"
	FundChainCodeKey         = "fund.chain_code"
	FundRPCURLKey            = "fund.rpc_url"
	FundReconnectIntervalKey = "fund.reconnect_interval"
	FundBlockConfirmationKey = "fund.block_confirmation"

	LogLevelKey       = "log.level"
	LogOutputPathsKey = "log.output_paths"
)

var ConfigFile string
var AppConfig *kclient.Config

// BindGlobalFlags binds the global flags to the command.
func BindGlobalFlags(fs *pflag.FlagSet) {
	fs.String(NodeEndpointKey, "", "the endpoint of the Kwil node")

	fs.String(FundWalletKey, "", "you wallet private key")
	fs.String(FundPoolAddressKey, "", "the address of the funding pool")
	fs.String(FundChainCodeKey, "", "the chain code of the funding pool chain")
	fs.String(FundRPCURLKey, "", "the provider url of the funding pool chain")
	// cli does not need to set these flags
	// fs.Int64(FundReconnectIntervalKey, 0, "the reconnect interval of the funding pool")
	// fs.Int64(FundBlockConfirmationKey, 0, "the block confirmation of the funding pool")

	// log flags
	fs.String(LogLevelKey, "", "the level of the Kwil log")
	// ignore the log.output_paths flag
	// fs.StringSlice(LogOutputPathsKey, []string{}, "the output path of the log (default: ['stdout']), use comma to separate multiple output paths")
}

// BindGlobalEnv binds the global flags to the environment variables.
func BindGlobalEnv(fs *pflag.FlagSet) {
	// node.endpoint maps to PREFIX_NODE_ENDPOINT
	viper.SetEnvPrefix(EnvPrefix)

	viper.BindEnv(NodeEndpointKey)
	viper.BindPFlag(NodeEndpointKey, fs.Lookup(NodeEndpointKey)) //flag override env

	viper.BindEnv(FundWalletKey)
	viper.BindPFlag(FundWalletKey, fs.Lookup(FundWalletKey))
	viper.BindEnv(FundPoolAddressKey)
	viper.BindPFlag(FundPoolAddressKey, fs.Lookup(FundPoolAddressKey))
	viper.BindEnv(FundChainCodeKey)
	viper.BindPFlag(FundChainCodeKey, fs.Lookup(FundChainCodeKey))
	viper.BindEnv(FundRPCURLKey)
	viper.BindPFlag(FundRPCURLKey, fs.Lookup(FundRPCURLKey))
	// viper.BindEnv(FundReconnectIntervalKey)
	// viper.BindPFlag(FundReconnectIntervalKey, fs.Lookup(FundReconnectIntervalKey))
	// viper.BindEnv("fund.block_confirmation")
	// viper.BindPFlag(FundBlockConfirmationKey, fs.Lookup(FundBlockConfirmationKey))

	// log key & env
	viper.BindEnv(LogLevelKey)
	viper.BindPFlag(LogLevelKey, fs.Lookup(LogLevelKey))
	viper.SetDefault(LogLevelKey, "info")
	viper.BindEnv(LogOutputPathsKey)
	viper.BindPFlag(LogOutputPathsKey, fs.Lookup(LogOutputPathsKey))
	viper.SetDefault(LogOutputPathsKey, []string{"stdout"})
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

	if err := viper.Unmarshal(&AppConfig, viper.DecodeHook(utils.StringPrivateKeyHookFunc())); err != nil {
		fmt.Fprintln(os.Stderr, "Error unmarshaling config file:", err)
	}
}
