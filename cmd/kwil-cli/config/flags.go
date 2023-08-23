package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	EnvPrefix = "KWIL_CLI"
)

var DefaultConfigFile string

func init() {
	viper.SetConfigType("json")
	viper.SetConfigName("config")
	viper.SetEnvPrefix(EnvPrefix)

	dirname, err := os.UserHomeDir()
	if err != nil {
		dirname = "/tmp"
	}

	configPath := filepath.Join(dirname, ".kwil_cli")

	DefaultConfigFile = filepath.Join(configPath, "config.json")

	viper.AddConfigPath(configPath)
	viper.AutomaticEnv()
}

const (
	privateKeyFlag = "private-key"
	grpcURLFlag    = "kwil-provider"

	privateKeyEnv = "private_key"
	grpcURLEnv    = "grpc_url"
)

func BindGlobalFlags(fs *pflag.FlagSet) {
	// Bind flags to environment variables
	fs.String(privateKeyFlag, "", "The private key of the wallet that will be used for signing")
	fs.String(grpcURLFlag, "", "The kwil provider endpoint")

	viper.BindPFlag(privateKeyEnv, fs.Lookup(privateKeyFlag))
	viper.BindPFlag(grpcURLEnv, fs.Lookup(grpcURLFlag))
}
