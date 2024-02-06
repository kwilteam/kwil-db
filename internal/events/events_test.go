//go:build pglive

package events

import (
	"context"
	"testing"

	"github.com/kwilteam/kwil-db/core/types"
	dbtest "github.com/kwilteam/kwil-db/internal/sql/pg/test"

	"github.com/stretchr/testify/require"
)

func Test_EventStore(t *testing.T) {
	type testcase struct {
		name string
		fn   func(t *testing.T, e *EventStore)
	}
	tests := []testcase{
		{
			name: "standard storage and retrieval",
			fn: func(t *testing.T, e *EventStore) {
				ctx := context.Background()

				err := e.Store(ctx, []byte("hello"), "test")
				require.NoError(t, err)

				events, err := e.GetEvents(ctx)
				require.NoError(t, err)

				require.Len(t, events, 1)
				require.Equal(t, []byte("hello"), events[0].Body)
				require.Equal(t, "test", events[0].Type)

				err = e.DeleteEvent(ctx, events[0].ID())
				require.NoError(t, err)

				events, err = e.GetEvents(ctx)
				require.NoError(t, err)

				require.Len(t, events, 0)
			},
		},
		{
			name: "idempotent storage",
			fn: func(t *testing.T, e *EventStore) {
				ctx := context.Background()

				err := e.Store(ctx, []byte("hello"), "test")
				require.NoError(t, err)

				err = e.Store(ctx, []byte("hello"), "test")
				require.NoError(t, err)

				events, err := e.GetEvents(ctx)
				require.NoError(t, err)

				require.Len(t, events, 1)
			},
		},
		{
			name: "deleting non-existent event",
			fn: func(t *testing.T, e *EventStore) {
				ctx := context.Background()

				err := e.DeleteEvent(ctx, types.NewUUIDV5([]byte("hello")))
				require.NoError(t, err)
			},
		},
		{
			name: "using kv scoping",
			fn: func(t *testing.T, e *EventStore) {
				ctx := context.Background()

				kv := e.KV([]byte("hello"))
				kvCopy := e.KV([]byte("hello"))
				kv2 := e.KV([]byte("hello2"))

				err := kv.Set(ctx, []byte("key"), []byte("value"))
				require.NoError(t, err)

				const sync = true
				value, err := kv.Get(ctx, []byte("key"), sync)
				require.NoError(t, err)
				require.Equal(t, []byte("value"), value)

				value, err = kvCopy.Get(ctx, []byte("key"), sync)
				require.NoError(t, err)
				require.Equal(t, []byte("value"), value)

				value, err = kv2.Get(ctx, []byte("key"), sync)
				require.NoError(t, err)
				require.Nil(t, value)
			},
		},
		{
			name: "marking received",
			fn: func(t *testing.T, e *EventStore) {
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

				err = e.MarkReceived(ctx, event.ID())
				require.NoError(t, err)

				// GetEvents should still return the event
				events, err = e.GetEvents(ctx)
				require.NoError(t, err)
				require.Len(t, events, 1)

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

			db, cleanUp, err := dbtest.NewTestPool(ctx, []string{schemaName})
			require.NoError(t, err)
			defer cleanUp()

			e, err := NewEventStore(ctx, db)
			require.NoError(t, err)

			defer db.Execute(ctx, dropEventsTable)

			tt.fn(t, e)
		})
	}
}
