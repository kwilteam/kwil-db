//go:build pglive

package voting_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"testing"

	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/internal/accounts"
	"github.com/kwilteam/kwil-db/internal/sql"
	dbtest "github.com/kwilteam/kwil-db/internal/sql/pg/test"
	"github.com/kwilteam/kwil-db/internal/voting"

	"github.com/stretchr/testify/require"
)

const examplePayloadType = "example"

func init() {
	err := voting.RegisterPayload(&exampleResolutionPayload{})
	if err != nil {
		panic(err)
	}
}

func Test_Votes(t *testing.T) {
	type testCase struct {
		name string
		fn   func(t *testing.T, v *voting.VoteProcessor, ds *voting.Datastores, db sql.DB)
	}

	testCases := []testCase{
		{
			name: "successful usage, single user",
			fn: func(t *testing.T, v *voting.VoteProcessor, ds *voting.Datastores, db sql.DB) {
				// add one voter, create vote, approve, and resolve
				ctx := context.Background()

				// add voter
				err := v.UpdateVoter(ctx, db, []byte("voter1"), 10)
				require.NoError(t, err)

				// create vote with body
				body := &exampleResolutionPayload{
					UniqueID: "unique_id",
					Account:  []byte("account1"),
					Amount:   100,
				}
				bts, err := body.MarshalBinary()
				require.NoError(t, err)
				event := &types.VotableEvent{
					Body: bts,
					Type: examplePayloadType,
				}

				err = v.CreateResolution(ctx, db, event, 10000)
				require.NoError(t, err)

				// approve vote
				// expiration does not matter here since it only matters for the first vote
				err = v.Approve(ctx, db, event.ID(), 10323, []byte("voter1"))
				require.NoError(t, err)

				// resolve vote
				processed, err := v.ProcessConfirmedResolutions(ctx, db)
				require.NoError(t, err)
				require.Len(t, processed, 1)

				// check that the account was credited
				acc, err := ds.Accounts.GetAccount(ctx, db, []byte("account1"))
				require.NoError(t, err)
				require.Equal(t, big.NewInt(100), acc.Balance)
			},
		},
		{
			name: "vote before adding body",
			fn: func(t *testing.T, v *voting.VoteProcessor, ds *voting.Datastores, db sql.DB) {
				// add one voter, create vote, approve, and resolve
				ctx := context.Background()

				// add voter
				err := v.UpdateVoter(ctx, db, []byte("voter1"), 10)
				require.NoError(t, err)

				// create vote with body
				body := &exampleResolutionPayload{
					UniqueID: "unique_id",
					Account:  []byte("account1"),
					Amount:   100,
				}
				bts, err := body.MarshalBinary()
				require.NoError(t, err)
				event := &types.VotableEvent{
					Body: bts,
					Type: examplePayloadType,
				}

				// approve vote, before creating vote
				// expiration does not matter here since it only matters for the first vote
				err = v.Approve(ctx, db, event.ID(), 10323, []byte("voter1"))
				require.NoError(t, err)

				// now create the vote
				err = v.CreateResolution(ctx, db, event, 10000)
				require.NoError(t, err)

				// resolve vote
				processed, err := v.ProcessConfirmedResolutions(ctx, db)
				require.NoError(t, err)
				require.Len(t, processed, 1)

				// check that the account was credited
				acc, err := ds.Accounts.GetAccount(ctx, db, []byte("account1"))
				require.NoError(t, err)
				require.Equal(t, big.NewInt(100), acc.Balance)
			},
		},
		{
			name: "vote without providing body does not confirm",
			fn: func(t *testing.T, v *voting.VoteProcessor, ds *voting.Datastores, db sql.DB) {
				// add one voter, create vote, approve, and resolve
				ctx := context.Background()

				// add voter
				err := v.UpdateVoter(ctx, db, []byte("voter1"), 10)
				require.NoError(t, err)

				// create vote with body
				body := &exampleResolutionPayload{
					UniqueID: "unique_id",
					Account:  []byte("account1"),
					Amount:   100,
				}
				bts, err := body.MarshalBinary()
				require.NoError(t, err)
				event := &types.VotableEvent{
					Body: bts,
					Type: examplePayloadType,
				}

				// approve vote, before creating vote
				// expiration does not matter here since it only matters for the first vote
				err = v.Approve(ctx, db, event.ID(), 10323, []byte("voter1"))
				require.NoError(t, err)

				// resolve vote
				processed, err := v.ProcessConfirmedResolutions(ctx, db)
				require.NoError(t, err)
				require.Len(t, processed, 0)

				// check that the account was not credited
				acc, err := ds.Accounts.GetAccount(ctx, db, []byte("account1"))
				require.NoError(t, err)
				require.Equal(t, big.NewInt(0), acc.Balance)
			},
		},
		{
			name: "insufficient voting power",
			fn: func(t *testing.T, v *voting.VoteProcessor, ds *voting.Datastores, db sql.DB) {
				// add one voter, create vote, approve, and resolve
				ctx := context.Background()

				// add voter 1
				err := v.UpdateVoter(ctx, db, []byte("voter1"), 10)
				require.NoError(t, err)

				// add voter 2
				err = v.UpdateVoter(ctx, db, []byte("voter2"), 1000)
				require.NoError(t, err)

				// create vote with body
				body := &exampleResolutionPayload{
					UniqueID: "unique_id",
					Account:  []byte("account1"),
					Amount:   100,
				}
				bts, err := body.MarshalBinary()
				require.NoError(t, err)
				event := &types.VotableEvent{
					Body: bts,
					Type: examplePayloadType,
				}

				err = v.CreateResolution(ctx, db, event, 10000)
				require.NoError(t, err)

				// approve vote
				// expiration does not matter here since it only matters for the first vote
				err = v.Approve(ctx, db, event.ID(), 10323, []byte("voter1"))
				require.NoError(t, err)

				// resolve votes, it will fail since voter 2 did not approve
				processed, err := v.ProcessConfirmedResolutions(ctx, db)
				require.NoError(t, err)
				require.Len(t, processed, 0)

				// check that the resolution still exists
				res, err := v.GetResolutionVoteInfo(ctx, db, event.ID())
				require.NoError(t, err)

				require.Equal(t, event.ID(), res.ID)
				if !bytes.EqualFold(bts, res.Body) {
					require.Equal(t, bts, res.Body) // will fail since the bytes are not equal
				}
				require.Equal(t, examplePayloadType, res.Type)
				require.Equal(t, int64(10000), res.Expiration)
				require.Equal(t, int64(10), res.ApprovedPower)

				// get votes by category
				resolutions, err := v.GetVotesByCategory(ctx, db, examplePayloadType)
				require.NoError(t, err)
				require.Len(t, resolutions, 1)

				requireEqualResolutions(t, resolutions[0], res)
			},
		},
		{
			name: "ByCategory does not panic when no votes exist",
			fn: func(t *testing.T, v *voting.VoteProcessor, ds *voting.Datastores, db sql.DB) {
				ctx := context.Background()

				// get votes by category
				resolutions, err := v.GetVotesByCategory(ctx, db, examplePayloadType)
				require.NoError(t, err)
				require.Len(t, resolutions, 0)
			},
		},
		{
			name: "Get and ByCategory do not panic with no body",
			fn: func(t *testing.T, v *voting.VoteProcessor, ds *voting.Datastores, db sql.DB) {
				ctx := context.Background()

				// add voter
				err := v.UpdateVoter(ctx, db, []byte("voter1"), 10)
				require.NoError(t, err)

				// create vote with body
				body := &exampleResolutionPayload{
					UniqueID: "unique_id",
					Account:  []byte("account1"),
					Amount:   100,
				}
				bts, err := body.MarshalBinary()
				require.NoError(t, err)
				event := &types.VotableEvent{
					Body: bts,
					Type: examplePayloadType,
				}

				// approve vote
				// expiration does not matter here since it only matters for the first vote
				err = v.Approve(ctx, db, event.ID(), 10323, []byte("voter1"))
				require.NoError(t, err)

				// get votes by category, should fail since categories do not get defined until
				// the body is set
				resolutions, err := v.GetVotesByCategory(ctx, db, examplePayloadType)
				require.NoError(t, err)
				require.Len(t, resolutions, 0)

				// get vote by id
				res, err := v.GetResolutionVoteInfo(ctx, db, event.ID())
				require.NoError(t, err)

				// check id is same
				require.Equal(t, event.ID(), res.ID)
				// body is nil, expiration is same, approved power is same, type is nil b/c body is not set
				require.Nil(t, res.Body)
				require.Equal(t, int64(10323), res.Expiration)
				require.Equal(t, int64(10), res.ApprovedPower)
			},
		},
		{
			name: "manipulating voting power",
			fn: func(t *testing.T, v *voting.VoteProcessor, ds *voting.Datastores, db sql.DB) {
				ctx := context.Background()

				// add voter
				err := v.UpdateVoter(ctx, db, []byte("voter1"), 10)
				require.NoError(t, err)

				// add power
				err = v.UpdateVoter(ctx, db, []byte("voter1"), 10)
				require.NoError(t, err)

				// get power
				power, err := v.GetVoterPower(ctx, db, []byte("voter1"))
				require.NoError(t, err)

				require.Equal(t, int64(20), power)

				// subtract power
				err = v.UpdateVoter(ctx, db, []byte("voter1"), -10)
				require.NoError(t, err)

				// get power
				power, err = v.GetVoterPower(ctx, db, []byte("voter1"))
				require.NoError(t, err)

				require.Equal(t, int64(10), power)

				// ensure power cannot go to 0
				err = v.UpdateVoter(ctx, db, []byte("voter1"), -10)
				if err == nil {
					t.Fatal("expected error")
				}

				// remove
				err = v.UpdateVoter(ctx, db, []byte("voter1"), 0)
				require.NoError(t, err)

				// get power
				power, err = v.GetVoterPower(ctx, db, []byte("voter1"))
				require.NoError(t, err)

				require.Equal(t, int64(0), power)
			},
		},
		{
			name: "non-existent voter cannot vote",
			fn: func(t *testing.T, v *voting.VoteProcessor, ds *voting.Datastores, db sql.DB) {
				ctx := context.Background()

				// create vote with body
				body := &exampleResolutionPayload{
					UniqueID: "unique_id",
					Account:  []byte("account1"),
					Amount:   100,
				}
				bts, err := body.MarshalBinary()
				require.NoError(t, err)
				event := &types.VotableEvent{
					Body: bts,
					Type: examplePayloadType,
				}

				// approve vote
				// expiration does not matter here since it only matters for the first vote
				err = v.Approve(ctx, db, event.ID(), 10323, []byte("voter1"))
				if err == nil {
					t.Fatal("expected error")
				}
			},
		},
		{
			name: "expiration works",
			fn: func(t *testing.T, v *voting.VoteProcessor, ds *voting.Datastores, db sql.DB) {
				// create 3 votes
				// expire on 2
				ctx := context.Background()

				// add voter
				err := v.UpdateVoter(ctx, db, []byte("voter1"), 10)
				require.NoError(t, err)

				// payload
				body1 := &exampleResolutionPayload{
					UniqueID: "unique_id",
					Account:  []byte("account1"),
					Amount:   100,
				}
				bts1, err := body1.MarshalBinary()
				require.NoError(t, err)
				event1 := &types.VotableEvent{
					Body: bts1,
					Type: examplePayloadType,
				}

				body2 := &exampleResolutionPayload{
					UniqueID: "unique_id2",
					Account:  []byte("account1"),
					Amount:   100,
				}
				bts2, err := body2.MarshalBinary()
				require.NoError(t, err)
				event2 := &types.VotableEvent{
					Body: bts2,
					Type: examplePayloadType,
				}

				body3 := &exampleResolutionPayload{
					UniqueID: "unique_id3",
					Account:  []byte("account1"),
					Amount:   100,
				}
				bts3, err := body3.MarshalBinary()
				require.NoError(t, err)
				event3 := &types.VotableEvent{
					Body: bts3,
					Type: examplePayloadType,
				}

				// create vote 1
				err = v.CreateResolution(ctx, db, event1, 2)
				require.NoError(t, err)

				err = v.CreateResolution(ctx, db, event2, 3)
				require.NoError(t, err)

				err = v.CreateResolution(ctx, db, event3, 4)
				require.NoError(t, err)

				// get votes by category
				resolutions, err := v.GetVotesByCategory(ctx, db, examplePayloadType)
				require.NoError(t, err)

				require.Len(t, resolutions, 3)

				// expire
				err = v.Expire(ctx, db, 3)
				require.NoError(t, err)

				// get votes by category
				resolutions, err = v.GetVotesByCategory(ctx, db, examplePayloadType)
				require.NoError(t, err)

				require.Len(t, resolutions, 1)
			},
		},
		{
			name: "double approve does nothing",
			fn: func(t *testing.T, v *voting.VoteProcessor, ds *voting.Datastores, db sql.DB) {
				ctx := context.Background()

				// add voter
				err := v.UpdateVoter(ctx, db, []byte("voter1"), 10)
				require.NoError(t, err)

				// payload
				body := &exampleResolutionPayload{
					UniqueID: "unique_id",
					Account:  []byte("account1"),
					Amount:   100,
				}
				bts, err := body.MarshalBinary()
				require.NoError(t, err)
				event := &types.VotableEvent{
					Body: bts,
					Type: examplePayloadType,
				}

				// approve vote twice
				err = v.Approve(ctx, db, event.ID(), 10323, []byte("voter1"))
				require.NoError(t, err)

				err = v.Approve(ctx, db, event.ID(), 10323, []byte("voter1"))
				require.NoError(t, err)
			},
		},
		{
			name: "approval correctly indicates if it contains a body",
			fn: func(t *testing.T, v *voting.VoteProcessor, ds *voting.Datastores, db sql.DB) {
				ctx := context.Background()

				// add voters
				err := v.UpdateVoter(ctx, db, []byte("voter1"), 10)
				require.NoError(t, err)

				// body
				body := &exampleResolutionPayload{
					UniqueID: "unique_id",
					Account:  []byte("account1"),
					Amount:   100,
				}
				bts, err := body.MarshalBinary()
				require.NoError(t, err)
				event := &types.VotableEvent{
					Body: bts,
					Type: examplePayloadType,
				}

				res, err := v.ContainsBodyOrFinished(ctx, db, event.ID())
				require.False(t, res)
				require.NoError(t, err)

				// approve vote
				err = v.Approve(ctx, db, event.ID(), 10323, []byte("voter1"))
				require.NoError(t, err)

				hasBody, err := v.ContainsBodyOrFinished(ctx, db, event.ID())
				require.NoError(t, err)
				require.False(t, hasBody)

				// create vote
				err = v.CreateResolution(ctx, db, event, 10000)
				require.NoError(t, err)

				hasBody, err = v.ContainsBodyOrFinished(ctx, db, event.ID())
				require.NoError(t, err)
				require.True(t, hasBody)
			},
		},
		{
			name: "test HasVoted",
			fn: func(t *testing.T, v *voting.VoteProcessor, ds *voting.Datastores, db sql.DB) {
				ctx := context.Background()

				body := &exampleResolutionPayload{
					UniqueID: "unique_id",
					Account:  []byte("account1"),
					Amount:   100,
				}
				bts, err := body.MarshalBinary()
				require.NoError(t, err)

				event := &types.VotableEvent{
					Body: bts,
					Type: examplePayloadType,
				}

				// hasVoted, no voter
				hasVoted, err := v.HasVoted(ctx, db, event.ID(), []byte("voter1"))
				require.NoError(t, err)
				require.False(t, hasVoted)

				// add voter
				err = v.UpdateVoter(ctx, db, []byte("voter1"), 10)
				require.NoError(t, err)

				// hasVoted, no vote
				hasVoted, err = v.HasVoted(ctx, db, event.ID(), []byte("voter1"))
				require.NoError(t, err)
				require.False(t, hasVoted)

				// approve vote
				err = v.Approve(ctx, db, event.ID(), 10323, []byte("voter1"))
				require.NoError(t, err)

				// hasVoted, vote
				hasVoted, err = v.HasVoted(ctx, db, event.ID(), []byte("voter1"))
				require.NoError(t, err)
				require.True(t, hasVoted)
			},
		},
		{
			name: "voting and giving a body for a finalized vote does nothing",
			fn: func(t *testing.T, v *voting.VoteProcessor, ds *voting.Datastores, db sql.DB) {
				ctx := context.Background()

				// add voters
				err := v.UpdateVoter(ctx, db, []byte("voter1"), 10)
				require.NoError(t, err)

				// body
				body := &exampleResolutionPayload{
					UniqueID: "unique_id",
					Account:  []byte("account1"),
					Amount:   100,
				}
				bts, err := body.MarshalBinary()
				require.NoError(t, err)
				event := &types.VotableEvent{
					Body: bts,
					Type: examplePayloadType,
				}

				// create vote
				err = v.CreateResolution(ctx, db, event, 10000)
				require.NoError(t, err)

				// approve vote
				err = v.Approve(ctx, db, event.ID(), 10323, []byte("voter1"))
				require.NoError(t, err)

				// resolve vote
				processed, err := v.ProcessConfirmedResolutions(ctx, db)
				require.NoError(t, err)
				require.Len(t, processed, 1)

				// give body
				err = v.CreateResolution(ctx, db, event, 10000)
				require.NoError(t, err)

				// approve vote
				err = v.Approve(ctx, db, event.ID(), 10323, []byte("voter1"))
				require.NoError(t, err)

				// get votes by category
				resolutions, err := v.GetVotesByCategory(ctx, db, examplePayloadType)
				require.NoError(t, err)
				require.Len(t, resolutions, 0)

				// get vote by id
				_, err = v.GetResolutionVoteInfo(ctx, db, event.ID())
				require.ErrorIs(t, err, voting.ErrResolutionNotFound)
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			// Can't use votingSchemaName because this is `package voting_test`
			// Either we convert this to `package voting`, use
			// voting.DropAllTables, or use a literal here:
			db, err := dbtest.NewTestDB(t)
			require.NoError(t, err)
			defer db.Close()

			dbTx, err := db.BeginTx(ctx)
			require.NoError(t, err)
			defer dbTx.Rollback(ctx) // always rollback to ensure cleanup

			ds := &voting.Datastores{
				Accounts: &mockAccountStore{
					accounts: map[string]*accounts.Account{},
				},
				Databases: nil,
			}

			// voting.DropAllTables(ctx, db)

			v, err := voting.NewVoteProcessor(ctx, dbTx, ds.Accounts, ds.Databases, 500000, log.NewStdOut(log.DebugLevel))
			if err != nil {
				t.Fatal(err)
			}

			// defer func() {
			// 	if err := voting.DropAllTables(ctx, db); err != nil {
			// 		t.Errorf("test table cleanup failed: %v", err)
			// 	}
			// }()

			tt.fn(t, v, ds, dbTx)
		})
	}
}

// requireEqualResolutions is a helper function to compare two resolutions.
// 1 is a resolution, the other is a resolution status
func requireEqualResolutions(t *testing.T, res1 *voting.Resolution, res2 *voting.ResolutionVoteInfo) {
	require.Equal(t, res1.ID, res2.ID)
	if !bytes.EqualFold(res1.Body, res2.Body) {
		require.Equal(t, res1.Body, res2.Body) // will fail since the bytes are not equal
	}
	require.Equal(t, res1.Type, res2.Type)
	require.Equal(t, res1.Expiration, res2.Expiration)
}

type mockAccountStore struct {
	accounts map[string]*accounts.Account
}

func (m *mockAccountStore) GetAccount(ctx context.Context, _ sql.DB, identifier []byte) (*accounts.Account, error) {
	acc, ok := m.accounts[string(identifier)]
	if !ok {
		acc = &accounts.Account{
			Identifier: identifier,
			Balance:    big.NewInt(0),
			Nonce:      0,
		}
		m.accounts[string(identifier)] = acc
	}

	return acc, nil
}

func (m *mockAccountStore) Credit(ctx context.Context, _ sql.DB, account []byte, amount *big.Int) error {
	acc, ok := m.accounts[string(account)]
	if !ok {
		acc = &accounts.Account{
			Identifier: account,
			Balance:    big.NewInt(0),
			Nonce:      0,
		}
		m.accounts[string(account)] = acc
	}

	acc.Balance = new(big.Int).Add(acc.Balance, amount)

	return nil
}

// exampleResolutionPayload is an example payload that can be used for testing
// we can use json encoding since it is a local unit test
type exampleResolutionPayload struct {
	UniqueID string `json:"unique_id"` // could be a transaction hash from a different chain
	Account  []byte `json:"account"`
	Amount   int64  `json:"amount"`
}

func (e *exampleResolutionPayload) Apply(ctx context.Context, db sql.DB, datastores voting.Datastores, logger log.Logger) error {
	if e.Account == nil {
		return fmt.Errorf("account is required")
	}

	if e.Amount <= 0 {
		return fmt.Errorf("amount must be greater than 0")
	}

	return datastores.Accounts.Credit(ctx, db, e.Account, big.NewInt(e.Amount))
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
