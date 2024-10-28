package types

import (
	"context"
	"errors"
)

var ErrNotFound = errors.New("not found")

// type TxIndex interface {
// 	Get(Hash) (int64, []byte)
// 	Have(Hash) bool
// 	Store(Hash, int64, []byte)
// }

type BlockStore interface {
	TxGetter

	Best() (int64, Hash, Hash)
	Have(Hash) bool
	Get(Hash) (*Block, Hash, error)
	// GetRaw(Hash) (int64, []byte)
	GetByHeight(int64) (Hash, *Block, Hash, error) // note: we can impl GetBlockHeader easily too
	// GetRawByHeight(int64) (Hash, []byte)
	Store(*Block, Hash) error
	PreFetch(Hash) (bool, func()) // should be app level instead
}

type TxGetter interface {
	GetTx(Hash) (int64, []byte, error)
	HaveTx(Hash) bool
}

type MemPool interface {
	Size() int
	ReapN(int) ([]Hash, [][]byte) // Reap(n int, maxBts int) ([]Hash, [][]byte)
	Get(Hash) []byte
	Store(Hash, []byte)
	PeekN(n int) []NamedTx
	// Check([]byte)
	PreFetch(txid Hash) bool // should be app level instead
}

type QualifiedBlock struct { // basically just caches the hash
	Block    *Block
	Hash     Hash
	Proposed bool
	AppHash  *Hash
}

type Execution interface {
	ExecBlock(blk *Block) (commit func(context.Context, bool) error, appHash Hash, res []TxResult, err error)
}

type NamedTx struct {
	Hash Hash
	Tx   []byte
}
