package events

import (
	"context"
	"testing"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/internal/sql"
	"github.com/kwilteam/kwil-db/internal/sql/sqlite"
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

			conn, err := sqlite.Open(ctx, ":memory:", sql.OpenCreate|sql.OpenMemory)
			require.NoError(t, err)

			es, err := NewEventStore(ctx, &db{conn})
			require.NoError(t, err)

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

type db struct {
	*sqlite.Connection
}

func (d *db) Execute(ctx context.Context, stmt string, args map[string]any) (*sql.ResultSet, error) {
	res, err := d.Connection.Execute(ctx, stmt, args)
	if err != nil {
		return nil, err
	}
	defer res.Finish()

	return res.ResultSet()
}

func (d *db) Query(ctx context.Context, query string, args map[string]any) (*sql.ResultSet, error) {
	res, err := d.Connection.Execute(ctx, query, args)
	if err != nil {
		return nil, err
	}
	defer res.Finish()

	return res.ResultSet()
}

func (d *db) Get(ctx context.Context, key []byte, sync bool) ([]byte, error) {
	return d.Connection.Get(ctx, key)
}
