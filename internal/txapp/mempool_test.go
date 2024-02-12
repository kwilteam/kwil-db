package txapp

import (
	"context"
	"math/big"
	"testing"

	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/transactions"
	"github.com/kwilteam/kwil-db/internal/accounts"
	"github.com/kwilteam/kwil-db/internal/sql"
	"github.com/stretchr/testify/assert"
)

func Test_Mempool(t *testing.T) {
	m := &mempool{
		accountStore: &mockAccountsModule{},
		accounts:     make(map[string]*accounts.Account),
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

type mockAccountsModule struct{}

var _ AccountReader = (*mockAccountsModule)(nil)

func (m *mockAccountsModule) GetAccount(ctx context.Context, _ sql.DB, acctID []byte) (*accounts.Account, error) {
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

type mockDb struct{}

func (m *mockDb) AccessMode() sql.AccessMode {
	return sql.ReadOnly
}

func (m *mockDb) BeginTx(ctx context.Context) (sql.Tx, error) {
	return &mockTx{m}, nil
}

func (m *mockDb) Execute(ctx context.Context, stmt string, args ...any) (*sql.ResultSet, error) {
	return nil, nil
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
