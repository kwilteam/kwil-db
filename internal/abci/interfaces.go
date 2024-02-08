package abci

import (
	"context"
	"math/big"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/transactions"

	// NOTE: we are defining interfaces, but using the types defined in the
	// packages that provide the concrete implementations. This is a bit
	// backwards, but it at least allows us to stub out for testing.

	"github.com/kwilteam/kwil-db/internal/abci/snapshots"
	"github.com/kwilteam/kwil-db/internal/txapp"
)

// KVStore is an interface for a basic key-value store
type KVStore interface {
	Get(key []byte) ([]byte, error)
	Set(key []byte, value []byte) error
}

// SnapshotModule is an interface for a struct that implements snapshotting
type SnapshotModule interface {
	// Checks if databases are to be snapshotted at a particular height
	IsSnapshotDue(height uint64) bool

	// Starts the snapshotting process, Locking databases need to be handled outside this fn
	CreateSnapshot(height uint64) error

	// Lists all the available snapshots in the snapshotstore and returns the snapshot metadata
	ListSnapshots() ([]snapshots.Snapshot, error)

	// Returns the snapshot chunk of index chunkId at a given height
	LoadSnapshotChunk(height uint64, format uint32, chunkID uint32) []byte
}

// DBBootstrapModule is an interface for a struct that implements bootstrapping
type DBBootstrapModule interface {
	// Offers a snapshot (metadata) to the bootstrapper and decides whether to accept the snapshot or not
	OfferSnapshot(snapshot *snapshots.Snapshot) error

	// Offers a snapshot Chunk to the bootstrapper, once all the chunks corresponding to the snapshot are received, the databases are restored from the chunks
	ApplySnapshotChunk(chunk []byte, index uint32) ([]uint32, snapshots.Status, error)

	// Signifies the end of the db restoration
	IsDBRestored() bool
}

// TxApp is an application that can process transactions.
// It has methods for beginning and ending blocks, applying transactions,
// and managing a mempool
type TxApp interface {
	// accounts -> string([]accound_identifier) : *big.Int(balance)
	GenesisInit(ctx context.Context, validators []*types.Validator, accounts map[string]*big.Int, initialHeight int64) error
	ApplyMempool(ctx context.Context, tx *transactions.Transaction) error
	// Begin signals that a new block has begun.
	Begin(ctx context.Context) error
	Finalize(ctx context.Context, blockHeight int64) (apphash []byte, validatorUpgrades []*types.Validator, err error)
	Commit(ctx context.Context) error
	Execute(ctx txapp.TxContext, tx *transactions.Transaction) *txapp.TxResponse
	ProposerTxs(ctx context.Context, txNonce uint64) ([]*transactions.Transaction, error)
	UpdateValidator(ctx context.Context, validator []byte, power int64) error
	GetValidators(ctx context.Context) ([]*types.Validator, error)
	AccountInfo(ctx context.Context, acctID []byte, getUncommitted bool) (balance *big.Int, nonce int64, err error)
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
