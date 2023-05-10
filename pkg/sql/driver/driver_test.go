package driver_test

import (
	"fmt"
	"github.com/kwilteam/kwil-db/pkg/sql/driver"
	"sync"
	"testing"
	"time"
)

const (
	createTestTable = `CREATE TABLE test_table (id INTEGER PRIMARY KEY, name TEXT);`
	dropTestTable   = `DROP TABLE IF EXISTS test_table;`
	insertTestRow   = `INSERT INTO test_table (id, name) VALUES ($id, $name);`
	updateTestRow   = `UPDATE test_table SET name = $name WHERE id = $id;`
	deleteTestRow   = `DELETE FROM test_table WHERE id = $id;`
)

func Test_Driver(t *testing.T) {
	conn, err := createTestDB()
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	defer conn.ReleaseLock()

	// test insert
	err = conn.Execute(insertTestRow, 1, "test1")
	if err != nil {
		t.Errorf("failed to insert: %v", err)
	}

	// test insert named
	err = conn.ExecuteNamed(insertTestRow, map[string]interface{}{
		"$name":   "test2",
		"$id":     2,
		"@caller": "test", // testing flag override
	}, nil)
	if err != nil {
		t.Errorf("failed to insert: %v", err)
	}

	// read back the rows
	type Row struct {
		id   int64
		name string
	}

	rows := []Row{}
	err = conn.Query("SELECT id, name FROM test_table;", func(stmt *driver.Statement) error {
		var row Row
		row.id = stmt.GetInt64("id")
		row.name = stmt.GetText("name")

		rows = append(rows, row)
		return nil
	})
	if err != nil {
		t.Errorf("failed to query: %v", err)
	}

	if len(rows) != 2 {
		t.Errorf("expected 2 rows, got %d", len(rows))
	}

	if rows[0].id != 1 || rows[0].name != "test1" {
		t.Errorf("expected row 1 to be {1, test1}, got {%d, %s}", rows[0].id, rows[0].name)
	}

	// test update
	err = conn.Execute(updateTestRow, "test1-updated", 1)
	if err != nil {
		t.Errorf("failed to update: %v", err)
	}

	// test update named
	err = conn.ExecuteNamed(updateTestRow, map[string]interface{}{
		"$name": "test2-updated",
		"$id":   2,
	}, nil)
	if err != nil {
		t.Errorf("failed to update: %v", err)
	}

	rows = []Row{}
	err = conn.Query("SELECT id, name FROM test_table;", func(stmt *driver.Statement) error {
		var row Row
		row.id = stmt.GetInt64("id")
		row.name = stmt.GetText("name")

		rows = append(rows, row)
		return nil
	})
	if err != nil {
		t.Errorf("failed to query: %v", err)
	}

	if len(rows) != 2 {
		t.Errorf("expected 2 rows, got %d", len(rows))
	}

	if rows[0].id != 1 || rows[0].name != "test1-updated" {
		t.Errorf("expected row 1 to be {1, test1-updated}, got {%d, %s}", rows[0].id, rows[0].name)
	}

	ro, err := conn.CopyReadOnly()
	if err != nil {
		t.Errorf("failed to copy read only: %v", err)
	}

	var rows2 []Row
	err = ro.Query("SELECT id, name FROM test_table;", func(stmt *driver.Statement) error {
		var row Row
		row.id = stmt.GetInt64("id")
		row.name = stmt.GetText("name")

		rows2 = append(rows2, row)
		return nil
	})
	if err != nil {
		t.Errorf("failed to query: %v", err)
	}

	if rows2[0].id != 1 || rows2[0].name != "test1-updated" {
		t.Errorf("expected row to be {0, }, got {%d, %s}", rows2[0].id, rows2[0].name)
	}

	// test delete
	err = conn.Execute(deleteTestRow, 1)
	if err != nil {
		t.Errorf("failed to delete: %v", err)
	}

	// test delete named
	err = conn.ExecuteNamed(deleteTestRow, map[string]interface{}{
		"$id": 2,
	}, nil)
	if err != nil {
		t.Errorf("failed to delete: %v", err)
	}

	rows = []Row{}
	err = conn.Query("SELECT id, name FROM test_table;", func(stmt *driver.Statement) error {
		var row Row
		row.id = stmt.GetInt64("id")
		row.name = stmt.GetText("name")

		rows = append(rows, row)
		return nil
	})
	if err != nil {
		t.Errorf("failed to query: %v", err)
	}

	if len(rows) != 0 {
		t.Errorf("expected 0 rows, got %d", len(rows))
	}
}

// this is to test a weird edge case i think i fixed
func Test_SingleVal(t *testing.T) {
	conn, err := createTestDB()
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	defer conn.ReleaseLock()

	// test insert
	err = conn.Execute(insertTestRow, 6482, "test431")
	if err != nil {
		t.Errorf("failed to insert: %v", err)
	}

	// read back the rows
	type Row struct {
		id   int64
		name string
	}

	rows := []Row{}
	err = conn.Query("SELECT id, name FROM test_table;", func(stmt *driver.Statement) error {
		var row Row
		row.id = stmt.GetInt64("id")
		row.name = stmt.GetText("name")

		rows = append(rows, row)
		return nil
	})
	if err != nil {
		t.Errorf("failed to query: %v", err)
	}

	if len(rows) != 1 {
		t.Errorf("expected 1 rows, got %d", len(rows))
	}

	if rows[0].id != 6482 || rows[0].name != "test431" {
		t.Errorf("expected row 1 to be {6482, test431}, got {%d, %s}", rows[0].id, rows[0].name)
	}
}

func Test_RapidWrite(t *testing.T) {
	conn, err := createTestDB()
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	defer conn.ReleaseLock()

	wg := sync.WaitGroup{}
	wg.Add(100)

	err = conn.Execute(insertTestRow, 1000000, "test1000000")
	if err != nil {
		t.Errorf("failed to insert: %v", err)
	}

	for i := 0; i < 100; i++ {
		go func(i int) {
			idpref := i * 1000
			for j := 0; j < 10; j++ {
				id := idpref + j
				err = conn.Execute(insertTestRow, id, fmt.Sprintf("test%d", id))
				if err != nil {
					t.Errorf("failed to insert: %v", err)
				}
			}

			wg.Done()
		}(i)
	}

	wg.Wait()

}

func Test_Savepoints(t *testing.T) {
	// TEST 1 rollback
	conn, err := createTestDB()
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	defer conn.ReleaseLock()

	sp, err := conn.Begin()
	if err != nil {
		t.Fatal(err)
	}

	err = sp.Execute(insertTestRow, 1, "test1")
	if err != nil {
		t.Errorf("failed to insert: %v", err)
	}

	err = sp.Rollback()
	if err != nil {
		t.Fatal(err)
	}

	rows := []int64{}
	err = conn.Query("SELECT id FROM test_table;", func(stmt *driver.Statement) error {
		rows = append(rows, stmt.GetInt64("id"))
		return nil
	})
	if err != nil {
		t.Errorf("failed to query: %v", err)
	}

	if len(rows) != 0 {
		t.Errorf("expected 0 rows, got %d", len(rows))
	}

	// TEST 2 changeset

	// test savepoint with commit
	sp2, err := conn.Begin()
	if err != nil {
		t.Fatal(err)
	}

	err = conn.Execute(insertTestRow, 1, "test1")
	if err != nil {
		t.Errorf("failed to insert: %v", err)
	}

	// now commit
	err = sp2.Commit()
	if err != nil {
		t.Fatal(err)
	}

	rows = []int64{}
	err = conn.Query("SELECT id FROM test_table;", func(stmt *driver.Statement) error {
		rows = append(rows, stmt.GetInt64("id"))
		return nil
	})
	if err != nil {
		t.Errorf("failed to query: %v", err)
	}

	if len(rows) != 1 {
		t.Errorf("expected 1 row, got %d", len(rows))
	}

	// now delete the row
	err = conn.Execute(deleteTestRow, 1)
	if err != nil {
		t.Errorf("failed to delete: %v", err)
	}

	rows = []int64{}
	err = conn.Query("SELECT id FROM test_table;", func(stmt *driver.Statement) error {
		rows = append(rows, stmt.GetInt64("id"))
		return nil
	})
	if err != nil {
		t.Errorf("failed to query: %v", err)
	}

	if len(rows) != 0 {
		t.Errorf("expected 0 rows, got %d", len(rows))
	}
}

func createTestDB() (*driver.Connection, error) {
	conn, err := driver.OpenConn("test_database_NOBODYBETTERGUESSTHISDBID", driver.WithInjectableVars([]*driver.InjectableVar{
		{
			Name:       "@caller",
			DefaultVal: "test",
		},
	})) // added random string to avoid collisions
	if err != nil {
		return nil, err
	}

	err = conn.AcquireLock()
	if err != nil {
		return nil, err
	}

	// wipe the database
	err = conn.Execute(dropTestTable)
	if err != nil {
		return nil, err
	}

	err = conn.Execute(createTestTable)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func Test_Injectable(t *testing.T) {
	conn, err := createTestDB()
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	defer conn.ReleaseLock()

	// test insert
	err = conn.Execute(insertTestRow, 1, "test1")
	if err != nil {
		t.Errorf("failed to insert: %v", err)
	}

	// test insert named
	err = conn.ExecuteNamed(insertTestRow, map[string]interface{}{
		"$name":   "test2",
		"$id":     2,
		"@caller": "test", // testing flag override
	}, nil)
	if err != nil {
		t.Errorf("failed to insert: %v", err)
	}

	// query
	err = conn.QueryNamed("SELECT * FROM test_table WHERE id = $id;", func(s *driver.Statement) error {
		if s.GetInt64("id") != 2 {
			t.Errorf("expected id to be 2, got %d", s.GetInt64("id"))
		}

		if s.GetText("name") != "test2" {
			t.Errorf("expected name to be test2, got %s", s.GetText("name"))
		}

		return nil
	},
		map[string]interface{}{
			"$id":     2,
			"@caller": "test", // testing injected
		})

	if err != nil {
		t.Errorf("failed to query: %v", err)
	}
}

func Test_ReadOnly(t *testing.T) {
	db, err := createTestDB()
	if err != nil {
		t.Fatal(err)
	}

	defer db.Close()

	ro, err := db.CopyReadOnly()
	if err != nil {
		t.Fatal(err)
	}

	defer ro.Close()

	// test insert
	err = ro.Execute(insertTestRow, 1, "test1")
	if err == nil {
		t.Errorf("expected error, got nil")
	}

	// read
	err = ro.QueryNamed("SELECT * FROM test_table WHERE id = $id;", func(s *driver.Statement) error {
		if s.GetInt64("id") != 1 {
			t.Errorf("expected id to be 1, got %d", s.GetInt64("id"))
		}

		if s.GetText("name") != "test1" {
			t.Errorf("expected name to be test1, got %s", s.GetText("name"))
		}

		return nil
	},
		map[string]interface{}{
			"$id": 1,
		})

	if err != nil {
		t.Errorf("failed to query: %v", err)
	}
}

func Test_ReadExec(t *testing.T) {
	conn, err := createTestDB()
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	defer conn.ReleaseLock()

	// test insert
	err = conn.Execute(insertTestRow, 1, "test1")
	if err != nil {
		t.Errorf("failed to insert: %v", err)
	}

	retVals := []int64{}
	// test read
	err = conn.ExecuteNamed("SELECT * FROM test_table", nil, func(s *driver.Statement) error {
		retVals = append(retVals, s.GetInt64("id"))
		return nil
	})
	if err != nil {
		t.Errorf("failed to insert: %v", err)
	}

	if len(retVals) != 1 {
		t.Errorf("expected 1 row, got %d", len(retVals))
	}

	// drop
	err = conn.Execute(dropTestTable)
	if err != nil {
		t.Errorf("failed to drop: %v", err)
	}
}

// previous error where db can crash if you query a non-existent table
func Test_NonexistentTable(t *testing.T) {
	conn, err := createTestDB()
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	defer conn.ReleaseLock()

	// test read
	err = conn.QueryNamed("SELECT * FROM test_table23w24e2", func(s *driver.Statement) error {
		return nil
	}, nil)
	if err == nil {
		t.Errorf("expected error, got nil")
	}
	// try read with execute
	err = conn.ExecuteNamed("SELECT * FROM test_table23w24e2", nil, func(s *driver.Statement) error {
		return nil
	})
	if err == nil {
		t.Errorf("expected error, got nil")
	}

	// check that table does not exist
	exists, err := conn.TableExists("test_table23w24e2")
	if err != nil {
		t.Errorf("failed to check table exists: %v", err)
	}

	if exists {
		t.Errorf("expected table to not exist")
	}
}

func Test_Reopen(t *testing.T) {
	conn, err := createTestDB()
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	defer conn.ReleaseLock()

	// test insert named
	err = conn.ExecuteNamed(insertTestRow, map[string]interface{}{
		"$name":   "test2",
		"$id":     2,
		"@caller": "test", // testing flag override
	}, nil)
	if err != nil {
		t.Errorf("failed to insert: %v", err)
	}

	// read-only
	roConn, err := conn.CopyReadOnly()
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 10; i++ {
		err = roConn.ReOpen()
		if err != nil {
			t.Errorf("failed to reopen: %v", err)
		}

		time.Sleep(100 * time.Millisecond)

		// insert another row
		err = conn.ExecuteNamed(insertTestRow, map[string]interface{}{
			"$name":   "test2",
			"$id":     i * 100,
			"@caller": "test", // testing flag override
		}, nil)
		if err != nil {
			t.Errorf("failed to insert: %v", err)
		}

		// read back the rows
		type Row struct {
			id   int64
			name string
		}

		rows := []Row{}
		err = conn.Query("SELECT id, name FROM test_table;", func(stmt *driver.Statement) error {
			var row Row
			row.id = stmt.GetInt64("id")
			row.name = stmt.GetText("name")

			rows = append(rows, row)
			return nil
		})
		if err != nil {
			t.Errorf("failed to query: %v", err)
		}

		if len(rows) != i+2 {
			t.Errorf("expected i+2 rows, got %d", len(rows))
		}
	}
}

func Test_Transaction_Failure(t *testing.T) {
	conn, err := createTestDB()
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	defer conn.ReleaseLock()

	readOnly, err := conn.CopyReadOnly()
	if err != nil {
		t.Fatal(err)
	}

	sp, err := conn.Begin()
	if err != nil {
		t.Fatal(err)
	}

	// insert another row
	err = conn.ExecuteNamed(insertTestRow, map[string]interface{}{
		"$name":   "test1",
		"$id":     1,
		"@caller": "test", // testing flag override
	}, nil)
	if err != nil {
		t.Errorf("failed to insert: %v", err)
	}

	// insert another row
	err = conn.ExecuteNamed(insertTestRow, map[string]interface{}{
		"$name":   "test2",
		"$id":     2,
		"@caller": "test", // testing flag override
	}, nil)
	if err != nil {
		t.Errorf("failed to insert: %v", err)
	}

	// insert a row that should fail
	err = conn.ExecuteNamed(insertTestRow, map[string]interface{}{
		"$name":   "test2",
		"$id":     "nfcjkde",
		"@caller": "test", // testing flag override
	}, nil)
	if err == nil {
		t.Errorf("expected error, got nil")
	}

	// read back the rows
	type Row struct {
		id   int64
		name string
	}

	rows := []Row{}
	err = readOnly.Query("SELECT id, name FROM test_table;", func(stmt *driver.Statement) error {
		var row Row
		row.id = stmt.GetInt64("id")
		row.name = stmt.GetText("name")

		rows = append(rows, row)
		return nil
	})
	if err != nil {
		t.Errorf("failed to query: %v", err)
	}

	if len(rows) != 0 {
		t.Errorf("expected 0 rows, got %d", len(rows))
	}

	// read from the original connection
	rows = []Row{}
	err = conn.Query("SELECT id, name FROM test_table;", func(stmt *driver.Statement) error {
		var row Row
		row.id = stmt.GetInt64("id")
		row.name = stmt.GetText("name")

		rows = append(rows, row)
		return nil
	})
	if err != nil {
		t.Errorf("failed to query: %v", err)
	}

	if len(rows) != 2 {
		t.Errorf("expected 2 rows, got %d", len(rows))
	}

	err = sp.Rollback()
	if err != nil {
		t.Errorf("failed to rollback: %v", err)
	}

	rows = []Row{}
	err = conn.Query("SELECT id, name FROM test_table;", func(stmt *driver.Statement) error {
		var row Row
		row.id = stmt.GetInt64("id")
		row.name = stmt.GetText("name")

		rows = append(rows, row)
		return nil
	})
	if err != nil {
		t.Errorf("failed to query: %v", err)
	}

	if len(rows) != 0 {
		t.Errorf("expected 2 rows, got %d", len(rows))
	}
}

func Test_Nested_Savepoints(t *testing.T) {
	conn, err := createTestDB()
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	defer conn.ReleaseLock()

	sp1, err := conn.Begin()
	if err != nil {
		t.Fatal(err)
	}

	sp2, err := conn.Begin()
	if err != nil {
		t.Fatal(err)
	}

	err = sp2.Rollback()
	if err != nil {
		t.Fatal(err)
	}

	if conn.AutocommitEnabled() {
		t.Errorf("expected autocommit to be disabled since ther is an active savepoint")
	}

	err = sp1.Rollback()
	if err != nil {
		t.Fatal(err)
	}

	if !conn.AutocommitEnabled() {
		t.Errorf("expected autocommit to be enabled since there are no active savepoints")
	}
}
