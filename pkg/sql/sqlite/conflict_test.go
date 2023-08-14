package sqlite_test

import (
	"bytes"
	"context"
	"fmt"
	"testing"
)

func Test_ChangesetConflict_Insert(t *testing.T) {
	db, td := openRealDB()
	defer td()
	ctx := context.Background()

	sp, err := db.Savepoint()
	if err != nil {
		t.Fatal(err)
	}

	ses, err := db.CreateSession()
	if err != nil {
		t.Fatal(err)
	}
	defer ses.Delete()

	// insert a row
	err = db.Execute("INSERT INTO users (id, name, age) VALUES ($id, $name, $age)", map[string]interface{}{
		"$id":   1,
		"$name": "John",
		"$age":  20,
	})
	if err != nil {
		t.Fatal(err)
	}

	// generate changeset
	cs, err := ses.GenerateChangeset()
	if err != nil {
		t.Fatal(err)
	}

	csBytes := cs.Export()
	err = cs.Close()
	if err != nil {
		t.Fatal(err)
	}

	err = sp.Rollback()
	if err != nil {
		t.Fatal(err)
	}

	results, err := db.Query(ctx, "SELECT * FROM users")
	if err != nil {
		t.Fatal(err)
	}
	defer results.Finish()

	records, err := results.ExportRecords()
	if err != nil {
		t.Fatal(err)
	}

	if err != nil {
		t.Fatal(err)
	}

	if len(records) != 0 {
		t.Fatal("expected 0 rows")
	}

	// apply changeset
	err = db.ApplyChangeset(bytes.NewBuffer(csBytes))
	if err != nil {
		t.Fatal(err)
	}

	// check that the row is there
	results, err = db.Query(ctx, "SELECT * FROM users")
	if err != nil {
		t.Fatal(err)
	}

	records, err = results.ExportRecords()
	if err != nil {
		t.Fatal(err)
	}

	if len(records) != 1 {
		t.Fatal("expected 1 row")
	}

	// Reapply changeset
	err = db.ApplyChangeset(bytes.NewBuffer(csBytes))
	if err != nil {
		fmt.Println("Failed while reapplying changset: ", err)
		t.Fatal(err)
	}

	// check that the row is there
	results, err = db.Query(ctx, "SELECT * FROM users")
	if err != nil {
		t.Fatal(err)
	}

	records, err = results.ExportRecords()
	if err != nil {
		t.Fatal(err)
	}

	if len(records) != 1 {
		t.Fatal("expected 1 row")
	}
}

func Test_ChangesetConflict_Update(t *testing.T) {
	db, td := openRealDB()
	defer td()
	ctx := context.Background()

	err := db.Execute("INSERT INTO users (id, name, age) VALUES ($id, $name, $age)", map[string]interface{}{
		"$id":   1,
		"$name": "John",
		"$age":  20,
	})
	if err != nil {
		t.Fatal(err)
	}

	sp, err := db.Savepoint()
	if err != nil {
		t.Fatal(err)
	}

	ses, err := db.CreateSession()
	if err != nil {
		t.Fatal(err)
	}
	defer ses.Delete()

	// insert a row
	err = db.Execute("UPDATE users SET name = $name WHERE id = $id", map[string]interface{}{
		"$id":   1,
		"$name": "Jane",
	})
	if err != nil {
		t.Fatal(err)
	}

	// generate changeset
	cs, err := ses.GenerateChangeset()
	if err != nil {
		t.Fatal(err)
	}

	csBytes := cs.Export()
	err = cs.Close()
	if err != nil {
		t.Fatal(err)
	}

	err = sp.Rollback()
	if err != nil {
		t.Fatal(err)
	}

	results, err := db.Query(ctx, "SELECT * FROM users")
	if err != nil {
		t.Fatal(err)
	}
	defer results.Finish()

	records, err := results.ExportRecords()
	if err != nil {
		t.Fatal(err)
	}

	if err != nil {
		t.Fatal(err)
	}

	if len(records) != 1 {
		t.Fatal("expected 0 rows")
	}

	// apply changeset
	err = db.ApplyChangeset(bytes.NewBuffer(csBytes))
	if err != nil {
		t.Fatal(err)
	}

	// check that the row is there
	results, err = db.Query(ctx, "SELECT * FROM users")
	if err != nil {
		t.Fatal(err)
	}

	records, err = results.ExportRecords()
	if err != nil {
		t.Fatal(err)
	}

	if len(records) != 1 {
		t.Fatal("expected 1 row")
	}

	// Reapply changeset
	err = db.ApplyChangeset(bytes.NewBuffer(csBytes))
	if err != nil {
		fmt.Println("Failed while reapplying changset: ", err)
		t.Fatal(err)
	}

	// check that the row is there
	results, err = db.Query(ctx, "SELECT * FROM users")
	if err != nil {
		t.Fatal(err)
	}

	records, err = results.ExportRecords()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(records)
	if len(records) != 1 {
		t.Fatal("expected 1 row")
	}
}

func Test_ChangesetConflict_Delete(t *testing.T) {
	db, td := openRealDB()
	defer td()
	ctx := context.Background()

	err := db.Execute("INSERT INTO users (id, name, age) VALUES ($id, $name, $age)", map[string]interface{}{
		"$id":   1,
		"$name": "John",
		"$age":  20,
	})
	if err != nil {
		t.Fatal(err)
	}

	sp, err := db.Savepoint()
	if err != nil {
		t.Fatal(err)
	}

	ses, err := db.CreateSession()
	if err != nil {
		t.Fatal(err)
	}
	defer ses.Delete()

	// delete a row
	err = db.Execute("DELETE FROM users WHERE id = $id", map[string]interface{}{
		"$id": 1,
	})
	if err != nil {
		t.Fatal(err)
	}

	// generate changeset
	cs, err := ses.GenerateChangeset()
	if err != nil {
		t.Fatal(err)
	}

	csBytes := cs.Export()
	err = cs.Close()
	if err != nil {
		t.Fatal(err)
	}

	err = sp.Rollback()
	if err != nil {
		t.Fatal(err)
	}

	results, err := db.Query(ctx, "SELECT * FROM users")
	if err != nil {
		t.Fatal(err)
	}
	defer results.Finish()

	records, err := results.ExportRecords()
	if err != nil {
		t.Fatal(err)
	}

	if err != nil {
		t.Fatal(err)
	}

	if len(records) != 1 {
		t.Fatal("expected 1 rows")
	}

	// apply changeset
	err = db.ApplyChangeset(bytes.NewBuffer(csBytes))
	if err != nil {
		t.Fatal(err)
	}

	// check that the row is there
	results, err = db.Query(ctx, "SELECT * FROM users")
	if err != nil {
		t.Fatal(err)
	}

	records, err = results.ExportRecords()
	if err != nil {
		t.Fatal(err)
	}

	if len(records) != 0 {
		t.Fatal("expected 1 row")
	}

	// Reapply changeset
	err = db.ApplyChangeset(bytes.NewBuffer(csBytes))
	if err != nil {
		fmt.Println("Failed while reapplying changset: ", err)
		t.Fatal(err)
	}

	// check that the row is there
	results, err = db.Query(ctx, "SELECT * FROM users")
	if err != nil {
		t.Fatal(err)
	}

	records, err = results.ExportRecords()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(records)
	if len(records) != 0 {
		t.Fatal("expected 1 row")
	}
}

func Test_ChangesetConflict_Multiplerows(t *testing.T) {
	db, td := openRealDB()
	defer td()
	ctx := context.Background()

	err := db.Execute("INSERT INTO users (id, name, age) VALUES ($id, $name, $age)", map[string]interface{}{
		"$id":   1,
		"$name": "John",
		"$age":  20,
	})
	if err != nil {
		t.Fatal(err)
	}

	err = db.Execute("INSERT INTO users (id, name, age) VALUES ($id, $name, $age)", map[string]interface{}{
		"$id":   2,
		"$name": "Jen",
		"$age":  25,
	})
	if err != nil {
		t.Fatal(err)
	}

	sp, err := db.Savepoint()
	if err != nil {
		t.Fatal(err)
	}

	ses, err := db.CreateSession()
	if err != nil {
		t.Fatal(err)
	}
	defer ses.Delete()

	// Update row1 & Delete row2
	err = db.Execute("UPDATE users SET name = $name WHERE id = $id", map[string]interface{}{
		"$id":   1,
		"$name": "Jane",
	})
	if err != nil {
		t.Fatal(err)
	}

	err = db.Execute("DELETE FROM users WHERE id = $id", map[string]interface{}{
		"$id": 2,
	})
	if err != nil {
		t.Fatal(err)
	}

	// generate changeset
	cs, err := ses.GenerateChangeset()
	if err != nil {
		t.Fatal(err)
	}

	csBytes := cs.Export()
	err = cs.Close()
	if err != nil {
		t.Fatal(err)
	}

	err = sp.Rollback()
	if err != nil {
		t.Fatal(err)
	}

	results, err := db.Query(ctx, "SELECT * FROM users")
	if err != nil {
		t.Fatal(err)
	}
	defer results.Finish()

	records, err := results.ExportRecords()
	if err != nil {
		t.Fatal(err)
	}

	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(records)
	if len(records) != 2 {
		t.Fatal("expected 2 rows")
	}

	// apply changeset
	err = db.ApplyChangeset(bytes.NewBuffer(csBytes))
	if err != nil {
		t.Fatal(err)
	}

	// check that the row is there
	results, err = db.Query(ctx, "SELECT * FROM users")
	if err != nil {
		t.Fatal(err)
	}

	records, err = results.ExportRecords()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(records)
	if len(records) != 1 {
		t.Fatal("expected 1 row")
	}

	// Reapply changeset
	err = db.ApplyChangeset(bytes.NewBuffer(csBytes))
	if err != nil {
		fmt.Println("Failed while reapplying changset: ", err)
		t.Fatal(err)
	}

	// check that the row is there
	results, err = db.Query(ctx, "SELECT * FROM users")
	if err != nil {
		t.Fatal(err)
	}

	records, err = results.ExportRecords()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(records)
	if len(records) != 1 {
		t.Fatal("expected 1 row")
	}
}
