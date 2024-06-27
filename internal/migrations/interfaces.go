package migrations

import (
	"context"

	"github.com/kwilteam/kwil-db/common/sql"
	"github.com/kwilteam/kwil-db/internal/statesync"
)

type Snapshotter interface {
	CreateSnapshot(ctx context.Context, height uint64, snapshotID string) error
	LoadSnapshotChunk(height uint64, format uint32, chunkIdx uint32) ([]byte, error)
	ListSnapshots() []*statesync.Snapshot
}

type Database interface {
	sql.SnapshotTxMaker
}
