package blockprocessor

import (
	"context"
	"math/big"

	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/core/types"
	ktypes "github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/node/migrations"
	"github.com/kwilteam/kwil-db/node/snapshotter"
	"github.com/kwilteam/kwil-db/node/txapp"
	"github.com/kwilteam/kwil-db/node/types/sql"
	"github.com/kwilteam/kwil-db/node/voting"
)

// DB is the interface for the main SQL database. All queries must be executed
// from within a transaction. A DB can create read transactions or the special
// two-phase outer write transaction.
type DB interface {
	sql.TxMaker // for out-of-consensus writes e.g. setup and meta table writes
	sql.PreparedTxMaker
	sql.ReadTxMaker
	sql.SnapshotTxMaker
}

type Accounts interface {
	Updates() []*ktypes.Account
}

type ValidatorModule interface {
	GetValidators() []*ktypes.Validator
	ValidatorUpdates() map[string]*ktypes.Validator
	LoadValidatorSet(ctx context.Context, db sql.Executor) error
}

type TxApp interface {
	Begin(ctx context.Context, height int64) error
	Execute(ctx *common.TxContext, db sql.DB, tx *ktypes.Transaction) *txapp.TxResponse
	Finalize(ctx context.Context, db sql.DB, block *common.BlockContext) (approvedJoins, expiredJoins []*types.AccountID, err error)
	Commit() error
	Rollback()
	GenesisInit(ctx context.Context, db sql.DB, validators []*ktypes.Validator, genesisAccounts []*ktypes.Account, initialHeight int64, initialDBOwner string, chain *common.ChainContext) error
	ApplyMempool(ctx *common.TxContext, db sql.DB, tx *types.Transaction) error

	Price(ctx context.Context, dbTx sql.DB, tx *ktypes.Transaction, chainContext *common.ChainContext) (*big.Int, error)
	AccountInfo(ctx context.Context, dbTx sql.DB, identifier *types.AccountID, pending bool) (balance *big.Int, nonce int64, err error)
}

// Question:
// Blockstore: Blocks, Txs, Results, AppHash (for each block)
// What is replaying a block from the blockstore? -> do we still have the results and apphash?
// Do we overwrite the results? or skip adding it to the blockstore?

// SnapshotModule is an interface for a struct that implements snapshotting
type SnapshotModule interface {
	// Lists all the available snapshots in the snapshotstore and returns the snapshot metadata
	ListSnapshots() []*snapshotter.Snapshot

	// Returns the snapshot chunk of index chunkId at a given height
	LoadSnapshotChunk(height uint64, format uint32, chunkID uint32) ([]byte, error)

	// CreateSnapshot creates a snapshot of the current state.
	CreateSnapshot(ctx context.Context, height uint64, snapshotID string, schemas, excludedTables []string, excludeTableData []string) error

	// IsSnapshotDue returns true if a snapshot is due at the given height.
	IsSnapshotDue(height uint64) bool

	Enabled() bool
}

// EventStore allows the BlockProcessor to read events from the event store.
type EventStore interface {
	// GetUnbroadcastedEvents filters out the events observed by the validator
	// that are not previously broadcasted.
	GetUnbroadcastedEvents(ctx context.Context) ([]*types.UUID, error)

	// MarkBroadcasted marks list of events as broadcasted.
	MarkBroadcasted(ctx context.Context, ids []*types.UUID) error
}

var (
	// getEvents gets all events, even if they have been
	// marked received
	getEvents = voting.GetEvents
)

type MigratorModule interface {
	NotifyHeight(ctx context.Context, block *common.BlockContext, db migrations.Database, tx sql.Executor) error
	StoreChangesets(height int64, changes <-chan any) error
	PersistLastChangesetHeight(ctx context.Context, tx sql.Executor, height int64) error
	GetMigrationMetadata(ctx context.Context, status types.MigrationStatus) (*types.MigrationMetadata, error)
}

type BlockStore interface {
	GetByHeight(height int64) (types.Hash, *ktypes.Block, *ktypes.CommitInfo, error)
}
