package txapp

import (
	"context"
	"math/big"
	"testing"

	sql "github.com/kwilteam/kwil-db/common/sql"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/transactions"
	"github.com/stretchr/testify/assert"
)

func Test_MempoolWithoutGas(t *testing.T) {
	m := &mempool{
		accounts: make(map[string]*types.Account),
	}

	ctx := context.Background()
	db := &mockDb{}
	rebroadcast := &mockRebroadcast{}

	// Successful transaction A: 1
	err := m.applyTransaction(ctx, newTx(t, 1, "A"), db, rebroadcast)
	assert.NoError(t, err)
	assert.EqualValues(t, m.accounts["A"].Nonce, 1)

	// Successful transaction A: 2
	err = m.applyTransaction(ctx, newTx(t, 2, "A"), db, rebroadcast)
	assert.NoError(t, err)
	assert.EqualValues(t, m.accounts["A"].Nonce, 2)

	// Duplicate nonce failure
	err = m.applyTransaction(ctx, newTx(t, 2, "A"), db, rebroadcast)
	assert.Error(t, err)
	assert.EqualValues(t, m.accounts["A"].Nonce, 2)

	// Invalid order
	err = m.applyTransaction(ctx, newTx(t, 4, "A"), db, rebroadcast)
	assert.Error(t, err)
	assert.EqualValues(t, m.accounts["A"].Nonce, 2)

	err = m.applyTransaction(ctx, newTx(t, 3, "A"), db, rebroadcast)
	assert.NoError(t, err)
	assert.EqualValues(t, m.accounts["A"].Nonce, 3)

	// Recheck nonce 4 transaction
	err = m.applyTransaction(ctx, newTx(t, 4, "A"), db, rebroadcast)
	assert.NoError(t, err)
	assert.EqualValues(t, m.accounts["A"].Nonce, 4)
}

func Test_MempoolWithGas(t *testing.T) {
	m := &mempool{
		accounts:   make(map[string]*types.Account),
		gasEnabled: true,
	}

	ctx := context.Background()
	db := &mockDb{}
	rebroadcast := &mockRebroadcast{}

	// Transaction from Unknown sender should fail
	tx := newTx(t, 1, "A")
	err := m.applyTransaction(ctx, tx, db, rebroadcast)
	assert.Error(t, err)

	// Resubmitting the same transaction should fail
	err = m.applyTransaction(ctx, tx, db, rebroadcast)
	assert.Error(t, err)

	// Credit account A
	m.accounts["A"] = &types.Account{
		Identifier: []byte("A"),
		Balance:    big.NewInt(100),
		Nonce:      0,
	}

	// Successful transaction A: 1
	err = m.applyTransaction(ctx, tx, db, rebroadcast)
	assert.NoError(t, err)
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

type mockDb struct{}

func (m *mockDb) BeginTx(ctx context.Context) (sql.Tx, error) {
	return &mockTx{m}, nil
}

func (m *mockDb) Execute(ctx context.Context, stmt string, args ...any) (*sql.ResultSet, error) {
	return &sql.ResultSet{}, nil
}

type mockTx struct {
	*mockDb
}

func (m *mockTx) Commit(ctx context.Context) error {
	return nil
}

func (m *mockTx) Rollback(ctx context.Context) error {
	return nil
}

type mockOuterTx struct {
	*mockTx
}

func (m *mockOuterTx) Precommit(ctx context.Context) ([]byte, error) {
	return nil, nil
}

type mockRebroadcast struct{}

func (m *mockRebroadcast) MarkRebroadcast(ctx context.Context, ids []types.UUID) error {
	return nil
}
