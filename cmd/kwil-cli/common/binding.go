package common

import (
	"kwil/kwil/client"
	"time"

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

	fs.String(client.ChainCodeFlag, "", "chain code")
	fs.String(client.PrivateKeyFlag, "", "private key")
	fs.String(client.FundingPoolFlag, "", "funding pool")
	fs.String(client.NodeAddressFlag, "", "node address")
	fs.String(client.EthProviderFlag, "", "eth provider")
}

// BindGlobalEnv binds the global flags to the environment variables.
func BindGlobalEnv(fs *pflag.FlagSet) {
	viper.BindEnv(client.ChainCodeFlag, client.ChainCodeEnv)
	viper.BindPFlag(client.ChainCodeFlag, fs.Lookup(client.ChainCodeFlag))

	viper.BindEnv(client.PrivateKeyFlag, client.PrivateKeyEnv)
	viper.BindPFlag(client.PrivateKeyFlag, fs.Lookup(client.PrivateKeyFlag))

	viper.BindEnv(client.FundingPoolFlag, client.FundingPoolEnv)
	viper.BindPFlag(client.FundingPoolFlag, fs.Lookup(client.FundingPoolFlag))

	viper.BindEnv(client.NodeAddressFlag, client.NodeAddressEnv)
	viper.BindPFlag(client.NodeAddressFlag, fs.Lookup(client.NodeAddressFlag))

	viper.BindEnv(client.EthProviderFlag, client.EthProviderEnv)
	viper.BindPFlag(client.EthProviderFlag, fs.Lookup(client.EthProviderFlag))

	viper.BindEnv(client.DialTimeoutFlag, client.DialTimeoutEnv)
	viper.BindPFlag(client.DialTimeoutFlag, fs.Lookup(client.DialTimeoutFlag))

	viper.BindEnv(client.EndpointFlag, client.EndpointEnv)
	viper.BindPFlag(client.EndpointFlag, fs.Lookup(client.EndpointFlag))

	viper.BindEnv(client.ApiKeyFlag, client.ApiKeyEnv)
	viper.BindPFlag(client.ApiKeyFlag, fs.Lookup(client.ApiKeyFlag))
}

/*
func BindKwilEnv(cmd *cobra.Command) {
	fs := cmd.PersistentFlags()

	viper.BindEnv(client.DialTimeoutFlag, client.DialTimeoutEnv)
	viper.BindPFlag(client.DialTimeoutFlag, fs.Lookup(client.DialTimeoutFlag))

	viper.BindEnv(client.EndpointFlag, client.EndpointEnv)
	viper.BindPFlag(client.EndpointFlag, fs.Lookup(client.EndpointFlag))

	viper.BindEnv(client.ApiKeyFlag, client.ApiKeyEnv)
	viper.BindPFlag(client.ApiKeyFlag, fs.Lookup(client.ApiKeyFlag))
}

func BindKwilFlags(cmd *cobra.Command) {
	fs := cmd.PersistentFlags()

	fs.Duration(client.DialTimeoutFlag, 5*time.Second, "timeout for requests")
	fs.String(client.EndpointFlag, "", "the endpoint of the Kwil node")
	fs.String(client.ApiKeyFlag, "", "your api key")
}

func BindChainEnv(cmd *cobra.Command) {
	fs := cmd.PersistentFlags()

	viper.BindEnv(client.ChainCodeFlag, client.ChainCodeEnv)
	viper.BindPFlag(client.ChainCodeFlag, fs.Lookup(client.ChainCodeFlag))

	viper.BindEnv(client.PrivateKeyFlag, client.PrivateKeyEnv)
	viper.BindPFlag(client.PrivateKeyFlag, fs.Lookup(client.PrivateKeyFlag))

	viper.BindEnv(client.FundingPoolFlag, client.FundingPoolEnv)
	viper.BindPFlag(client.FundingPoolFlag, fs.Lookup(client.FundingPoolFlag))

	viper.BindEnv(client.NodeAddressFlag, client.NodeAddressEnv)
	viper.BindPFlag(client.NodeAddressFlag, fs.Lookup(client.NodeAddressFlag))

	viper.BindEnv(client.EthProviderFlag, client.EthProviderEnv)
	viper.BindPFlag(client.EthProviderFlag, fs.Lookup(client.EthProviderFlag))
}

func BindChainFlags(cmd *cobra.Command) {
	fs := cmd.PersistentFlags()

	fs.String(client.ChainCodeFlag, "", "chain code")
	fs.String(client.PrivateKeyFlag, "", "private key")
	fs.String(client.FundingPoolFlag, "", "funding pool")
	fs.String(client.NodeAddressFlag, "", "node address")
	fs.String(client.EthProviderFlag, "", "eth provider")
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
