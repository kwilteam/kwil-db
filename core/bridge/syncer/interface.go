package syncer

import (
	"context"

	"github.com/kwilteam/kwil-db/core/types/chain"
)

// TODO: Come up with a better name for this interface and the package

type BlockSyncer interface {
	Listen(ctx context.Context, blocks chan<- int64) error
	LatestBlock(ctx context.Context) (*chain.Header, error)
	Close() error
}
