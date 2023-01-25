package provider

import (
	"context"
	"fmt"
	"kwil/x/chain/provider/dto"
	"kwil/x/chain/provider/evm"
	"kwil/x/chain/types"
	"math/big"

	"github.com/ethereum/go-ethereum/ethclient"
)

func New(endpoint string, chainCode types.ChainCode) (ChainProvider, error) {
	switch chainCode {
	case types.ETHEREUM:
		return evm.New(endpoint, chainCode)
	case types.GOERLI:
		return evm.New(endpoint, chainCode)
	default:
		return nil, fmt.Errorf("unsupported chain code: %s", fmt.Sprint(chainCode))
	}
}

type ChainProvider interface {
	HeaderByNumber(ctx context.Context, number *big.Int) (*dto.Header, error)
	SubscribeNewHead(ctx context.Context, ch chan<- dto.Header) (dto.Subscription, error)
	ChainCode() types.ChainCode
	AsEthClient() (*ethclient.Client, error)
	Endpoint() string
}
