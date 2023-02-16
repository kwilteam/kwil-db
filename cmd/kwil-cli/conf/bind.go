package conf

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func BindGlobalFlags(fs *pflag.FlagSet) {
	fs.String(KwilProviderFlag, "", "the address of the Kwild node server")

	fs.String(PrivateKeyFlag, "", "your wallet private key")

	fs.String(ChainProviderFlag, "", "the address of the chain provider")
}

func BindGlobalEnv(fs *pflag.FlagSet) {
	viper.SetEnvPrefix(EnvPrefix)

	viper.BindEnv(KwilProviderRpcUrlKey)
	viper.BindPFlag(KwilProviderRpcUrlKey, fs.Lookup(KwilProviderFlag))

	viper.BindEnv(WalletPrivateKeyKey)
	viper.BindPFlag(WalletPrivateKeyKey, fs.Lookup(PrivateKeyFlag))

	viper.BindEnv(ClientChainProviderRpcUrlKey)
	viper.BindPFlag(ClientChainProviderRpcUrlKey, fs.Lookup(ChainProviderFlag))

	envs := []string{
		KwilProviderRpcUrlKey,
		WalletPrivateKeyKey,
	}

	for _, v := range envs {
		viper.BindEnv(v)
		viper.BindPFlag(v, fs.Lookup(v))
	}
}
