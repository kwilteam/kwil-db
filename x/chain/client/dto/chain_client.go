package dto

import (
	"context"
	provider "kwil/x/chain/provider/dto"
)

type ChainClient interface {
	Listen(ctx context.Context, blocks chan<- provider.Header) error
	GetLatestBlock(ctx context.Context) (*provider.Header, error)
}
