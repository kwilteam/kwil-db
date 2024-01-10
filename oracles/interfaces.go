package oracles

import (
	"context"

	"github.com/kwilteam/kwil-db/internal/sql"
)

type EventStore interface {
	KV(prefix []byte) sql.KVStore
	Store(ctx context.Context, data []byte, eventType string) error
}
