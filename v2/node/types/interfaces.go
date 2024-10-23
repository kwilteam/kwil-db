package types

type TxIndex interface {
	Get(Hash) []byte
	Store(Hash, []byte)
}

type BlockStore interface {
	Have(Hash) bool
	Get(Hash) (int64, []byte)
	Store(Hash, int64, []byte)
	PreFetch(Hash) bool // should be app level instead
}

type MemPool interface {
	Size() int
	ReapN(int) ([]Hash, [][]byte)
	Get(Hash) []byte
	Store(Hash, []byte)
	FeedN(n int) <-chan NamedTx
	// Check([]byte)
	PreFetch(txid Hash) bool // should be app level instead
}

type QualifiedBlock struct { // basically just caches the hash
	Block *Block
	Hash  Hash
}

type TxResult struct {
	Code   uint16
	Log    string
	Events []Event
}

type Event struct{}

type Execution interface {
	ExecBlock(blk *Block) (commit func() error, appHash Hash, res []TxResult, err error)
}

type NamedTx struct {
	ID Hash
	Tx []byte
}
