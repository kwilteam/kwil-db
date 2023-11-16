package tokenbridge

import (
	"context"

	"github.com/kwilteam/kwil-db/core/types/chain"
)

type EventStore interface {
	AddLocalEvent(ctx context.Context, event *chain.Event) error
	AddExternalEvent(ctx context.Context, event *chain.Event) error
	LastProcessedBlock(ctx context.Context) (int64, error)
	SetLastProcessedBlock(ctx context.Context, height int64) error
}
