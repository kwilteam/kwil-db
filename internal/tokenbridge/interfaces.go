package tokenbridge

import (
	"context"
	"math/big"

	"github.com/kwilteam/kwil-db/core/types/chain"
)

type EventStore interface {
	AddLocalEvent(ctx context.Context, event *chain.Event) error
	LastProcessedBlock(ctx context.Context) (int64, error)
	SetLastProcessedBlock(ctx context.Context, height int64) error
}

type DepositsModule interface {
	AddDeposit(ctx context.Context, eventID string, spender string, amount *big.Int, observer string) error
	LastProcessedBlock(ctx context.Context) (int64, error)
	SetLastProcessedBlock(ctx context.Context, height int64) error
}
