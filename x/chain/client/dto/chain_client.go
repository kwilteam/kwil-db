package dto

import (
	"context"
	provider "kwil/x/chain/provider/dto"
)

type ChainClient interface {
	Listen(ctx context.Context, blocks chan<- int64) error
	GetLatestBlock(ctx context.Context) (*provider.Header, error)
}
