package sqlite_test

import (
	"context"
	"fmt"
	"sync"
	"testing"

	sql "github.com/kwilteam/kwil-db/internal/sql"
	"github.com/kwilteam/kwil-db/internal/sql/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Pool(t *testing.T) {
	type testcase struct {
		name              string
		persistentReaders int
		maximumReaders    int
		fn                func(*testing.T, *sqlite.Pool)
	}

	tests := []testcase{
		{
			name:              "basic concurrent readers",
			persistentReaders: 2,
			maximumReaders:    3,
			fn: func(t *testing.T, p *sqlite.Pool) {
				ctx := context.Background()

				sp, err := p.Savepoint()
				assert.NoError(t, err)
				defer sp.Rollback()

				ses, err := p.CreateSession()
				assert.NoError(t, err)
				defer ses.Delete()

				start := make(chan struct{})
				done := sync.WaitGroup{}

				done.Add(1)
				go func() {
					<-start

					_, err := p.Execute(ctx, "insert into users (id, name, age) values (1, 'John', 20);", nil)
					assert.NoError(t, err)

					err = p.Set(ctx, []byte("key"), []byte("value"))
					assert.NoError(t, err)

					// try KV, with checking for committed and uncommitted

					val, err := p.Get(ctx, []byte("key"), true)
					assert.NoError(t, err)

					assert.Equal(t, val, []byte("value"))

					val, err = p.Get(ctx, []byte("key"), false)
					assert.NoError(t, err)

					assert.Nil(t, val)

					err = sp.Commit()
					assert.NoError(t, err)

					val, err = p.Get(ctx, []byte("key"), false)
					assert.NoError(t, err)

					assert.Equal(t, val, []byte("value"))

					done.Done()
				}()

				for i := 0; i < 3; i++ {
					done.Add(1)
					go func() {
						<-start

						_, err := p.Query(ctx, "select * from users;", nil)
						assert.NoError(t, err)

						done.Done()
					}()
				}

				close(start)
				done.Wait()
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			deleteTempDir()
			defer deleteTempDir()

			ctx := context.Background()
			conn, err := sqlite.Open(ctx, fmt.Sprintf("%s/%s", tempDir, "testdb"), sql.OpenCreate)
			if err != nil {
				t.Fatal(err)
			}

			err = createUserTable(ctx, conn)
			require.NoError(t, err)

			err = conn.Close()
			require.NoError(t, err)

			p, err := sqlite.NewPool(ctx, fmt.Sprintf("%s/%s", tempDir, "testdb"), test.persistentReaders, test.maximumReaders, false)
			require.NoError(t, err)

			test.fn(t, p)

			err = p.Close()
			require.NoError(t, err)

		})
	}
}

func Test_RunningReadersClose(t *testing.T) {
	ctx := context.Background()
	deleteTempDir()
	defer deleteTempDir()

	p, err := sqlite.NewPool(ctx, fmt.Sprintf("%s/%s", tempDir, "testdb"), 1, 2, true)
	require.NoError(t, err)

	_, err = p.Execute(ctx, createUsersTableStmt, nil)
	require.NoError(t, err)

	for i := 0; i < 100; i++ {
		_, err = p.Execute(ctx, fmt.Sprintf("insert into users (id, name, age) values (%d, 'John', 20);", i), nil)
		require.NoError(t, err)
	}

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		for i := 0; i < 100000000000; i++ { // will timeout if not closed
			p.Query(ctx, "select * from users;", nil)
			if i == 0 {
				wg.Done()
			}
		}
	}()

	wg.Wait()

	err = p.Close()
	require.NoError(t, err)

}

func Test_Open(t *testing.T) {
	type testcase struct {
		name              string
		persistentReaders int
		maximumReaders    int
		err               error
	}

	tests := []testcase{
		{
			name:              "more persistent readers than maximum readers fails",
			persistentReaders: 2,
			maximumReaders:    1,
			err:               sqlite.ErrPersistentGreaterThanMaxReaders,
		},
		{
			name:           "maximum readers less than 1 fails",
			maximumReaders: 0,
			err:            sqlite.ErrMaxReaders,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Cleanup(deleteTempDir)
			ctx := context.Background()
			conn, err := sqlite.Open(ctx, fmt.Sprintf("%s/%s", tempDir, "testdb"), sql.OpenCreate)
			if err != nil {
				t.Fatal(err)
			}

			err = createUserTable(ctx, conn)
			require.NoError(t, err)

			err = conn.Close()
			require.NoError(t, err)

			p, err := sqlite.NewPool(ctx, fmt.Sprintf("%s/%s", tempDir, "testdb"), test.persistentReaders, test.maximumReaders, false)
			if test.err != nil {
				require.ErrorIs(t, err, test.err)
				return
			} else {
				require.NoError(t, err)
			}

			err = p.Close()
			require.NoError(t, err)
		})
	}
}
