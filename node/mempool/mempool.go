package mempool

import (
	"context"
	"slices"
	"sync"

	ktypes "github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/node/types"
)

// mempool is an index of unconfirmed transactions

type Mempool struct {
	mtx      sync.RWMutex
	txns     map[types.Hash]*ktypes.Transaction
	txQ      []types.NamedTx
	fetching map[types.Hash]bool
	// acctTxns map[string][]types.NamedTx
}

func New() *Mempool {
	return &Mempool{
		txns:     make(map[types.Hash]*ktypes.Transaction),
		fetching: make(map[types.Hash]bool),
	}
}

func (mp *Mempool) Have(txid types.Hash) bool { // this is racy
	mp.mtx.RLock()
	defer mp.mtx.RUnlock()
	_, have := mp.txns[txid]
	return have
}

func (mp *Mempool) Remove(txid types.Hash) {
	mp.mtx.Lock()
	defer mp.mtx.Unlock()
	mp.remove(txid)
}

func (mp *Mempool) remove(txid types.Hash) {
	idx := slices.IndexFunc(mp.txQ, func(a types.NamedTx) bool {
		return a.Hash == txid
	})
	if idx == -1 {
		return
	}
	mp.txQ = slices.Delete(mp.txQ, idx, idx+1) // remove txQ[idx]
	delete(mp.txns, txid)
}

// Store adds a transaction to the mempool. If the transaction is already in the
// mempool, it returns true indicating that the transaction is already in the mempool.
func (mp *Mempool) Store(txid types.Hash, tx *ktypes.Transaction) (found bool) {
	mp.mtx.Lock()
	defer mp.mtx.Unlock()
	delete(mp.fetching, txid)

	if tx == nil { // legacy semantics for removal
		mp.remove(txid)
		return false
	}

	if _, ok := mp.txns[txid]; ok {
		return true // already have it
	}

	mp.txns[txid] = tx
	mp.txQ = append(mp.txQ, types.NamedTx{
		Hash: txid,
		Tx:   tx,
	})
	return false
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
	return len(mp.txQ)
}

func (mp *Mempool) Get(txid types.Hash) *ktypes.Transaction {
	mp.mtx.RLock()
	defer mp.mtx.RUnlock()
	return mp.txns[txid]
}

// ReapN extracts the first n transactions in the queue
func (mp *Mempool) ReapN(n int) []types.NamedTx {
	mp.mtx.Lock()
	defer mp.mtx.Unlock()
	n = min(n, len(mp.txQ))
	txns := slices.Clone(mp.txQ[:n])
	mp.txQ = mp.txQ[n:]
	for _, tx := range txns {
		delete(mp.txns, tx.Hash)
	}
	return txns
}

func (mp *Mempool) PeekN(n int) []types.NamedTx {
	mp.mtx.RLock()
	defer mp.mtx.RUnlock()
	n = min(n, len(mp.txns))
	txns := make([]types.NamedTx, n)
	copy(txns, mp.txQ)
	return txns
}

type CheckFn func(ctx context.Context, tx *ktypes.Transaction) error

func (mp *Mempool) RecheckTxs(ctx context.Context, fn CheckFn) {
	mp.mtx.RLock()
	defer mp.mtx.RUnlock()

	for _, tx := range mp.txQ {
		if err := fn(ctx, tx.Tx); err != nil {
			mp.remove(tx.Hash)
		}
	}
}

func (mp *Mempool) TxsAvailable() bool {
	mp.mtx.RLock()
	defer mp.mtx.RUnlock()
	return len(mp.txQ) > 0
}
