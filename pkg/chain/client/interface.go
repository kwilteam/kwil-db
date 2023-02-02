package client

import (
	"context"
	"github.com/ethereum/go-ethereum/ethclient"
	provider "kwil/pkg/chain/provider/dto"
	"kwil/pkg/chain/types"
)

type ChainClient interface {
	Listen(ctx context.Context, blocks chan<- int64) error
	GetLatestBlock(ctx context.Context) (*provider.Header, error)
	ChainCode() types.ChainCode
	AsEthClient() (*ethclient.Client, error)
	Close() error
}
