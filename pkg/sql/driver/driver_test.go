package driver_test

import (
	"fmt"
	"kwil/pkg/sql/driver"
	"sync"
	"testing"
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
		"$name": "test2",
		"$id":   2,
	})
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
	})
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

	// test delete
	err = conn.Execute(deleteTestRow, 1)
	if err != nil {
		t.Errorf("failed to delete: %v", err)
	}

	// test delete named
	err = conn.ExecuteNamed(deleteTestRow, map[string]interface{}{
		"$id": 2,
	})
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
	conn, err := createTestDB()
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	defer conn.ReleaseLock()

	sp, err := conn.Savepoint()
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

	// test savepoint with commit
	sp, err = conn.Savepoint()
	if err != nil {
		t.Fatal(err)
	}

	err = sp.Execute(insertTestRow, 1, "test1")
	if err != nil {
		t.Errorf("failed to insert: %v", err)
	}

	// generate changeset
	changeset, err := sp.GetChangeset()
	if err != nil {
		t.Fatal(err)
	}

	// now commit
	err = sp.Commit()
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

	// apply changeset
	err = sp.ApplyChangeset(changeset)
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
}

func createTestDB() (*driver.Connection, error) {
	conn, err := driver.OpenConn("test_database_NOBODYBETTERGUESSTHISDBID") // added random string to avoid collisions
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
