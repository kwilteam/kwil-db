package mempool

import (
	"p2p/node/types"
	"sync"
)

// mempool is an index of unconfirmed transactions

type Mempool struct {
	mtx      sync.RWMutex
	txns     map[types.Hash][]byte
	fetching map[types.Hash]bool
}

func New() *Mempool {
	return &Mempool{
		txns:     make(map[types.Hash][]byte),
		fetching: make(map[types.Hash]bool),
	}
}

func (mp *Mempool) Have(txid types.Hash) bool { // this is racy
	mp.mtx.RLock()
	defer mp.mtx.RUnlock()
	_, have := mp.txns[txid]
	return have
}

func (mp *Mempool) Store(txid types.Hash, raw []byte) {
	mp.mtx.Lock()
	defer mp.mtx.Unlock()
	if raw == nil {
		delete(mp.txns, txid)
		delete(mp.fetching, txid)
		return
	}
	mp.txns[txid] = raw
}

func (mp *Mempool) PreFetch(txid types.Hash) bool { // probably make node business
	mp.mtx.Lock()
	defer mp.mtx.Unlock()
	if _, have := mp.txns[txid]; have {
		return false // don't need it
	}

	if fetching := mp.fetching[txid]; fetching {
		return false // already getting it
	}
	mp.fetching[txid] = true

	return true // go get it
}

func (mp *Mempool) Size() int {
	mp.mtx.RLock()
	defer mp.mtx.RUnlock()
	return len(mp.txns)
}

func (mp *Mempool) Get(txid types.Hash) []byte {
	mp.mtx.RLock()
	defer mp.mtx.RUnlock()
	return mp.txns[txid]
}

func (mp *Mempool) ReapN(n int) ([]types.Hash, [][]byte) {
	mp.mtx.Lock()
	defer mp.mtx.Unlock()
	n = min(n, len(mp.txns))
	txids := make([]types.Hash, 0, n)
	txns := make([][]byte, 0, n)
	for txid, rawTx := range mp.txns {
		delete(mp.txns, txid)
		txids = append(txids, txid)
		txns = append(txns, rawTx)
		if len(txids) == cap(txids) {
			break
		}
	}
	return txids, txns
}

func (mp *Mempool) PeekN(n int) []types.NamedTx {
	mp.mtx.RLock()
	defer mp.mtx.RUnlock()
	n = min(n, len(mp.txns))
	txns := make([]types.NamedTx, 0, n)
	for txid, rawTx := range mp.txns {
		txns = append(txns, types.NamedTx{ID: txid, Tx: rawTx})
		if len(txns) == n {
			break
		}
	}

	return txns
}
