package consensus

import (
	"context"

	ktypes "github.com/kwilteam/kwil-db/core/types"
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
	ReapN(maxSize int) ([]types.Hash, [][]byte)
	Store(txid types.Hash, tx []byte)
}

// BlockStore includes both txns and blocks
type BlockStore interface {
	// GetBlockByHeight(height int64) (types.Block, error)
	Best() (int64, types.Hash, types.Hash)
	Store(block *types.Block, appHash types.Hash) error
	// Have(blkid types.Hash) bool
	Get(blkid types.Hash) (*types.Block, types.Hash, error)
	GetByHeight(height int64) (types.Hash, *types.Block, types.Hash, error)
	StoreResults(hash types.Hash, results []ktypes.TxResult) error
	// Results(hash types.Hash) ([]types.TxResult, error)
}

type BlockProcessor interface {
	InitChain(ctx context.Context, req *ktypes.InitChainRequest) error
	ExecuteBlock(ctx context.Context, req *ktypes.BlockExecRequest) (*ktypes.BlockExecResult, error)
	Commit(ctx context.Context, height int64, appHash types.Hash, syncing bool) error
	Close() error
	Rollback(ctx context.Context, height int64, appHash ktypes.Hash) error
}
