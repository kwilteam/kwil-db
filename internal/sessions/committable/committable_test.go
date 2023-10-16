package committable_test

import (
	"context"
	"crypto/sha256"
	"testing"

	"github.com/kwilteam/kwil-db/internal/sql"
	"github.com/stretchr/testify/require"

	"github.com/kwilteam/kwil-db/internal/sessions/committable"
)

func Test_Committable(t *testing.T) {
	type testcase struct {
		name    string
		initial map[string][]byte
		fn      func(t *testing.T, c *committable.SavepointCommittable)
	}

	tests := []testcase{
		{
			name: "proper usage",
			fn: func(t *testing.T, c *committable.SavepointCommittable) {
				ctx := context.Background()
				key := []byte("key")

				err := c.Begin(ctx, key)
				require.NoError(t, err)

				require.False(t, c.Skip())

				val := []byte("val")
				err = c.Register(val)
				require.NoError(t, err)

				hash := sha256.New()
				hash.Write(val)

				id, err := c.Commit(ctx, key)
				require.NoError(t, err)

				require.Equal(t, hash.Sum(nil), id)

				require.False(t, c.Skip())
			},
		},
		{
			name: "commit with no begin",
			fn: func(t *testing.T, c *committable.SavepointCommittable) {
				ctx := context.Background()
				key := []byte("key")

				require.False(t, c.Skip())

				_, err := c.Commit(ctx, key)
				require.Error(t, err)
			},
		},
		{
			name: "register during no session",
			fn: func(t *testing.T, c *committable.SavepointCommittable) {
				err := c.Register([]byte("val"))
				require.Error(t, err)
			},
		},
		{
			name: "recovery, already committed",
			initial: map[string][]byte{
				string(committable.IdempotencyKeyKey): []byte("key"),
				string(committable.ApphashKey):        []byte("apphash"),
			},
			fn: func(t *testing.T, c *committable.SavepointCommittable) {
				key := []byte("key")
				ctx := context.Background()

				err := c.BeginRecovery(ctx, key)
				require.NoError(t, err)

				require.True(t, c.Skip())

				err = c.Register([]byte("val"))
				require.NoError(t, err)

				id, err := c.Commit(ctx, key)
				require.NoError(t, err)

				require.Equal(t, []byte("apphash"), id)
			},
		},
		{
			name: "recovery, not committed",
			initial: map[string][]byte{
				string(committable.IdempotencyKeyKey): []byte("oldKey"),
				string(committable.ApphashKey):        []byte("apphash"),
			},
			fn: func(t *testing.T, c *committable.SavepointCommittable) {
				key := []byte("key")
				ctx := context.Background()

				err := c.BeginRecovery(ctx, key)
				require.NoError(t, err)

				require.False(t, c.Skip())

				err = c.Register([]byte("val"))
				require.NoError(t, err)

				hash := sha256.New()
				hash.Write([]byte("val"))

				id, err := c.Commit(ctx, key)
				require.NoError(t, err)

				require.Equal(t, hash.Sum(nil), id)
			},
		},
		{
			name: "id function instead of hash",
			fn: func(t *testing.T, c *committable.SavepointCommittable) {
				fn := func() ([]byte, error) {
					return []byte("apphash"), nil
				}
				c.SetIDFunc(fn)

				ctx := context.Background()
				key := []byte("key")

				err := c.Begin(ctx, key)
				require.NoError(t, err)

				require.False(t, c.Skip())

				val := []byte("val")
				err = c.Register(val)
				require.Error(t, err)

				id, err := c.Commit(ctx, key)
				require.NoError(t, err)

				require.Equal(t, []byte("apphash"), id)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.initial == nil {
				tt.initial = make(map[string][]byte)
			}
			store := &mockStore{
				values: tt.initial,
			}

			c := committable.New(store)

			tt.fn(t, c)
		})
	}
}

type mockStore struct {
	values map[string][]byte
}

func (m *mockStore) Get(ctx context.Context, key []byte, sync bool) ([]byte, error) {
	_, ok := m.values[string(key)]
	if !ok {
		return nil, nil
	}

	return m.values[string(key)], nil
}

func (m *mockStore) Savepoint() (sql.Savepoint, error) {
	return &mockSavepoint{}, nil
}

func (m *mockStore) Set(ctx context.Context, key []byte, value []byte) error {
	m.values[string(key)] = value
	return nil
}

type mockSavepoint struct{}

func (m *mockSavepoint) Commit() error {
	return nil
}

func (m *mockSavepoint) Rollback() error {
	return nil
}
