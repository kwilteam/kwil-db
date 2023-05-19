package sqlite_test

import (
	"context"
	"testing"

	"github.com/kwilteam/kwil-db/pkg/engine2/dto/mocks"
	dbi "github.com/kwilteam/kwil-db/pkg/engine2/sqldb"
	"github.com/kwilteam/kwil-db/pkg/engine2/sqldb/sqlite"
	"github.com/stretchr/testify/assert"
)

func Test_WriteAndRead(t *testing.T) {
	db, cleanup := openTestDB(t)
	defer cleanup()

	err := createTestSchema(db)
	if err != nil {
		t.Fatal(err)
	}

	stmt, err := createTestAction(db)
	if err != nil {
		t.Fatal(err)
	}

	insertTestUser(t, stmt, &user{
		id:   1,
		name: "foo",
		age:  20,
	})
	insertTestUser(t, stmt, &user{
		id:   2,
		name: "bar",
		age:  30,
	})

	result, err := db.Query(context.Background(), "SELECT * FROM users;", nil)
	if err != nil {
		t.Fatal(err)
	}

	records := result.Records()

	if len(records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(records))
	}
}

func Test_MetadataTracking(t *testing.T) {
	db, cleanup := openTestDB(t)
	defer cleanup()

	err := createTestSchema(db)
	if err != nil {
		t.Fatal(err)
	}

	_, err = createTestAction(db)
	if err != nil {
		t.Fatal(err)
	}

	actions, err := db.ListActions(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, mocks.ActionInsertUser, actions[0])

	tables, err := db.ListTables(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if len(tables) != 1 {
		t.Fatalf("expected 1 table, got %d", len(tables))
	}

	assert.Equal(t, mocks.TableUsers, tables[0])

}

func openTestDB(t *testing.T) (dbi.DB, func()) {

	db, err := sqlite.NewSqliteStore("testdb")
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

func createTestSchema(db dbi.DB) error {
	ctx := context.Background()

	err := db.CreateTable(ctx, mocks.TableUsers)
	if err != nil {
		return err
	}

	return nil
}

func createTestAction(db dbi.DB) (dbi.Statement, error) {

	ctx := context.Background()
	err := db.StoreAction(ctx, mocks.ActionInsertUser)
	if err != nil {
		return nil, err
	}

	stmt, err := db.Prepare(mocks.ActionInsertUser.Statements[0])
	if err != nil {
		return nil, err
	}

	return stmt, nil
}

func insertTestUser(t *testing.T, stmt dbi.Statement, usr *user) {
	_, err := stmt.Execute(map[string]any{
		"$id":   usr.id,
		"$name": usr.name,
		"$age":  usr.age,
	})
	if err != nil {
		t.Fatal(err)
	}
}

type user struct {
	id   int64
	name string
	age  int64
}
