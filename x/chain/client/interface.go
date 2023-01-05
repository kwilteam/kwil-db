package client

import (
	"context"
	"kwil/x/chain"
	provider "kwil/x/chain/provider/dto"

	"github.com/ethereum/go-ethereum/ethclient"
)

type ChainClient interface {
	Listen(ctx context.Context, blocks chan<- int64) error
	GetLatestBlock(ctx context.Context) (*provider.Header, error)
	ChainCode() chain.ChainCode
	AsEthClient() (*ethclient.Client, error)
}
