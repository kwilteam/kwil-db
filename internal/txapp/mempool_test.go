package txapp

import (
	"context"
	"math/big"
	"testing"

	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/kwilteam/kwil-db/core/types/transactions"
	"github.com/kwilteam/kwil-db/internal/accounts"
	"github.com/stretchr/testify/assert"
)

func Test_Mempool(t *testing.T) {
	m := &mempool{
		accountStore: &mockAccountsModule{},
		accounts:     make(map[string]*accounts.Account),
	}
	ctx := context.Background()

	// Successful transaction A: 1
	err := m.applyTransaction(ctx, newTx(t, 1, "A"))
	assert.NoError(t, err)
	assert.EqualValues(t, m.accounts["A"].Nonce, 1)

	// Successful transaction A: 2
	err = m.applyTransaction(ctx, newTx(t, 2, "A"))
	assert.NoError(t, err)
	assert.EqualValues(t, m.accounts["A"].Nonce, 2)

	// Duplicate nonce failure
	err = m.applyTransaction(ctx, newTx(t, 2, "A"))
	assert.Error(t, err)
	assert.EqualValues(t, m.accounts["A"].Nonce, 2)

	// Invalid order
	err = m.applyTransaction(ctx, newTx(t, 4, "A"))
	assert.Error(t, err)
	assert.EqualValues(t, m.accounts["A"].Nonce, 2)

	err = m.applyTransaction(ctx, newTx(t, 3, "A"))
	assert.NoError(t, err)
	assert.EqualValues(t, m.accounts["A"].Nonce, 3)

	// Recheck nonce 4 transaction
	err = m.applyTransaction(ctx, newTx(t, 4, "A"))
	assert.NoError(t, err)
	assert.EqualValues(t, m.accounts["A"].Nonce, 4)
}

type mockAccountsModule struct{}

func (m *mockAccountsModule) GetAccount(ctx context.Context, acctID []byte) (*accounts.Account, error) {
	return &accounts.Account{
		Nonce:      0,
		Balance:    big.NewInt(0),
		Identifier: acctID,
	}, nil
}

func newTx(t *testing.T, nonce uint64, sender string) *transactions.Transaction {
	return &transactions.Transaction{
		Signature: &auth.Signature{},
		Body: &transactions.TransactionBody{
			Description: "test",
			Payload:     []byte(`random payload`),
			Fee:         big.NewInt(0),
			Nonce:       nonce,
		},
		Sender: []byte(sender),
	}
}
