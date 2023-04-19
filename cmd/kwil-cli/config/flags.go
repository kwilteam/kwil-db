package config

import (
	"fmt"
	"os"

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

	configPath := fmt.Sprintf("%s/.kwil_cli/", dirname)

	DefaultConfigFile = fmt.Sprintf("%s/config.json", configPath)

	viper.AddConfigPath(configPath)
	viper.AutomaticEnv()
}

const (
	privateKeyFlag  = "private-key"
	grpcURLFlag     = "kwil-provider"
	clientChainFlag = "client-chain-rpc-url"

	privateKeyEnv  = "private_key"
	grpcURLEnv     = "grpc_url"
	clientChainEnv = "client_chain_rpc_url"
)

func BindGlobalFlags(fs *pflag.FlagSet) {
	// Bind flags to environment variables
	fs.String(privateKeyFlag, "", "The private key of the wallet that will be used for signing")
	fs.String(grpcURLFlag, "", "The kwil provider endpoint")
	fs.String(clientChainFlag, "", "The client chain RPC URL")

	viper.BindPFlag(privateKeyEnv, fs.Lookup(privateKeyFlag))
	viper.BindPFlag(grpcURLEnv, fs.Lookup(grpcURLFlag))
	viper.BindPFlag(clientChainEnv, fs.Lookup(clientChainFlag))
}
