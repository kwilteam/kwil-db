// go:build pglive

package voting_test

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	"github.com/kwilteam/kwil-db/common/sql"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/extensions/resolutions"
	dbtest "github.com/kwilteam/kwil-db/internal/sql/pg/test"
	"github.com/kwilteam/kwil-db/internal/voting"
	"github.com/stretchr/testify/require"
)

const testType = "test"

func init() {
	err := resolutions.RegisterResolution(testType, resolutions.ResolutionConfig{})
	if err != nil {
		panic(err)
	}
}

func TestMain(m *testing.M) {
	// pg.UseLogger(log.NewStdOut(log.InfoLevel))
	m.Run()
}

func Test_Voting(t *testing.T) {
	type testcase struct {
		name          string
		startingPower map[string]int64 // starting power for any validators
		fn            func(t *testing.T, db sql.DB)
	}

	tests := []testcase{
		{
			name: "successful creationg and voting",
			startingPower: map[string]int64{
				"a": 100,
				"b": 100,
			},
			fn: func(t *testing.T, db sql.DB) {
				ctx := context.Background()

				err := voting.CreateResolution(ctx, db, testEvent, 10, []byte("a"))
				require.NoError(t, err)

				err = voting.ApproveResolution(ctx, db, testEvent.ID(), 10, []byte("a"))
				require.NoError(t, err)

				err = voting.ApproveResolution(ctx, db, testEvent.ID(), 10, []byte("b"))
				require.NoError(t, err)

				events, err := voting.GetResolutionsByThresholdAndType(ctx, db, testConfirmationThreshold, testType)
				require.NoError(t, err)

				require.Len(t, events, 1)

				require.Equal(t, testEvent.Body, events[0].Body)
				require.Equal(t, testEvent.Type, events[0].Type)
				require.Equal(t, testEvent.ID(), events[0].ID)
				require.Equal(t, int64(10), events[0].ExpirationHeight)
				require.False(t, events[0].DoubleProposerVote)
				require.Equal(t, int64(200), events[0].ApprovedPower)
			},
		},
		{
			name: "validator management",
			fn: func(t *testing.T, db sql.DB) {
				// I add power here because this is part of the domain of validator management
				// if test setup changes, this test will still be valid
				err := voting.SetValidatorPower(context.Background(), db, []byte("a"), 100)
				require.NoError(t, err)

				err = voting.SetValidatorPower(context.Background(), db, []byte("b"), 100)
				require.NoError(t, err)

				voters, err := voting.GetValidators(context.Background(), db)
				require.NoError(t, err)

				require.Len(t, voters, 2)

				voterAPower, err := voting.GetValidatorPower(context.Background(), db, []byte("a"))
				require.NoError(t, err)

				require.Equal(t, int64(100), voterAPower)
			},
		},
		{
			name: "deletion and processed",
			startingPower: map[string]int64{
				"a": 100,
			},
			fn: func(t *testing.T, db sql.DB) {
				ctx := context.Background()

				err := voting.CreateResolution(ctx, db, testEvent, 10, []byte("a"))
				require.NoError(t, err)

				err = voting.DeleteResolutions(ctx, db, testEvent.ID())
				require.NoError(t, err)

				processed, err := voting.IsProcessed(ctx, db, testEvent.ID())
				require.NoError(t, err)

				require.False(t, processed)

				err = voting.MarkProcessed(ctx, db, testEvent.ID())
				require.NoError(t, err)

				processed, err = voting.IsProcessed(ctx, db, testEvent.ID())
				require.NoError(t, err)

				require.True(t, processed)
			},
		},
		{
			name: "reading resolution info",
			startingPower: map[string]int64{
				"a": 100,
				"b": 100,
			},
			fn: func(t *testing.T, db sql.DB) {
				ctx := context.Background()

				// validator 1 will approve first here, to test that it is properly ordered

				err := voting.ApproveResolution(ctx, db, testEvent.ID(), 10, []byte("a"))
				require.NoError(t, err)

				err = voting.CreateResolution(ctx, db, testEvent, 10, []byte("a"))
				require.NoError(t, err)

				err = voting.ApproveResolution(ctx, db, testEvent.ID(), 10, []byte("b"))
				require.NoError(t, err)

				info, err := voting.GetResolutionInfo(ctx, db, testEvent.ID())
				require.NoError(t, err)

				infoSlice, err := voting.GetResolutionsByType(ctx, db, testType)
				require.NoError(t, err)
				require.Len(t, infoSlice, 1)

				require.EqualValues(t, testEvent.ID(), infoSlice[0].ID)

				info2Slice, err := voting.GetResolutionIDsByTypeAndProposer(ctx, db, testType, []byte("a"))
				require.NoError(t, err)
				require.Len(t, info2Slice, 1)

				require.Equal(t, infoSlice[0].ID, info2Slice[0])

				require.Equal(t, testEvent.Body, info.Body)
				require.Equal(t, testEvent.Type, info.Type)
				require.Equal(t, testEvent.ID(), info.ID)
				require.Equal(t, int64(10), info.ExpirationHeight)
				require.True(t, info.DoubleProposerVote)
				require.Equal(t, int64(200), info.ApprovedPower)

				hasValidator1Info := false
				hasValidator2Info := false

				for _, voter := range info.Voters {
					if string(voter.PubKey) == "a" && voter.Power == 100 {
						hasValidator1Info = true
					}

					if string(voter.PubKey) == "b" && voter.Power == 100 {
						hasValidator2Info = true
					}
				}
				if !hasValidator1Info || !hasValidator2Info {
					t.Errorf("expected to find both validators in the voters list")
				}
			},
		},
		{
			name: "test expiration",
			startingPower: map[string]int64{
				"a": 100,
			},
			fn: func(t *testing.T, db sql.DB) {
				ctx := context.Background()

				err := voting.CreateResolution(ctx, db, testEvent, 10, []byte("a"))
				require.NoError(t, err)

				expired, err := voting.GetExpired(ctx, db, 10)
				require.NoError(t, err)
				require.Equal(t, 1, len(expired))

				resolutionInfo, err := voting.GetResolutionInfo(ctx, db, testEvent.ID())
				require.NoError(t, err)

				require.EqualValues(t, resolutionInfo, expired[0])
			},
		},
		{
			name: "many resolutions test",
			startingPower: map[string]int64{
				"a": 100,
			},
			fn: func(t *testing.T, db sql.DB) {
				ctx := context.Background()

				events := make([]*types.VotableEvent, 3)
				ids := make([]types.UUID, 3)
				for i := 0; i < 3; i++ {
					events[i] = &types.VotableEvent{
						Body: []byte("test" + fmt.Sprint(i)),
						Type: testType,
					}

					ids[i] = events[i].ID()
				}

				// we will create and approve 1,
				// create 2,
				// and only approve 3

				err := voting.CreateResolution(ctx, db, events[0], 10, []byte("a"))
				require.NoError(t, err)
				err = voting.ApproveResolution(ctx, db, events[0].ID(), 10, []byte("a"))
				require.NoError(t, err)

				err = voting.CreateResolution(ctx, db, events[1], 10, []byte("a"))
				require.NoError(t, err)

				err = voting.ApproveResolution(ctx, db, events[2].ID(), 10, []byte("a"))
				require.NoError(t, err)

				containsBody, err := voting.ResolutionsContainBody(ctx, db, ids...)
				require.NoError(t, err)

				require.True(t, containsBody[0])
				require.True(t, containsBody[1])
				require.False(t, containsBody[2])

				// delete and process all
				err = voting.DeleteResolutions(ctx, db, ids...)
				require.NoError(t, err)
				err = voting.MarkProcessed(ctx, db, ids...)
				require.NoError(t, err)

				// check that they are all processed
				processed, err := voting.ManyAreProcessed(ctx, db, ids...)
				require.NoError(t, err)

				for _, p := range processed {
					require.True(t, p)
				}
			},
		},
		{
			name: "no resolutions",
			startingPower: map[string]int64{
				"a": 100,
			},
			fn: func(t *testing.T, db sql.DB) {
				ctx := context.Background()

				containsBody, err := voting.ResolutionsContainBody(ctx, db, types.NewUUIDV5([]byte("ss")))
				require.NoError(t, err)

				require.False(t, containsBody[0])

				processed, err := voting.ManyAreProcessed(ctx, db, types.NewUUIDV5([]byte("ss")))
				require.NoError(t, err)

				require.False(t, processed[0])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			db, err := dbtest.NewTestDB(t)
			require.NoError(t, err)
			defer db.Close()

			dbTx, err := db.BeginTx(ctx)
			require.NoError(t, err)
			defer dbTx.Rollback(ctx) // always rollback to ensure cleanup

			err = voting.InitializeVoteStore(ctx, dbTx)
			require.NoError(t, err)

			for addr, power := range tt.startingPower {
				err = voting.SetValidatorPower(ctx, dbTx, []byte(addr), power)
				require.NoError(t, err)
			}

			tt.fn(t, dbTx)
		})
	}
}

var testEvent = &types.VotableEvent{
	Body: []byte("test"),
	Type: testType,
}

var testConfirmationThreshold = big.NewRat(2, 3)
