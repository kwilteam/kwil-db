package abci

import (
	"context"
	"math/big"

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
	// GenesisInit is used outside of the consensus thread, creating it's own
	// transient outer DB transaction.
	GenesisInit(ctx context.Context, validators []*types.Validator, accounts []*types.Account, initialHeight int64, appHash []byte) error

	// Read-only methods. Do NOT use these from consensus connection methods.
	ChainInfo(ctx context.Context) (int64, []byte, error)
	GetValidators(ctx context.Context) ([]*types.Validator, error)
	AccountInfo(ctx context.Context, acctID []byte, unconfirmed bool) (balance *big.Int, nonce int64, err error)
	ApplyMempool(ctx context.Context, tx *transactions.Transaction) error

	// ProposerTxs is used when a validator prepares a block proposal This is
	// prior to the Begin->Commit cycle, and as such does not use the DB
	// transaction or a writer connection.
	ProposerTxs(ctx context.Context, txNonce uint64, maxTxSize int64, proposerAddr []byte) ([][]byte, error)

	// Begin signals that a new block has begun. The following methods are
	// expected to use either an encompassing DB transaction started with Begin
	// and ended with Commit.
	Begin(ctx context.Context) error
	Execute(ctx txapp.TxContext, tx *transactions.Transaction) *txapp.TxResponse
	UpdateValidator(ctx context.Context, validator []byte, power int64) error
	Finalize(ctx context.Context, blockHeight int64) (appHash []byte, validatorUpgrades []*types.Validator, err error)
	Commit(ctx context.Context) error

	// ConsensusAccountInfo and ConsensusValidators are used in two different
	// contexts, but always from the ABCI consensus connection. During block
	// execution (between Begin and Commit) the active write transaction is used
	// to read uncommitted data. From PrepareProposal and ProcessProposal, an
	// ephemeral read-only transaction is created.
	ConsensusAccountInfo(ctx context.Context, acctID []byte) (balance *big.Int, nonce int64, err error)
	ConsensusValidators(ctx context.Context) ([]*types.Validator, error)

	// Reload reloads the state of engine and txapp.
	Reload(ctx context.Context) error
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
