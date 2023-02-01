package types

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	ChainCodeFlag = "chain-code"
	ChainCodeEnv  = "KWIL_CHAIN_CODE"

	//@yaiba TODO: rename to provider
	EthProviderFlag = "eth-provider"
	EthProviderEnv  = "KWIL_ETH_PROVIDER"

	PrivateKeyFlag = "private-key"
	PrivateKeyEnv  = "KWIL_PRIVATE_KEY"

	RequiredConfirmationsFlag = "required-confirmations"
	RequiredConfirmationsEnv  = "KWIL_REQUIRED_CONFIRMATIONS"

	ReconnectionIntervalFlag = "reconnection-interval"
	ReconnectionIntervalEnv  = "KWIL_RECONNECTION_INTERVAL"
)

func BindChainFlags(fs *pflag.FlagSet) {

	fs.String(ChainCodeFlag, "", "chain code")
	viper.BindPFlag(ChainCodeFlag, fs.Lookup(ChainCodeFlag))

	fs.String(EthProviderFlag, "", "eth provider")
	viper.BindPFlag(EthProviderFlag, fs.Lookup(EthProviderFlag))

	fs.String(PrivateKeyFlag, "", "private key")
	viper.BindPFlag(PrivateKeyFlag, fs.Lookup(PrivateKeyFlag))

	fs.String(RequiredConfirmationsFlag, "12", "required confirmations")
	viper.BindPFlag(RequiredConfirmationsFlag, fs.Lookup(RequiredConfirmationsFlag))

	fs.String(ReconnectionIntervalFlag, "30", "reconnection interval")
	viper.BindPFlag(ReconnectionIntervalFlag, fs.Lookup(ReconnectionIntervalFlag))
}

func BindChainEnv() {
	viper.BindEnv(ChainCodeFlag, ChainCodeEnv)
	viper.BindEnv(PrivateKeyFlag, PrivateKeyEnv)
	viper.BindEnv(EthProviderFlag, EthProviderEnv)
	viper.BindEnv(RequiredConfirmationsFlag, RequiredConfirmationsEnv)
	viper.BindEnv(ReconnectionIntervalFlag, ReconnectionIntervalEnv)
}
