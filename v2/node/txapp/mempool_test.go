package txapp

import (
	"context"
	"kwil/crypto/auth"
	"kwil/node/types/sql"
	"kwil/types"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_MempoolWithoutGas(t *testing.T) {
	accounts := mockAccount{}
	m := &mempool{
		accounts:   make(map[string]*types.Account),
		accountMgr: &accounts,
	}

	ctx := context.Background()
	db := &mockDb{}
	rebroadcast := &mockRebroadcast{}

	txCtx := &types.TxContext{
		Ctx: ctx,
		BlockContext: &types.BlockContext{
			ChainContext: &types.ChainContext{
				NetworkParameters: &types.NetworkParameters{
					DisabledGasCosts: true,
				},
			},
		},
	}

	// Successful transaction A: 1
	err := m.applyTransaction(txCtx, newTx(t, 1, "A"), db, rebroadcast)
	assert.NoError(t, err)
	assert.EqualValues(t, m.accounts["A"].Nonce, 1)

	// Successful transaction A: 2
	err = m.applyTransaction(txCtx, newTx(t, 2, "A"), db, rebroadcast)
	assert.NoError(t, err)
	assert.EqualValues(t, m.accounts["A"].Nonce, 2)

	// Duplicate nonce failure
	err = m.applyTransaction(txCtx, newTx(t, 2, "A"), db, rebroadcast)
	assert.Error(t, err)
	assert.EqualValues(t, m.accounts["A"].Nonce, 2)

	// Invalid order
	err = m.applyTransaction(txCtx, newTx(t, 4, "A"), db, rebroadcast)
	assert.Error(t, err)
	assert.EqualValues(t, m.accounts["A"].Nonce, 2)

	err = m.applyTransaction(txCtx, newTx(t, 3, "A"), db, rebroadcast)
	assert.NoError(t, err)
	assert.EqualValues(t, m.accounts["A"].Nonce, 3)

	// Recheck nonce 4 transaction
	err = m.applyTransaction(txCtx, newTx(t, 4, "A"), db, rebroadcast)
	assert.NoError(t, err)
	assert.EqualValues(t, m.accounts["A"].Nonce, 4)
}

func Test_MempoolWithGas(t *testing.T) {
	m := &mempool{
		accounts:   make(map[string]*types.Account),
		accountMgr: &mockAccount{},
	}

	txCtx := &types.TxContext{
		Ctx: context.Background(),
		BlockContext: &types.BlockContext{
			ChainContext: &types.ChainContext{
				NetworkParameters: &types.NetworkParameters{
					DisabledGasCosts: false,
				},
			},
		},
	}

	db := &mockDb{}
	rebroadcast := &mockRebroadcast{}

	// Transaction from Unknown sender should fail
	tx := newTx(t, 1, "A")
	err := m.applyTransaction(txCtx, tx, db, rebroadcast)
	assert.Error(t, err)

	// Resubmitting the same transaction should fail
	err = m.applyTransaction(txCtx, tx, db, rebroadcast)
	assert.Error(t, err)

	// Credit account A
	m.accounts["A"] = &types.Account{
		Identifier: []byte("A"),
		Balance:    big.NewInt(100),
		Nonce:      0,
	}

	// Successful transaction A: 1
	err = m.applyTransaction(txCtx, tx, db, rebroadcast)
	assert.NoError(t, err)
}

func newTx(_ *testing.T, nonce uint64, sender string) *types.Transaction {
	return &types.Transaction{
		Signature: &auth.Signature{},
		Body: &types.TransactionBody{
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

type mockRebroadcast struct{}

func (m *mockRebroadcast) MarkRebroadcast(ctx context.Context, ids []*types.UUID) error {
	return nil
}
