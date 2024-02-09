//go:build pglive

package events

import (
	"context"
	"testing"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/internal/sql"
	dbtest "github.com/kwilteam/kwil-db/internal/sql/pg/test"

	"github.com/stretchr/testify/require"
)

func Test_EventStore(t *testing.T) {
	type testcase struct {
		name string
		// we have to use an outerTx here because we are testing commits from different connections
		// to the event store
		fn func(t *testing.T, e *EventStore, consensusTx sql.Tx)
	}
	tests := []testcase{
		{
			name: "standard storage and retrieval",
			fn: func(t *testing.T, e *EventStore, consensusTx sql.Tx) {
				ctx := context.Background()

				err := e.Store(ctx, []byte("hello"), "test")
				require.NoError(t, err)

				events, err := GetEvents(ctx, consensusTx)
				require.NoError(t, err)

				require.Len(t, events, 1)
				require.Equal(t, []byte("hello"), events[0].Body)
				require.Equal(t, "test", events[0].Type)

				err = DeleteEvent(ctx, consensusTx, events[0].ID())
				require.NoError(t, err)

				events, err = GetEvents(ctx, consensusTx)
				require.NoError(t, err)

				require.Len(t, events, 0)
			},
		},
		{
			name: "idempotent storage",
			fn: func(t *testing.T, e *EventStore, consensusTx sql.Tx) {
				ctx := context.Background()

				err := e.Store(ctx, []byte("hello"), "test")
				require.NoError(t, err)

				err = e.Store(ctx, []byte("hello"), "test")
				require.NoError(t, err)

				events, err := GetEvents(ctx, consensusTx)
				require.NoError(t, err)

				require.Len(t, events, 1)
			},
		},
		{
			name: "deleting non-existent event",
			fn: func(t *testing.T, e *EventStore, consensusTx sql.Tx) {
				ctx := context.Background()

				err := DeleteEvent(ctx, consensusTx, types.NewUUIDV5([]byte("hello")))
				require.NoError(t, err)
			},
		},
		{
			name: "using kv scoping",
			fn: func(t *testing.T, e *EventStore, consensusTx sql.Tx) {
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
			name: "marking received",
			fn: func(t *testing.T, e *EventStore, consensusTx sql.Tx) {
				ctx := context.Background()

				event := &types.VotableEvent{
					Body: []byte("hello"),
					Type: "test",
				}

				err := e.Store(ctx, event.Body, event.Type)
				require.NoError(t, err)

				// GetUnreceivedEvents should return the event
				events, err := e.GetUnreceivedEvents(ctx)
				require.NoError(t, err)
				require.Len(t, events, 1)

				err = MarkReceived(ctx, consensusTx, event.ID())
				require.NoError(t, err)

				// GetEvents should still return the event
				events, err = GetEvents(ctx, consensusTx)
				require.NoError(t, err)
				require.Len(t, events, 1)

				err = consensusTx.Commit(ctx)
				require.NoError(t, err)

				// GetUnreceivedEvents should not return the event
				events, err = e.GetUnreceivedEvents(ctx)
				require.NoError(t, err)
				require.Len(t, events, 0)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			db, cleanup, err := dbtest.NewTestPool(ctx, []string{SchemaName}) // db is the event store specific connection
			require.NoError(t, err)
			defer cleanup()

			e, err := NewEventStore(ctx, db, &mockVoteStore{})
			require.NoError(t, err)

			// create a second db connection to emulate the consensus db
			consensusDB, cleanup2, err := dbtest.NewTestPool(ctx, nil) // don't need to delete schema since we will never commit
			require.NoError(t, err)
			defer cleanup2()

			consensusTx, err := consensusDB.BeginTx(ctx)
			require.NoError(t, err)
			defer consensusTx.Rollback(ctx) // always rollback, to clean up

			tt.fn(t, e, consensusTx)
		})
	}
}

type mockVoteStore struct {
}

func (m *mockVoteStore) IsProcessed(ctx context.Context, db sql.DB, resolutionID types.UUID) (bool, error) {
	return false, nil
}
