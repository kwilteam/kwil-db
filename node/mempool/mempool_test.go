package mempool

import (
	"math/big"
	"testing"

	"github.com/kwilteam/kwil-db/core/crypto/auth"
	ktypes "github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/node/types"
	"github.com/stretchr/testify/assert"
)

func newTx(nonce uint64, sender string) *ktypes.Transaction {
	return &ktypes.Transaction{
		Signature: &auth.Signature{},
		Body: &ktypes.TransactionBody{
			Description: "test",
			Payload:     []byte(`random payload`),
			Fee:         big.NewInt(0),
			Nonce:       nonce,
		},
		Sender: []byte(sender),
	}
}

func Test_MempoolRemove(t *testing.T) {
	m := New()

	// Setup test transactions
	tx1 := types.NamedTx{
		Hash: types.Hash{1, 2, 3},
		Tx:   newTx(1, "A"),
	}
	tx2 := types.NamedTx{
		Hash: types.Hash{4, 5, 6},
		Tx:   newTx(2, "B"),
	}

	// Add transactions to mempool
	m.Store(tx1.Hash, tx1.Tx)
	m.Store(tx2.Hash, tx2.Tx)

	// Test removing existing transaction
	m.Remove(tx1.Hash)
	assert.Len(t, m.txQ, 1)
	assert.Len(t, m.txns, 1)
	assert.Equal(t, m.txQ[0].Hash, tx2.Hash)
	_, exists := m.txns[tx1.Hash]
	assert.False(t, exists)

	// Test removing non-existent transaction
	nonExistentHash := types.Hash{9}
	m.Remove(nonExistentHash)
	assert.Len(t, m.txQ, 1)
	assert.Len(t, m.txns, 1)
	assert.Equal(t, m.txQ[0].Hash, tx2.Hash)

	// Test removing last transaction
	m.Remove(tx2.Hash)
	assert.Empty(t, m.txQ)
	assert.Empty(t, m.txns)
}

func Test_MempoolReapN(t *testing.T) {
	m := New()

	// Setup test transactions
	tx1 := types.NamedTx{
		Hash: types.Hash{1, 2, 3},
		Tx:   newTx(1, "A"),
	}
	tx2 := types.NamedTx{
		Hash: types.Hash{4, 5, 6},
		Tx:   newTx(2, "B"),
	}
	tx3 := types.NamedTx{
		Hash: types.Hash{7, 8, 9},
		Tx:   newTx(3, "C"),
	}

	// Test reaping from empty mempool
	emptyReap := m.ReapN(1)
	assert.Empty(t, emptyReap)

	// Add transactions to mempool
	m.Store(tx1.Hash, tx1.Tx)
	m.Store(tx2.Hash, tx2.Tx)
	m.Store(tx3.Hash, tx3.Tx)

	// Test reaping more transactions than available
	overReap := m.ReapN(5)
	assert.Len(t, overReap, 3)
	assert.Equal(t, overReap[0].Hash, tx1.Hash)
	assert.Equal(t, overReap[1].Hash, tx2.Hash)
	assert.Equal(t, overReap[2].Hash, tx3.Hash)
	assert.Empty(t, m.txQ)
	assert.Empty(t, m.txns)

	// Refill mempool
	m.Store(tx1.Hash, tx1.Tx)
	m.Store(tx2.Hash, tx2.Tx)
	m.Store(tx3.Hash, tx3.Tx)

	// Test partial reaping
	partialReap := m.ReapN(2)
	assert.Len(t, partialReap, 2)
	assert.Equal(t, partialReap[0].Hash, tx1.Hash)
	assert.Equal(t, partialReap[1].Hash, tx2.Hash)
	assert.Len(t, m.txQ, 1)
	assert.Len(t, m.txns, 1)

	// Test reaping remaining transaction
	finalReap := m.ReapN(1)
	assert.Len(t, finalReap, 1)
	assert.Equal(t, finalReap[0].Hash, tx3.Hash)
	assert.Empty(t, m.txQ)
	assert.Empty(t, m.txns)

	// Test reaping with zero count
	zeroReap := m.ReapN(0)
	assert.Empty(t, zeroReap)
}
