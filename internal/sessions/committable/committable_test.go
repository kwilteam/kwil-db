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
		fn      func(t *testing.T, c *committable.Committable)
	}

	tests := []testcase{
		{
			name: "proper usage",
			fn: func(t *testing.T, c *committable.Committable) {
				ctx := context.Background()
				key := []byte("key")

				err := c.Begin(ctx, key)
				require.NoError(t, err)

				val := []byte("val")
				err = c.Register(val)
				require.NoError(t, err)

				hash := sha256.New()
				hash.Write(val)

				id, err := c.Commit(ctx, key)
				require.NoError(t, err)

				require.Equal(t, hash.Sum(nil), id)
			},
		},
		{
			name: "commit with no begin",
			fn: func(t *testing.T, c *committable.Committable) {
				ctx := context.Background()
				key := []byte("key")

				_, err := c.Commit(ctx, key)
				require.Error(t, err)
			},
		},
		{
			name: "register during no session",
			fn: func(t *testing.T, c *committable.Committable) {
				err := c.Register([]byte("val"))
				require.Error(t, err)
			},
		},
		{
			name: "id function instead of hash",
			fn: func(t *testing.T, c *committable.Committable) {
				fn := func() ([]byte, error) {
					return []byte("apphash"), nil
				}
				c.SetIDFunc(fn)

				ctx := context.Background()
				key := []byte("key")

				err := c.Begin(ctx, key)
				require.NoError(t, err)

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

func (m *mockStore) Begin(ctx context.Context) (sql.TxCloser, error) {
	return &mockSavepoint{}, nil
}

func (m *mockStore) Set(ctx context.Context, key []byte, value []byte) error {
	m.values[string(key)] = value
	return nil
}

type mockSavepoint struct{}

func (m *mockSavepoint) Commit(ctx context.Context) error {
	return nil
}

func (m *mockSavepoint) Precommit(ctx context.Context) ([]byte, error) {
	return nil, nil
}

func (m *mockSavepoint) Rollback(ctx context.Context) error {
	return nil
}
