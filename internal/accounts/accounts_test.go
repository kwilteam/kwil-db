package accounts

import (
	"context"
	"errors"
	"math/big"
	"testing"

	sql "github.com/kwilteam/kwil-db/common/sql"
	"github.com/kwilteam/kwil-db/core/types"

	"github.com/stretchr/testify/require"
)

type mockTx struct {
	*mockDB
}

func (m *mockTx) Commit(ctx context.Context) error {
	return nil
}

func (m *mockTx) Rollback(ctx context.Context) error {
	return nil
}

type mockDB struct {
	accts map[string]*types.Account //string([]byte{a,c,c,t})
}

func newDB() *mockDB {
	return &mockDB{
		accts: make(map[string]*types.Account),
	}
}

func (m *mockDB) AccessMode() sql.AccessMode {
	return sql.ReadWrite // not use in these tests
}

func (m *mockDB) BeginTx(ctx context.Context) (sql.Tx, error) {
	return &mockTx{m}, nil
}

func (m *mockDB) Execute(ctx context.Context, stmt string, args ...any) (*sql.ResultSet, error) {
	// mock some expected queries from internal functions
	switch stmt {
	case sqlCreateAccount: // via createAccount and createAccountWithNonce
		id := args[0].([]byte)
		bal, ok := big.NewInt(0).SetString(args[1].(string), 10)
		if !ok {
			return nil, errors.New("not a string balance")
		}
		m.accts[string(id)] = &types.Account{
			Identifier: id,
			Nonce:      args[2].(int64),
			Balance:    bal,
		}
		return &sql.ResultSet{
			Status: sql.CommandTag{
				RowsAffected: 1,
				Text:         `INSERT ...`,
			},
		}, nil
	case sqlUpdateAccount: // via updateAccount
		bal, ok := big.NewInt(0).SetString(args[0].(string), 10)
		if !ok {
			return nil, errors.New("not a string balance")
		}
		id, nonce := args[2].([]byte), args[1].(int64)

		acct, ok := m.accts[string(id)]
		if !ok {
			return &sql.ResultSet{
				Status: sql.CommandTag{
					RowsAffected: 0,
					Text:         `UPDATE ...`,
				},
			}, nil
		}

		acct.Balance = bal
		acct.Nonce = nonce

		return &sql.ResultSet{
			Status: sql.CommandTag{
				RowsAffected: 1,
				Text:         `UPDATE ...`,
			},
		}, nil
	case sqlGetAccount: // via getAccount
		id := args[0].([]byte)
		acct, ok := m.accts[string(id)]
		if !ok {
			return &sql.ResultSet{}, nil // not ErrNoRows since we don't use Scan in pg
		}
		return &sql.ResultSet{
			Columns: []string{"balance", "nonce"},
			Rows: [][]any{
				{acct.Balance.String(), acct.Nonce},
			},
		}, nil
	default:
		return nil, errors.New("bad query")
	}
}

var (
	account1 = []byte("account1")
	account2 = []byte("account2")
)

type acctsTestCase struct {
	name string
	fn   func(t *testing.T, db sql.DB)
}

// once we have a way to increase balances in accounts, we will have to add tests
// for spending a valid amount
var acctsTestCases = []acctsTestCase{
	{
		name: "credit and debit",
		fn: func(t *testing.T, db sql.DB) {
			ctx := context.Background()

			err := Credit(ctx, db, account1, big.NewInt(100))
			require.NoError(t, err)

			err = Credit(ctx, db, account1, big.NewInt(-100))
			require.NoError(t, err)
		},
	},
	{
		name: "debit non-existent account",
		fn: func(t *testing.T, db sql.DB) {
			ctx := context.Background()

			err := Credit(ctx, db, account1, big.NewInt(-100))
			require.ErrorIs(t, err, ErrNegativeBalance)
		},
	},
	{
		name: "credit and over-debit",
		fn: func(t *testing.T, db sql.DB) {
			ctx := context.Background()

			err := Credit(ctx, db, account1, big.NewInt(100))
			require.NoError(t, err)

			err = Credit(ctx, db, account1, big.NewInt(-101))
			require.ErrorIs(t, err, ErrNegativeBalance)
		},
	},
	{
		name: "transfer to nonexistent account",
		fn: func(t *testing.T, db sql.DB) {
			ctx := context.Background()

			err := Credit(ctx, db, account1, big.NewInt(100))
			require.NoError(t, err)

			err = Transfer(ctx, db, account1, account2, big.NewInt(100))
			require.NoError(t, err)

			acc, err := GetAccount(ctx, db, account1)
			require.NoError(t, err)

			require.Equal(t, big.NewInt(0), acc.Balance)

			acc, err = GetAccount(ctx, db, account2)
			require.NoError(t, err)

			require.Equal(t, big.NewInt(100), acc.Balance)
		},
	},
	{
		name: "transfer to existing account",
		fn: func(t *testing.T, db sql.DB) {
			ctx := context.Background()

			err := Credit(ctx, db, account1, big.NewInt(100))
			require.NoError(t, err)

			err = Credit(ctx, db, account2, big.NewInt(100))
			require.NoError(t, err)

			err = Transfer(ctx, db, account1, account2, big.NewInt(50))
			require.NoError(t, err)

			acc, err := GetAccount(ctx, db, account1)
			require.NoError(t, err)

			require.Equal(t, big.NewInt(50), acc.Balance)

			acc, err = GetAccount(ctx, db, account2)
			require.NoError(t, err)

			require.Equal(t, big.NewInt(150), acc.Balance)
		},
	},
	{
		name: "transfer negative amount",
		fn: func(t *testing.T, db sql.DB) {
			ctx := context.Background()

			err := Credit(ctx, db, account1, big.NewInt(100))
			require.NoError(t, err)

			err = Transfer(ctx, db, account1, account2, big.NewInt(-50))
			require.ErrorIs(t, err, ErrNegativeTransfer)
		},
	},
	{
		name: "transfer more than you have",
		fn: func(t *testing.T, db sql.DB) {
			ctx := context.Background()

			err := Credit(ctx, db, account1, big.NewInt(100))
			require.NoError(t, err)

			err = Transfer(ctx, db, account1, account2, big.NewInt(150))
			require.ErrorIs(t, err, ErrInsufficientFunds)
		},
	},
	{
		name: "get non existent account",
		fn: func(t *testing.T, db sql.DB) {
			ctx := context.Background()

			acc, err := GetAccount(ctx, db, account1)
			require.NoError(t, err)

			require.Equal(t, big.NewInt(0), acc.Balance)
			require.Equal(t, int64(0), acc.Nonce)
		},
	},
	{
		name: "spend from non existent account",
		fn: func(t *testing.T, db sql.DB) {
			ctx := context.Background()

			err := Spend(ctx, db, account1, big.NewInt(100), 1)
			require.ErrorIs(t, err, ErrAccountNotFound)
		},
	},
	{
		name: "spend more than you have",
		fn: func(t *testing.T, db sql.DB) {
			ctx := context.Background()

			err := Credit(ctx, db, account1, big.NewInt(100))
			require.NoError(t, err)

			err = Spend(ctx, db, account1, big.NewInt(101), 1)
			require.ErrorIs(t, err, ErrInsufficientFunds)

			acc, err := GetAccount(ctx, db, account1)
			require.NoError(t, err)

			require.Equal(t, big.NewInt(100), acc.Balance)
		},
	},
	{
		name: "spend with invalid nonce",
		fn: func(t *testing.T, db sql.DB) {
			ctx := context.Background()

			err := Credit(ctx, db, account1, big.NewInt(100))
			require.NoError(t, err)

			err = Spend(ctx, db, account1, big.NewInt(50), 2)
			require.ErrorIs(t, err, ErrInvalidNonce)

			acc, err := GetAccount(ctx, db, account1)
			require.NoError(t, err)

			require.Equal(t, big.NewInt(100), acc.Balance)
		},
	},
	{
		name: "valid spend",
		fn: func(t *testing.T, db sql.DB) {
			ctx := context.Background()

			err := Credit(ctx, db, account1, big.NewInt(100))
			require.NoError(t, err)

			err = Spend(ctx, db, account1, big.NewInt(50), 1)
			require.NoError(t, err)

			acc, err := GetAccount(ctx, db, account1)
			require.NoError(t, err)

			require.Equal(t, big.NewInt(50), acc.Balance)
		},
	},
	{
		name: "spend 0 on non-existent account",
		fn: func(t *testing.T, db sql.DB) {
			ctx := context.Background()

			err := Spend(ctx, db, account1, big.NewInt(0), 1)
			require.NoError(t, err)

			acc, err := GetAccount(ctx, db, account1)
			require.NoError(t, err)

			require.Equal(t, big.NewInt(0), acc.Balance)
			require.Equal(t, int64(1), acc.Nonce)
		},
	},
}

func Test_Accounts(t *testing.T) {
	for _, tc := range acctsTestCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			db := newDB()
			tx, _ := db.BeginTx(ctx)

			tc.fn(t, tx)
		})
	}
}
