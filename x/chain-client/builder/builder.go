package builder

import (
	"fmt"
	"kwil/x/chain-client/dto"
)

// the builder is used to build the contract interface and event listener for the specified chain.
// there is already a decent amount of initialization logic required for the chain client,
// and it will only get more complex as we add ABI bindings for erc20, erc721, etc.

type chainBuilder struct {
	chainCode       dto.ChainCode
	rpc             string
	contractAddress string
	privateKey      string
}

type ChainComponents struct {
	Listener dto.Listener
	Escrow   dto.EscrowContract
}

func NewChainBuilder() ChainBuilder {
	return &chainBuilder{}
}

type ChainBuilder interface {
	ChainCode(dto.ChainCode) ChainBuilder
	RPCProvider(string) ChainBuilder
	ContractAddress(string) ChainBuilder
	PrivateKey(string) ChainBuilder
	Build() (*ChainComponents, error)
}

func (b *chainBuilder) ChainCode(cc dto.ChainCode) ChainBuilder {
	b.chainCode = cc
	return b
}

func (b *chainBuilder) RPCProvider(rpcProvider string) ChainBuilder {
	b.rpc = rpcProvider
	return b
}

func (b *chainBuilder) ContractAddress(contractAddress string) ChainBuilder {
	b.contractAddress = contractAddress
	return b
}

func (b *chainBuilder) PrivateKey(privateKey string) ChainBuilder {
	b.privateKey = privateKey
	return b
}

func (b *chainBuilder) Build() (*ChainComponents, error) {
	switch b.chainCode {
	case dto.ETHEREUM:
		return b.buildEVM(b.chainCode.ToChainId())
	case dto.GOERLI:
		return b.buildEVM(b.chainCode.ToChainId())
	default:
		return nil, fmt.Errorf("unsupported chain code: %s", fmt.Sprint(b.chainCode))
	}
}
