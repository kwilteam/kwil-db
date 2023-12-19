package events_test

import (
	"context"
	"testing"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/internal/events"
	"github.com/kwilteam/kwil-db/internal/sql"
	"github.com/kwilteam/kwil-db/internal/sql/sqlite"
	"github.com/stretchr/testify/require"
)

func Test_EventStore(t *testing.T) {
	type testcase struct {
		name string
		fn   func(t *testing.T, e *events.EventStore)
	}
	tests := []testcase{
		{
			name: "standard storage and retrieval",
			fn: func(t *testing.T, e *events.EventStore) {
				ctx := context.Background()

				err := e.Store(ctx, []byte("hello"), "test")
				require.NoError(t, err)

				events, err := e.GetEvents(ctx)
				require.NoError(t, err)

				require.Len(t, events, 1)
				require.Equal(t, []byte("hello"), events[0].Data)
				require.Equal(t, "test", events[0].EventType)

				err = e.DeleteEvent(ctx, events[0].ID())
				require.NoError(t, err)

				events, err = e.GetEvents(ctx)
				require.NoError(t, err)

				require.Len(t, events, 0)
			},
		},
		{
			name: "idempotent storage",
			fn: func(t *testing.T, e *events.EventStore) {
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
			fn: func(t *testing.T, e *events.EventStore) {
				ctx := context.Background()

				err := e.DeleteEvent(ctx, types.NewUUIDV5([]byte("hello")))
				require.NoError(t, err)
			},
		},
		{
			name: "using kv scoping",
			fn: func(t *testing.T, e *events.EventStore) {
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			conn, err := sqlite.Open(ctx, ":memory:", sql.OpenCreate|sql.OpenMemory)
			require.NoError(t, err)

			e, err := events.NewEventStore(ctx, &db{conn})
			require.NoError(t, err)
			tt.fn(t, e)
		})
	}
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
