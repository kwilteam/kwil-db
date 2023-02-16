package config

const (
	EnvPrefix         = "KCLI"
	DefaultConfigName = "config"
	DefaultConfigDir  = ".kwil_cli"
	DefaultConfigType = "yaml"
)

// viper keys
const (
	KwilProviderRpcUrlKey = "node.rpc_url"
	KwilProviderFlag      = "kwil-provider"

	WalletPrivateKeyKey = "wallet.private_key"
	PrivateKeyFlag      = "private-key"

	ClientChainProviderRpcUrlKey = "chain.rpc_url"
	ChainProviderFlag            = "client-chain-provider"
)

var ConfigFile string
var Config CliConfig
