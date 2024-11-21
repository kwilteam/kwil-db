package migrations

import (
	"context"

	"github.com/kwilteam/kwil-db/common/sql"
	"github.com/kwilteam/kwil-db/internal/statesync"
	"github.com/kwilteam/kwil-db/internal/txapp"
)

type Snapshotter interface {
	CreateSnapshot(ctx context.Context, height uint64, snapshotID string, schemas, excludedTables []string, excludedTableData []string) error
	LoadSnapshotChunk(height uint64, format uint32, chunkIdx uint32) ([]byte, error)
	ListSnapshots() []*statesync.Snapshot
}

type Database interface {
	sql.TxMaker
	sql.ReadTxMaker
	sql.SnapshotTxMaker
}

type SpendTracker interface {
	GetBlockSpends() []*txapp.Spend
}
