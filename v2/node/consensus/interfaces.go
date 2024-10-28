package consensus

import (
	"p2p/node/types"
)

type Mempool interface {
	ReapN(maxSize int) ([]types.Hash, [][]byte)
	Store(txid types.Hash, tx []byte)
}

type BlockStore interface {
	// GetBlockByHeight(height int64) (types.Block, error)
	Store(block *types.Block, appHash types.Hash) error
	Have(blkid types.Hash) bool
}

type BlockExecutor interface {
	Execute(tx []byte) (txResult, error)
	Commit() error
}
