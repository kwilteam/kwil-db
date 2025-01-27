package types

import (
	"errors"
	"time"

	"github.com/kwilteam/kwil-db/core/types"
)

var ErrNotFound = types.ErrNotFound

var (
	HashBytes          = types.HashBytes
	ErrTxNotFound      = errors.New("tx not available")
	ErrTxAlreadyExists = errors.New("transaction already exists")
	ErrBlkNotFound     = errors.New("block not available")
	ErrStillProcessing = errors.New("block still being executed")
	ErrNoResponse      = errors.New("stream closed without response")
)

const HashLen = types.HashLen

type Hash = types.Hash
type BlockStore interface {
	BlockGetter
	BlockStorer
	TxGetter
	BlockResultsStorer

	Best() (height int64, blkHash, appHash Hash, stamp time.Time)

	PreFetch(Hash) (bool, func()) // should be app level instead (TODO: remove)

	Close() error
}

type BlockGetter interface {
	Have(Hash) bool
	Get(Hash) (*types.Block, *CommitInfo, error)
	GetByHeight(int64) (Hash, *types.Block, *CommitInfo, error) // note: we can impl GetBlockHeader easily too
}

type RawGetter interface {
	GetRaw(Hash) ([]byte, error)
	GetRawByHeight(int64) (Hash, []byte)
}

type BlockStorer interface {
	Store(*types.Block, *CommitInfo) error
}

type BlockResultsStorer interface {
	StoreResults(hash Hash, results []types.TxResult) error
	Results(hash Hash) ([]types.TxResult, error)
	Result(hash Hash, idx uint32) (*types.TxResult, error)
}

type TxGetter interface {
	GetTx(txHash types.Hash) (raw *types.Transaction, height int64, blkHash types.Hash, blkIdx uint32, err error)
	HaveTx(Hash) bool
}

type MemPool interface {
	Size() int
	ReapN(int) []NamedTx
	Get(Hash) *types.Transaction
	Remove(Hash)
	Store(Hash, *types.Transaction)
	PeekN(n int) []NamedTx
	// Check([]byte)
	PreFetch(txid Hash) bool // should be app level instead
}

type QualifiedBlock struct { // basically just caches the hash
	Block    *types.Block
	Hash     Hash
	Proposed bool
	AppHash  *Hash
}

type NamedTx struct {
	Hash Hash
	Tx   *types.Transaction
}
