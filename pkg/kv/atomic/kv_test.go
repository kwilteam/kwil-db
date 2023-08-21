package atomic_test

import (
	"context"
	"testing"

	"github.com/kwilteam/kwil-db/pkg/kv"
	"github.com/kwilteam/kwil-db/pkg/kv/atomic"
	kvTesting "github.com/kwilteam/kwil-db/pkg/kv/atomic/testing"
	"github.com/stretchr/testify/assert"
)

// tests basic KV functionality; anything that is not the sessions.Committable interface
func Test_BasicKV(t *testing.T) {
	dbType := kvTesting.TestKVFlagInMemory
	type testCase struct {
		name     string
		testFunc func(t *testing.T, db *atomic.AtomicKV)
	}

	testCases := []testCase{
		{
			name: "successful set, get, delete",
			testFunc: func(t *testing.T, db *atomic.AtomicKV) {
				err := db.Set(ser("key"), ser("value"))
				assert.NoError(t, err)

				value, err := db.Get(ser("key"))
				assert.NoError(t, err)

				assert.Equal(t, ser("value"), value)

				err = db.Delete(ser("key"))
				assert.NoError(t, err)

				_, err = db.Get(ser("key"))
				assert.ErrorIs(t, err, kv.ErrKeyNotFound)
			},
		},
		{
			name: "nested prefixes can manually append to the key",
			testFunc: func(t *testing.T, db *atomic.AtomicKV) {
				prefDb := db.NewPrefix(ser("prefix"))
				doublePrefDb := prefDb.NewPrefix(ser("prefix2"))

				err := doublePrefDb.Set(ser("key"), ser("value"))
				assert.NoError(t, err)

				value, err := doublePrefDb.Get(ser("key"))
				assert.NoError(t, err)
				value2, err := prefDb.Get(ser("prefix2", "key"))
				assert.NoError(t, err)
				value3, err := db.Get(ser("prefix", "prefix2", "key"))
				assert.NoError(t, err)

				assert.Equal(t, value, value2)
				assert.Equal(t, value2, value3)

				err = doublePrefDb.Delete(ser("key"))
				assert.NoError(t, err)

				_, err = doublePrefDb.Get(ser("key"))
				assert.ErrorIs(t, err, kv.ErrKeyNotFound)
				_, err = prefDb.Get(ser("prefix2", "key"))
				assert.ErrorIs(t, err, kv.ErrKeyNotFound)
				_, err = db.Get(ser("prefix", "prefix2", "key"))
				assert.ErrorIs(t, err, kv.ErrKeyNotFound)
			},
		},
		{
			name: "successful atomic commit",
			testFunc: func(t *testing.T, db *atomic.AtomicKV) {
				ctx := context.Background()

				// begin with some data
				err := db.Set([]byte("original"), []byte("value"))
				assert.NoError(t, err)

				err = db.BeginCommit(ctx)
				assert.NoError(t, err)

				err = db.Set([]byte("key"), []byte("value"))
				assert.NoError(t, err)

				err = db.Delete([]byte("original"))
				assert.NoError(t, err)

				appender := newCommitAppender()

				testIds(t, db)

				err = db.EndCommit(ctx, appender.Append)
				assert.NoError(t, err)

				err = db.BeginApply(ctx)
				assert.NoError(t, err)

				// there should only be one, but this may change
				for _, commit := range appender.commits {
					err = db.Apply(ctx, commit)
					assert.NoError(t, err)
				}

				err = db.EndApply(ctx)
				assert.NoError(t, err)

				value, err := db.Get([]byte("key"))
				assert.NoError(t, err)

				assert.Equal(t, []byte("value"), value)

				_, err = db.Get([]byte("original"))
				assert.ErrorIs(t, err, kv.ErrKeyNotFound)
			},
		},
		{
			name: "canceled atomic commit",
			testFunc: func(t *testing.T, db *atomic.AtomicKV) {
				ctx := context.Background()
				err := db.BeginCommit(ctx)
				assert.NoError(t, err)

				err = db.Set([]byte("key"), []byte("value"))
				assert.NoError(t, err)

				appender := newCommitAppender()

				err = db.EndCommit(ctx, appender.Append)
				assert.NoError(t, err)

				err = db.BeginApply(ctx)
				assert.NoError(t, err)

				// there should only be one, but this may change
				for _, commit := range appender.commits {
					err = db.Apply(ctx, commit)
					assert.NoError(t, err)
				}

				db.Cancel(ctx)

				_, err = db.Get([]byte("key"))
				assert.ErrorIs(t, err, kv.ErrKeyNotFound)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db, td, err := kvTesting.OpenTestKv(context.Background(), "kv_test", dbType)
			if err != nil {
				t.Fatal(err)
			}
			defer td()

			tc.testFunc(t, db)
		})
	}
}

func testIds(t *testing.T, db *atomic.AtomicKV) {
	ctx := context.Background()
	lastId, err := db.ID(ctx)
	assert.NoError(t, err)

	for i := 0; i < 100; i++ {
		id, err := db.ID(ctx)
		assert.NoError(t, err)

		assert.Equal(t, lastId, id)
		lastId = id
	}
}

func newCommitAppender() *commitAppender {
	return &commitAppender{
		commits: make([][]byte, 0),
	}
}

type commitAppender struct {
	commits [][]byte
}

func (c *commitAppender) Append(v []byte) error {
	c.commits = append(c.commits, v)
	return nil
}

// ser is a helper function to convert a list of strings to a byte slice
func ser(str ...string) []byte {
	var result []byte
	for _, s := range str {
		result = append(result, []byte(s)...)
	}
	return result
}
