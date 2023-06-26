package sqlite_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/kwilteam/kwil-db/pkg/sql/sqlite"
)

func Test_Query_With_Opts(t *testing.T) {
	ctx := context.Background()
	conn, teardown := openRealDB()
	defer teardown()

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
	results, err := stmt.Start(ctx, sqlite.WithNamedArgs(map[string]interface{}{
		"$id":   1,
		"$name": "John",
		"$age":  30,
	}),
	)
	if err != nil {
		t.Fatal(err)
	}

	err = results.Finish()
	if err != nil {
		t.Fatal(err)
	}

	// read it back
	results, err = conn.Query(ctx, "SELECT * FROM users")
	if err != nil {
		t.Fatal(err)
	}

	records, err := results.ExportRecords()
	if err != nil {
		t.Fatal(err)
	}

	if len(records) != 1 {
		t.Fatalf("expected 1 row, got %d", len(records))
	}
	if len(records[0]) != 3 {
		t.Fatalf("expected 3 columns, got %d", len(records[0]))
	}

	// Test ResultFunc
	results, err = conn.Query(ctx, "SELECT * FROM users")
	if err != nil {
		t.Fatal(err)
	}

	records, err = results.ExportRecords()
	if err != nil {
		t.Fatal(err)
	}

	if len(records) != 1 {
		t.Fatalf("expected 1 row, got %d", len(records))
	}

	// test numbered args
	// read it back
	results, err = conn.Query(ctx, "SELECT * FROM users WHERE id = $1", sqlite.WithArgs("1"))
	if err != nil {
		t.Fatal(err)
	}

	records, err = results.ExportRecords()
	if err != nil {
		t.Fatal(err)
	}

	if len(records) != 1 {
		t.Fatalf("expected 1 row, got %d", len(records))
	}

	// test named args
	// read it back
	results, err = conn.Query(ctx, "SELECT * FROM users WHERE id = $id",
		sqlite.WithNamedArgs(map[string]interface{}{
			"$id": 1,
		}),
	)
	if err != nil {
		t.Fatal(err)
	}

	records, err = results.ExportRecords()
	if err != nil {
		t.Fatal(err)
	}

	if len(records) != 1 {
		t.Fatalf("expected 1 row, got %d", len(records))
	}
	counter := 0
	// test result func
	results, err = conn.Query(ctx, "SELECT * FROM users")
	if err != nil {
		t.Fatal(err)
	}
	records, err = results.ExportRecords()
	if err != nil {
		t.Fatal(err)
	}

	if len(records) != 1 {
		t.Fatalf("expected 1 row, got %d", counter)
	}

	if records[0]["name"] != "John" {
		t.Fatalf("expected name to be John, got %s", records[0]["name"])
	}

	if records[0]["age"] != int64(30) {
		t.Fatalf("expected age to be 30, got %d", records[0]["age"])
	}

	if records[0]["id"] != int64(1) {
		t.Fatalf("expected id to be 1, got %d", records[0]["id"])
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
	ctx := context.Background()
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
	results, err := stmt.Start(ctx, sqlite.WithNamedArgs(map[string]interface{}{
		"$id":   1,
		"$name": "John",
		"$age":  30,
	}),
	)
	if err != nil {
		t.Fatal(err)
	}

	err = results.Finish()
	if err != nil {
		t.Fatal(err)
	}

	// query users
	results, err = conn.Query(ctx, "SELECT * FROM users")
	if err != nil {
		t.Fatal(err)
	}

	records, err := results.ExportRecords()
	if err != nil {
		t.Fatal(err)
	}

	if len(records) != 0 {
		t.Fatalf("expected 0 rows since insert is not committed, got %d", len(records))
	}

	// rollback
	err = sp.Rollback()
	if err != nil {
		t.Fatal(err)
	}

	// query users
	results, err = conn.Query(context.Background(), "SELECT * FROM users")
	if err != nil {
		t.Fatal(err)
	}

	records, err = results.ExportRecords()
	if err != nil {
		t.Fatal(err)
	}

	if len(records) != 0 {
		t.Fatalf("expected 0 rows since insert is not committed, got %d", len(records))
	}

	// re-insert users
	sp, err = conn.Savepoint()
	if err != nil {
		t.Fatal(err)
	}

	// insert users
	results, err = stmt.Start(ctx, sqlite.WithNamedArgs(map[string]interface{}{
		"$id":   1,
		"$name": "John",
		"$age":  30,
	}),
	)
	if err != nil {
		t.Fatal(err)
	}

	err = results.Finish()
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
	results, err = conn.Query(ctx, "SELECT * FROM users")
	if err != nil {
		t.Fatal(err)
	}

	records2, err := results.ExportRecords()
	if err != nil {
		t.Fatal(err)
	}

	if len(records2) != 1 {
		t.Errorf("expected 1 row")
	}
}

/*
	func Test_Global_Variables(t *testing.T) {
		ctx := context.Background()
		conn, td := openRealDB()
		defer td()

		// prepare statement
		stmt, err := conn.Prepare("INSERT INTO users (id, name, age) VALUES ($id, @caller, @block)")
		if err != nil {
			t.Fatal(err)
		}

		// test defaults
		results, err := stmt.Start(ctx, sqlite.WithNamedArgs(map[string]interface{}{
			"$id":     1,
			"@caller": "0xbennan",
			"@block":  420,
		}),
		)
		if err != nil {
			t.Fatal(err)
		}

		err = results.Finish()
		if err != nil {
			t.Fatal(err)
		}

		// query users
		results, err = conn.Query(context.Background(), "SELECT * FROM users")
		if err != nil {
			t.Fatal(err)
		}

		records, err := results.ExportRecords()
		if err != nil {
			t.Fatal(err)
		}

		if len(records) != 1 {
			t.Fatalf("expected 1 row, got %d", len(records))
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
		err = stmt.Execute(sqlite.WithNamedArgs(map[string]interface{}{
			"$id":     2,
			"@caller": "0xjohndoe",
			"@block":  69,
		}),
		)
		if err != nil {
			t.Fatal(err)
		}

		// query users
		results = &sqlite.ResultSet{}
		err = conn.Query(context.Background(), "SELECT * FROM users")
		if err != nil {
			t.Fatal(err)
		}

		if len(records) != 2 {
			t.Fatalf("expected 2 rows, got %d", len(records))
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
*/
func openRealDB() (conn *sqlite.Connection, teardown func() error) {
	conn, td, err := sqlite.OpenDbWithTearDown()
	if err != nil {
		panic(err)
	}

	initTables(conn)

	return conn, td
}

func openRealDBWithAttached() (conn *sqlite.Connection, teardown func() error) {
	conn1, err := sqlite.OpenConn("testdb", sqlite.WithPath("./tmp/"), sqlite.WithConnectionPoolSize(1), sqlite.WithAttachedDatabase("test_attach", "attachdb"))
	if err != nil {
		panic(err)
	}

	err = conn1.Delete()
	if err != nil {
		panic(err)
	}

	conn, err = sqlite.OpenConn("testdb", sqlite.WithPath("./tmp/"), sqlite.WithAttachedDatabase("test_attach", "attachdb"))
	if err != nil {
		panic(err)
	}

	initTables(conn)

	return conn, conn.Delete
}

func Test_Reads(t *testing.T) {
	ctx := context.Background()
	conn, td := openRealDB()
	defer td()
	// prepare statement
	stmt, err := conn.Prepare("INSERT INTO users (id, name, age) VALUES ($id, $name, $age)")
	if err != nil {
		t.Fatal(err)
	}

	// insert users
	results, err := stmt.Start(ctx,
		sqlite.WithNamedArgs(map[string]interface{}{
			"$id":   1,
			"$name": "John",
			"$age":  30,
		}),
	)
	if err != nil {
		t.Fatal(err)
	}

	err = results.Finish()
	if err != nil {
		t.Fatal(err)
	}

	// try to insert user with a query
	results, err = conn.Query(ctx, "INSERT INTO users (id, name, age) VALUES ($id, $name, $age)",
		sqlite.WithNamedArgs(map[string]interface{}{
			"$id":   2,
			"$name": "Jane",
			"$age":  25,
		}),
	)
	if err != nil {
		t.Fatal(err)
	}

	err = results.Finish()
	if err == nil {
		t.Fatal("expected error")
	}

	// test injection
	_, err = conn.Query(ctx, "SELECT * FROM users; INSERT INTO users VALUES (4, 'bb', 3);")
	if err == nil {
		t.Errorf("expected error")
	}

}

// testing statement prepare with two statements that are the same
// usually, this returns 1 prepared statement, but our implementation
// returns 2
func Test_Preparation(t *testing.T) {
	db, td := openRealDB()
	defer td()

	stmt, err := db.Prepare("SELECT * FROM users;")
	if err != nil {
		t.Fatal(err)
	}

	stmt2, err := db.Prepare("SELECT * FROM users;")
	if err != nil {
		t.Fatal(err)
	}

	err = stmt.Finalize()
	if err != nil {
		t.Fatal(err)
	}

	err = stmt2.Finalize()
	if err != nil {
		t.Fatal(err)
	}
}

func Test_CustomFunction(t *testing.T) {
	db, td := openRealDB()
	defer td()

	// testing the custom error('msg') function
	stmt, err := db.Prepare("SELECT ERROR('msg');")
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	results, err := stmt.Start(ctx)
	if err != nil {
		t.Fatal(err)
	}
	closeErr1 := results.Finish()
	if closeErr1 == nil {
		t.Errorf("expected error")
	}

	results, err = stmt.Start(ctx)
	if err != nil {
		t.Fatal(err)
	}
	closeErr2 := results.Finish()
	if closeErr2 == nil {
		t.Errorf("expected error")
	}

	if closeErr1.Error() != closeErr2.Error() {
		t.Fatalf("expected errors to be the same, got %s and %s", closeErr1.Error(), closeErr2.Error())
	}

	results, err = db.Query(ctx, "SELECT error('msg');")
	if err != nil {
		t.Fatal(err)
	}
	closeErr := results.Finish()
	if closeErr == nil {
		t.Errorf("expected no error")
	}

	// try inserting data with an error, within a savepoint
	sp, err := db.Savepoint()
	if err != nil {
		t.Fatal(err)
	}

	stmt, err = db.Prepare("INSERT INTO users (id, name, age) VALUES ($id, $name, $age) RETURNING ERROR('msg');")
	if err != nil {
		t.Fatal(err)
	}

	results, err = stmt.Start(ctx, sqlite.WithNamedArgs(map[string]interface{}{
		"$id":   1,
		"$name": "John",
		"$age":  30,
	}))
	if err != nil {
		t.Fatal(err)
	}
	finishErr := results.Finish()
	if finishErr == nil {
		t.Errorf("expected error")
	}

	err = sp.Rollback()
	if err != nil {
		t.Fatal(err)
	}

	// ensure that no user was inserted
	results, err = db.Query(ctx, "SELECT * FROM users")
	if err != nil {
		t.Fatal(err)
	}

	records, err := results.ExportRecords()
	if err != nil {
		t.Fatal(err)
	}

	if len(records) != 0 {
		t.Errorf("expected 0 rows, got %d", len(records))
	}
}

func Test_Attach(t *testing.T) {
	ctx := context.Background()
	err, attachedTeardown := createAttachDB()
	if err != nil {
		t.Fatal(err)
	}
	defer attachedTeardown()

	db, td := openRealDBWithAttached()
	defer td()

	// try preparing and executing a statement
	stmt, err := db.Prepare("INSERT INTO test_attach.users (id, name, age) VALUES ($id, $name, $age)")
	if err != nil {
		t.Fatal(err)
	}

	results, err := stmt.Start(ctx, sqlite.WithNamedArgs(map[string]interface{}{
		"$id":   1,
		"$name": "John",
		"$age":  30,
	}))
	if err != nil {
		t.Fatal(err)
	}

	err = results.Finish()
	if err == nil {
		t.Errorf("expected error when inserting into attached database")
	}

	// test that we can read from the attached database
	// if it is not registered it will panic
	results, err = db.Query(ctx, "SELECT * FROM test_attach.users")
	if err != nil {
		t.Fatal(err)
	}

	err = results.Finish()
	if err != nil {
		t.Fatal(err)
	}
}

const (
	usersTable = `CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY NOT NULL,
		name TEXT NOT NULL,
		age INTEGER NOT NULL
	) WITHOUT ROWID, STRICT;`
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
		panic("expected 1 table, got " + fmt.Sprint(len(tables)))
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

const attachedDBFileName = "attachdb"

func createAttachDB() (error, func() error) {
	conn1, err := sqlite.OpenConn(attachedDBFileName, sqlite.WithPath("./tmp/"), sqlite.WithConnectionPoolSize(1))
	if err != nil {
		return err, nil
	}

	err = conn1.Delete()
	if err != nil {
		return err, nil
	}

	conn, err := sqlite.OpenConn(attachedDBFileName, sqlite.WithPath("./tmp/"))
	if err != nil {
		return err, nil
	}

	initTables(conn)

	return nil, conn.Delete
}
