package consensus

import (
	"context"

	ktypes "github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/node/mempool"
	"github.com/kwilteam/kwil-db/node/types"
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

type Mempool interface {
	PeekN(maxSize int) []types.NamedTx
	Remove(txid types.Hash)
	RecheckTxs(ctx context.Context, checkFn mempool.CheckFn)
}

// BlockStore includes both txns and blocks
type BlockStore interface {
	// GetBlockByHeight(height int64) (types.Block, error)
	Best() (int64, types.Hash, types.Hash)
	Store(block *ktypes.Block, appHash types.Hash) error
	// Have(blkid types.Hash) bool
	Get(blkid types.Hash) (*ktypes.Block, types.Hash, error)
	GetByHeight(height int64) (types.Hash, *ktypes.Block, types.Hash, error)
	StoreResults(hash types.Hash, results []ktypes.TxResult) error
	// Results(hash types.Hash) ([]types.TxResult, error)
}

type BlockProcessor interface {
	InitChain(ctx context.Context) (int64, []byte, error)
	ExecuteBlock(ctx context.Context, req *ktypes.BlockExecRequest) (*ktypes.BlockExecResult, error)
	Commit(ctx context.Context, req *ktypes.CommitRequest) error
	Rollback(ctx context.Context, height int64, appHash ktypes.Hash) error
	Close() error

	CheckTx(ctx context.Context, tx *ktypes.Transaction, recheck bool) error

	GetValidators() []*ktypes.Validator
}
