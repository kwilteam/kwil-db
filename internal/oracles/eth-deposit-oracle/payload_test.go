package deposit_oracle

import (
	"context"
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/internal/accounts"
	"github.com/kwilteam/kwil-db/internal/sql"
	"github.com/kwilteam/kwil-db/internal/voting"

	"github.com/stretchr/testify/require"
)

const (
	testAccount     = "0xabcd"
	trimTestAccount = "abcd"
)

var (
	validAC = &AccountCredit{
		Account:   testAccount,
		Amount:    big.NewInt(100),
		TxHash:    "test-txhash",
		BlockHash: "test-blockhash",
		ChainID:   "test-chainid",
	}
)

func Test_ValidPayload(t *testing.T) {
	ds := voting.Datastores{
		Accounts: &mockAccountStore{
			accounts: make(map[string]*accounts.Account),
		},
		Databases: nil,
	}

	ac := validAC
	db := &mockDb{}

	err := ac.Apply(context.Background(), db, ds, log.NewStdOut(log.InfoLevel))
	require.NoError(t, err)

	// check that the account was credited
	bts, err := hex.DecodeString(trimTestAccount)
	require.NoError(t, err)

	acc, err := ds.Accounts.GetAccount(context.Background(), db, bts)
	require.NoError(t, err)
	require.Equal(t, big.NewInt(100), acc.Balance)

}

type mockAccountStore struct {
	accounts map[string]*accounts.Account
}

func (mas *mockAccountStore) GetAccount(ctx context.Context, _ sql.DB, identifier []byte) (*accounts.Account, error) {
	id := hex.EncodeToString(identifier)
	acct, ok := mas.accounts[id]
	if !ok {
		return nil, accounts.ErrAccountNotFound
	}
	return acct, nil
}

func (mas *mockAccountStore) Credit(ctx context.Context, _ sql.DB, identifier []byte, amount *big.Int) error {
	id := hex.EncodeToString(identifier)
	acct, ok := mas.accounts[id]
	if !ok {
		acct = &accounts.Account{
			Identifier: identifier,
			Balance:    big.NewInt(0),
			Nonce:      0,
		}
		mas.accounts[id] = acct
	}

	acct.Balance.Add(acct.Balance, amount)

	return nil
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
