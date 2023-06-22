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

	err := db.Execute("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT) STRICT, WITHOUT ROWID;")
	if err != nil {
		t.Fatal(err)
	}

	insertStmt, err := db.Prepare("INSERT INTO test (id, name) VALUES ($id, $name);")
	if err != nil {
		t.Fatal(err)
	}

	_, err = insertStmt.Execute(map[string]interface{}{
		"$id":   1,
		"$name": "test",
	})
	if err != nil {
		t.Fatal(err)
	}

	results, err := db.Query(ctx, "SELECT * FROM test;")
	if err != nil {
		t.Fatal(err)
	}

	resultMaps, err := client.ResultsfromReader(results)
	if err != nil {
		t.Fatal(err)
	}

	if len(resultMaps) != 1 {
		t.Fatalf("expected 1 result, got %d", len(resultMaps))
	}

	exists, err := db.TableExists(ctx, "test")
	if err != nil {
		t.Fatal(err)
	}

	if !exists {
		t.Fatal("expected table to exist")
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
