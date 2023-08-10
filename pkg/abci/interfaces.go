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
	IsSnapshotDue(height uint64) bool
	CreateSnapshot(height uint64) error
	ListSnapshots() ([]snapshots.Snapshot, error)
	LoadSnapshotChunk(height uint64, format uint32, chunkID uint32) []byte
}

type DBBootstrapModule interface {
	//(ctx context.Context, snapshot []Snapshot) error
	ApplySnapshotChunk(chunk []byte, index uint32) ([]uint32, error)
	OfferSnapshot(snapshot *snapshots.Snapshot) error
	IsDBRestored() bool
}
