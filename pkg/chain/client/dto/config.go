package dto

type Config struct {
	ChainCode         int64  `mapstructure:"chain_code"`
	ReconnectInterval int64  `mapstructure:"reconnect_interval"`
	BlockConfirmation int64  `mapstructure:"block_confirmation"`
	RpcUrl            string `mapstructure:"rpc_url"`
	PublicRpcUrl      string `mapstructure:"public_rpc_url"`
}
