package consensus

import (
	"context"

	"github.com/kwilteam/kwil-db/common"
	ktypes "github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/node/txapp"
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
	Store(block *types.Block, appHash types.Hash) error
	// Have(blkid types.Hash) bool
	Get(blkid types.Hash) (*types.Block, types.Hash, error)
	GetByHeight(height int64) (types.Hash, *types.Block, types.Hash, error)
	StoreResults(hash types.Hash, results []ktypes.TxResult) error
	// Results(hash types.Hash) ([]types.TxResult, error)
}

type BlockExecutor interface {
	Execute(ctx context.Context, tx []byte) (ktypes.TxResult, error)
	Precommit() (types.Hash, error)
	Commit(func() error) error
	Rollback() error
}

type Accounts interface {
	Updates() []*ktypes.Account
}

type Validators interface {
	GetValidators() []*ktypes.Validator
	ValidatorUpdates() map[string]*ktypes.Validator
}

type TxApp interface {
	Begin(ctx context.Context, height int64) error
	Execute(ctx *common.TxContext, db sql.DB, tx *ktypes.Transaction) *txapp.TxResponse
	Finalize(ctx context.Context, db sql.DB, block *common.BlockContext) (finalValidators []*ktypes.Validator, err error)
	Commit() error
}

// Question:
// Blockstore: Blocks, Txs, Results, AppHash (for each block)
// What is replaying a block from the blockstore? -> do we still have the results and apphash?
// Do we overwrite the results? or skip adding it to the blockstore?
