package sqlite_test

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/kwilteam/kwil-db/pkg/sql/sqlite"
)

func Test_Query_With_Opts(t *testing.T) {
	conn := openMemoryDB()

	defer conn.Close(nil)

	// prepare invalid statement
	_, err := conn.Prepare("INSERT INTOewfnw users (id, name, age) VALUES ($id, $name, $age)")
	if err == nil {
		t.Fatal("expected error")
	}

	// prepare statement with trailing bytes
	_, err = conn.Prepare("INSERT INTO users (id, name, age) VALUES ($id, $name, $age); DROP TABLE users")
	if err == nil {
		t.Fatal("expected error")
	}

	// prepare statement
	stmt, err := conn.Prepare("INSERT INTO users (id, name, age) VALUES ($id, $name, $age)")
	if err != nil {
		t.Fatal(err)
	}

	// insert users
	err = stmt.Execute(&sqlite.ExecOpts{
		NamedArgs: map[string]interface{}{
			"$id":   1,
			"$name": "John",
			"$age":  30,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	// test ResultSet read back
	results := &sqlite.ResultSet{}
	// read it back
	err = conn.Query(ctx, "SELECT * FROM users", &sqlite.ExecOpts{
		ResultSet: results,
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(results.Rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(results.Rows))
	}
	if len(results.Rows[0]) != 3 {
		t.Fatalf("expected 3 columns, got %d", len(results.Rows[0]))
	}

	// Test ResultFunc
	result := &sqlite.ResultSet{}
	err = conn.Query(ctx, "SELECT * FROM users", &sqlite.ExecOpts{
		ResultSet: result,
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(result.Rows))
	}

	// test numbered args
	results = &sqlite.ResultSet{}
	// read it back
	err = conn.Query(ctx, "SELECT * FROM users WHERE id = $1", &sqlite.ExecOpts{
		ResultSet: results,
		Args:      []interface{}{"1"},
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(results.Rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(results.Rows))
	}

	// test named args
	results = &sqlite.ResultSet{}
	// read it back
	err = conn.Query(ctx, "SELECT * FROM users WHERE id = $id", &sqlite.ExecOpts{
		ResultSet: results,
		NamedArgs: map[string]interface{}{
			"$id": 1,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(results.Rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(results.Rows))
	}

	var userName string
	var userAge int64
	var userId int64
	counter := 0
	// test result func
	err = conn.Query(ctx, "SELECT * FROM users", &sqlite.ExecOpts{
		ResultFunc: func(stmt *sqlite.Statement) error {
			counter++
			userName = stmt.GetText("name")
			userAge = stmt.GetInt64("age")
			userId = stmt.GetInt64("id")
			return nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if counter != 1 {
		t.Fatalf("expected 1 row, got %d", counter)
	}

	if userName != "John" {
		t.Fatalf("expected name to be John, got %s", userName)
	}

	if userAge != 30 {
		t.Fatalf("expected age to be 30, got %d", userAge)
	}

	if userId != 1 {
		t.Fatalf("expected id to be 1, got %d", userId)
	}

	// delete the database
	err = conn.Delete()
	if err != nil {
		t.Fatal(err)
	}

	// check that the database is gone
	_, err = conn.ListTables(ctx)
	if err == nil {
		t.Fatal("expected error")
	}
}

func Test_Database_Wal(t *testing.T) {
	conn, teardown := openRealDB()
	defer teardown()

	sp, err := conn.Savepoint()
	if err != nil {
		t.Fatal(err)
	}

	// prepare statement
	stmt, err := conn.Prepare("INSERT INTO users (id, name, age) VALUES ($id, $name, $age)")
	if err != nil {
		t.Fatal(err)
	}

	// insert users
	err = stmt.Execute(&sqlite.ExecOpts{
		NamedArgs: map[string]interface{}{
			"$id":   1,
			"$name": "John",
			"$age":  30,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	// query users
	results := &sqlite.ResultSet{}
	err = conn.Query(context.Background(), "SELECT * FROM users", &sqlite.ExecOpts{
		ResultSet: results,
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(results.Rows) != 0 {
		t.Fatalf("expected 0 rows since insert is not committed, got %d", len(results.Rows))
	}

	// rollback
	err = sp.Rollback()
	if err != nil {
		t.Fatal(err)
	}

	// query users
	results = &sqlite.ResultSet{}
	err = conn.Query(context.Background(), "SELECT * FROM users", &sqlite.ExecOpts{
		ResultSet: results,
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(results.Rows) != 0 {
		t.Fatalf("expected 0 rows since insert is not committed, got %d", len(results.Rows))
	}

	// re-insert users
	sp, err = conn.Savepoint()
	if err != nil {
		t.Fatal(err)
	}

	// insert users
	err = stmt.Execute(&sqlite.ExecOpts{
		NamedArgs: map[string]interface{}{
			"$id":   1,
			"$name": "John",
			"$age":  30,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	err = stmt.Finalize()
	if err != nil {
		t.Fatal(err)
	}

	err = sp.CommitAndCheckpoint()
	if err != nil {
		t.Fatal(err)
	}

	// query users
	results = &sqlite.ResultSet{}
	err = conn.Query(context.Background(), "SELECT * FROM users", &sqlite.ExecOpts{
		ResultSet: results,
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(results.Rows) != 1 {
		t.Fatal("expected 1 row")
	}

	res2 := results.Records()
	if len(res2) != 1 {
		t.Fatal("expected 1 row")
	}
}

func Test_Global_Variables(t *testing.T) {
	conn := openMemoryDB(sqlite.WithGlobalVariables(map[string]any{
		"@caller": "0xbennan",
		"@block":  420,
	}))

	defer conn.Close(nil)

	// prepare statement
	stmt, err := conn.Prepare("INSERT INTO users (id, name, age) VALUES ($id, @caller, @block)")
	if err != nil {
		t.Fatal(err)
	}

	// test defaults
	err = stmt.Execute(&sqlite.ExecOpts{
		NamedArgs: map[string]interface{}{
			"$id": 1,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	// query users
	results := &sqlite.ResultSet{}
	err = conn.Query(context.Background(), "SELECT * FROM users", &sqlite.ExecOpts{
		ResultSet: results,
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(results.Rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(results.Rows))
	}

	results.Next()
	retrievedNamed, ok := results.GetColumn("name").(string)
	if !ok {
		t.Fatalf("expected string, got %T", results.GetColumn("name"))
	}

	if retrievedNamed != "0xbennan" {
		t.Fatalf("expected 0xbennan, got %s", retrievedNamed)
	}

	retrievedAge, ok := results.GetColumn("age").(int64)
	if !ok {
		t.Fatalf("expected int64, got %T", results.GetColumn("age"))
	}

	if retrievedAge != 420 {
		t.Fatalf("expected 420, got %d", retrievedAge)
	}

	// now we test that the global variables can be overwritten
	err = stmt.Execute(&sqlite.ExecOpts{
		NamedArgs: map[string]interface{}{
			"$id":     2,
			"@caller": "0xjohndoe",
			"@block":  69,
		}})
	if err != nil {
		t.Fatal(err)
	}

	// query users
	results = &sqlite.ResultSet{}
	err = conn.Query(context.Background(), "SELECT * FROM users", &sqlite.ExecOpts{
		ResultSet: results,
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(results.Rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(results.Rows))
	}

	results.Next()
	results.Next() // skip first row.  sqlite's ordering is deterministic, even without an ORDER BY clause
	retrievedNamed, ok = results.GetColumn("name").(string)
	if !ok {
		t.Fatalf("expected string, got %T", results.GetColumn("name"))
	}

	if retrievedNamed != "0xjohndoe" {
		t.Fatalf("expected 0xjohndoe, got %s", retrievedNamed)
	}

	retrievedAge, ok = results.GetColumn("age").(int64)
	if !ok {
		t.Fatalf("expected int64, got %T", results.GetColumn("age"))
	}

	if retrievedAge != 69 {
		t.Fatalf("expected 69, got %d", retrievedAge)
	}

	err = stmt.Finalize()
	if err != nil {
		t.Fatal(err)
	}
}

func openRealDB() (conn *sqlite.Connection, teardown func() error) {
	conn, err := sqlite.OpenConn("testdb", sqlite.WithPath("./tmp/"))
	if err != nil {
		panic(err)
	}

	err = conn.Delete()
	if err != nil {
		panic(err)
	}

	conn, err = sqlite.OpenConn("testdb", sqlite.WithPath("./tmp/"))
	if err != nil {
		panic(err)
	}

	initTables(conn)

	return conn, conn.Delete
}

// this runs the same insert 100 times, and tests each result 100 times
func Test_Order_Determinism(t *testing.T) {

	rand.Seed(time.Now().UnixNano())

	for i := 0; i < 100; i++ {
		wg := sync.WaitGroup{}
		wg.Add(1)
		func() {
			conn := openMemoryDB(sqlite.WithConnectionPoolSize(1))

			defer func() { // since delete takes a while, we have to wait for it to finish before we can close the connection
				err := conn.Delete()
				if err != nil {
					t.Fatal(err)
				}
				wg.Done()
			}()

			// prepare statement
			stmt, err := conn.Prepare("INSERT INTO users (id, name, age) VALUES ($id, $name, $age)")
			if err != nil {
				t.Fatal(err)
			}

			users := []map[string]interface{}{
				{
					"$id":   1,
					"$name": "bennan",
					"$age":  420,
				},
				{
					"$id":   2,
					"$name": "luke",
					"$age":  69,
				},
				{
					"$id":   3,
					"$name": "gavin",
					"$age":  65,
				},
				{
					"$id":   4,
					"$name": "luis",
					"$age":  61023,
				},
			}

			for _, user := range users {
				err = stmt.Execute(&sqlite.ExecOpts{
					NamedArgs: user,
				})
				if err != nil {
					t.Fatal(err)
				}
			}

			for i := 0; i < 100; i++ {
				// query users
				results := &sqlite.ResultSet{}
				err = conn.Query(context.Background(), "SELECT * FROM users as u1 CROSS JOIN users AS u2", &sqlite.ExecOpts{
					ResultSet: results,
				})
				if err != nil {
					t.Fatal(err)
				}

				idCounter := int64(1)
				for results.Next() {
					rec := results.GetRecord()

					id, ok := rec["id"].(int64)
					if !ok {
						t.Fatalf("expected int64, got %T", rec["id"])
					}

					if (idCounter-id)%4 != 0 {
						t.Fatalf("expected %d, got %d", idCounter, id)
					}

					idCounter++
				}
			}

		}()
		wg.Wait()
	}
}

func Test_Reads(t *testing.T) {
	conn, td := openRealDB()
	defer td()

	defer conn.Delete()

	// prepare statement
	stmt, err := conn.Prepare("INSERT INTO users (id, name, age) VALUES ($id, $name, $age)")
	if err != nil {
		t.Fatal(err)
	}

	// insert users
	err = stmt.Execute(&sqlite.ExecOpts{
		NamedArgs: map[string]interface{}{
			"$id":   1,
			"$name": "John",
			"$age":  30,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	// try to insert user with a query
	err = conn.Query(context.Background(), "INSERT INTO users (id, name, age) VALUES ($id, $name, $age)", &sqlite.ExecOpts{
		NamedArgs: map[string]interface{}{
			"$id":   2,
			"$name": "Jane",
			"$age":  25,
		},
	})
	if err == nil {
		t.Fatal("expected error")
	}

	// test injection
	err = conn.Query(context.Background(), "SELECT * FROM users; INSERT INTO users VALUES (4, 'bb', 3)", &sqlite.ExecOpts{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func openMemoryDB(opts ...sqlite.ConnectionOption) *sqlite.Connection {
	opts = append(opts, sqlite.InMemory())

	conn, err := sqlite.OpenConn("testdb", opts...)
	if err != nil {
		panic(err)
	}

	initTables(conn)

	return conn
}

const (
	usersTable = `CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY NOT NULL,
		name TEXT NOT NULL,
		age INTEGER NOT NULL
	);`
)

func initTables(conn *sqlite.Connection) {
	err := conn.Execute(usersTable)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	tables, err := conn.ListTables(ctx)
	if err != nil {
		panic(err)
	}

	if len(tables) != 1 {
		panic("expected 1 table, got " + fmt.Sprint((len(tables))))
	}

	if tables[0] != "users" {
		panic("expected users table")
	}

	// also test if table exists
	exists, err := conn.TableExists(ctx, "users")
	if err != nil {
		panic(err)
	}

	if !exists {
		panic("expected users table to exist")
	}
}
