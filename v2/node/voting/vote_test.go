//go:build pglive

package voting

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	"kwil/extensions/resolutions"
	dbtest "kwil/node/pg/test"
	"kwil/node/types/sql"
	"kwil/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testType = "test"

func init() {
	err := resolutions.RegisterResolution(testType, resolutions.ModAdd, resolutions.ResolutionConfig{})
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
		fn            func(t *testing.T, db sql.DB, v *VoteStore)
	}

	tests := []testcase{
		{
			name: "successful creating and voting",
			startingPower: map[string]int64{
				"a": 100,
				"b": 100,
			},
			fn: func(t *testing.T, db sql.DB, v *VoteStore) {
				ctx := context.Background()

				err := CreateResolution(ctx, db, dummyEvent, 10, []byte("a"))
				require.Error(t, err)

				// Can't approve non-existent resolutions
				err = ApproveResolution(ctx, db, testEvent.ID(), []byte("a"))
				require.Error(t, err)

				err = CreateResolution(ctx, db, testEvent, 10, []byte("a"))
				require.NoError(t, err)

				// duplicate creation should fail
				err = CreateResolution(ctx, db, testEvent, 10, []byte("a"))
				require.Error(t, err)

				err = ApproveResolution(ctx, db, testEvent.ID(), []byte("a"))
				require.NoError(t, err)

				err = ApproveResolution(ctx, db, testEvent.ID(), []byte("b"))
				require.NoError(t, err)

				events, err := GetResolutionsByThresholdAndType(ctx, db, testConfirmationThreshold, testType, 200)
				require.NoError(t, err)

				require.Len(t, events, 1)

				require.Equal(t, testEvent.Body, events[0].Body)
				require.Equal(t, testEvent.Type, events[0].Type)
				require.Equal(t, testEvent.ID(), events[0].ID)
				require.Equal(t, int64(10), events[0].ExpirationHeight)
				require.Equal(t, int64(200), events[0].ApprovedPower)
			},
		},
		{
			name: "validator management",
			fn: func(t *testing.T, db sql.DB, v *VoteStore) {
				// I add power here because this is part of the domain of validator management
				// if test setup changes, this test will still be valid
				ctx := context.Background()
				err := v.SetValidatorPower(ctx, db, []byte("a"), 100)
				require.NoError(t, err)

				err = v.SetValidatorPower(ctx, db, []byte("b"), 100)
				require.NoError(t, err)

				voters, err := v.GetValidators()
				require.NoError(t, err)
				// Before commit
				require.Len(t, voters, 0)

				err = v.Commit()
				require.NoError(t, err)

				// After commit
				voters, err = v.GetValidators()
				require.NoError(t, err)
				require.Len(t, voters, 2)

				voterAPower, err := v.GetValidatorPower(ctx, db, []byte("a"))
				require.NoError(t, err)

				require.Equal(t, int64(100), voterAPower)
			},
		},
		{
			name: "deletion and processed",
			startingPower: map[string]int64{
				"a": 100,
			},
			fn: func(t *testing.T, db sql.DB, v *VoteStore) {
				ctx := context.Background()

				err := CreateResolution(ctx, db, testEvent, 10, []byte("a"))
				require.NoError(t, err)

				// verify that the resolution exists
				exists, err := ResolutionExists(ctx, db, testEvent.ID())
				require.NoError(t, err)
				require.True(t, exists)

				err = DeleteResolutions(ctx, db, testEvent.ID())
				require.NoError(t, err)

				// verify that the resolution no longer exists
				exists, err = ResolutionExists(ctx, db, testEvent.ID())
				require.NoError(t, err)
				require.False(t, exists)

				processed, err := IsProcessed(ctx, db, testEvent.ID())
				require.NoError(t, err)

				require.False(t, processed)

				err = MarkProcessed(ctx, db, testEvent.ID())
				require.NoError(t, err)

				processed, err = IsProcessed(ctx, db, testEvent.ID())
				require.NoError(t, err)

				// Resolution creation should fail if the resolution is already processed
				err = CreateResolution(ctx, db, testEvent, 10, []byte("a"))
				require.Error(t, err)

				require.True(t, processed)
			},
		},
		{
			name: "reading resolution info",
			startingPower: map[string]int64{
				"a": 100,
				"b": 100,
			},
			fn: func(t *testing.T, db sql.DB, v *VoteStore) {
				ctx := context.Background()

				// Voters can't approve non-existent resolutions
				err := ApproveResolution(ctx, db, testEvent.ID(), []byte("a"))
				require.Error(t, err)

				err = CreateResolution(ctx, db, testEvent, 10, []byte("a"))
				require.NoError(t, err)

				// verify that the resolution exists
				exists, err := ResolutionExists(ctx, db, testEvent.ID())
				require.NoError(t, err)
				require.True(t, exists)

				err = ApproveResolution(ctx, db, testEvent.ID(), []byte("a"))
				require.NoError(t, err)

				err = ApproveResolution(ctx, db, testEvent.ID(), []byte("b"))
				require.NoError(t, err)

				info, err := GetResolutionInfo(ctx, db, testEvent.ID())
				require.NoError(t, err)

				infoSlice, err := GetResolutionsByType(ctx, db, testType)
				require.NoError(t, err)
				require.Len(t, infoSlice, 1)

				require.EqualValues(t, testEvent.ID(), infoSlice[0].ID)

				info2Slice, err := GetResolutionIDsByTypeAndProposer(ctx, db, testType, []byte("a"))
				require.NoError(t, err)
				require.Len(t, info2Slice, 1)

				require.Equal(t, infoSlice[0].ID, info2Slice[0])

				require.Equal(t, testEvent.Body, info.Body)
				require.Equal(t, testEvent.Type, info.Type)
				require.Equal(t, testEvent.ID(), info.ID)
				require.Equal(t, int64(10), info.ExpirationHeight)
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
			fn: func(t *testing.T, db sql.DB, v *VoteStore) {
				ctx := context.Background()

				err := CreateResolution(ctx, db, testEvent, 10, []byte("a"))
				require.NoError(t, err)

				expired, err := GetExpired(ctx, db, 10)
				require.NoError(t, err)
				require.Equal(t, 1, len(expired))

				resolutionInfo, err := GetResolutionInfo(ctx, db, testEvent.ID())
				require.NoError(t, err)

				require.EqualValues(t, resolutionInfo, expired[0])
			},
		},
		{
			name: "many resolutions test",
			startingPower: map[string]int64{
				"a": 100,
			},
			fn: func(t *testing.T, db sql.DB, v *VoteStore) {
				ctx := context.Background()

				events := make([]*types.VotableEvent, 3)
				ids := make([]*types.UUID, 3)
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

				err := CreateResolution(ctx, db, events[0], 10, []byte("a"))
				require.NoError(t, err)
				err = ApproveResolution(ctx, db, events[0].ID(), []byte("a"))
				require.NoError(t, err)

				err = CreateResolution(ctx, db, events[1], 10, []byte("a"))
				require.NoError(t, err)

				err = ApproveResolution(ctx, db, events[2].ID(), []byte("a"))
				require.Error(t, err)

				// check that none are processed
				notProcessed, err := FilterNotProcessed(ctx, db, ids)
				require.NoError(t, err)
				assert.Equal(t, len(notProcessed), 3)

				// delete and process all
				err = DeleteResolutions(ctx, db, ids...)
				require.NoError(t, err)
				err = MarkProcessed(ctx, db, ids...)
				require.NoError(t, err)

				// check that they are all processed
				notProcessed, err = FilterNotProcessed(ctx, db, ids)
				require.NoError(t, err)

				assert.Equal(t, len(notProcessed), 0)
			},
		},
		{
			name: "no resolutions",
			startingPower: map[string]int64{
				"a": 100,
			},
			fn: func(t *testing.T, db sql.DB, v *VoteStore) {
				ctx := context.Background()

				processed, err := FilterNotProcessed(ctx, db, []*types.UUID{types.NewUUIDV5([]byte("ss"))})
				require.NoError(t, err)

				require.Equal(t, len(processed), 1)

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

			v, err := initializeVoteStore(ctx, dbTx)
			require.NoError(t, err)

			for addr, power := range tt.startingPower {
				err = v.SetValidatorPower(ctx, dbTx, []byte(addr), power)
				require.NoError(t, err)
			}

			tt.fn(t, dbTx, v)
		})
	}
}

var testEvent = &types.VotableEvent{
	Body: []byte("test"),
	Type: testType,
}

var dummyEvent = &types.VotableEvent{
	Body: []byte("test"),
	Type: "blah",
}
var testConfirmationThreshold = big.NewRat(2, 3)
