package sqlite_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/kwilteam/kwil-db/pkg/sql/sqlite"
)

func Test_ForeignKey(t *testing.T) {
	db, td := openRealDB()
	defer td()

	err := db.Execute(createTblFk1)
	if err != nil {
		t.Error(err)
	}

	err = db.Execute(createTblFk2)
	if err != nil {
		t.Error(err)
	}

	// insert user
	insertUserStmt, err := db.Prepare("INSERT INTO wallets (id, username) VALUES ($id, $username)")
	if err != nil {
		t.Error(err)
	}

	testUsername := "test_user"

	err = insertUserStmt.Execute(sqlite.WithNamedArgs(map[string]interface{}{
		"$id":       1,
		"$username": testUsername,
	}))
	if err != nil {
		t.Error(err)
	}

	// insert post
	insertPostStmt, err := db.Prepare("INSERT INTO posts (id, user_id, title) VALUES ($id, $user_id, $title)")
	if err != nil {
		t.Error(err)
	}

	err = insertPostStmt.Execute(sqlite.WithNamedArgs(map[string]interface{}{
		"$id":      1,
		"$user_id": 1,
		"$title":   "test_post",
	}))
	if err != nil {
		t.Error(err)
	}

	// update user
	updateUserStmt, err := db.Prepare("UPDATE wallets SET id = $id WHERE username = $username")
	if err != nil {
		t.Error(err)
	}

	err = updateUserStmt.Execute(sqlite.WithNamedArgs(map[string]interface{}{
		"$id":       2,
		"$username": testUsername,
	}))
	if err != nil {
		t.Error(err)
	}

	ctx := context.Background()
	var resultSet sqlite.ResultSet
	err = db.Query(ctx, "SELECT * FROM posts", sqlite.WithResultSet(&resultSet))
	if err != nil {
		t.Error(err)
	}

	res := resultSet.Records()
	firstPost := res[0]
	if fmt.Sprint(firstPost["user_id"]) != fmt.Sprint(2) {
		t.Errorf("expected user_id to be 2, got %v", firstPost["user_id"])
	}

	// now delete user
	deleteUserStmt, err := db.Prepare("DELETE FROM wallets")
	if err != nil {
		t.Error(err)
	}

	err = deleteUserStmt.Execute(sqlite.WithNamedArgs(map[string]interface{}{
		"$username": testUsername,
	}))
	if err != nil {
		t.Error(err)
	}
	// check that user is deleted
	var userResults sqlite.ResultSet
	err = db.Query(ctx, "SELECT * FROM wallets", sqlite.WithResultSet(&userResults))
	if err != nil {
		t.Error(err)
	}

	res = userResults.Records()
	if len(res) != 0 {
		t.Errorf("expected 0 user, got %d", len(res))
	}

	// check if post is deleted
	var postResults sqlite.ResultSet
	err = db.Query(ctx, "SELECT * FROM posts", sqlite.WithResultSet(&resultSet))
	if err != nil {
		t.Error(err)
	}

	res = postResults.Records()
	if len(res) != 0 {
		t.Errorf("expected 0 post, got %d", len(res))
	}
}

const (
	createTblFk1 = ` CREATE TABLE IF NOT EXISTS wallets (
		id INTEGER PRIMARY KEY,
		username TEXT NOT NULL
	) WITHOUT ROWID, STRICT;`
	createTblFk2 = ` CREATE TABLE IF NOT EXISTS posts (
		id INTEGER PRIMARY KEY,
		user_id INTEGER NOT NULL,
		title TEXT NOT NULL,
		FOREIGN KEY (user_id) REFERENCES wallets(id) ON UPDATE CASCADE ON DELETE CASCADE
	) WITHOUT ROWID, STRICT;`
)
