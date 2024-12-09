package types

import (
	"context"
	"errors"

	"github.com/kwilteam/kwil-db/core/types"
)

var ErrNotFound = errors.New("not found")

type BlockStore interface {
	BlockGetter
	BlockStorer
	TxGetter
	BlockResultsStorer

	Best() (int64, Hash, Hash)

	PreFetch(Hash) (bool, func()) // should be app level instead (TODO: remove)

	Close() error
}

type BlockGetter interface {
	Have(Hash) bool
	Get(Hash) (*Block, Hash, error)
	GetByHeight(int64) (Hash, *Block, Hash, error) // note: we can impl GetBlockHeader easily too
}

type RawGetter interface {
	GetRaw(Hash) ([]byte, error)
	GetRawByHeight(int64) (Hash, []byte)
}

type BlockStorer interface {
	Store(*Block, Hash) error
}

type BlockResultsStorer interface {
	StoreResults(hash Hash, results []types.TxResult) error
	Results(hash Hash) ([]types.TxResult, error)
	Result(hash Hash, idx uint32) (*types.TxResult, error)
}

type TxGetter interface {
	GetTx(txHash types.Hash) (raw []byte, height int64, blkHash types.Hash, blkIdx uint32, err error)
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
	ExecBlock(blk *Block) (commit func(context.Context, bool) error, appHash Hash, res []types.TxResult, err error)
}

type NamedTx struct {
	Hash Hash
	Tx   []byte
}
