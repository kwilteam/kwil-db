package sqlite_test

import (
	"context"
	"testing"

	"kwil/pkg/sql/sqlite"
)

func Test_Query_With_Opts(t *testing.T) {
	conn := openMemoryDB()

	defer conn.Close(nil)

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
	count := 0
	err = conn.Query(ctx, "SELECT * FROM users", &sqlite.ExecOpts{
		ResultFunc: func(stmt *sqlite.Statement) error {
			count++
			return nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if count != 1 {
		t.Fatalf("expected 1 row, got %d", count)
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

func openMemoryDB() *sqlite.Connection {
	conn, err := sqlite.OpenConn("testdb", sqlite.InMemory())
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
		panic("expected 1 table")
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
