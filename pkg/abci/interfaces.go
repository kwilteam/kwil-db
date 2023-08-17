package abci

import (
	"context"

	"github.com/kwilteam/kwil-db/pkg/engine/types"
	"github.com/kwilteam/kwil-db/pkg/snapshots"
	"github.com/kwilteam/kwil-db/pkg/transactions"
)

type DatasetsModule interface {
	Deploy(ctx context.Context, schema *types.Schema, tx *transactions.Transaction) (*transactions.TransactionStatus, error)
	Drop(ctx context.Context, dbid string, tx *transactions.Transaction) (*transactions.TransactionStatus, error)
	Execute(ctx context.Context, dbid string, action string, args [][]any, tx *transactions.Transaction) (*transactions.TransactionStatus, error)
}

// Should be implemented in pkg/modules/validators
type ValidatorModule interface {
	ValidatorJoin(ctx context.Context, address string, tx *transactions.Transaction) (*transactions.TransactionStatus, error)
	ValidatorApprove(ctx context.Context, address string, approvedBy string, tx *transactions.Transaction) (*transactions.TransactionStatus, error)
}

// AtomicCommitter is an interface for a struct that implements atomic commits across multiple stores
type AtomicCommitter interface {
	Begin(ctx context.Context) error
	Commit(ctx context.Context, applyCallback func(error)) (commitId []byte, err error)
}

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

type DBBootstrapModule interface {
	// Offers a snapshot (metadata) to the bootstrapper and decides whether to accept the snapshot or not
	OfferSnapshot(snapshot *snapshots.Snapshot) error

	// Offers a snapshot Chunk to the bootstrapper, once all the chunks corresponding to the snapshot are received, the databases are restored from the chunks
	ApplySnapshotChunk(chunk []byte, index uint32) ([]uint32, snapshots.Status, error)

	// Signifies the end of the db restoration
	IsDBRestored() bool
}
