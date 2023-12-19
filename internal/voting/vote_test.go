package voting_test

import (
	"bytes"
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
	err := voting.RegisterPaylod(&exampleResolutionPayload{})
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
				err := v.UpdateVoter(ctx, []byte("voter1"), 10)
				require.NoError(t, err)

				// create vote with body
				body := &exampleResolutionPayload{
					UniqueID: "unique_id",
					Account:  []byte("account1"),
					Amount:   100,
				}
				bts, err := body.MarshalBinary()
				require.NoError(t, err)

				uuid := types.NewUUIDV5(bts)

				alreadyProcessed, err := v.AlreadyProcessed(ctx, uuid)
				require.NoError(t, err)
				require.False(t, alreadyProcessed)

				err = v.CreateVote(ctx, bts, examplePayloadType, 10000)
				require.NoError(t, err)

				alreadyProcessed, err = v.AlreadyProcessed(ctx, uuid)
				require.NoError(t, err)
				require.False(t, alreadyProcessed)

				// approve vote
				// expiration does not matter here since it only matters for the first vote
				err = v.Approve(ctx, uuid, 10323, []byte("voter1"))
				require.NoError(t, err)

				alreadyProcessed, err = v.AlreadyProcessed(ctx, uuid)
				require.NoError(t, err)
				require.False(t, alreadyProcessed)

				// resolve vote
				err = v.ProcessConfirmedResolutions(ctx)
				require.NoError(t, err)

				alreadyProcessed, err = v.AlreadyProcessed(ctx, uuid)
				require.NoError(t, err)
				require.True(t, alreadyProcessed)

				// check that the account was credited
				acc, err := ds.Accounts.Account(ctx, []byte("account1"))
				require.NoError(t, err)
				require.Equal(t, big.NewInt(100), acc.Balance)
			},
		},
		{
			name: "vote before adding body",
			fn: func(t *testing.T, v *voting.VoteProcessor, ds *voting.Datastores) {
				// add one voter, create vote, approve, and resolve
				ctx := context.Background()

				// add voter
				err := v.UpdateVoter(ctx, []byte("voter1"), 10)
				require.NoError(t, err)

				// create vote with body
				body := &exampleResolutionPayload{
					UniqueID: "unique_id",
					Account:  []byte("account1"),
					Amount:   100,
				}
				bts, err := body.MarshalBinary()
				require.NoError(t, err)

				uuid := types.NewUUIDV5(bts)

				// approve vote, before creating vote
				// expiration does not matter here since it only matters for the first vote
				err = v.Approve(ctx, uuid, 10323, []byte("voter1"))
				require.NoError(t, err)

				// now create the vote
				err = v.CreateVote(ctx, bts, examplePayloadType, 10000)
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
		{
			name: "vote without providing body does not confirm",
			fn: func(t *testing.T, v *voting.VoteProcessor, ds *voting.Datastores) {
				// add one voter, create vote, approve, and resolve
				ctx := context.Background()

				// add voter
				err := v.UpdateVoter(ctx, []byte("voter1"), 10)
				require.NoError(t, err)

				// create vote with body
				body := &exampleResolutionPayload{
					UniqueID: "unique_id",
					Account:  []byte("account1"),
					Amount:   100,
				}
				bts, err := body.MarshalBinary()
				require.NoError(t, err)

				uuid := types.NewUUIDV5(bts)

				// approve vote, before creating vote
				// expiration does not matter here since it only matters for the first vote
				err = v.Approve(ctx, uuid, 10323, []byte("voter1"))
				require.NoError(t, err)

				// resolve vote
				err = v.ProcessConfirmedResolutions(ctx)
				require.NoError(t, err)

				// check that the account was credited
				acc, err := ds.Accounts.Account(ctx, []byte("account1"))
				require.NoError(t, err)
				require.Equal(t, big.NewInt(0), acc.Balance)
			},
		},
		{
			name: "insufficient voting power",
			fn: func(t *testing.T, v *voting.VoteProcessor, ds *voting.Datastores) {
				// add one voter, create vote, approve, and resolve
				ctx := context.Background()

				// add voter 1
				err := v.UpdateVoter(ctx, []byte("voter1"), 10)
				require.NoError(t, err)

				// add voter 2
				err = v.UpdateVoter(ctx, []byte("voter2"), 1000)
				require.NoError(t, err)

				// create vote with body
				body := &exampleResolutionPayload{
					UniqueID: "unique_id",
					Account:  []byte("account1"),
					Amount:   100,
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

				// resolve votes, it will fail since voter 2 did not approve
				err = v.ProcessConfirmedResolutions(ctx)
				require.NoError(t, err)

				// check that the resolution still exists
				res, err := v.GetResolution(ctx, uuid)
				require.NoError(t, err)

				require.Equal(t, uuid, res.ID)
				if !bytes.EqualFold(bts, res.Body) {
					require.Equal(t, bts, res.Body) // will fail since the bytes are not equal
				}
				require.Equal(t, examplePayloadType, res.Type)
				require.Equal(t, int64(10000), res.Expiration)
				require.Equal(t, int64(10), res.ApprovedPower)

				// get votes by category
				resolutions, err := v.GetVotesByCategory(ctx, examplePayloadType)
				require.NoError(t, err)
				require.Len(t, resolutions, 1)

				requireEqualResolutions(t, resolutions[0], res)
			},
		},
		{
			name: "ByCategory does not panic when no votes exist",
			fn: func(t *testing.T, v *voting.VoteProcessor, ds *voting.Datastores) {
				ctx := context.Background()

				// get votes by category
				resolutions, err := v.GetVotesByCategory(ctx, examplePayloadType)
				require.NoError(t, err)
				require.Len(t, resolutions, 0)
			},
		},
		{
			name: "Get and ByCategory do not panic with no body",
			fn: func(t *testing.T, v *voting.VoteProcessor, ds *voting.Datastores) {
				ctx := context.Background()

				// add voter
				err := v.UpdateVoter(ctx, []byte("voter1"), 10)
				require.NoError(t, err)

				// create vote with body
				body := &exampleResolutionPayload{
					UniqueID: "unique_id",
					Account:  []byte("account1"),
					Amount:   100,
				}
				bts, err := body.MarshalBinary()
				require.NoError(t, err)

				uuid := types.NewUUIDV5(bts)

				// approve vote
				// expiration does not matter here since it only matters for the first vote
				err = v.Approve(ctx, uuid, 10323, []byte("voter1"))
				require.NoError(t, err)

				// get votes by category, should fail since categories do not get defined until
				// the body is set
				resolutions, err := v.GetVotesByCategory(ctx, examplePayloadType)
				require.NoError(t, err)
				require.Len(t, resolutions, 0)

				// get vote by id
				res, err := v.GetResolution(ctx, uuid)
				require.NoError(t, err)

				// check id is same
				require.Equal(t, uuid, res.ID)
				// body is nil, expiration is same, approved power is same, type is nil b/c body is not set
				require.Nil(t, res.Body)
				require.Equal(t, int64(10323), res.Expiration)
				require.Equal(t, int64(10), res.ApprovedPower)
			},
		},
		{
			name: "manipulating voting power",
			fn: func(t *testing.T, v *voting.VoteProcessor, ds *voting.Datastores) {
				ctx := context.Background()

				// add voter
				err := v.UpdateVoter(ctx, []byte("voter1"), 10)
				require.NoError(t, err)

				// add power
				err = v.UpdateVoter(ctx, []byte("voter1"), 10)
				require.NoError(t, err)

				// get power
				power, err := v.GetVoterPower(ctx, []byte("voter1"))
				require.NoError(t, err)

				require.Equal(t, int64(20), power)

				// subtract power
				err = v.UpdateVoter(ctx, []byte("voter1"), -10)
				require.NoError(t, err)

				// get power
				power, err = v.GetVoterPower(ctx, []byte("voter1"))
				require.NoError(t, err)

				require.Equal(t, int64(10), power)

				// ensure power cannot go to 0
				err = v.UpdateVoter(ctx, []byte("voter1"), -10)
				if err == nil {
					t.Fatal("expected error")
				}

				// remove
				err = v.UpdateVoter(ctx, []byte("voter1"), 0)
				require.NoError(t, err)

				// get power
				power, err = v.GetVoterPower(ctx, []byte("voter1"))
				require.NoError(t, err)

				require.Equal(t, int64(0), power)
			},
		},
		{
			name: "non-existent voter cannot vote",
			fn: func(t *testing.T, v *voting.VoteProcessor, ds *voting.Datastores) {
				ctx := context.Background()

				// create vote with body
				body := &exampleResolutionPayload{
					UniqueID: "unique_id",
					Account:  []byte("account1"),
					Amount:   100,
				}
				bts, err := body.MarshalBinary()
				require.NoError(t, err)

				uuid := types.NewUUIDV5(bts)

				// approve vote
				// expiration does not matter here since it only matters for the first vote
				err = v.Approve(ctx, uuid, 10323, []byte("voter1"))
				if err == nil {
					t.Fatal("expected error")
				}
			},
		},
		{
			name: "expiration works",
			fn: func(t *testing.T, v *voting.VoteProcessor, ds *voting.Datastores) {
				// create 3 votes
				// expire on 2
				ctx := context.Background()

				// add voter
				err := v.UpdateVoter(ctx, []byte("voter1"), 10)
				require.NoError(t, err)

				// payload
				body1 := &exampleResolutionPayload{
					UniqueID: "unique_id",
					Account:  []byte("account1"),
					Amount:   100,
				}
				bts1, err := body1.MarshalBinary()
				require.NoError(t, err)

				body2 := &exampleResolutionPayload{
					UniqueID: "unique_id2",
					Account:  []byte("account1"),
					Amount:   100,
				}
				bts2, err := body2.MarshalBinary()
				require.NoError(t, err)
				uuid2 := types.NewUUIDV5(bts2)

				body3 := &exampleResolutionPayload{
					UniqueID: "unique_id3",
					Account:  []byte("account1"),
					Amount:   100,
				}
				bts3, err := body3.MarshalBinary()
				require.NoError(t, err)

				// create vote 1
				err = v.CreateVote(ctx, bts1, examplePayloadType, 2)
				require.NoError(t, err)

				err = v.CreateVote(ctx, bts2, examplePayloadType, 3)
				require.NoError(t, err)

				err = v.CreateVote(ctx, bts3, examplePayloadType, 4)
				require.NoError(t, err)

				// get votes by category
				resolutions, err := v.GetVotesByCategory(ctx, examplePayloadType)
				require.NoError(t, err)

				require.Len(t, resolutions, 3)

				alreadyProcessed, err := v.AlreadyProcessed(ctx, uuid2)
				require.NoError(t, err)
				require.False(t, alreadyProcessed)

				// expire
				err = v.Expire(ctx, 3)
				require.NoError(t, err)

				alreadyProcessed, err = v.AlreadyProcessed(ctx, uuid2)
				require.NoError(t, err)
				require.True(t, alreadyProcessed)

				// get votes by category
				resolutions, err = v.GetVotesByCategory(ctx, examplePayloadType)
				require.NoError(t, err)

				require.Len(t, resolutions, 1)
			},
		},
		{
			name: "double approve does nothing",
			fn: func(t *testing.T, v *voting.VoteProcessor, ds *voting.Datastores) {
				ctx := context.Background()

				// add voter
				err := v.UpdateVoter(ctx, []byte("voter1"), 10)
				require.NoError(t, err)

				// payload
				body := &exampleResolutionPayload{
					UniqueID: "unique_id",
					Account:  []byte("account1"),
					Amount:   100,
				}
				bts, err := body.MarshalBinary()
				require.NoError(t, err)
				uuid := types.NewUUIDV5(bts)

				// approve vote twice
				err = v.Approve(ctx, uuid, 10323, []byte("voter1"))
				require.NoError(t, err)

				err = v.Approve(ctx, uuid, 10323, []byte("voter1"))
				require.NoError(t, err)
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

// requireEqualResolutions is a helper function to compare two resolutions.
// 1 is a resolution, the other is a resolution status
func requireEqualResolutions(t *testing.T, res1 *voting.Resolution, res2 *voting.ResolutionStatus) {
	require.Equal(t, res1.ID, res2.ID)
	if !bytes.EqualFold(res1.Body, res2.Body) {
		require.Equal(t, res1.Body, res2.Body) // will fail since the bytes are not equal
	}
	require.Equal(t, res1.Type, res2.Type)
	require.Equal(t, res1.Expiration, res2.Expiration)
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
	UniqueID string `json:"unique_id"` // could be a transaction hash from a different chain
	Account  []byte `json:"account"`
	Amount   int64  `json:"amount"`
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

func (e *exampleResolutionPayload) Type() string {
	return examplePayloadType
}
