package client

import (
	"context"
	provider "kwil/x/chain/provider/dto"
	"kwil/x/chain/types"

	"github.com/ethereum/go-ethereum/ethclient"
)

type ChainClient interface {
	Listen(ctx context.Context, blocks chan<- int64) error
	GetLatestBlock(ctx context.Context) (*provider.Header, error)
	ChainCode() types.ChainCode
	AsEthClient() (*ethclient.Client, error)
}
