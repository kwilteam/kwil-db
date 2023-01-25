package common

import (
	"kwil/kwil/client"
	"time"

	chain "kwil/x/chain/types"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func BindConfigFlags(cmd *cobra.Command) {
	//fs := cmd.PersistentFlags()

	//fs.String(client.ConfigFileFlag, "", "config file (default is $HOME/.kwil/config/cli.toml)")
}

func BindGlobalFlags(fs *pflag.FlagSet) {
	fs.Duration(client.DialTimeoutFlag, 5*time.Second, "timeout for requests")
	fs.String(client.EndpointFlag, "", "the endpoint of the Kwil node")
	fs.String(client.ApiKeyFlag, "", "your api key")

	chain.BindChainFlags(fs)

}

// BindGlobalEnv binds the global flags to the environment variables.
func BindGlobalEnv(fs *pflag.FlagSet) {
	viper.BindEnv(client.DialTimeoutFlag, client.DialTimeoutEnv)
	viper.BindPFlag(client.DialTimeoutFlag, fs.Lookup(client.DialTimeoutFlag))

	viper.BindEnv(client.EndpointFlag, client.EndpointEnv)
	viper.BindPFlag(client.EndpointFlag, fs.Lookup(client.EndpointFlag))

	viper.BindEnv(client.ApiKeyFlag, client.ApiKeyEnv)
	viper.BindPFlag(client.ApiKeyFlag, fs.Lookup(client.ApiKeyFlag))

	chain.BindChainEnv()
}

/*
func BindKwilEnv(cmd *cobra.Command) {
	viper.BindEnv(client.DialTimeoutFlag, client.DialTimeoutEnv)
	viper.BindEnv(client.EndpointFlag, client.EndpointEnv)
	viper.BindEnv(client.ApiKeyFlag, client.ApiKeyEnv)
}

func BindKwilFlags(cmd *cobra.Command) {
	fs := cmd.PersistentFlags()

	fs.Duration(client.DialTimeoutFlag, 5*time.Second, "timeout for requests")
	viper.BindPFlag(client.DialTimeoutFlag, fs.Lookup(client.DialTimeoutFlag))

	fs.String(client.EndpointFlag, "", "the endpoint of the Kwil node")
	viper.BindPFlag(client.EndpointFlag, fs.Lookup(client.EndpointFlag))

	fs.String(client.ApiKeyFlag, "", "your api key")
	viper.BindPFlag(client.ApiKeyFlag, fs.Lookup(client.ApiKeyFlag))
}

func MaybeSetFlag(cmd *cobra.Command, name, envVal string) error {
	fl := cmd.Flag(name)
	if fl == nil {
		return nil
	}
	if fl.Changed {
		return nil
	}
	if envVal == "" {
		return nil
	}
	return cmd.Flags().Set(name, envVal)
}
*/
