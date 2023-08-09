package client_test

import (
	"context"
	"testing"

	client "github.com/kwilteam/kwil-db/pkg/sql/client"
)

// this test is very basic, but this package is tiny and will likely get moved into sqlite soon
func Test_Client(t *testing.T) {
	ctx := context.Background()

	db, cleanup := openTestDB(t)
	defer cleanup()

	err := db.Execute(ctx, "CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT) STRICT, WITHOUT ROWID;", nil)
	if err != nil {
		t.Fatal(err)
	}

	insertStmt, err := db.Prepare("INSERT INTO test (id, name) VALUES ($id, $name);")
	if err != nil {
		t.Fatal(err)
	}

	_, err = insertStmt.Execute(ctx, map[string]interface{}{
		"$id":   1,
		"$name": "test",
	})
	if err != nil {
		t.Fatal(err)
	}

	results, err := db.Query(ctx, "SELECT * FROM test;", nil)
	if err != nil {
		t.Fatal(err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	exists, err := db.TableExists(ctx, "test")
	if err != nil {
		t.Fatal(err)
	}

	if !exists {
		t.Fatal("expected table to exist")
	}
}

func Test_Session(t *testing.T) {
	ctx := context.Background()

	db, cleanup := openTestDB(t)
	defer cleanup()

	err := db.Execute(ctx, "CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT) STRICT, WITHOUT ROWID;", nil)
	if err != nil {
		t.Fatal(err)
	}

	ses, err := db.CreateSession()
	if err != nil {
		t.Fatal(err)
	}
	defer ses.Delete()

	insertStmt, err := db.Prepare("INSERT INTO test (id, name) VALUES ($id, $name);")
	if err != nil {
		t.Fatal(err)
	}

	_, err = insertStmt.Execute(ctx, map[string]interface{}{
		"$id":   1,
		"$name": "test",
	})
	if err != nil {
		t.Fatal(err)
	}

	results, err := ses.GenerateChangeset()
	if err != nil {
		t.Fatal(err)
	}

	if len(results) != 26 {
		t.Fatalf("expected 26 result, got %d", len(results))
	}
}

func Test_Step(t *testing.T) {
	ctx := context.Background()

	db, cleanup := openTestDB(t)
	defer cleanup()

	err := db.Execute(ctx, "CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT) STRICT, WITHOUT ROWID;", nil)
	if err != nil {
		t.Fatal(err)
	}

	err = db.Execute(ctx, "CREATE TABLE test2 (id INTEGER PRIMARY KEY, name TEXT) STRICT, WITHOUT ROWID;", nil)
	if err != nil {
		t.Fatal(err)
	}

	insertStmt, err := db.Prepare("INSERT INTO test2 (id, name) VALUES ($id, $name);")
	if err != nil {
		t.Fatal(err)
	}

	_, err = insertStmt.Execute(ctx, map[string]interface{}{
		"$id":   1,
		"$name": "test",
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = insertStmt.Execute(ctx, map[string]interface{}{
		"$id":   2,
		"$name": "test2",
	})
	if err != nil {
		t.Fatal(err)
	}

	inert2Stmt, err := db.Prepare("INSERT INTO test2 (id, name) SELECT id, name FROM test ORDER BY id;")
	if err != nil {
		t.Fatal(err)
	}

	_, err = inert2Stmt.Execute(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
}

func openTestDB(t *testing.T) (*client.SqliteClient, func()) {

	db, err := client.NewSqliteStore("testdb")
	if err != nil {
		t.Fatal(err)
	}

	return db, func() {
		err = db.Delete()
		if err != nil {
			t.Fatal(err)
		}
	}
}
