package config

import (
	"github.com/ethereum/go-ethereum/accounts/abi"
)

type Config struct {
	ClientChain ClientChain `json:"client_chain" mapstructure:"client_chain"`
	Wallets     Wallets     `json:"wallets" mapstructure:"wallets"`
	Storage     Storage     `json:"storage" mapstructure:"storage"`
	Log         struct {
		Human bool `default:"false" json:"human" mapstructure:"human"`
		Debug bool `default:"false" mapstructure:"debug"`
	}
	Api        Api             `json:"api" mapstructure:"api"`
	Cost       Cost            `json:"cost" mapstructure:"cost"`
	Auth       Auth            `json:"auth" mapstructure:"auth"`
	Friendlist []string        `json:"friends" mapstructure:"friends"`
	Friends    map[string]bool `json:"-" mapstructure:"-"`
	Peers      []string        `json:"peers" mapstructure:"peers"`
	PricePath  string          `json:"price_path" mapstructure:"price_path"`
}

func (c *Config) GetPricePath() string {
	return c.PricePath
}

func (c *Config) GetPeers() []string {
	return c.Peers
}

func (c *Config) GetFriends() []string {
	return c.Friendlist
}

func (c *Config) GetChainID() int {
	return c.ClientChain.ChainID
}

// kvPath
func (c *Config) GetKVPath() string {
	return c.Storage.Badger.Path
}

// contract abi
func (c *Config) GetContractABI() abi.ABI {
	return c.ClientChain.DepositContract.ABI
}

// deposit address
func (c *Config) GetDepositAddress() string {
	return c.ClientChain.DepositContract.Address
}

// required block confirmations
func (c *Config) GetReqConfirmations() int {
	return c.ClientChain.RequiredConfirmations
}

// buffer size
func (c *Config) GetBufferSize() int {
	return c.ClientChain.MaxBufferSize
}

// block timeout
func (c *Config) GetBlockTimeout() int {
	return c.ClientChain.BlockTimeout
}

// lowest height
func (c *Config) GetLowestHeight() int64 {
	return int64(c.ClientChain.LowestHeight)
}

// key name
func (c *Config) GetKeyName() string {
	return c.Wallets.Ethereum.KeyName
}

// private key path
func (c *Config) GetPrivKeyPath() string {
	return c.Wallets.Ethereum.PrivKeyPath
}

// isFriend
func (c *Config) IsFriend(friend string) bool {
	return c.Friends[friend]
}

type Auth struct {
	ExpirationTime int `json:"token_expiration_time" mapstructure:"token_expiration_time"`
}

type Cost struct {
	Database DatabaseCosts `json:"database" mapstructure:"database"`
	Ddl      DDLCosts      `json:"ddl" mapstructure:"ddl"`
}

type Storage struct {
	Badger Badger `json:"badger" mapstructure:"badger"`
}

type Badger struct {
	Path string `json:"path" mapstructure:"path"`
}

type Api struct {
	Port        int `json:"port" mapstructure:"port"`
	TimeoutTime int `json:"timeout_time" mapstructure:"timeout_time"`
}

type DatabaseCosts struct {
	Create string `json:"create" mapstructure:"create"`
	Delete string `json:"delete" mapstructure:"delete"`
}

type DDLCosts struct {
	Table TableDDLCosts `json:"table" mapstructure:"table"`
	Role  RoleDDLCosts  `json:"role" mapstructure:"role"`
	Query QueryDDLCosts `json:"query" mapstructure:"query"`
}

type TableDDLCosts struct {
	Create string `json:"create" mapstructure:"create"`
	Delete string `json:"delete" mapstructure:"delete"`
	Modify string `json:"modify" mapstructure:"modify"`
}

type RoleDDLCosts struct {
	Create string `json:"create" mapstructure:"create"`
	Delete string `json:"delete" mapstructure:"delete"`
	Modify string `json:"modify" mapstructure:"modify"`
}

type QueryDDLCosts struct {
	Create string `json:"create" mapstructure:"create"`
	Delete string `json:"delete" mapstructure:"delete"`
}

type ClientChain struct {
	ChainID               int             `json:"chain_id" mapstructure:"id"`
	Endpoint              string          `json:"endpoint" mapstructure:"endpoint"`
	DepositContract       DepositContract `json:"deposit_contract" mapstructure:"deposit_contract"`
	BlockTimeout          int             `json:"block_timeout" mapstructure:"block_timeout"`
	RequiredConfirmations int             `json:"required_confirmations" mapstructure:"required_confirmations"`
	MaxBufferSize         int             `json:"max_block_buffer_size" mapstructure:"max_block_buffer_size"`
	LowestHeight          int             `json:"lowest_height" mapstructure:"lowest_height"`
}

type Wallets struct {
	Ethereum EthereumWallet `json:"ethereum" mapstructure:"ethereum"`
	Cosmos   CosmosWallet   `json:"cosmos" mapstructure:"cosmos"`
}

type EthereumWallet struct {
	Address     string `json:"address" mapstructure:"address"`
	PrivKeyPath string `json:"private_key_path" mapstructure:"private_key_path"`
	KeyName     string `json:"name_on_keyring" mapstructure:"name_on_keyring"`
}

type CosmosWallet struct {
	AddressPrefix string `json:"address_prefix" mapstructure:"address_prefix"`
	MnemonicPath  string `json:"mnemonic_path" mapstructure:"mnemonic_path"`
	KeyName       string `json:"name_on_keyring" mapstructure:"name_on_keyring"`
}

type DepositContract struct {
	Address string   `json:"address" mapstructure:"address"` // Default is currently the USDC address
	ABI     abi.ABI  `json:"abi"`
	ABIPath string   `json:"abi_path" mapstructure:"abi_path"`
	Events  []string `json:"events" mapstructure:"events"`
}
