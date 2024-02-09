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
		fn   func(t *testing.T, v *voting.VoteProcessor, ds *voting.Datastores)
	}

	testCases := []testCase{
		{
			name: "Required Power",
			fn: func(t *testing.T, v *voting.VoteProcessor, ds *voting.Datastores) {
				ctx := context.Background()

				// add 4 voters
				for i := 0; i < 4; i++ {
					voterid := fmt.Sprintf("voter%d", i)
					err := v.UpdateVoter(ctx, []byte(voterid), 1)
					require.NoError(t, err)
				}

				thresholdPower, err := v.RequiredPower(ctx, 2, 3)
				require.NoError(t, err)
				// threshold power is >= 2/3 * 4 = 3
				require.Equal(t, int64(3), thresholdPower)

				// add 2 more voter
				for i := 4; i < 6; i++ {
					voterid := fmt.Sprintf("voter%d", i)
					err := v.UpdateVoter(ctx, []byte(voterid), 1)
					require.NoError(t, err)
				}

				thresholdPower, err = v.RequiredPower(ctx, 2, 3)
				require.NoError(t, err)
				// threshold power is 2/3 * 6 = 4
				require.Equal(t, int64(4), thresholdPower)
			},
		},
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
				event := &types.VotableEvent{
					Body: bts,
					Type: examplePayloadType,
				}

				err = v.CreateResolution(ctx, event, 10000, []byte("voter1"))
				require.NoError(t, err)

				// approve vote
				// expiration does not matter here since it only matters for the first vote
				err = v.Approve(ctx, event.ID(), 10323, []byte("voter1"))
				require.NoError(t, err)

				// Get resolution info
				res, err := v.GetResolutionVoteInfo(ctx, event.ID())
				require.NoError(t, err)
				// submittedBodyAndID should be false
				require.False(t, res.SubmittedBodyAndID)

				// resolve vote
				processed, err := v.ProcessConfirmedResolutions(ctx)
				require.NoError(t, err)
				require.Len(t, processed, 1)

				// check that the account was credited
				acc, err := ds.Accounts.GetAccount(ctx, []byte("account1"))
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
				event := &types.VotableEvent{
					Body: bts,
					Type: examplePayloadType,
				}

				// approve vote, before creating vote
				// expiration does not matter here since it only matters for the first vote
				err = v.Approve(ctx, event.ID(), 10323, []byte("voter1"))
				require.NoError(t, err)

				// Get resolution info
				res, err := v.GetResolutionVoteInfo(ctx, event.ID())
				require.NoError(t, err)
				// voteBody and voteBodyProposer should be null
				require.Nil(t, res.Body)
				require.Nil(t, res.VoteBodyProposer)
				require.Equal(t, 1, len(res.Voters))

				// now create the vote
				err = v.CreateResolution(ctx, event, 10000, []byte("voter1"))
				require.NoError(t, err)

				// Get resolution info
				res, err = v.GetResolutionVoteInfo(ctx, event.ID())
				require.NoError(t, err)
				// submittedBodyAndID should be true
				require.True(t, res.SubmittedBodyAndID)

				// resolve vote
				processed, err := v.ProcessConfirmedResolutions(ctx)
				require.NoError(t, err)
				require.Len(t, processed, 1)

				// check that the account was credited
				acc, err := ds.Accounts.GetAccount(ctx, []byte("account1"))
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
				event := &types.VotableEvent{
					Body: bts,
					Type: examplePayloadType,
				}

				// approve vote, before creating vote
				// expiration does not matter here since it only matters for the first vote
				err = v.Approve(ctx, event.ID(), 10323, []byte("voter1"))
				require.NoError(t, err)

				// resolve vote
				processed, err := v.ProcessConfirmedResolutions(ctx)
				require.NoError(t, err)
				require.Len(t, processed, 0)

				// check that the account was not credited
				acc, err := ds.Accounts.GetAccount(ctx, []byte("account1"))
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
				event := &types.VotableEvent{
					Body: bts,
					Type: examplePayloadType,
				}

				err = v.CreateResolution(ctx, event, 10000, []byte("voter1"))
				require.NoError(t, err)

				// approve vote
				// expiration does not matter here since it only matters for the first vote
				err = v.Approve(ctx, event.ID(), 10323, []byte("voter1"))
				require.NoError(t, err)

				// resolve votes, it will fail since voter 2 did not approve
				processed, err := v.ProcessConfirmedResolutions(ctx)
				require.NoError(t, err)
				require.Len(t, processed, 0)

				// check that the resolution still exists
				res, err := v.GetResolutionVoteInfo(ctx, event.ID())
				require.NoError(t, err)

				require.Equal(t, event.ID(), res.ID)
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
				event := &types.VotableEvent{
					Body: bts,
					Type: examplePayloadType,
				}

				// approve vote
				// expiration does not matter here since it only matters for the first vote
				err = v.Approve(ctx, event.ID(), 10323, []byte("voter1"))
				require.NoError(t, err)

				// get votes by category, should fail since categories do not get defined until
				// the body is set
				resolutions, err := v.GetVotesByCategory(ctx, examplePayloadType)
				require.NoError(t, err)
				require.Len(t, resolutions, 0)

				// get vote by id
				res, err := v.GetResolutionVoteInfo(ctx, event.ID())
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
			fn: func(t *testing.T, v *voting.VoteProcessor, ds *voting.Datastores) {
				ctx := context.Background()

				// add voter
				err := v.UpdateVoter(ctx, []byte("voter1"), 10)
				require.NoError(t, err)

				// Update power
				err = v.UpdateVoter(ctx, []byte("voter1"), 20)
				require.NoError(t, err)

				// get power
				power, err := v.GetVoterPower(ctx, []byte("voter1"))
				require.NoError(t, err)

				require.Equal(t, int64(20), power)

				// subtract power
				err = v.UpdateVoter(ctx, []byte("voter1"), 10)
				require.NoError(t, err)

				// get power
				power, err = v.GetVoterPower(ctx, []byte("voter1"))
				require.NoError(t, err)

				require.Equal(t, int64(10), power)

				// Update to negative power, this should delete the voter
				err = v.UpdateVoter(ctx, []byte("voter1"), -10)
				require.NoError(t, err)

				// get power
				power, err = v.GetVoterPower(ctx, []byte("voter1"))
				require.NoError(t, err)
				require.Equal(t, int64(0), power)

				// Update power to 10
				err = v.UpdateVoter(ctx, []byte("voter1"), 10)
				require.NoError(t, err)

				// get power
				power, err = v.GetVoterPower(ctx, []byte("voter1"))
				require.NoError(t, err)
				require.Equal(t, int64(10), power)

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
				event := &types.VotableEvent{
					Body: bts,
					Type: examplePayloadType,
				}

				// approve vote
				// expiration does not matter here since it only matters for the first vote
				err = v.Approve(ctx, event.ID(), 10323, []byte("voter1"))
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
				err = v.CreateResolution(ctx, event1, 2, []byte("voter1"))
				require.NoError(t, err)

				err = v.CreateResolution(ctx, event2, 3, []byte("voter1"))
				require.NoError(t, err)

				err = v.CreateResolution(ctx, event3, 4, []byte("voter1"))
				require.NoError(t, err)

				// get votes by category
				resolutions, err := v.GetVotesByCategory(ctx, examplePayloadType)
				require.NoError(t, err)

				require.Len(t, resolutions, 3)

				// expire
				err = v.Expire(ctx, 3)
				require.NoError(t, err)

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
				event := &types.VotableEvent{
					Body: bts,
					Type: examplePayloadType,
				}

				// approve vote twice
				err = v.Approve(ctx, event.ID(), 10323, []byte("voter1"))
				require.NoError(t, err)

				err = v.Approve(ctx, event.ID(), 10323, []byte("voter1"))
				require.NoError(t, err)
			},
		},
		{
			name: "ContainsBody",
			fn: func(t *testing.T, v *voting.VoteProcessor, ds *voting.Datastores) {
				ctx := context.Background()

				// add voters
				err := v.UpdateVoter(ctx, []byte("voter1"), 10)
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

				res, err := v.ContainsBody(ctx, event.ID())
				require.NoError(t, err)
				require.False(t, res)

				// approve vote
				err = v.Approve(ctx, event.ID(), 10323, []byte("voter1"))
				require.NoError(t, err)

				res, err = v.ContainsBody(ctx, event.ID())
				require.NoError(t, err)
				require.False(t, res)

				// create vote
				err = v.CreateResolution(ctx, event, 10000, []byte("voter1"))
				require.NoError(t, err)

				res, err = v.ContainsBody(ctx, event.ID())
				require.NoError(t, err)
				require.True(t, res)
			},
		},
		{
			name: "approval correctly indicates if it contains a body",
			fn: func(t *testing.T, v *voting.VoteProcessor, ds *voting.Datastores) {
				ctx := context.Background()

				// add voters
				err := v.UpdateVoter(ctx, []byte("voter1"), 10)
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

				res, err := v.ContainsBodyOrFinished(ctx, event.ID())
				require.False(t, res)
				require.NoError(t, err)

				// approve vote
				err = v.Approve(ctx, event.ID(), 10323, []byte("voter1"))
				require.NoError(t, err)

				hasBody, err := v.ContainsBodyOrFinished(ctx, event.ID())
				require.NoError(t, err)
				require.False(t, hasBody)

				// create vote
				err = v.CreateResolution(ctx, event, 10000, []byte("voter1"))
				require.NoError(t, err)

				hasBody, err = v.ContainsBodyOrFinished(ctx, event.ID())
				require.NoError(t, err)
				require.True(t, hasBody)
			},
		},
		{
			name: "test HasVoted",
			fn: func(t *testing.T, v *voting.VoteProcessor, ds *voting.Datastores) {
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
				hasVoted, err := v.HasVoted(ctx, event.ID(), []byte("voter1"))
				require.NoError(t, err)
				require.False(t, hasVoted)

				// add voter
				err = v.UpdateVoter(ctx, []byte("voter1"), 10)
				require.NoError(t, err)

				// hasVoted, no vote
				hasVoted, err = v.HasVoted(ctx, event.ID(), []byte("voter1"))
				require.NoError(t, err)
				require.False(t, hasVoted)

				// approve vote
				err = v.Approve(ctx, event.ID(), 10323, []byte("voter1"))
				require.NoError(t, err)

				// hasVoted, vote
				hasVoted, err = v.HasVoted(ctx, event.ID(), []byte("voter1"))
				require.NoError(t, err)
				require.True(t, hasVoted)
			},
		},
		{
			name: "voting and giving a body for a finalized vote does nothing",
			fn: func(t *testing.T, v *voting.VoteProcessor, ds *voting.Datastores) {
				ctx := context.Background()

				// add voters
				err := v.UpdateVoter(ctx, []byte("voter1"), 10)
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
				err = v.CreateResolution(ctx, event, 10000, []byte("voter1"))
				require.NoError(t, err)

				// approve vote
				err = v.Approve(ctx, event.ID(), 10323, []byte("voter1"))
				require.NoError(t, err)

				// resolve vote
				processed, err := v.ProcessConfirmedResolutions(ctx)
				require.NoError(t, err)
				require.Len(t, processed, 1)

				// give body
				err = v.CreateResolution(ctx, event, 10000, []byte("voter1"))
				require.NoError(t, err)

				// approve vote
				err = v.Approve(ctx, event.ID(), 10323, []byte("voter1"))
				require.NoError(t, err)

				// get votes by category
				resolutions, err := v.GetVotesByCategory(ctx, examplePayloadType)
				require.NoError(t, err)
				require.Len(t, resolutions, 0)

				// get vote by id
				_, err = v.GetResolutionVoteInfo(ctx, event.ID())
				require.ErrorIs(t, err, voting.ErrResolutionNotFound)
			},
		},
		{
			// Check if the voters are refunded if the resolution is not approved and more than 1/3rd voted.
			name: "Voters are not refunded on expiry of resolution if it has < 1/3rd voting power",
			fn: func(t *testing.T, v *voting.VoteProcessor, ds *voting.Datastores) {
				ctx := context.Background()

				// add voters
				for i := 0; i < 6; i++ {
					err := v.UpdateVoter(ctx, []byte(fmt.Sprintf("voter%d", i)), 10)
					require.NoError(t, err)
				}

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
				err = v.CreateResolution(ctx, event, 10, []byte("voter1"))
				require.NoError(t, err)

				// approve vote
				err = v.Approve(ctx, event.ID(), 10, []byte("voter1"))
				require.NoError(t, err)

				err = v.Expire(ctx, 20)
				require.NoError(t, err)

				// check that the proposer was not refunded as resolution didnt receive 1/3rd votes.
				acc, err := ds.Accounts.GetAccount(ctx, []byte("voter1"))
				require.NoError(t, err)
				require.Equal(t, big.NewInt(0), acc.Balance)
			},
		},
		{
			// Check if the voters are refunded if the resolution is not approved and more than 1/3rd voted.
			name: "refund voters upon resolution expiry if it has >=1/3rd of voting power",
			fn: func(t *testing.T, v *voting.VoteProcessor, ds *voting.Datastores) {
				ctx := context.Background()

				// add voters
				for i := 0; i < 6; i++ {
					err := v.UpdateVoter(ctx, []byte(fmt.Sprintf("voter%d", i)), 10)
					require.NoError(t, err)
				}

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
				err = v.CreateResolution(ctx, event, 10, []byte("voter1"))
				require.NoError(t, err)

				// approve vote
				err = v.Approve(ctx, event.ID(), 10, []byte("voter1"))
				require.NoError(t, err)

				err = v.Approve(ctx, event.ID(), 10, []byte("voter2"))
				require.NoError(t, err)

				err = v.Expire(ctx, 20)
				require.NoError(t, err)

				// check that the proposer was not refunded as resolution didnt receive 1/3rd votes.
				voterFee := big.NewInt(voting.ValidatorVoteIDPrice)
				acc, err := ds.Accounts.GetAccount(ctx, []byte("voter2"))
				require.NoError(t, err)
				require.Equal(t, voterFee, acc.Balance)

				proposerFee := big.NewInt(int64(len(bts)) * voting.ValidatorVoteBodyBytePrice)
				acc, err = ds.Accounts.GetAccount(ctx, []byte("voter1"))
				require.NoError(t, err)
				require.Equal(t, proposerFee, acc.Balance)
			},
		},
		{
			// If proposer has sent both the VoteID and VoteBody, then the proposer should be refunded for both the transaction costs upon expiry.
			name: "refund proposer for both VoteBody and VoteID upon resolution expiry if resolution has >=1/3rd of voting power",
			fn: func(t *testing.T, v *voting.VoteProcessor, ds *voting.Datastores) {
				ctx := context.Background()

				// add voters
				for i := 0; i < 6; i++ {
					err := v.UpdateVoter(ctx, []byte(fmt.Sprintf("voter%d", i)), 10)
					require.NoError(t, err)
				}

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

				// voter1 approves vote
				err = v.Approve(ctx, event.ID(), 10, []byte("voter1"))
				require.NoError(t, err)

				// voter1 creates resolution
				err = v.CreateResolution(ctx, event, 10, []byte("voter1"))
				require.NoError(t, err)

				err = v.Approve(ctx, event.ID(), 10, []byte("voter2"))
				require.NoError(t, err)

				err = v.Expire(ctx, 20)
				require.NoError(t, err)

				// check that the proposer was not refunded as resolution didnt receive 1/3rd votes.
				voterFee := big.NewInt(voting.ValidatorVoteIDPrice)
				acc, err := ds.Accounts.GetAccount(ctx, []byte("voter2"))
				require.NoError(t, err)
				require.Equal(t, voterFee, acc.Balance)

				proposerFee := big.NewInt(int64(len(bts))*voting.ValidatorVoteBodyBytePrice + voting.ValidatorVoteIDPrice)
				acc, err = ds.Accounts.GetAccount(ctx, []byte("voter1"))
				require.NoError(t, err)
				require.Equal(t, proposerFee, acc.Balance)
			},
		},
		{
			// If proposer has sent both the VoteID and VoteBody, then the proposer should be refunded for both the transaction costs upon approval of resolution.
			name: "refund proposer for both VoteBody and VoteID upon resolution approval",
			fn: func(t *testing.T, v *voting.VoteProcessor, ds *voting.Datastores) {
				ctx := context.Background()

				// add voters
				for i := 0; i < 3; i++ {
					err := v.UpdateVoter(ctx, []byte(fmt.Sprintf("voter%d", i)), 10)
					require.NoError(t, err)
				}

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

				// voter1 approves vote
				err = v.Approve(ctx, event.ID(), 10, []byte("voter1"))
				require.NoError(t, err)

				// voter1 creates resolution
				err = v.CreateResolution(ctx, event, 10, []byte("voter1"))
				require.NoError(t, err)

				err = v.Approve(ctx, event.ID(), 10, []byte("voter2"))
				require.NoError(t, err)

				ids, err := v.ProcessConfirmedResolutions(ctx)
				require.NoError(t, err)
				require.Len(t, ids, 1)

				// check that the proposer was not refunded as resolution didnt receive 1/3rd votes.
				voterFee := big.NewInt(voting.ValidatorVoteIDPrice)
				acc, err := ds.Accounts.GetAccount(ctx, []byte("voter2"))
				require.NoError(t, err)
				require.Equal(t, voterFee, acc.Balance)

				proposerFee := big.NewInt(int64(len(bts))*voting.ValidatorVoteBodyBytePrice + voting.ValidatorVoteIDPrice)
				acc, err = ds.Accounts.GetAccount(ctx, []byte("voter1"))
				require.NoError(t, err)
				require.Equal(t, proposerFee, acc.Balance)
			},
		},
		{
			name: "Get and ByCategory without body",
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
				event := &types.VotableEvent{
					Body: bts,
					Type: examplePayloadType,
				}

				// approve vote
				// expiration does not matter here since it only matters for the first vote
				err = v.Approve(ctx, event.ID(), 10323, []byte("voter1"))
				require.NoError(t, err)

				// get votes by category, should fail since categories do not get defined until
				// the body is set
				resolutions, err := v.GetVotesByCategory(ctx, examplePayloadType)
				require.NoError(t, err)
				require.Len(t, resolutions, 0)

				// get vote by id
				res, err := v.GetResolutionVoteInfo(ctx, event.ID())
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
			name: "Get and ByCategory with body and voters",
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
				event := &types.VotableEvent{
					Body: bts,
					Type: examplePayloadType,
				}

				// approve vote
				// expiration does not matter here since it only matters for the first vote
				err = v.Approve(ctx, event.ID(), 10323, []byte("voter1"))
				require.NoError(t, err)

				// get votes by category, should fail since categories do not get defined until
				// the body is set
				resolutions, err := v.GetVotesByCategory(ctx, examplePayloadType)
				require.NoError(t, err)
				require.Len(t, resolutions, 0)

				// get vote by id
				res, err := v.GetResolutionVoteInfo(ctx, event.ID())
				require.NoError(t, err)

				// check id is same
				require.Equal(t, event.ID(), res.ID)
				// body is nil, expiration is same, approved power is same, type is nil b/c body is not set
				require.Nil(t, res.Body)
				require.Equal(t, int64(10323), res.Expiration)
				require.Equal(t, int64(10), res.ApprovedPower)
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			// Can't use votingSchemaName because this is `package voting_test`
			// Either we convert this to `package voting`, use
			// voting.DropAllTables, or use a literal here:
			db, cleanUp, err := dbtest.NewTestPool(ctx, []string{`kwild_voting`})
			require.NoError(t, err)
			defer cleanUp()

			ds := &voting.Datastores{
				Accounts: &mockAccountStore{
					accounts: map[string]*accounts.Account{},
				},
				Databases: nil,
			}

			// voting.DropAllTables(ctx, db)

			v, err := voting.NewVoteProcessor(ctx, db, ds.Accounts, ds.Databases, voting.Threshold{Num: 2, Denom: 3}, log.NewStdOut(log.DebugLevel))
			if err != nil {
				t.Fatal(err)
			}
			// defer func() {
			// 	if err := voting.DropAllTables(ctx, db); err != nil {
			// 		t.Errorf("test table cleanup failed: %v", err)
			// 	}
			// }()

			tt.fn(t, v, ds)
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

func (m *mockAccountStore) GetAccount(ctx context.Context, identifier []byte) (*accounts.Account, error) {
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

func (m *mockAccountStore) Credit(ctx context.Context, account []byte, amount *big.Int) error {
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

func (e *exampleResolutionPayload) Apply(ctx context.Context, datastores voting.Datastores, proposer []byte, voters []voting.Voter, logger log.Logger) error {
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
