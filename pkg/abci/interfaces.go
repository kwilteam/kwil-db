package abci

import (
	"context"

	modDataset "github.com/kwilteam/kwil-db/pkg/modules/datasets"
	modVal "github.com/kwilteam/kwil-db/pkg/modules/validators"

	"github.com/kwilteam/kwil-db/pkg/engine/types"
	"github.com/kwilteam/kwil-db/pkg/snapshots"
	"github.com/kwilteam/kwil-db/pkg/transactions"
	"github.com/kwilteam/kwil-db/pkg/validators"
)

type DatasetsModule interface {
	Deploy(ctx context.Context, schema *types.Schema, tx *transactions.Transaction) (*modDataset.ExecutionResponse, error)
	Drop(ctx context.Context, dbid string, tx *transactions.Transaction) (*modDataset.ExecutionResponse, error)
	Execute(ctx context.Context, dbid string, action string, args [][]any, tx *transactions.Transaction) (*modDataset.ExecutionResponse, error)
}

// ValidatorModule handles the processing of validator approve/join/leave
// transactions, punishment, preparation of validator updates to be applied when
// a block is finalized, and performing transaction accounting (e.g. fee and
// nonce checks).
//
// NOTE: this may be premature abstraction since we are designing this function
// for the needs for an abci/types.Application, yet using standard or Kwil
// types. But if a different blockchain package is used, this is unlikely to be
// what it needs, but it is as generic as possible.
type ValidatorModule interface {
	// GenesisInit configures the initial set of validators for the genesis
	// block. This is only called once for a new chain.
	GenesisInit(ctx context.Context, vals []*validators.Validator) error

	// CurrentSet returns the current validator set. This is used on app
	// construction to initialize the addr=>pubkey mapping.
	CurrentSet(ctx context.Context) ([]*validators.Validator, error)

	// Punish may be used at the start of block processing when byzantine
	// validators are listed by the consensus client (no transaction).
	Punish(ctx context.Context, validator []byte, power int64) error

	// Join creates a join request for a prospective validator.
	Join(ctx context.Context, joiner []byte, power int64, tx *transactions.Transaction) (*modVal.ExecutionResponse, error)
	// Leave processes a leave request for a validator.
	Leave(ctx context.Context, joiner []byte, tx *transactions.Transaction) (*modVal.ExecutionResponse, error)
	// Approve records an approval transaction from a current validator. The
	// approver is the tx Sender.
	Approve(ctx context.Context, joiner []byte, tx *transactions.Transaction) (*modVal.ExecutionResponse, error)

	// Finalize is used at the end of block processing to retrieve the validator
	// updates to be provided to the consensus client for the next block. This
	// is not idempotent. The modules working list of updates is reset until
	// subsequent join/approves are processed for the next block.
	Finalize(ctx context.Context) []*validators.Validator // end of block processing requires providing list of updates to the node's consensus client
}

// AtomicCommitter is an interface for a struct that implements atomic commits across multiple stores
type AtomicCommitter interface {
	ClearWal(ctx context.Context) error
	Begin(ctx context.Context) error
	ID(ctx context.Context) ([]byte, error)
	Commit(ctx context.Context, applyCallback func(error)) error
}

// KVStore is an interface for a basic key-value store
type KVStore interface {
	Get(key []byte) ([]byte, error)
	Set(key []byte, value []byte) error
	Delete(key []byte) error
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
