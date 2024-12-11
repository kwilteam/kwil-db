package blockprocessor

import (
	"context"
	"math/big"

	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/core/types"
	ktypes "github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/node/snapshotter"
	"github.com/kwilteam/kwil-db/node/txapp"
	"github.com/kwilteam/kwil-db/node/types/sql"
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
}

type TxApp interface {
	Begin(ctx context.Context, height int64) error
	Execute(ctx *common.TxContext, db sql.DB, tx *ktypes.Transaction) *txapp.TxResponse
	Finalize(ctx context.Context, db sql.DB, block *common.BlockContext) (finalValidators []*ktypes.Validator, err error)
	Commit() error
	GenesisInit(ctx context.Context, db sql.DB, validators []*ktypes.Validator, genesisAccounts []*ktypes.Account, initialHeight int64, chain *common.ChainContext) error
	ApplyMempool(ctx *common.TxContext, db sql.DB, tx *types.Transaction) error

	Price(ctx context.Context, dbTx sql.DB, tx *ktypes.Transaction, chainContext *common.ChainContext) (*big.Int, error)
	AccountInfo(ctx context.Context, dbTx sql.DB, identifier []byte, pending bool) (balance *big.Int, nonce int64, err error)
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
