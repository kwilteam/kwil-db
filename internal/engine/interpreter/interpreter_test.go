package interpreter_test

import (
	"context"
	"testing"

	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/internal/engine/interpreter"
	"github.com/stretchr/testify/require"
)

const defaultCaller = "owner"

type execAction struct {
	namespace string
	action    string
	values    []any
}

func Test_SQL(t *testing.T) {
	type testcase struct {
		name string // name of the test
		// array of sql statements, first element is the namespace, second is the sql statement
		// they can begin with {namespace}sql, or just sql
		sql         []string
		execSQL     string  // sql to return the results. Either this or execAction must be set
		results     [][]any // table of results
		err         error   // expected error, can be nil. Errors _MUST_ occur on the exec. This is a direct match
		errContains string  // expected error message, can be empty. Errors _MUST_ occur on the exec. This is a substring match
	}

	createUsersTable := `
CREATE TABLE users (
	id INT PRIMARY KEY,
	name TEXT,
	age INT
);
	`

	createPostsTable := `
CREATE TABLE posts (
	id INT PRIMARY KEY,
	owner_id INT NOT NULL REFERENCES users(id),
	content TEXT,
	created_at INT
);
	`

	tests := []testcase{
		{
			name: "insert and select",
			sql: []string{
				"INSERT INTO users (id, name, age) VALUES (1, 'Alice', 30);",
			},
			execSQL: "SELECT name, age FROM users;",
			results: [][]any{
				{"Alice", int64(30)},
			},
		},
		{
			name: "create namespace, add table, add record, alter table, select",
			sql: []string{
				"CREATE NAMESPACE test;",
				"{test}CREATE TABLE users (id INT PRIMARY KEY, name TEXT, age INT);",
				"{test}INSERT INTO users (id, name, age) VALUES (1, 'Bob', 30);",
				"{test}ALTER TABLE users DROP COLUMN age;",
			},
			execSQL: "{test}SELECT * FROM users;",
			results: [][]any{
				{int64(1), "Bob"},
			},
		},
		{
			name: "foreign key across namespaces",
			sql: []string{
				"CREATE NAMESPACE test1;",
				"CREATE NAMESPACE test2;",
				"{test1}CREATE TABLE users (id INT PRIMARY KEY, name TEXT);",
				`{test2}CREATE TABLE posts (id INT PRIMARY KEY,
				owner_id INT NOT NULL REFERENCES test1.users(id) ON UPDATE CASCADE ON DELETE CASCADE,
				content TEXT, created_at INT);`,
				"{test1}INSERT INTO users (id, name) VALUES (1, 'Alice'), (2, 'Bob');",
				"{test2}INSERT INTO posts (id, owner_id, content, created_at) VALUES (1, 1, 'Hello', @height), (2, 2, 'World', @height);",
				"{test1}DELETE FROM users WHERE id = 1;",
			},
			execSQL: `{test2}SELECT * FROM posts;`,
			results: [][]any{
				{int64(2), int64(2), "World", int64(1)},
			},
		},
		{
			name: "update and delete",
			sql: []string{
				"INSERT INTO users (id, name, age) VALUES (1, 'Alice', 30), (2, 'Bob', 40);",
				"UPDATE users SET age = 50 WHERE name = 'Alice';",
				"DELETE FROM users WHERE age = 40;",
			},
			execSQL: "SELECT name, age FROM users;",
			results: [][]any{
				{"Alice", int64(50)},
			},
		},
		{
			name: "recursive common table expression",
			execSQL: `
			with recursive r as (
				select 1 as n
				union all
				select n+1 from r where n < 6
			)
			select * from r;
			`,
			results: [][]any{
				{int64(1)}, {int64(2)}, {int64(3)}, {int64(4)}, {int64(5)}, {int64(6)},
			},
		},
		{
			name: "alter table add column",
			sql: []string{
				"ALTER TABLE users ADD COLUMN email TEXT;",
				"INSERT INTO users (id, name, age, email) VALUES (1, 'Alice', 30, 'alice@kwil.com');",
			},
			execSQL: "SELECT name, age, email FROM users;",
			results: [][]any{
				{"Alice", int64(30), "alice@kwil.com"},
			},
		},
		{
			name: "alter table drop column",
			sql: []string{
				"INSERT INTO users (id, name, age) VALUES (1, 'Alice', 30);",
				"ALTER TABLE users DROP COLUMN age;",
			},
			execSQL: "SELECT * FROM users;",
			results: [][]any{
				{1, "Alice"},
			},
		},

		// Setting a column to be NOT NULL
		{
			name: "alter table set column not null",
			sql: []string{
				"ALTER TABLE users ALTER COLUMN name SET NOT NULL;",
			},
			execSQL:     "INSERT INTO users (id, name, age) VALUES (1, null, 30);",
			errContains: "violates not-null constraint (SQLSTATE 23502)",
		},

		// Setting a default on a column
		{
			name: "alter table set column default",
			sql: []string{
				"ALTER TABLE users ALTER COLUMN age SET DEFAULT 25;",
				"INSERT INTO users (id, name) VALUES (1, 'Alice');",
			},
			execSQL: "SELECT id, name, age FROM users;",
			results: [][]any{
				{int64(1), "Alice", int64(25)},
			},
		},

		// Removing a default from a column
		{
			name: "alter table drop column default",
			sql: []string{
				"ALTER TABLE users ALTER COLUMN age SET DEFAULT 25;",
				"ALTER TABLE users ALTER COLUMN age DROP DEFAULT;",
				"INSERT INTO users (id, name) VALUES (1, 'Alice');",
			},
			execSQL: "SELECT id, name, age FROM users;",
			results: [][]any{
				{int64(1), "Alice", nil}, // Age will be NULL since the default is removed
			},
		},

		// Removing NOT NULL from a column
		{
			name: "alter table drop column not null",
			sql: []string{
				"ALTER TABLE users ALTER COLUMN name SET NOT NULL;",
				"ALTER TABLE users ALTER COLUMN name DROP NOT NULL;",
				"INSERT INTO users (id, age) VALUES (1, 30);",
			},
			execSQL: "SELECT id, name, age FROM users;",
			results: [][]any{
				{int64(1), nil, int64(30)},
			},
		},

		// Renaming a column
		{
			name: "alter table rename column",
			sql: []string{
				"ALTER TABLE users RENAME COLUMN name TO full_name;",
				"INSERT INTO users (id, full_name, age) VALUES (1, 'Alice', 30);",
			},
			execSQL: "SELECT full_name, age FROM users;",
			results: [][]any{
				{"Alice", int64(30)},
			},
		},

		// Renaming a table
		{
			name: "alter table rename table",
			sql: []string{
				"ALTER TABLE users RENAME TO app_users;",
				"INSERT INTO app_users (id, name, age) VALUES (1, 'Alice', 30);",
			},
			execSQL: "SELECT name, age FROM app_users;",
			results: [][]any{
				{"Alice", int64(30)},
			},
		},
	}

	// we can set up the interpreter and db here and just create and rollback a tx for each test case.
	// This drastically reduces the time taken to run the tests
	ctx := context.Background()
	interp, db, cleanup, err := interpreter.NewTestInterpeter(t)
	require.NoError(t, err)
	defer cleanup()

	err = interp.SetOwner(ctx, db, defaultCaller)
	require.NoError(t, err)

	err = interp.Execute(newTxCtx(), db, createUsersTable, nil)
	require.NoError(t, err)

	err = interp.Execute(newTxCtx(), db, createPostsTable, nil)
	require.NoError(t, err)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tx, err := db.BeginTx(ctx)
			require.NoError(t, err)
			defer tx.Rollback(ctx)
			var values [][]any

			// I wrap this in a function so that I can
			for _, sql := range test.sql {
				err = interp.Execute(newTxCtx(), tx, sql, func(v []interpreter.Value) error {
					row := make([]any, len(v))
					for i, val := range v {
						row[i] = val.RawValue()
					}
					values = append(values, row)
					return nil
				})
				require.NoError(t, err)
			}

			if test.execSQL != "" {
				err = interp.Execute(newTxCtx(), tx, test.execSQL, func(v []interpreter.Value) error {
					row := make([]any, len(v))
					for i, val := range v {
						row[i] = val.RawValue()
					}
					values = append(values, row)
					return nil
				})
				if test.err != nil {
					require.Error(t, err)
					require.ErrorIs(t, err, test.err)
				} else if test.errContains != "" {
					require.Contains(t, err.Error(), test.errContains)
				} else {
					require.NoError(t, err)
				}
			}

			require.Equal(t, len(test.results), len(values))
			for i, row := range values {
				require.Equal(t, len(test.results[i]), len(row))
				for j, val := range row {
					require.EqualValues(t, test.results[i][j], val)
				}
			}
		})
	}
}

func newTxCtx() *common.TxContext {
	return &common.TxContext{
		Ctx: context.Background(),
		BlockContext: &common.BlockContext{
			Height: 1,
			ChainContext: &common.ChainContext{
				NetworkParameters: &common.NetworkParameters{},
				MigrationParams:   &common.MigrationContext{},
			},
		},
		Caller:        defaultCaller,
		Signer:        []byte(defaultCaller),
		Authenticator: "test_authenticator",
	}
}
