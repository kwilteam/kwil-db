package abci

import (
	"context"
	"math/big"

	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/common/sql"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/transactions"

	// NOTE: we are defining interfaces, but using the types defined in the
	// packages that provide the concrete implementations. This is a bit
	// backwards, but it at least allows us to stub out for testing.

	"github.com/kwilteam/kwil-db/internal/statesync"
	"github.com/kwilteam/kwil-db/internal/txapp"
)

// SnapshotModule is an interface for a struct that implements snapshotting
type SnapshotModule interface {
	// Lists all the available snapshots in the snapshotstore and returns the snapshot metadata
	ListSnapshots() []*statesync.Snapshot

	// Returns the snapshot chunk of index chunkId at a given height
	LoadSnapshotChunk(height uint64, format uint32, chunkID uint32) ([]byte, error)

	// CreateSnapshot creates a snapshot of the current state.
	CreateSnapshot(ctx context.Context, height uint64, snapshotID string) error

	// IsSnapshotDue returns true if a snapshot is due at the given height.
	IsSnapshotDue(height uint64) bool
}

// DBBootstrapModule is an interface for a struct that implements bootstrapping
type StateSyncModule interface {
	// Offers a snapshot (metadata) to the bootstrapper and decides whether to accept the snapshot or not
	OfferSnapshot(ctx context.Context, snapshot *statesync.Snapshot) error

	// Offers a snapshot Chunk to the bootstrapper, once all the chunks corresponding to the snapshot are received, the databases are restored from the chunks
	ApplySnapshotChunk(ctx context.Context, chunk []byte, index uint32) (bool, error)
}

// TxApp is an application that can process transactions.
// It has methods for beginning and ending blocks, applying transactions,
// and managing a mempool
type TxApp interface {
	AccountInfo(ctx context.Context, db sql.DB, acctID []byte, getUnconfirmed bool) (balance *big.Int, nonce int64, err error)
	ApplyMempool(ctx *common.TxContext, db sql.DB, tx *transactions.Transaction) error
	Begin(ctx context.Context, height int64) error
	Commit(ctx context.Context)
	Execute(ctx txapp.TxContext, db sql.DB, tx *transactions.Transaction) *txapp.TxResponse
	Finalize(ctx context.Context, db sql.DB, block *common.BlockContext) (finalValidators []*types.Validator, err error)
	GenesisInit(ctx context.Context, db sql.DB, validators []*types.Validator, genesisAccounts []*types.Account, initialHeight int64, chain *common.ChainContext) error
	GetValidators(ctx context.Context, db sql.DB) ([]*types.Validator, error)
	ProposerTxs(ctx context.Context, db sql.DB, txNonce uint64, maxTxsSize int64, block *common.BlockContext) ([][]byte, error)
	Reload(ctx context.Context, db sql.DB) error
	UpdateValidator(ctx context.Context, db sql.DB, validator []byte, power int64) error
	Price(ctx context.Context, db sql.DB, tx *transactions.Transaction, chainCtx *common.ChainContext) (*big.Int, error)
}

// ConsensusParams returns kwil specific consensus parameters.
// I made this its own separate interface (instead of adding it to AbciConfig)
// since this should be dynamic and changeable via voting.
type ConsensusParams interface {
	// VotingPeriod is the vote expiration period
	// for validator joins and resolutions.
	// We may want these to be separate in the future.
	VotingPeriod() int64
}

// DB is the interface for the main SQL database. All queries must be executed
// from within a transaction. A DB can create read transactions or the special
// two-phase outer write transaction.
type DB interface {
	sql.TxMaker // for out-of-consensus writes e.g. setup and meta table writes
	sql.PreparedTxMaker
	sql.ReadTxMaker
	sql.SnapshotTxMaker
}
