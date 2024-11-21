package badger_test

import (
	"context"
	"testing"

	badgerTesting "github.com/kwilteam/kwil-db/internal/kv/badger/testing"
)

// testing double write does not produce an error
func Test_BadgerKV(t *testing.T) {
	db, td, err := badgerTesting.NewTestBadgerDB(context.Background(), "test", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer td()

	err = db.Set([]byte("key"), []byte("value"))
	if err != nil {
		t.Fatal(err)
	}

	err = db.Set([]byte("key"), []byte("value2"))
	if err != nil {
		t.Fatal(err)
	}

	val, err := db.Get([]byte("key"))
	if err != nil {
		t.Fatal(err)
	}

	if string(val) != "value2" {
		t.Fatal("expected value2")
	}

	err = db.Delete([]byte("key"))
	if err != nil {
		t.Fatal(err)
	}

	// try deleting non-existent key
	err = db.Delete([]byte("key"))
	if err != nil {
		t.Fatal(err)
	}
}
