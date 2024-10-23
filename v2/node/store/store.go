package store

import (
	"p2p/node/types"
	"sync"

	"github.com/dgraph-io/badger/v4"
)

type blockStore struct {
	mtx      sync.RWMutex
	idx      map[types.Hash]int64
	fetching map[types.Hash]bool

	memStore map[types.Hash][]byte // TODO: disk
	db       *badger.DB            // <-disk
}

func NewBlockStore(dir string) (*blockStore, error) {
	opts := badger.DefaultOptions(dir)
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	return &blockStore{
		idx:      make(map[types.Hash]int64),
		fetching: make(map[types.Hash]bool),
		memStore: make(map[types.Hash][]byte),
		db:       db,
	}, nil
}

func (bki *blockStore) Have(blkid types.Hash) bool { // this is racy
	bki.mtx.RLock()
	defer bki.mtx.RUnlock()
	_, have := bki.idx[blkid]
	return have
}

func (bki *blockStore) Store(blkid types.Hash, height int64, raw []byte) {
	bki.mtx.Lock()
	defer bki.mtx.Unlock()
	delete(bki.fetching, blkid)
	if height == -1 {
		delete(bki.idx, blkid)
		return
	}
	bki.idx[blkid] = height

	bki.memStore[blkid] = raw
	err := bki.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte("blk:"+blkid.String()), raw)
	})
	if err != nil {
		panic(err)
	}
}

func (bki *blockStore) PreFetch(blkid types.Hash) bool {
	bki.mtx.Lock()
	defer bki.mtx.Unlock()
	if _, have := bki.idx[blkid]; have {
		return false // don't need it
	}

	if fetching := bki.fetching[blkid]; fetching {
		return false // already getting it
	}
	bki.fetching[blkid] = true

	return true // go get it
}

func (bki *blockStore) size() int {
	bki.mtx.RLock()
	defer bki.mtx.RUnlock()
	return len(bki.idx)
}

func (bki *blockStore) Get(blkid types.Hash) (int64, []byte) {
	bki.mtx.RLock()
	defer bki.mtx.RUnlock()
	h, have := bki.idx[blkid]
	if !have {
		return -1, nil
	}
	// raw, have := bki.memStore[blkid]
	// if !have {
	// 	return -1, nil
	// }
	var raw []byte
	err := bki.db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("blk:" + blkid.String()))
		if err != nil {
			return err
		}
		raw, err = item.ValueCopy(nil)
		return err
	})
	if err != nil {
		panic(err)
	}
	return h, raw
}

type transactionIndex struct {
	mtx   sync.RWMutex
	txids map[types.Hash][]byte
}

func NewTransactionIndex() *transactionIndex {
	return &transactionIndex{
		txids: make(map[types.Hash][]byte),
	}
}

func (txi *transactionIndex) Have(txid types.Hash) bool { // this is racy
	txi.mtx.RLock()
	defer txi.mtx.RUnlock()
	_, have := txi.txids[txid]
	return have
}

func (txi *transactionIndex) Store(txid types.Hash, raw []byte) {
	txi.mtx.Lock()
	defer txi.mtx.Unlock()
	if raw == nil {
		delete(txi.txids, txid)
		return
	}
	txi.txids[txid] = raw
}

func (txi *transactionIndex) Get(txid types.Hash) []byte {
	txi.mtx.RLock()
	defer txi.mtx.RUnlock()
	return txi.txids[txid]
}
