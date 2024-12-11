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
