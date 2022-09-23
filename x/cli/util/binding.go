package util

import (
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func BindKwilFlags(fs *pflag.FlagSet) {
	fs.Duration("dial-timeout", 5*time.Second, "timeout for requests")
	viper.BindPFlag("dial-timeout", fs.Lookup("dial-timeout"))

	fs.String("endpoint", "", "the endpoint of the Kwil node")
	viper.BindPFlag("endpoint", fs.Lookup("endpoint"))
	viper.BindEnv("endpoint", "KWIL_ENDPOINT")

	fs.String("api-key", "", "your api key")
	viper.BindPFlag("api-key", fs.Lookup("api-key"))
	viper.BindEnv("api-key", "KWIL_API_KEY")
}

func BindChainFlags(fs *pflag.FlagSet) {
	fs.String("chain-id", "", "chain id")
	fs.String("private-key", "", "private key")
	fs.String("funding-pool", "", "funding pool")
	fs.String("node-address", "", "node address")
	fs.String("eth-provider", "", "eth provider")

	viper.BindPFlag("chain-id", fs.Lookup("chain-id"))
	viper.BindPFlag("private-key", fs.Lookup("private-key"))
	viper.BindPFlag("funding-pool", fs.Lookup("funding-pool"))
	viper.BindPFlag("node-address", fs.Lookup("node-address"))
	viper.BindPFlag("eth-provider", fs.Lookup("eth-provider"))
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
