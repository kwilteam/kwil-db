package sqlite_test

import (
	"context"
	"testing"
)

// common error is blob vals are returned as nil
func Test_ReadBytes(t *testing.T) {
	ctx := context.Background()
	conn, teardown := openRealDB()
	defer teardown()

	err := conn.Execute("CREATE TABLE test (id INTEGER PRIMARY KEY, data BLOB);", nil)
	if err != nil {
		t.Fatal(err)
	}

	err = conn.Execute("INSERT INTO test (id, data) VALUES (1, $data);", map[string]interface{}{
		"$data": []byte("test"),
	})
	if err != nil {
		t.Fatal(err)
	}

	results, err := conn.Query(ctx, "SELECT * FROM test;", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer results.Finish()

	var byteData []byte
	for {
		rowReturned, err := results.Next()
		if err != nil {
			t.Fatal(err)
		}

		if !rowReturned {
			break
		}

		data, ok := results.GetRecord()["data"].([]byte)
		if !ok {
			t.Fatal("expected data to be []byte")
		}

		byteData = data
	}

	if string(byteData) != "test" {
		t.Fatalf("expected data to be 'test', got %s", string(byteData))
	}
}
