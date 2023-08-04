package sqlite_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/kwilteam/kwil-db/pkg/sql/sqlite"
)

// This test tests that changesets can be generated, applied, and inverted.
func Test_ChangesetApply(t *testing.T) {
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
}

func Test_ChangesetIgnoresIntermediateOperations(t *testing.T) {
	db, td := openRealDB()
	defer td()

	ses, err := db.CreateSession()
	if err != nil {
		t.Fatal(err)
	}
	defer ses.Delete()

	// insert some rows
	err = insertUsers(db, []*user{
		{
			id:   1,
			name: "John",
			age:  20,
		},
		{
			id:   2,
			name: "Jane",
			age:  21,
		},
		{
			id:   3,
			name: "Jack",
			age:  22,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	// update one
	err = db.Execute("UPDATE users SET name = $name WHERE id = $id", map[string]interface{}{
		"$name": "Jill",
		"$id":   2,
	})
	if err != nil {
		t.Fatal(err)
	}

	// delete one
	err = db.Execute("DELETE FROM users WHERE id = $id", map[string]interface{}{
		"$id": 3,
	})
	if err != nil {
		t.Fatal(err)
	}

	// generate changeset
	cs, err := ses.GenerateChangeset()
	if err != nil {
		t.Fatal(err)
	}
	defer cs.Close()

	// changeset should have two inserts containing John and Jill
	counter := 0
	for {
		rowReturned, err := cs.Next()
		if err != nil {
			t.Fatal(err)
		}

		if !rowReturned {
			break
		}

		op, err := cs.Operation()
		if err != nil {
			t.Fatal(err)
		}

		if op.Type != sqlite.OpInsert {
			t.Fatal("expected insert")
		}

		if op.TableName != "users" {
			t.Fatal("expected users")
		}

		if op.NumColumns != 3 {
			t.Fatal("expected 3 columns")
		}

		if op.Indirect {
			t.Fatal("expected direct")
		}

		if counter == 0 {
			// should be John
			name, err := cs.New(1)
			if err != nil {
				t.Fatal(err)
			}

			if name.Text() != "John" {
				t.Fatal("expected John")
			}
		}

		if counter == 1 {
			// should be Jill
			name, err := cs.New(1)
			if err != nil {
				t.Fatal(err)
			}

			if name.Text() != "Jill" {
				t.Fatal("expected Jill")
			}
		}

		counter++
	}
}
func Test_DELETEME(t *testing.T) {
	db, td := openRealDB()
	defer td()

	// insert some rows
	err := insertUsers(db, []*user{
		{
			id:   1,
			name: "John",
			age:  20,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	ses, err := db.CreateSession()
	if err != nil {
		t.Fatal(err)
	}
	defer ses.Delete()

	// update one
	err = db.Execute("UPDATE users SET name = $name, age = $age WHERE id = $id", map[string]interface{}{
		"$age":  21,
		"$name": "Jill",
		"$id":   1,
	})
	if err != nil {
		t.Fatal(err)
	}

	cs, err := ses.GenerateChangeset()
	if err != nil {
		t.Fatal(err)
	}
	defer cs.Close()

	for {
		rowReturned, err := cs.Next()
		if err != nil {
			t.Fatal(err)
		}

		if !rowReturned {
			break
		}

		val1, _ := cs.New(1)
		val2, _ := cs.New(2)
		val3, _ := cs.New(3)

		if !val1.Changed() {
			t.Errorf("expected val1 to be changed")
		}
		if !val2.Changed() {
			t.Errorf("expected val2 to be changed")
		}
		if val3.Changed() {
			t.Errorf("expected val3 to not be changed")
		}
	}
}

type user struct {
	id   int
	name string
	age  int
}

func insertUsers(c *sqlite.Connection, newUsers []*user) error {
	for _, newUser := range newUsers {
		err := c.Execute("INSERT INTO users (id, name, age) VALUES ($id, $name, $age)", map[string]interface{}{
			"$id":   newUser.id,
			"$name": newUser.name,
			"$age":  newUser.age,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// This test tests that sessions operations persist across connection threads even if they are deleted.
func Test_SessionPersistence(t *testing.T) {
	db, td := openRealDB()
	defer td()
	ctx := context.Background()

	ses, err := db.CreateSession()
	if err != nil {
		t.Fatal(err)
	}

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

	err = cs.Close()
	if err != nil {
		t.Fatal(err)
	}

	// delete the session
	err = ses.Delete()
	if err != nil {
		t.Fatal(err)
	}

	// check that the row is still there
	results, err := db.Query(ctx, "SELECT COUNT(*) FROM users")
	if err != nil {
		t.Fatal(err)
	}

	records, err := results.ExportRecords()
	if err != nil {
		t.Fatal(err)
	}

	if len(records) != 1 {
		t.Fatal("expected 1 row")
	}
}
