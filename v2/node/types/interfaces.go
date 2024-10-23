package types

type TxIndex interface {
	Get(Hash) []byte
	Store(Hash, []byte)
}

type BlockStore interface {
	Have(Hash) bool
	Get(Hash) (int64, []byte)
	Store(Hash, int64, []byte)
	PreFetch(Hash) bool // maybe app level instead
}

type MemPool interface {
	Size() int
	ReapN(int) ([]Hash, [][]byte)
	Get(Hash) []byte
	Store(Hash, []byte)
	FeedN(n int) <-chan NamedTx
	// Check([]byte)
	PreFetch(txid Hash) bool
}

type Execution interface {
	ExecBlock(blk *Block) (commit func() error, appHash Hash, err error)
}

type NamedTx struct {
	ID Hash
	Tx []byte
}
