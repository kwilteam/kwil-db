//go:build pglive

package events

import (
	"context"
	"fmt"
	"testing"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/extensions/resolutions"
	"github.com/kwilteam/kwil-db/internal/sql/pg"
	dbtest "github.com/kwilteam/kwil-db/internal/sql/pg/test"
	"github.com/kwilteam/kwil-db/internal/voting"

	"github.com/stretchr/testify/require"
)

func init() {
	err := resolutions.RegisterResolution("test", resolutions.ResolutionConfig{})
	if err != nil {
		panic(err)
	}
}

func Test_EventStore(t *testing.T) {
	type testcase struct {
		name string
		// we have to use an outerTx here because we are testing commits from different connections
		// to the event store
		fn func(t *testing.T, e *EventStore, consensusTx *pg.Pool)
	}
	tests := []testcase{
		{
			name: "standard storage and retrieval",
			fn: func(t *testing.T, e *EventStore, db *pg.Pool) {
				ctx := context.Background()

				err := e.Store(ctx, []byte("hello"), "test")
				require.NoError(t, err)

				events, err := GetEvents(ctx, db)
				require.NoError(t, err)

				require.Len(t, events, 1)
				require.Equal(t, []byte("hello"), events[0].Body)
				require.Equal(t, "test", events[0].Type)

				err = DeleteEvent(ctx, db, events[0].ID())
				require.NoError(t, err)

				events, err = GetEvents(ctx, db)
				require.NoError(t, err)

				require.Len(t, events, 0)
			},
		},
		{
			name: "idempotent storage",
			fn: func(t *testing.T, e *EventStore, db *pg.Pool) {
				ctx := context.Background()

				err := e.Store(ctx, []byte("hello"), "test")
				require.NoError(t, err)

				err = e.Store(ctx, []byte("hello"), "test")
				require.NoError(t, err)

				events, err := GetEvents(ctx, db)
				require.NoError(t, err)

				require.Len(t, events, 1)
			},
		},
		{
			name: "deleting non-existent event",
			fn: func(t *testing.T, e *EventStore, db *pg.Pool) {
				ctx := context.Background()

				err := DeleteEvent(ctx, db, types.NewUUIDV5([]byte("hello")))
				require.NoError(t, err)
			},
		},
		{
			name: "using kv scoping",
			fn: func(t *testing.T, e *EventStore, db *pg.Pool) {
				ctx := context.Background()

				kv := e.KV([]byte("hello"))
				kvCopy := e.KV([]byte("hello"))
				kv2 := e.KV([]byte("hello2"))

				err := kv.Set(ctx, []byte("key"), []byte("value"))
				require.NoError(t, err)

				value, err := kv.Get(ctx, []byte("key"))
				require.NoError(t, err)
				require.Equal(t, []byte("value"), value)

				value, err = kvCopy.Get(ctx, []byte("key"))
				require.NoError(t, err)
				require.Equal(t, []byte("value"), value)

				value, err = kv2.Get(ctx, []byte("key"))
				require.NoError(t, err)
				require.Nil(t, value)
			},
		},
		{
			name: "marking broadcasted",
			fn: func(t *testing.T, e *EventStore, db *pg.Pool) {
				ctx := context.Background()

				tx, err := db.BeginTx(ctx)
				require.NoError(t, err)
				defer tx.Rollback(ctx)

				event := &types.VotableEvent{
					Body: []byte("hello"),
					Type: "test",
				}
				id := event.ID()

				err = e.Store(ctx, event.Body, event.Type)
				require.NoError(t, err)

				// GetUnreceivedEvents should return the event
				events, err := e.GetUnreceivedEvents(ctx)
				require.NoError(t, err)
				require.Len(t, events, 1)

				// Mark event as broadcasted
				err = e.MarkBroadcasted(ctx, []types.UUID{id})
				require.NoError(t, err)

				err = tx.Commit(ctx)
				require.NoError(t, err)

				tx2, err := db.BeginTx(ctx)
				require.NoError(t, err)
				defer tx2.Rollback(ctx)

				// GetEvents should still return the event
				events, err = GetEvents(ctx, tx2)
				require.NoError(t, err)
				require.Len(t, events, 1)

				// GetUnreceivedEvents should not return the event
				events, err = e.GetUnreceivedEvents(ctx)
				require.NoError(t, err)
				require.Len(t, events, 0)

				err = e.MarkRebroadcast(ctx, []types.UUID{id})
				require.NoError(t, err)

				err = tx2.Commit(ctx)
				require.NoError(t, err)

				// GetEvents should still return the event
				events, err = GetEvents(ctx, db)
				require.NoError(t, err)
				require.Len(t, events, 1)

				// GetUnreceivedEvents should not return the event
				events, err = e.GetUnreceivedEvents(ctx)
				require.NoError(t, err)
				require.Len(t, events, 1)
			},
		},
		{
			name: "get events which has no resolutions",
			fn: func(t *testing.T, e *EventStore, db *pg.Pool) {
				ctx := context.Background()

				// create 3 events
				for i := 0; i < 3; i++ {
					data := fmt.Sprintf("test%d", i)
					err := e.Store(ctx, []byte(data), "test")
					require.NoError(t, err)
				}

				// Get events which have no resolutions
				events, err := GetEvents(ctx, db)
				require.NoError(t, err)
				require.Len(t, events, 3)

				// create resolutions for 1 events
				err = voting.CreateResolution(ctx, db, events[0], 10, []byte("a"))
				require.NoError(t, err)

				// Get events which have no resolutions
				events, err = GetEvents(ctx, db)
				require.NoError(t, err)
				require.Len(t, events, 2)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			db, cleanup, err := dbtest.NewTestPool(ctx, []string{schemaName, "kwild_voting"}) // db is the event store specific connection
			require.NoError(t, err)
			defer cleanup()

			err = voting.InitializeVoteStore(ctx, db)
			require.NoError(t, err)

			e, err := NewEventStore(ctx, db) // needs BeginReadTx
			require.NoError(t, err)

			// create a second db connection to emulate the consensus db
			consensusDB, cleanup2, err := dbtest.NewTestPool(ctx, nil) // don't need to delete schema since we will never commit
			require.NoError(t, err)
			defer cleanup2()

			consensusTx, err := consensusDB.BeginTx(ctx)
			require.NoError(t, err)
			defer consensusTx.Rollback(ctx) // always rollback, to clean up

			tt.fn(t, e, consensusDB)
		})
	}
}
