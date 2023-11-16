package sqlite_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"testing"

	sql "github.com/kwilteam/kwil-db/internal/sql"
	"github.com/kwilteam/kwil-db/internal/sql/sqlite"
	"github.com/stretchr/testify/require"
)

func Test_WriterConn(t *testing.T) {
	type testcase struct {
		name  string
		flags sql.ConnectionFlag
		fn    func(*testing.T, *sqlite.Connection)
	}

	cases := []testcase{
		{
			name: "close database while read query is running",
			fn: func(t *testing.T, conn *sqlite.Connection) {
				ctx := context.Background()

				err := createUserTable(ctx, conn)
				require.NoError(t, err)

				// insert a user
				err = insertUser(ctx, conn, 1, "john", 20)
				require.NoError(t, err)

				// open a reader connection
				readerConn, err := openDB("test", sql.OpenReadOnly)
				require.NoError(t, err)

				// select all users
				res, err := selectUsers(ctx, readerConn)
				require.NoError(t, err)

				// close the database
				err = conn.Close()
				require.NoError(t, err)

				// check that we can still read from the reader connection
				results, err := getResults(res)
				require.NoError(t, err)

				require.Len(t, results, 1)
			},
		},
		{
			name: "commit savepoint while read query is running",
			fn: func(t *testing.T, conn *sqlite.Connection) {
				ctx := context.Background()

				err := createUserTable(ctx, conn)
				require.NoError(t, err)

				// insert a user
				err = insertUser(ctx, conn, 1, "john", 20)
				require.NoError(t, err)

				sp, err := conn.Savepoint()
				require.NoError(t, err)

				// insert another user
				err = insertUser(ctx, conn, 2, "jane", 20)
				require.NoError(t, err)

				// open a reader connection
				readerConn, err := openDB("test", sql.OpenReadOnly)
				require.NoError(t, err)

				// select all users
				res, err := selectUsers(ctx, readerConn)
				require.NoError(t, err)
				defer res.Finish()

				// commit savepoint
				err = sp.Commit()
				require.NoError(t, err)

				// close the database
				err = conn.Close()
				require.NoError(t, err)

			},
		},
		{
			name: "close database while query is running",
			fn: func(t *testing.T, conn *sqlite.Connection) {
				ctx := context.Background()

				err := createUserTable(ctx, conn)
				require.NoError(t, err)

				// insert a user
				err = insertUser(ctx, conn, 1, "john", 20)
				require.NoError(t, err)

				// select all users
				res, err := selectUsers(ctx, conn)
				require.NoError(t, err)

				// close the database
				err = conn.Close()
				require.NoError(t, err)

				err = res.Finish()
				require.ErrorIs(t, err, sqlite.ErrClosed)
			},
		},
		{
			name: "testing basic write, read, and foreign key",
			fn: func(t *testing.T, conn *sqlite.Connection) {
				ctx := context.Background()

				err := createUserTable(ctx, conn)
				require.NoError(t, err)

				err = createPostTable(ctx, conn)
				require.NoError(t, err)

				// insert a user
				err = insertUser(ctx, conn, 1, "john", 20)
				require.NoError(t, err)

				// insert another user
				err = insertUser(ctx, conn, 2, "jane", 20)
				require.NoError(t, err)

				// insert a post
				err = insertPost(ctx, conn, 1, "hello world", 1)
				require.NoError(t, err)

				// select all users
				res, err := selectUsers(ctx, conn)
				require.NoError(t, err)

				// check that we have 2 users
				results, err := getResults(res)
				require.NoError(t, err)

				require.Len(t, results, 2)

				if notEq(results[0]["id"], "1") || notEq(results[0]["name"], "john") || notEq(results[0]["age"], "20") {
					t.Errorf("expected user 1 to be john, 20, got %v", results[0])
				}

				if notEq(results[1]["id"], "2") || notEq(results[1]["name"], "jane") || notEq(results[1]["age"], "20") {
					t.Errorf("expected user 2 to be jane, 20, got %v", results[1])
				}

				// select all posts
				res, err = selectPosts(ctx, conn)
				require.NoError(t, err)

				// check that we have 1 post
				results, err = getResults(res)
				require.NoError(t, err)

				require.Len(t, results, 1)

				if notEq(results[0]["id"], "1") || notEq(results[0]["content"], "hello world") || notEq(results[0]["user_id"], "1") {
					t.Errorf("expected post 1 to be hello world, 1, got %v", results[0])
				}

				// delete user 1
				err = deleteUser(ctx, conn, 1)
				require.NoError(t, err)

				// select all users
				res, err = selectUsers(ctx, conn)
				require.NoError(t, err)

				// check that we have 1 user
				results, err = getResults(res)
				require.NoError(t, err)

				require.Len(t, results, 1)

				// check that post 1 is gone
				res, err = selectPosts(ctx, conn)
				require.NoError(t, err)

				// check that we have 0 posts
				results, err = getResults(res)
				require.NoError(t, err)

				require.Len(t, results, 0)
			},
		},
		{
			name: "testing savepoint",
			fn: func(t *testing.T, conn *sqlite.Connection) {
				ctx := context.Background()

				err := createUserTable(ctx, conn)
				require.NoError(t, err)

				sp, err := conn.Savepoint()
				require.NoError(t, err)

				// insert a user
				err = insertUser(ctx, conn, 1, "john", 20)
				require.NoError(t, err)

				// rollback
				err = sp.Rollback()
				require.NoError(t, err)

				// check that we have 0 users
				res, err := selectUsers(ctx, conn)
				require.NoError(t, err)

				results, err := getResults(res)
				require.NoError(t, err)

				require.Len(t, results, 0)

			},
		},
		{
			name: "table exists",
			fn: func(t *testing.T, conn *sqlite.Connection) {
				ctx := context.Background()

				err := createUserTable(ctx, conn)
				require.NoError(t, err)

				exists, err := conn.TableExists(ctx, "users")
				require.NoError(t, err)

				require.True(t, exists)
			},
		},
		{
			name: "reader connections",
			fn: func(t *testing.T, conn *sqlite.Connection) {
				ctx := context.Background()

				err := createUserTable(ctx, conn)
				require.NoError(t, err)

				// create a reader connection
				readerConn, err := openDB("test", sql.OpenReadOnly)
				require.NoError(t, err)

				// insert a user with reader connection
				err = insertUser(ctx, readerConn, 1, "john", 20)
				require.Error(t, err)

				// insert a user with writer connection
				err = insertUser(ctx, conn, 1, "john", 20)
				require.NoError(t, err)

				// select all users with reader connection
				res, err := selectUsers(ctx, readerConn)
				require.NoError(t, err)

				// check that we have 1 user
				results, err := getResults(res)
				require.NoError(t, err)

				require.Len(t, results, 1)

				if notEq(results[0]["id"], "1") || notEq(results[0]["name"], "john") || notEq(results[0]["age"], "20") {
					t.Errorf("expected user 1 to be john, 20, got %v", results[0])
				}

				// savepoint with writer connection
				sp, err := conn.Savepoint()
				require.NoError(t, err)

				// insert another user with writer connection
				err = insertUser(ctx, conn, 2, "jane", 20)
				require.NoError(t, err)

				// see if reader connection can see the new user
				res, err = selectUsers(ctx, readerConn)
				require.NoError(t, err)

				// check that we have 1 user on reader connection
				results, err = getResults(res)
				require.NoError(t, err)

				require.Len(t, results, 1)

				// commit savepoint
				err = sp.Commit()
				require.NoError(t, err)

				// see if reader connection can see the new user
				res, err = selectUsers(ctx, readerConn)
				require.NoError(t, err)

				// check that we have 2 users on reader connection
				results, err = getResults(res)
				require.NoError(t, err)

				require.Len(t, results, 2)
			},
		},
		{
			name: "testing reader restrictions",
			fn: func(t *testing.T, conn *sqlite.Connection) {
				reader, err := openDB("test", sql.OpenReadOnly)
				require.NoError(t, err)

				err = reader.ApplyChangeset(bytes.NewReader([]byte{}))
				require.Error(t, err)

				_, err = reader.CreateSession()
				require.Error(t, err)

				err = reader.DeleteDatabase()
				require.Error(t, err)

				err = reader.DisableForeignKey()
				require.Error(t, err)

				err = reader.EnableForeignKey()
				require.Error(t, err)

				_, err = reader.Savepoint()
				require.Error(t, err)
			},
		},
		{
			name: "testing multiple writers fails",
			fn: func(t *testing.T, conn *sqlite.Connection) {
				_, err := openDB("test", sql.OpenCreate)
				require.Error(t, err)
			},
		},
		{
			name: "using a writer without a closed result set errors",
			fn: func(t *testing.T, conn *sqlite.Connection) {
				ctx := context.Background()

				err := createUserTable(ctx, conn)
				require.NoError(t, err)

				err = insertUser(ctx, conn, 1, "john", 20)
				require.NoError(t, err)

				res, err := selectUsers(ctx, conn)
				require.NoError(t, err)

				_, err = selectUsers(ctx, conn)
				require.ErrorIs(t, err, sqlite.ErrInUse)

				err = res.Close()
				require.NoError(t, err)

				_, err = selectUsers(ctx, conn)
				require.NoError(t, err)
			},
		},
		{
			name: "cancelling context mid statement",
			fn: func(t *testing.T, conn *sqlite.Connection) {
				ctx, cancel := context.WithCancel(context.Background())

				err := createUserTable(ctx, conn)
				require.NoError(t, err)

				// insert a user
				err = insertUser(ctx, conn, 1, "john", 20)
				require.NoError(t, err)

				// select all users
				res, err := selectUsers(ctx, conn)
				require.NoError(t, err)

				cancel()

				err = res.Finish()
				require.ErrorIs(t, err, sqlite.ErrInterrupted)
			},
		},
		{
			name: "test kv",
			fn: func(t *testing.T, conn *sqlite.Connection) {
				ctx := context.Background()

				err := conn.Set(ctx, []byte("key"), []byte("value"))
				require.NoError(t, err)

				value, err := conn.Get(ctx, []byte("key"))
				require.NoError(t, err)

				require.Equal(t, []byte("value"), value)

				// get empty key
				value, err = conn.Get(ctx, []byte("key2"))
				require.NoError(t, err)

				require.Nil(t, value)
			},
		},
		{
			name: "kv conflict",
			fn: func(t *testing.T, conn *sqlite.Connection) {
				ctx := context.Background()

				err := conn.Set(ctx, []byte("key"), []byte("value"))
				require.NoError(t, err)

				err = conn.Set(ctx, []byte("key"), []byte("value2"))
				require.NoError(t, err)

				value, err := conn.Get(ctx, []byte("key"))
				require.NoError(t, err)

				require.Equal(t, []byte("value2"), value)
			},
		},
		{
			name: "kv non existent",
			fn: func(t *testing.T, conn *sqlite.Connection) {
				ctx := context.Background()

				value, err := conn.Get(ctx, []byte("key"))
				require.NoError(t, err)

				require.Nil(t, value)
			},
		},
		{
			name: "KV delete",
			fn: func(t *testing.T, conn *sqlite.Connection) {
				ctx := context.Background()

				err := conn.Set(ctx, []byte("key"), []byte("value"))
				require.NoError(t, err)

				err = conn.Delete(ctx, []byte("key"))
				require.NoError(t, err)

				value, err := conn.Get(ctx, []byte("key"))
				require.NoError(t, err)

				require.Nil(t, value)
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			defer deleteTempDir()
			if tt.flags&sql.OpenCreate == 0 {
				tt.flags |= sql.OpenCreate
			}

			conn, err := openDB("test", tt.flags)
			require.NoError(t, err)
			tt.fn(t, conn)
			err = conn.Close()
			require.NoError(t, err)

			// try calling close twice
			err = conn.Close()
			require.NoError(t, err)
		})
	}
}

// notEq returns true if a and b are not equal.
// It converts a and b to strings before comparing.
func notEq(a, b any) bool {
	return fmt.Sprintf("%v", a) != fmt.Sprintf("%v", b)
}

// we pass no OpenCreate flag, so it should fail
func Test_OpenNoCreate(t *testing.T) {
	_, err := openDB("test", 0)
	if err == nil {
		t.Fatal("expected error")
	}

	deleteTempDir()
}

func Test_InMemory(t *testing.T) {
	ctx := context.Background()
	conn, err := sqlite.Open(ctx, ":memory:", sql.OpenMemory)
	require.NoError(t, err)
	defer deleteTempDir()
	defer func() {
		err = conn.Close()
		require.NoError(t, err)
	}()

	res, err := conn.Execute(ctx, "SELECT 1+2", nil)
	require.NoError(t, err)

	rowReturned, err := res.Next()
	require.NoError(t, err)
	require.True(t, rowReturned)

	vals, err := res.Values()
	require.NoError(t, err)

	require.Len(t, vals, 1)

	if notEq(vals[0], int64(3)) {
		t.Errorf("expected 3, got %v", vals[0])
	}

	// test that the same db name can be opened twice if in memory
	_, err = sqlite.Open(ctx, ":memory:", sql.OpenMemory)
	require.NoError(t, err)
}

const tempDir = "./tmp"

func openDB(name string, flag sql.ConnectionFlag) (*sqlite.Connection, error) {
	ctx := context.Background()
	return sqlite.Open(ctx, fmt.Sprintf("%s/%s", tempDir, name), flag)
}
func deleteTempDir() {
	err := os.RemoveAll(tempDir)
	if err != nil {
		panic(err)
	}
}

const (
	createUsersTableStmt = `CREATE TABLE users (
		id INTEGER PRIMARY KEY NOT NULL,
		name TEXT NOT NULL,
		age INTEGER NOT NULL
	) WITHOUT ROWID, STRICT;`

	insertUserStmt = `INSERT INTO users (id, name, age) VALUES ($id, $name, $age);`

	selectUsersStmt = `SELECT * FROM users;`

	deleteUserStmt       = `DELETE FROM users WHERE id = $id;`
	createPostsTableStmt = `CREATE TABLE posts (
		id INTEGER PRIMARY KEY NOT NULL,
		content TEXT NOT NULL,
		user_id INTEGER NOT NULL,
		FOREIGN KEY (user_id) REFERENCES users (id) ON UPDATE CASCADE ON DELETE CASCADE
	) WITHOUT ROWID, STRICT;`

	insertPostStmt = `INSERT INTO posts (id, content, user_id) VALUES ($id, $content, $user_id);`

	selectPostsStmt = `SELECT * FROM posts;`
)

func createUserTable(ctx context.Context, conn *sqlite.Connection) error {
	res, err := conn.Execute(ctx, createUsersTableStmt, nil)
	if err != nil {
		return err
	}
	return res.Finish()
}

func createPostTable(ctx context.Context, conn *sqlite.Connection) error {
	res, err := conn.Execute(ctx, createPostsTableStmt, nil)
	if err != nil {
		return err
	}
	return res.Finish()
}

func insertUser(ctx context.Context, conn *sqlite.Connection, id int64, name string, age int64) error {
	res, err := conn.Execute(ctx, insertUserStmt, map[string]any{
		"$id":   id,
		"$name": name,
		"$age":  age,
	})
	if err != nil {
		return err
	}
	return res.Finish()
}

func insertPost(ctx context.Context, conn *sqlite.Connection, id int64, context string, userID int64) error {
	res, err := conn.Execute(ctx, insertPostStmt, map[string]any{
		"$id":      id,
		"$content": context,
		"$user_id": userID,
	})
	if err != nil {
		return err
	}
	return res.Finish()
}

func selectUsers(ctx context.Context, conn *sqlite.Connection) (sql.Result, error) {
	return conn.Execute(ctx, selectUsersStmt, nil)
}

func selectPosts(ctx context.Context, conn *sqlite.Connection) (sql.Result, error) {
	return conn.Execute(ctx, selectPostsStmt, nil)
}

func deleteUser(ctx context.Context, conn *sqlite.Connection, id int64) error {
	res, err := conn.Execute(ctx, deleteUserStmt, map[string]any{
		"$id": id,
	})
	if err != nil {
		return err
	}
	return res.Finish()
}

// getResults exports the results from a sqlite.Result.
func getResults(res sql.Result) ([]map[string]any, error) {
	results := make([]map[string]any, 0)
	for {
		rowReturned, err := res.Next()
		if err != nil {
			return nil, err
		}
		if !rowReturned {
			break
		}

		values, err := res.Values()
		if err != nil {
			return nil, err
		}

		record := make(map[string]any)

		for i, col := range res.Columns() {
			record[col] = values[i]
		}

		results = append(results, record)
	}

	return results, res.Close()
}
