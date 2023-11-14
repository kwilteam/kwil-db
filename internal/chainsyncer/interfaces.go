package chainsyncer

import (
	"context"

	"github.com/kwilteam/kwil-db/core/types/chain"
)

type EventStore interface {
	AddLocalEvent(ctx context.Context, event *chain.Event) error
}
