package sqlite_test

import (
	"context"
	"fmt"
	"path/filepath"
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
			tempDir := t.TempDir()

			ctx := context.Background()
			conn, err := sqlite.Open(ctx, filepath.Join(tempDir, "testdb"), sql.OpenCreate)
			if err != nil {
				t.Fatal(err)
			}

			err = createUserTable(ctx, conn)
			require.NoError(t, err)

			err = conn.Close()
			require.NoError(t, err)

			p, err := sqlite.NewPool(ctx, filepath.Join(tempDir, "testdb"), test.persistentReaders, test.maximumReaders, false)
			require.NoError(t, err)

			test.fn(t, p)

			err = p.Close()
			require.NoError(t, err)
		})
	}
}

func Test_RunningReadersClose(t *testing.T) {
	// This test ensures that Close() on a pool with active readers doing Query
	// works as expected.
	ctx := context.Background()
	tempDir := t.TempDir()

	p, err := sqlite.NewPool(ctx, filepath.Join(tempDir, "testdb"), 1, 2, true)
	require.NoError(t, err)

	_, err = p.Execute(ctx, createUsersTableStmt, nil)
	require.NoError(t, err)

	for i := 0; i < 100; i++ {
		_, err = p.Execute(ctx, fmt.Sprintf("insert into users (id, name, age) values (%d, 'John', 20);", i), nil)
		require.NoError(t, err)
	}

	var wg, wgLoop sync.WaitGroup
	wg.Add(1)
	wgLoop.Add(1)

	go func() {
		defer wgLoop.Done()
		for i := 0; ; i++ { // will timeout if not closed
			_, err := p.Query(ctx, "select * from users;", nil)
			if i == 0 {
				wg.Done() // let Close() happen concurrently
			}
			// There are various errors here depending on what it was doing when
			// Close was called, including "execution was interrupted", "context
			// canceled", "interrupted", etc.
			//
			// If the loop continues, subsequent Queries would receive
			// "connection pool forcefully closed".
			//
			// However, we are just going to break the loop so that the DB does
			// not continue to be in use after this test case.
			if err != nil {
				return
			}
		}
	}()

	wg.Wait()

	// Sleep a bit to discover other returns paths from Query => reader => ... => Open
	// This is the kind of real-life timing to expect from weak and virtualized
	// test runners used by CI systems like GitHub Actions.
	// time.Sleep(200 * time.Millisecond)

	err = p.Close()
	require.NoError(t, err)

	wgLoop.Wait()
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
			tempDir := t.TempDir()
			ctx := context.Background()
			conn, err := sqlite.Open(ctx, filepath.Join(tempDir, "testdb"), sql.OpenCreate)
			if err != nil {
				t.Fatal(err)
			}

			err = createUserTable(ctx, conn)
			require.NoError(t, err)

			err = conn.Close()
			require.NoError(t, err)

			p, err := sqlite.NewPool(ctx, filepath.Join(tempDir, "testdb"), test.persistentReaders, test.maximumReaders, false)
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
