//go:build pglive

package events

import (
	"context"
	"testing"

	"github.com/kwilteam/kwil-db/core/types"
	dbtest "github.com/kwilteam/kwil-db/internal/sql/pg/test"
	"github.com/stretchr/testify/require"
)

func Test_EventManager(t *testing.T) {
	type testcase struct {
		name string
		fn   func(t *testing.T, em *EventMgr, vs *mockVoteStore)
	}
	tests := []testcase{
		{
			name: "new event",
			fn: func(t *testing.T, em *EventMgr, vs *mockVoteStore) {
				ctx := context.Background()

				err := em.Store(ctx, []byte("hello"), "test")
				require.NoError(t, err)

				events, err := em.GetEvents(ctx)
				require.NoError(t, err)

				require.Len(t, events, 1)
				require.Equal(t, []byte("hello"), events[0].Body)
				require.Equal(t, "test", events[0].Type)
			},
		},
		{
			name: "duplicate event",
			fn: func(t *testing.T, em *EventMgr, vs *mockVoteStore) {
				ctx := context.Background()

				event := &types.VotableEvent{
					Body: []byte("hello"),
					Type: "test",
				}
				id := event.ID()

				// mark the id as processed
				vs.Processed(id)

				// Check if the event is processed
				isProcessed, err := em.votestore.IsProcessed(ctx, id)
				require.NoError(t, err)
				require.True(t, isProcessed)

				// Try storing the event again, as its already processed, it should not be stored
				err = em.Store(ctx, []byte("hello"), "test")
				require.NoError(t, err)

				events, err := em.GetEvents(ctx)
				require.NoError(t, err)

				require.Len(t, events, 0)

			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			db, cleanUp, err := dbtest.NewTestPool(ctx, []string{schemaName})
			require.NoError(t, err)
			defer cleanUp()

			es, err := NewEventStore(ctx, db)
			require.NoError(t, err)

			defer db.Execute(ctx, dropEventsTable)

			vs := NewMockVoteStore()
			em := NewEventMgr(es, vs)

			tc.fn(t, em, vs)
		})
	}
}

type mockVoteStore struct {
	processed map[types.UUID]bool
}

func NewMockVoteStore() *mockVoteStore {
	return &mockVoteStore{
		processed: make(map[types.UUID]bool),
	}
}

func (m *mockVoteStore) Processed(resolutionID types.UUID) {
	m.processed[resolutionID] = true
}

func (m *mockVoteStore) IsProcessed(ctx context.Context, resolutionID types.UUID) (bool, error) {
	return m.processed[resolutionID], nil
}
