package voting_test

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"testing"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/internal/sql"
	"github.com/kwilteam/kwil-db/internal/sql/sqlite"
	"github.com/kwilteam/kwil-db/internal/voting"
	"github.com/stretchr/testify/require"
)

const examplePayloadType = "example"

func init() {
	err := voting.RegisterPaylod(examplePayloadType, &exampleResolutionPayload{})
	if err != nil {
		panic(err)
	}
}

func Test_Votes(t *testing.T) {
	type testCase struct {
		name string
		fn   func(t *testing.T, v *voting.VoteProcessor, ds *voting.Datastores)
	}

	testCases := []testCase{
		{
			name: "successful usage, single user",
			fn: func(t *testing.T, v *voting.VoteProcessor, ds *voting.Datastores) {
				// add one voter, create vote, approve, and resolve
				ctx := context.Background()

				// add voter
				err := v.AddVoter(ctx, []byte("voter1"), 10)
				require.NoError(t, err)

				// create vote with body
				body := &exampleResolutionPayload{
					Account: []byte("account1"),
					Amount:  100,
				}
				bts, err := body.MarshalBinary()
				require.NoError(t, err)

				err = v.CreateVote(ctx, bts, examplePayloadType, 10000)
				require.NoError(t, err)

				uuid := types.NewUUIDV5(bts)

				// approve vote
				// expiration does not matter here since it only matters for the first vote
				err = v.Approve(ctx, uuid, 10323, []byte("voter1"))
				require.NoError(t, err)

				// resolve vote
				err = v.ProcessConfirmedResolutions(ctx)
				require.NoError(t, err)

				// check that the account was credited
				acc, err := ds.Accounts.Account(ctx, []byte("account1"))
				require.NoError(t, err)
				require.Equal(t, big.NewInt(100), acc.Balance)
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			conn, err := sqlite.Open(ctx, ":memory:", sql.OpenCreate)
			if err != nil {
				t.Fatal(err)
			}
			defer conn.Close()

			ds := &voting.Datastores{
				Accounts: &mockAccountStore{
					accounts: map[string]*types.Account{},
				},
				Databases: &db{conn: conn},
			}

			v, err := voting.NewVoteProcessor(ctx, ds.Databases, ds.Accounts, 500000)
			if err != nil {
				t.Fatal(err)
			}

			tt.fn(t, v, ds)
		})
	}
}

type db struct {
	conn *sqlite.Connection
}

func (d *db) Execute(ctx context.Context, stmt string, args map[string]any) (*sql.ResultSet, error) {
	res, err := d.conn.Execute(ctx, stmt, args)
	if err != nil {
		return nil, err
	}
	defer res.Finish()

	return res.ResultSet()
}

func (d *db) Query(ctx context.Context, query string, args map[string]any) (*sql.ResultSet, error) {
	res, err := d.conn.Execute(ctx, query, args)
	if err != nil {
		return nil, err
	}
	defer res.Finish()

	return res.ResultSet()
}

func (d *db) Savepoint() (sql.Savepoint, error) {
	return d.conn.Savepoint()
}

type mockAccountStore struct {
	accounts map[string]*types.Account
}

func (m *mockAccountStore) Account(ctx context.Context, identifier []byte) (*types.Account, error) {
	acc, ok := m.accounts[string(identifier)]
	if !ok {
		acc = &types.Account{
			Identifier: identifier,
			Balance:    big.NewInt(0),
			Nonce:      0,
		}
		m.accounts[string(identifier)] = acc
	}

	return acc, nil
}

func (m *mockAccountStore) Credit(ctx context.Context, account []byte, amount *big.Int) error {
	acc, ok := m.accounts[string(account)]
	if !ok {
		acc = &types.Account{
			Identifier: account,
			Balance:    big.NewInt(0),
			Nonce:      0,
		}
		m.accounts[string(account)] = acc
	}

	acc.Balance = new(big.Int).Add(acc.Balance, amount)

	return nil
}

func (m *mockAccountStore) Debit(ctx context.Context, account []byte, amount *big.Int) error {
	acc, ok := m.accounts[string(account)]
	if !ok {
		return fmt.Errorf("account %s not found", account)
	}

	if acc.Balance.Cmp(amount) < 0 {
		return fmt.Errorf("insufficient funds")
	}

	acc.Balance.Sub(acc.Balance, amount)

	return nil
}

func (m *mockAccountStore) Transfer(ctx context.Context, from []byte, to []byte, amount *big.Int) error {
	if err := m.Debit(ctx, from, amount); err != nil {
		return err
	}

	return m.Credit(ctx, to, amount)
}

// exampleResolutionPayload is an example payload that can be used for testing
// we can use json encoding since it is a local unit test
type exampleResolutionPayload struct {
	Account []byte `json:"account"`
	Amount  int64  `json:"amount"`
}

func (e *exampleResolutionPayload) Apply(ctx context.Context, datastores *voting.Datastores) error {
	if e.Account == nil {
		return fmt.Errorf("account is required")
	}

	if e.Amount <= 0 {
		return fmt.Errorf("amount must be greater than 0")
	}

	return datastores.Accounts.Credit(ctx, e.Account, big.NewInt(e.Amount))
}

func (e *exampleResolutionPayload) MarshalBinary() ([]byte, error) {
	return json.Marshal(e)
}

func (e *exampleResolutionPayload) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, e)
}
