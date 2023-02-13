package config

import (
	"fmt"
	"github.com/spf13/pflag"
	"kwil/internal/pkg/config"
	"kwil/pkg/kclient"
	"os"

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
	NodeAddrKey = "node.addr"

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

var defaultConfig = map[string]interface{}{
	"log": map[string]interface{}{
		"level":        "info",
		"output_paths": []string{"stdout"},
	},
	"fund": map[string]interface{}{
		"reconnect_interval": 30,
		"block_confirmation": 12,
	},
}

// BindGlobalFlags binds the global flags to the command.
func BindGlobalFlags(fs *pflag.FlagSet) {
	fs.String(NodeAddrKey, "", "the address of the Kwild node server")

	fs.String(FundWalletKey, "", "your wallet private key")
	fs.String(FundPoolAddressKey, "", "the address of the funding pool")
	fs.String(FundChainCodeKey, "", "the chain code of the funding pool chain")
	fs.String(FundRPCURLKey, "", "the provider url of the funding pool chain")

	// log flags
	fs.String(LogLevelKey, "", "the level of the log (default: info)")
}

// BindGlobalEnv binds the global flags to the environment variables.
func BindGlobalEnv(fs *pflag.FlagSet) {
	// node.endpoint maps to PREFIX_NODE_ENDPOINT
	viper.SetEnvPrefix(EnvPrefix)

	envs := []string{
		NodeAddrKey,
		FundWalletKey,
		FundPoolAddressKey,
		FundChainCodeKey,
		FundRPCURLKey,
		FundReconnectIntervalKey,
		FundBlockConfirmationKey,
		LogLevelKey,
		LogOutputPathsKey,
	}

	for _, v := range envs {
		viper.BindEnv(v)
		viper.BindPFlag(v, fs.Lookup(v))
	}
}

func LoadConfig() {
	config.LoadConfig(defaultConfig, ConfigFile, EnvPrefix, DefaultConfigDir, DefaultConfigName, DefaultConfigType)
	if err := viper.Unmarshal(&AppConfig, viper.DecodeHook(config.StringPrivateKeyHookFunc())); err != nil {
		fmt.Fprintln(os.Stderr, "Error unmarshaling config file:", err)
	}
}
