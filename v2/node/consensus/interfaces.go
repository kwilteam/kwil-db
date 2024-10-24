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
	Store(blkid types.Hash, height int64, raw []byte)
	Have(blkid types.Hash) bool
}

type Indexer interface {
	Store(txid types.Hash, tx []byte)
}

type BlockExecutor interface {
	Execute(tx []byte) (txResult, error)
	Commit() error
}
