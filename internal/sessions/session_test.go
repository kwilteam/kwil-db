package sessions_test

import (
	"bytes"
	"fmt"
	"testing"

	"context"

	"github.com/kwilteam/kwil-db/internal/kv"
	"github.com/kwilteam/kwil-db/internal/sessions"
	"github.com/stretchr/testify/assert"
)

func Test_Sessions(t *testing.T) {
	type testcase struct {
		name string
		// input mock committable
		committable *mockCommittable
		fn          func(t *testing.T, mc *sessions.MultiCommitter, kv *mockKV)
		// expected mock committable
		// can be nil if we don't expect any changes / care
		result *mockCommittable
	}

	testcases := []testcase{
		{
			name:        "basic commits",
			committable: &mockCommittable{},
			fn: func(t *testing.T, mc *sessions.MultiCommitter, kv *mockKV) {
				ctx := context.Background()
				key := []byte("test")

				err := mc.Begin(ctx, key)
				assert.NoError(t, err)

				_, err = mc.Commit(ctx, key)
				assert.NoError(t, err)

				key2 := []byte("test2")

				err = mc.Begin(ctx, key2)
				assert.NoError(t, err)

				_, err = mc.Commit(ctx, key2)
				assert.NoError(t, err)

				assert.Equal(t, len(kv.vals), 0)
			},
		},
		{
			name: "cancel",
			committable: &mockCommittable{
				errOnCommit: true,
			},
			fn: func(t *testing.T, mc *sessions.MultiCommitter, kv *mockKV) {
				ctx := context.Background()
				key := []byte("test")

				err := mc.Begin(ctx, key)
				assert.NoError(t, err)

				_, err = mc.Commit(ctx, key)
				assert.Error(t, err)

				assert.Equal(t, len(kv.vals), 1)
			},
		},
		{
			name: "recovery",
			committable: &mockCommittable{
				errOnCommit: true,
			},
			fn: func(t *testing.T, mc *sessions.MultiCommitter, kv *mockKV) {
				ctx := context.Background()
				key := []byte("test")

				err := mc.Begin(ctx, key)
				assert.NoError(t, err)

				_, err = mc.Commit(ctx, key)
				assert.Error(t, err)

				err = mc.Begin(ctx, key)
				assert.NoError(t, err)
			},
			result: &mockCommittable{
				inSession:   true,
				recovery:    true,
				errOnCommit: true,
				currentKey:  []byte("test"),
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			kv := &mockKV{
				vals: make(map[string][]byte),
			}

			committables := map[string]sessions.Committable{
				"test": tc.committable,
				// a second committable to test that we can have multiple committables
				"default": &mockCommittable{},
			}

			mc := sessions.NewCommitter(kv, committables)

			tc.fn(t, mc, kv)

			if tc.result != nil {
				assert.Equal(t, tc.result, committables["test"])
			}
		})
	}
}

type mockCommittable struct {
	inSession  bool
	recovery   bool
	currentKey []byte

	errOnBegin         bool
	errOnBeginRecovery bool
	errOnCommit        bool
}

func (m *mockCommittable) Begin(ctx context.Context, idempotencyKey []byte) error {
	if m.inSession {
		return fmt.Errorf("already in session")
	}

	if m.errOnBegin {
		return fmt.Errorf("intentional mock error on begin")
	}

	m.inSession = true

	m.currentKey = idempotencyKey

	return nil
}

func (m *mockCommittable) BeginRecovery(ctx context.Context, idempotencyKey []byte) error {
	if m.inSession {
		return fmt.Errorf("already in session")
	}

	if m.errOnBeginRecovery {
		return fmt.Errorf("intentional mock error on begin recovery")
	}

	m.inSession = true
	m.recovery = true

	m.currentKey = idempotencyKey

	return nil
}

func (m *mockCommittable) Cancel(ctx context.Context) error {
	m.inSession = false
	m.recovery = false
	m.currentKey = nil

	return nil
}

func (m *mockCommittable) Commit(ctx context.Context, idempotencyKey []byte) ([]byte, error) {
	if !m.inSession {
		return nil, fmt.Errorf("not in session")
	}

	if m.errOnCommit {
		return nil, fmt.Errorf("intentional mock error on commit")
	}

	if !bytes.Equal(m.currentKey, idempotencyKey) {
		return nil, fmt.Errorf("idempotency key mismatch")
	}

	m.inSession = false
	m.currentKey = nil

	return []byte("id"), nil
}

type mockKV struct {
	vals map[string][]byte
}

func (m *mockKV) Delete(key []byte) error {
	delete(m.vals, string(key))
	return nil
}

func (m *mockKV) Get(key []byte) ([]byte, error) {
	val, ok := m.vals[string(key)]
	if !ok {
		return nil, kv.ErrKeyNotFound
	}

	return val, nil
}

func (m *mockKV) Set(key []byte, value []byte) error {
	m.vals[string(key)] = value
	return nil
}
