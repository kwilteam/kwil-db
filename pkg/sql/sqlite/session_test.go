package sqlite_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/kwilteam/kwil-db/pkg/sql/sqlite"
	"github.com/stretchr/testify/require"
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

// there are known issues with updating primary keys with foreign key cascades, and changesets
func Test_ChangesetUpdatePrimaryKey(t *testing.T) {
	db, td := openRealDB()
	defer td()
	ctx := context.Background()

	err := db.Execute("CREATE TABLE posts (id INTEGER PRIMARY KEY, title TEXT, content TEXT, user_id INTEGER NOT NULL, FOREIGN KEY (user_id) REFERENCES users (id) ON UPDATE CASCADE ON DELETE CASCADE) WITHOUT ROWID, STRICT;")
	require.NoError(t, err)

	// insert a row
	err = db.Execute("INSERT INTO users (id, name, age) VALUES ($id, $name, $age)", map[string]interface{}{
		"$id":   1,
		"$name": "John",
		"$age":  20,
	})
	require.NoError(t, err)

	// insert post
	err = db.Execute("INSERT INTO posts (id, title, content, user_id) VALUES ($id, $title, $content, $user_id)", map[string]interface{}{
		"$id":      1,
		"$title":   "Hello",
		"$content": "World",
		"$user_id": 1,
	})
	require.NoError(t, err)

	ses, err := db.CreateSession()
	require.NoError(t, err)

	sp, err := db.Savepoint()
	require.NoError(t, err)

	err = db.Execute("UPDATE users SET id = $id, name = $name, age = $age WHERE id = $oldId", map[string]interface{}{
		"$id":    2,
		"$oldId": 1,
		"$name":  "Johnny",
		"$age":   21,
	})
	require.NoError(t, err)

	// generate changeset
	cs, err := ses.GenerateChangeset()
	require.NoError(t, err)

	err = sp.Rollback()
	require.NoError(t, err)

	err = db.DisableForeignKey()
	require.NoError(t, err)

	sp2, err := db.Savepoint()
	require.NoError(t, err)

	// apply changeset
	err = db.ApplyChangeset(bytes.NewBuffer(cs.Export()))
	require.NoError(t, err)

	err = sp2.Commit()
	require.NoError(t, err)

	err = db.EnableForeignKey()
	require.NoError(t, err)

	// check that the row is there
	results, err := db.Query(ctx, "SELECT * FROM users")
	require.NoError(t, err)

	records, err := results.ExportRecords()
	require.NoError(t, err)

	require.Equal(t, 1, len(records))
	require.Equal(t, int64(2), records[0]["id"].(int64))

	// check that the post is there
	results, err = db.Query(ctx, "SELECT * FROM posts")
	require.NoError(t, err)

	records, err = results.ExportRecords()
	require.NoError(t, err)

	require.Equal(t, 1, len(records))
	require.Equal(t, int64(2), records[0]["user_id"].(int64))
}
