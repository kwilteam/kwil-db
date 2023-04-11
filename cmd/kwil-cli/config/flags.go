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

var DefaultConfigPath string

func init() {
	viper.SetConfigType("json")
	viper.SetConfigName("config")
	viper.SetEnvPrefix(EnvPrefix)

	dirname, err := os.UserHomeDir()
	if err != nil {
		dirname = "/tmp"
	}

	DefaultConfigPath = fmt.Sprintf("%s/.kwil_cli/", dirname)

	viper.AddConfigPath(DefaultConfigPath)
}

const (
	privateKeyFlag  = "private-key"
	grpcURLFlag     = "kwil-provider"
	clientChainFlag = "client-chain-rpc-url"
)

func BindGlobalFlags(fs *pflag.FlagSet) {
	// Bind flags to environment variables
	fs.String(privateKeyFlag, "", "The private key of the wallet that will be used for signing")
	fs.String(grpcURLFlag, "", "The kwil provider endpoint")
	fs.String(clientChainFlag, "", "The client chain RPC URL")
}
