package client

import (
	"context"
	"kwil/pkg/chain/contracts"
	provider "kwil/pkg/chain/provider/dto"
	"kwil/pkg/chain/types"
)

type ChainClient interface {
	Listen(ctx context.Context, blocks chan<- int64) error
	GetLatestBlock(ctx context.Context) (*provider.Header, error)
	ChainCode() types.ChainCode
	Close() error
	Contracts() contracts.Contracter
}
