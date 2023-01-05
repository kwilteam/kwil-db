package provider

import (
	"context"
	"fmt"
	"kwil/x/chain"
	"kwil/x/chain/provider/dto"
	"kwil/x/chain/provider/evm"
	"math/big"

	"github.com/ethereum/go-ethereum/ethclient"
)

func New(endpoint string, chainCode chain.ChainCode) (ChainProvider, error) {
	switch chainCode {
	case chain.ETHEREUM:
		return evm.New(endpoint, chainCode)
	case chain.GOERLI:
		return evm.New(endpoint, chainCode)
	default:
		return nil, fmt.Errorf("unsupported chain code: %s", fmt.Sprint(chainCode))
	}
}

type ChainProvider interface {
	HeaderByNumber(ctx context.Context, number *big.Int) (*dto.Header, error)
	SubscribeNewHead(ctx context.Context, ch chan<- dto.Header) (dto.Subscription, error)
	ChainCode() chain.ChainCode
	AsEthClient() (*ethclient.Client, error)
	Endpoint() string
}
