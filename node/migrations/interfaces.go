package migrations

import (
	"context"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/node/accounts"
	"github.com/kwilteam/kwil-db/node/snapshotter"
	"github.com/kwilteam/kwil-db/node/types/sql"
)

type Snapshotter interface {
	CreateSnapshot(ctx context.Context, height uint64, snapshotID string, schemas, excludedTables []string, excludedTableData []string) error
	LoadSnapshotChunk(height uint64, format uint32, chunkIdx uint32) ([]byte, error)
	ListSnapshots() []*snapshotter.Snapshot
}

type Database interface {
	sql.TxMaker
	sql.ReadTxMaker
	sql.SnapshotTxMaker
}

type Accounts interface {
	GetBlockSpends() []*accounts.Spend
}

type Validators interface {
	GetValidators() []*types.Validator
}
