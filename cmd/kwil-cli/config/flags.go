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
	GlobalProviderFlag   = "provider"
	globalChainIDFlag    = "chain-id"
	globalConfigFileFlag = "config"

	// NOTE: viper key name are used for viper related operations
	// here they are same `mapstructure` names defined in the config struct
	viperPrivateKeyName = "private_key"
	viperProviderName   = "provider"
	viperChainID        = "chain_id"
	viperConfigFile     = "config"
)

var defaultConfigFile string
var DefaultConfigDir string
var configFile string

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
	fs.String(GlobalProviderFlag, cliCfg.Provider, "the Kwil provider RPC endpoint")
	fs.String(globalChainIDFlag, cliCfg.ChainID, "the expected/intended Kwil Chain ID")
	fs.StringVar(&configFile, globalConfigFileFlag, defaultConfigFile, "the path to the Kwil CLI persistent global settings file")

	// Bind flags to viper, named by the flag name
	viper.BindPFlag(viperPrivateKeyName, fs.Lookup(globalPrivateKeyFlag))
	viper.BindPFlag(viperProviderName, fs.Lookup(GlobalProviderFlag))
	viper.BindPFlag(viperChainID, fs.Lookup(globalChainIDFlag))

	// Add deprecated flag
	fs.String("kwil-provider", cliCfg.Provider, "the Kwil provider RPC endpoint")
	fs.MarkDeprecated("kwil-provider", "use '--provider' instead")
}
