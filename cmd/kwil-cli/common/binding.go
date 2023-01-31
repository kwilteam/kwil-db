package common

import (
	"kwil/pkg/grpc"
	"time"

	chain "kwil/x/chain/types"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func BindGlobalFlags(fs *pflag.FlagSet) {
	fs.Duration(grpc.DialTimeoutFlag, 5*time.Second, "timeout for requests")
	fs.String(grpc.EndpointFlag, "", "the endpoint of the Kwil node")
	fs.String(grpc.ApiKeyFlag, "", "your api key")

	// TODO: this was missing, not sure the best place for this to live?
	fs.String("funding-pool", "", "the address of the funding pool")

	chain.BindChainFlags(fs)

}

// BindGlobalEnv binds the global flags to the environment variables.
func BindGlobalEnv(fs *pflag.FlagSet) {
	viper.BindEnv(grpc.DialTimeoutFlag, grpc.DialTimeoutEnv)
	viper.BindPFlag(grpc.DialTimeoutFlag, fs.Lookup(grpc.DialTimeoutFlag))

	viper.BindEnv(grpc.EndpointFlag, grpc.EndpointEnv)
	viper.BindPFlag(grpc.EndpointFlag, fs.Lookup(grpc.EndpointFlag))

	viper.BindEnv(grpc.ApiKeyFlag, grpc.ApiKeyEnv)
	viper.BindPFlag(grpc.ApiKeyFlag, fs.Lookup(grpc.ApiKeyFlag))

	viper.BindEnv("funding-pool", "KWIL_FUNDING_POOL")
	viper.BindPFlag("funding-pool", fs.Lookup("funding-pool"))

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
