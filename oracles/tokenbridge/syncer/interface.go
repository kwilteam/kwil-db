package syncer

import (
	"context"
	"math/big"

	"github.com/kwilteam/kwil-db/core/types/chain"
)

// TODO: Come up with a better name for this interface and the package

type BlockSyncer interface {
	Listen(ctx context.Context, blocks chan<- int64) error
	LatestBlock(ctx context.Context) (*chain.Header, error)
	Close() error
}

type ChainClient interface {
	Close() error
	HeaderByNumber(ctx context.Context, number *big.Int) (*chain.Header, error)
	SubscribeNewHead(ctx context.Context, ch chan<- chain.Header) (chain.Subscription, error)
}
