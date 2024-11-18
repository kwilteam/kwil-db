package consensus

import (
	"context"

	"kwil/node/txapp"
	"kwil/node/types"
	"kwil/node/types/sql"
	ktypes "kwil/types"
)

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
}

type Validators interface {
	GetValidators() []*ktypes.Validator
	ValidatorUpdates() map[string]*ktypes.Validator
}

type TxApp interface {
	Begin(ctx context.Context, height int64) error
	Execute(ctx *ktypes.TxContext, db sql.DB, tx *ktypes.Transaction) *txapp.TxResponse
	Finalize(ctx context.Context, db sql.DB, block *ktypes.BlockContext) (finalValidators []*ktypes.Validator, err error)
	Commit() error
}

// Question:
// Blockstore: Blocks, Txs, Results, AppHash (for each block)
// What is replaying a block from the blockstore? -> do we still have the results and apphash?
// Do we overwrite the results? or skip adding it to the blockstore?
