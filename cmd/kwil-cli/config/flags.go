package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var cliCfg = DefaultKwilCliPersistedConfig()

const (
	defaultConfigDirName      = ".kwil-cli"
	defaultConfigFileName     = "config.json"
	AlternativeConfigHomePath = "/tmp"

	// NOTE: these flags below are also used as viper key names
	globalPrivateKeyFlag = "private-key"
	globalProviderFlag   = "provider"
	globalChainIDFlag    = "chain-id"
	// NOTE: viper key name are used for viper related operations
	// here they are same `mapstructure` names defined in the config struct
	viperPrivateKeyName = "private_key"
	viperProviderName   = "provider"
	viperChainID        = "chain_id"
)

var defaultConfigFile string
var DefaultConfigDir string

func init() {
	dirname, err := os.UserHomeDir()
	if err != nil {
		dirname = AlternativeConfigHomePath
	}

	configPath := filepath.Join(dirname, defaultConfigDirName)
	DefaultConfigDir = configPath
	defaultConfigFile = filepath.Join(configPath, defaultConfigFileName)
}

func BindGlobalFlags(fs *pflag.FlagSet) {
	// Bind flags to environment variables
	fs.String(globalPrivateKeyFlag, cliCfg.PrivateKey, "the private key of the wallet that will be used for signing")
	fs.String(globalProviderFlag, cliCfg.Provider, "the Kwil rpc provider HTTP endpoint")
	fs.String(globalChainIDFlag, cliCfg.ChainID, "the expected/intended Kwil Chain ID")

	// Bind flags to viper, named by the flag name
	viper.BindPFlag(viperPrivateKeyName, fs.Lookup(globalPrivateKeyFlag))
	viper.BindPFlag(viperProviderName, fs.Lookup(globalProviderFlag))
	viper.BindPFlag(viperChainID, fs.Lookup(globalChainIDFlag))
}
