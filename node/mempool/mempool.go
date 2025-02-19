package mempool

import (
	"context"
	"io"
	"slices"
	"sync"

	ktypes "github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/node/types"
)

// Mempool maintains a thread-safe pool of unconfirmed transactions with size limits.
type Mempool struct {
	mtx      sync.RWMutex
	txns     map[types.Hash]*sizedTx
	txQ      []types.NamedTx
	fetching map[types.Hash]bool
	// acctTxns map[string][]types.NamedTx
	currentSize int64 // bytes

	maxSize int64 // bytes
	// maxTxns int // number of transactions
}

type sizedTx struct {
	*ktypes.Transaction
	size int64
}

// New creates a new Mempool instance with a default max size of 200MB.
// See also SetMaxSize.
func New(sz int64) *Mempool {
	return &Mempool{
		txns:     make(map[types.Hash]*sizedTx),
		fetching: make(map[types.Hash]bool),
		maxSize:  sz,
	}
}

// SetMaxSize updates the maximum allowed size in bytes for the mempool.
func (mp *Mempool) SetMaxSize(maxBytes int64) {
	mp.mtx.Lock()
	defer mp.mtx.Unlock()
	mp.maxSize = maxBytes
}

// Have checks if a transaction with the given hash exists in the mempool.
func (mp *Mempool) Have(txid types.Hash) bool { // this is racy
	mp.mtx.RLock()
	defer mp.mtx.RUnlock()
	_, have := mp.txns[txid]
	return have
}

// Remove deletes a transaction from the mempool by its hash.
func (mp *Mempool) Remove(txid types.Hash) {
	mp.mtx.Lock()
	defer mp.mtx.Unlock()
	mp.remove(txid)
}

func (mp *Mempool) remove(txid types.Hash) {
	tx, have := mp.txns[txid]
	if !have {
		return
	}
	mp.currentSize -= tx.size

	delete(mp.txns, txid)

	idx := slices.IndexFunc(mp.txQ, func(a types.NamedTx) bool {
		return a.Hash == txid
	})
	if idx != -1 {
		mp.txQ = slices.Delete(mp.txQ, idx, idx+1) // remove txQ[idx]
	} // else there's a bug!
}

// Store adds a transaction to the mempool. Returns (found, rejected) where found indicates
// if the tx was already present, and rejected indicates if it was rejected due to size limits.
func (mp *Mempool) Store(txid types.Hash, tx *ktypes.Transaction) (found, rejected bool) {
	mp.mtx.Lock()
	defer mp.mtx.Unlock()

	delete(mp.fetching, txid)

	if tx == nil { // legacy semantics for removal
		mp.remove(txid)
		return false, false
	}

	if _, ok := mp.txns[txid]; ok {
		return true, false // already have it
	}

	sz, _ := tx.WriteTo(io.Discard) // TODO: make a SerializeSize method for Transaction

	if mp.currentSize+sz > mp.maxSize {
		return false, true
	}

	mp.currentSize += sz

	mp.txns[txid] = &sizedTx{
		Transaction: tx,
		size:        sz,
	}
	mp.txQ = append(mp.txQ, types.NamedTx{
		Hash: txid,
		Tx:   tx,
	})
	return false, false
}

// PreFetch marks a transaction as being fetched. Returns true if the tx should be fetched.
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

// Size returns the current total size in bytes and number of transactions in the mempool.
func (mp *Mempool) Size() (totalBytes, numTxns int) {
	mp.mtx.RLock()
	defer mp.mtx.RUnlock()
	return int(mp.currentSize), len(mp.txQ)
}

// Get retrieves a transaction by its hash, returns nil if not found.
func (mp *Mempool) Get(txid types.Hash) *ktypes.Transaction {
	mp.mtx.RLock()
	defer mp.mtx.RUnlock()
	tx, have := mp.txns[txid]
	if !have {
		return nil
	}
	return tx.Transaction
}

// ReapN removes and returns up to n transactions from the front of the queue.
func (mp *Mempool) ReapN(n int) []types.NamedTx {
	mp.mtx.Lock()
	defer mp.mtx.Unlock()
	n = min(n, len(mp.txQ))
	txns := slices.Clone(mp.txQ[:n])
	mp.txQ = mp.txQ[n:]
	for _, tx := range txns {
		szTx := mp.txns[tx.Hash]
		if szTx == nil {
			continue // bug, don't crash
		}
		mp.currentSize -= szTx.size
		delete(mp.txns, tx.Hash)
	}
	return txns
}

// PeekN returns up to n transactions from the front of the queue without
// removing them, the number of transactions returned may be less than n if the
// total size in bytes of the transactions exceeds szLimit.
func (mp *Mempool) PeekN(n, szLimit int) []types.NamedTx {
	mp.mtx.RLock()
	defer mp.mtx.RUnlock()
	n = min(n, len(mp.txns))
	var totalPickedSz int
	txns := make([]types.NamedTx, 0, n)
	for _, tx := range mp.txQ[:n] {
		if szLimit > 0 {
			txSz := int(tx.Tx.SerializeSize())
			if txSz+totalPickedSz > szLimit {
				break // no more checks since we are trying to keep order
			}
			totalPickedSz += txSz
		}
		txns = append(txns, tx)
	}
	return txns
}

// CheckFn is a function type for validating transactions.
type CheckFn func(ctx context.Context, tx *ktypes.Transaction) error

// RecheckTxs validates all transactions in the mempool using the provided check
// function, removing any that fail validation. This function will check the
// transaction queue in order.
func (mp *Mempool) RecheckTxs(ctx context.Context, fn CheckFn) {
	mp.mtx.Lock()
	defer mp.mtx.Unlock()

	type indexedTx struct {
		idx  int
		txid types.Hash
	}
	var toRemove []indexedTx
	for idx, tx := range mp.txQ { // must check in order
		if err := fn(ctx, tx.Tx); err != nil {
			toRemove = append(toRemove, indexedTx{idx: idx, txid: tx.Hash})
		}
	}

	if len(toRemove) == 0 {
		return
	}

	// Remove in reverse order to avoid shifting indices in the txQ slice.
	slices.Reverse(toRemove)

	for _, itx := range toRemove {
		tx, have := mp.txns[itx.txid]
		if have { // we should!
			mp.currentSize -= tx.size
		}
		delete(mp.txns, itx.txid)

		mp.txQ = slices.Delete(mp.txQ, itx.idx, itx.idx+1) // remove txQ[idx]
	}
}

// TxsAvailable returns true if there are any transactions in the mempool.
func (mp *Mempool) TxsAvailable() bool {
	mp.mtx.RLock()
	defer mp.mtx.RUnlock()
	return len(mp.txQ) > 0
}
