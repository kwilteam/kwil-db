package chain

import (
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

//goland:noinspection GoUnusedExportedFunction
func NewClientChain(chainID int, endpoint string, depositContract *DepositContract) *ClientChain {
	return &ClientChain{
		ChainID:         chainID,
		Endpoint:        endpoint,
		DepositContract: *depositContract,
	}
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
	KeyringFile string         `json:"keyring_file" mapstructure:"keyring_file"`
	Ethereum    EthereumWallet `json:"ethereum" mapstructure:"ethereum"`
	Cosmos      CosmosWallet   `json:"cosmos" mapstructure:"cosmos"`
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

func (c *ClientChain) GetChainID() int {
	return c.ChainID
}

func (c *ClientChain) GetEndpoint() string {
	return c.Endpoint
}

func (c *ClientChain) GetContractAddress() string {
	return c.DepositContract.Address
}

func (c *ClientChain) GetContractABI() *abi.ABI {
	return &c.DepositContract.ABI
}

func (c *ClientChain) GetDepositContract() DepositContract {
	return c.DepositContract
}

func (c *ClientChain) GetBlockTimeout() int {
	return c.BlockTimeout
}

func (c *ClientChain) GetRequiredConfirmations() int {
	return c.RequiredConfirmations
}

// Takes event names and gets topics based on the ABI
func (c *ClientChain) GetTopics(events []string) map[common.Hash]abi.Event {
	topics := make(map[common.Hash]abi.Event)
	for _, v := range events {
		event := c.GetContractABI().Events[v]
		topics[event.ID] = event
	}
	return topics
}

type DepositContract struct {
	Address string   `json:"address" mapstructure:"address"` // Default is currently the USDC address
	ABI     abi.ABI  `json:"abi"`
	ABIPath string   `json:"abi_path" mapstructure:"abi_path"`
	Events  []string `json:"events" mapstructure:"events"`
}

func (c *DepositContract) GetAddress() string {
	return c.Address
}

func (c *DepositContract) GetEvents() []string {
	return c.Events
}

func (c *DepositContract) GetABI() *abi.ABI {
	return &c.ABI
}

func (c *DepositContract) GetABIPath() string {
	return c.ABIPath
}
