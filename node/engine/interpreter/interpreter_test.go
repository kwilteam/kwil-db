//go:build pglive

package interpreter_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math"
	"reflect"
	"testing"

	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/decimal"
	"github.com/kwilteam/kwil-db/extensions/precompiles"
	"github.com/kwilteam/kwil-db/node/engine/interpreter"
	"github.com/kwilteam/kwil-db/node/pg"
	"github.com/kwilteam/kwil-db/node/types/sql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	defaultCaller    = "owner"
	createUsersTable = `
CREATE TABLE IF NOT EXISTS users (
	id INT PRIMARY KEY,
	name TEXT,
	age INT
);
	`

	createPostsTable = `
CREATE TABLE IF NOT EXISTS posts (
	id INT PRIMARY KEY,
	owner_id INT NOT NULL REFERENCES users(id),
	content TEXT,
	created_at INT
);`
)

func Test_SQL(t *testing.T) {
	type testcase struct {
		name string // name of the test
		// array of sql statements, first element is the namespace, second is the sql statement
		// they can begin with {namespace}sql, or just sql
		sql         []string
		execSQL     string         // sql to return the results. Either this or execAction must be set
		execVars    map[string]any // variables to pass to the execSQL
		results     [][]any        // table of results
		err         error          // expected error, can be nil. Errors _MUST_ occur on the exec. This is a direct match
		errContains string         // expected error message, can be empty. Errors _MUST_ occur on the exec. This is a substring match
		caller      string         // caller to use for the action, will default to defaultCaller
	}

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
		{
			name:    "drop default namespace",
			execSQL: "DROP NAMESPACE main;",
			err:     interpreter.ErrCannotDropBuiltinNamespace,
		},
		{
			name:    "drop info namespace",
			execSQL: "DROP NAMESPACE info;",
			err:     interpreter.ErrCannotDropBuiltinNamespace,
		},
		{
			name:    "drop non-existent namespace",
			execSQL: "DROP NAMESPACE some_ns;",
			err:     interpreter.ErrNamespaceNotFound,
		},
		{
			name: "global select permission - failure",
			sql: []string{
				// test_role starts with select b/c of default, but then loses it.
				"CREATE ROLE test_role;",
				"REVOKE select FROM default;",
				"GRANT test_role TO 'user'",
			},
			execSQL:     "SELECT * FROM users;",
			errContains: "permission denied for table users",
			caller:      "user",
			err:         interpreter.ErrDoesNotHavePriv,
		},
		{
			name: "global select permission - success",
			sql: []string{
				"CREATE ROLE test_role;",
				"REVOKE select FROM default;",
				"GRANT test_role TO 'user'",
				"GRANT select TO test_role;",
			},
			execSQL: "SELECT * FROM users;",
			results: [][]any{},
			caller:  "user",
		},
		{
			name: "namespaced select permission - failure",
			sql: []string{
				"CREATE NAMESPACE test;",
				"CREATE ROLE test_role;",
				"REVOKE select FROM default;",
				"GRANT test_role TO 'user'",
				"GRANT select ON test TO test_role;",
			},
			// they do not have permission to select from the users table, which is in the main namespace
			execSQL: "SELECT * FROM users;",
			err:     interpreter.ErrDoesNotHavePriv,
			caller:  "user",
		},
		{
			name: "namespaced select permission - success",
			sql: []string{
				"CREATE NAMESPACE test;",
				"CREATE ROLE test_role;",
				"REVOKE select FROM default;",
				"GRANT test_role TO 'user'",
				"GRANT select ON test TO test_role;",
				"{test}CREATE TABLE users (id INT PRIMARY KEY, name TEXT, age INT);",
			},
			execSQL: "{test}SELECT * FROM users;",
			results: [][]any{},
			caller:  "user",
		},
		{
			name:    "global insert permission - failure",
			execSQL: "INSERT INTO users (id, name, age) VALUES (1, 'Alice', 30);",
			err:     interpreter.ErrDoesNotHavePriv,
			caller:  "user",
		},
		{
			name: "global insert permission - success",
			sql: []string{
				"CREATE ROLE test_role;",
				"GRANT test_role TO 'user'",
				"GRANT insert TO test_role;",
			},
			execSQL: "INSERT INTO users (id, name, age) VALUES (1, 'Alice', 30);",
			results: [][]any{},
			caller:  "user",
		},
		// I wont test other namespace-able perms because they use the same logic
		{
			name: "drop role",
			sql: []string{
				"CREATE ROLE test_role;",
				"GRANT test_role TO 'user';",
				"GRANT iNSert TO test_role;",
				"DROP ROLE test_role;",
			},
			execSQL: "INSERT INTO users (id, name, age) VALUES (1, 'Alice', 30);",
			err:     interpreter.ErrDoesNotHavePriv,
			caller:  "user",
		},
		{
			name: "drop and recreate namespace",
			sql: []string{
				"CREATE NAMESPACE test;",
				"CREATE ROLE test_role;",
				"GRANT INSERT ON test TO test_role;",
				"GRANT test_role TO 'user';",
				"DROP NAMESPACE test;",
				"CREATE NAMESPACE test;",
			},
			execSQL: "{test}INSERT INTO users (id, name, age) VALUES (1, 'Alice', 30);",
			err:     interpreter.ErrDoesNotHavePriv,
			caller:  "user",
		},
		{
			name: "revoking global permission",
			sql: []string{
				"CREATE ROLE test_role;",
				"GRANT test_role TO 'user';",
				"GRANT insert TO test_role;",
				"REVOKE insert FROM test_role;",
			},
			execSQL: "INSERT INTO users (id, name, age) VALUES (1, 'Alice', 30);",
			err:     interpreter.ErrDoesNotHavePriv,
			caller:  "user",
		},
		{
			name: "revoking namespaced permission",
			sql: []string{
				"CREATE NAMESPACE test;",
				"CREATE ROLE test_role;",
				"GRANT test_role TO 'user';",
				"GRANT insert ON test TO test_role;",
				"REVOKE insert ON test FROM test_role;",
				"{test}CREATE TABLE users (id INT PRIMARY KEY, name TEXT, age INT);",
			},
			execSQL: "{test}INSERT INTO users (id, name, age) VALUES (1, 'Alice', 30);",
			err:     interpreter.ErrDoesNotHavePriv,
			caller:  "user",
		},
		{
			name: "revoke role",
			sql: []string{
				"CREATE ROLE test_role;",
				"GRANT test_role TO 'user';",
				"GRANT insert TO test_role;",
				"REVOKE test_role FROM 'user';",
			},
			execSQL: "INSERT INTO users (id, name, age) VALUES (1, 'Alice', 30);",
			err:     interpreter.ErrDoesNotHavePriv,
			caller:  "user",
		},
		{
			name: "grant owner role to user",
			sql: []string{
				"GRANT owner TO 'user';",
			},
			execSQL: "GRANT owner TO 'user2';",
			caller:  "user",
		},
		{
			name:    "grant owner role to user, parameterized",
			execSQL: `grant owner to $user;`,
			execVars: map[string]any{
				"user": "new_user",
			},
		},
		{
			name: "role cannot grant ownership if not owner, even if they have the roles priv",
			sql: []string{
				"CREATE ROLE test_role;",
				"GRANT ROLES TO test_role;",
				"GRANT test_role TO 'user';",
			},
			execSQL: `grant owner to 'user2';`,
			err:     interpreter.ErrDoesNotHavePriv,
			caller:  "user",
		},
		{
			name: "owner can revoke another owner",
			sql: []string{
				"GRANT owner TO 'user';",
			},
			execSQL: "REVOKE owner FROM 'user';",
		},
		// here we test that privileges are correctly enforced.
		// We do this by relying on the default role, which has no privileges
		// except for select and call.
		{
			name:    "default role cannot be dropped",
			execSQL: "DROP ROLE default;",
			err:     interpreter.ErrBuiltInRole,
		},
		{
			name:    "default role cannot create tables",
			execSQL: `CREATE TABLE tbl (col int primary key);`,
			err:     interpreter.ErrDoesNotHavePriv,
			caller:  "default",
		},
		{
			name:    "default role cannot alter tables",
			execSQL: "ALTER TABLE users ADD COLUMN email TEXT;",
			err:     interpreter.ErrDoesNotHavePriv,
			caller:  "default",
		},
		{
			name:    "default role cannot drop tables",
			execSQL: "DROP TABLE users;",
			err:     interpreter.ErrDoesNotHavePriv,
			caller:  "default",
		},
		{
			name:    "default role cannot create roles",
			execSQL: `CREATE ROLE test_role;`,
			err:     interpreter.ErrDoesNotHavePriv,
			caller:  "default",
		},
		{
			name:    "default role cannot drop roles",
			execSQL: `DROP ROLE test_role;`,
			err:     interpreter.ErrDoesNotHavePriv,
			caller:  "default",
		},
		{
			name:    "default role cannot create namespaces",
			execSQL: `CREATE NAMESPACE ns;`,
			err:     interpreter.ErrDoesNotHavePriv,
			caller:  "default",
		},
		{
			name: "default role cannot drop namespaces",
			sql: []string{
				"CREATE NAMESPACE ns;",
			},
			execSQL: `DROP NAMESPACE ns;`,
			err:     interpreter.ErrDoesNotHavePriv,
			caller:  "default",
		},
		{
			name:    "default role cannot create actions",
			execSQL: `CREATE ACTION act() public {};`,
			err:     interpreter.ErrDoesNotHavePriv,
			caller:  "default",
		},
		{
			name: "default role cannot drop actions",
			sql: []string{
				"CREATE ACTION act() public {};",
			},
			execSQL: `DROP ACTION act;`,
			err:     interpreter.ErrDoesNotHavePriv,
			caller:  "default",
		},
		{
			name: "default role cannot assign roles",
			sql: []string{
				"CREATE ROLE test_role;",
			},
			execSQL: `GRANT test_role TO 'default_user';`,
			err:     interpreter.ErrDoesNotHavePriv,
			caller:  "default",
		},
		{
			name: "default role cannot revoke roles",
			sql: []string{
				"CREATE ROLE test_role;",
				"GRANT test_role TO 'default_user';",
			},
			execSQL: `REVOKE test_role FROM 'default_user';`,
			err:     interpreter.ErrDoesNotHavePriv,
			caller:  "default",
		},
		{
			name: "default role cannot assign privileges",
			sql: []string{
				"CREATE ROLE test_role;",
			},
			execSQL: `GRANT select ON users TO test_role;`,
			err:     interpreter.ErrDoesNotHavePriv,
			caller:  "default",
		},
		{
			name: "default role cannot revoke privileges",
			sql: []string{
				"CREATE ROLE test_role;",
				"GRANT select ON main TO test_role;",
			},
			execSQL: `REVOKE select ON main FROM test_role;`,
			err:     interpreter.ErrDoesNotHavePriv,
			caller:  "default",
		},
		{
			name:    "default role cannot use extensions",
			execSQL: "USE test AS test_ext;",
			err:     interpreter.ErrDoesNotHavePriv,
			caller:  "default",
		},
		{
			name:    "default role cannot unuse extensions",
			execSQL: "UNUSE test_ext;",
			err:     interpreter.ErrDoesNotHavePriv,
			caller:  "default",
		},
		{
			name:    "default role cannot insert",
			execSQL: "INSERT INTO users (id, name, age) VALUES (1, 'Alice', 30);",
			err:     interpreter.ErrDoesNotHavePriv,
			caller:  "default",
		},
		{
			name:    "default role cannot update",
			execSQL: "UPDATE users SET age = 50 WHERE name = 'Alice';",
			err:     interpreter.ErrDoesNotHavePriv,
			caller:  "default",
		},
		{
			name:    "default role cannot delete",
			execSQL: "DELETE FROM users WHERE age = 40;",
			err:     interpreter.ErrDoesNotHavePriv,
			caller:  "default",
		},
		// testing that the admin cannot perform disallowed operations
		// (e.g. dropping the info namespace)
		{
			name:    "admin cannot drop info namespace",
			execSQL: "DROP NAMESPACE info;",
			err:     interpreter.ErrCannotDropBuiltinNamespace,
		},
		{
			name:    "admin cannot drop main namespace",
			execSQL: `DROP NAMESPACE main;`,
			err:     interpreter.ErrCannotDropBuiltinNamespace,
		},
		{
			name:    "admin cannot add table to info namespace",
			execSQL: `{info}CREATE TABLE tbl (col int primary key);`,
			err:     interpreter.ErrCannotMutateInfoNamespace,
		},
		{
			name:    "admin cannot add action to info namespace",
			execSQL: `{info}CREATE ACTION act() public {};`,
			err:     interpreter.ErrCannotMutateInfoNamespace,
		},
		{
			name:    "admin cannot drop table from info namespace",
			execSQL: `{info}DROP TABLE namespaces;`,
			err:     interpreter.ErrCannotMutateInfoNamespace,
		},
		{
			// this should always fail because it is a postgres view, but I want to make sure
			// the error is correctly caught by the engine
			name:    "admin cannot insert into info namespace",
			execSQL: `{info}INSERT INTO namespaces (name, type) VALUES ('test', 'SYSTEM');`,
			err:     interpreter.ErrCannotMutateInfoNamespace,
		},
		// testing info schema
		// I only have one test here because sql_test.go tests all of the info schema,
		// this is simply to ensure that it can be accessed by the engine.
		{
			name:    "read namespaces",
			execSQL: "SELECT name, type FROM info.namespaces;",
			results: [][]any{
				{"info", "SYSTEM"},
				{"main", "SYSTEM"},
			},
		},
	}

	db, err := newTestDB()
	require.NoError(t, err)
	defer db.Close()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			tx, err := db.BeginTx(ctx)
			require.NoError(t, err)
			defer tx.Rollback(ctx) // always rollback

			interp := newTestInterp(t, tx, test.sql)

			var values [][]any
			err = interp.Execute(newEngineCtx(test.caller), tx, test.execSQL, test.execVars, func(v *common.Row) error {
				values = append(values, v.Values)
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

// Test_CreateAndDelete tests creating and dropping different objects,
// as well as how created objects are read from the database on startup.
func Test_CreateAndDelete(t *testing.T) {
	type testcase struct {
		name   string // name of the test
		create string
		drop   string
		verify string // must return 0 rows
	}

	tests := []testcase{
		{
			name:   "create and drop table",
			create: "CREATE TABLE tbl (col INT PRIMARY KEY);",
			drop:   "DROP TABLE tbl;",
			verify: "SELECT * FROM info.tables where name = 'tbl' AND namespace = 'main';",
		},
		{
			name:   "create and drop role",
			create: "CREATE ROLE test_role;",
			drop:   "DROP ROLE test_role;",
			verify: "SELECT * FROM info.roles where name = 'test_role';",
		},
		{
			name:   "create and drop namespace",
			create: "CREATE NAMESPACE test;",
			drop:   "DROP NAMESPACE test;",
			verify: "SELECT * FROM info.namespaces where name = 'test';",
		},
		{
			name:   "create and drop index",
			create: "CREATE INDEX idx ON users (name);",
			drop:   "DROP INDEX idx;",
			verify: "SELECT * FROM info.indexes where name = 'idx' AND namespace = 'main';",
		},
		{
			name:   "create and drop action with no parameters",
			create: "CREATE ACTION act() public {};",
			drop:   "DROP ACTION act;",
			verify: "SELECT * FROM info.actions where name = 'act' AND namespace = 'main';",
		},
	}

	db, err := newTestDB()
	require.NoError(t, err)
	defer db.Close()

	for _, test := range tests {
		// the first run through, we will test creating and dropping tables
		t.Run(test.name+"_1", func(t *testing.T) {
			ctx := context.Background()
			tx, err := db.BeginTx(ctx)
			require.NoError(t, err)
			defer tx.Rollback(ctx) // always rollback

			interp := newTestInterp(t, tx, nil)

			err = interp.Execute(newEngineCtx(defaultCaller), tx, test.create, nil, nil)
			require.NoError(t, err)

			err = interp.Execute(newEngineCtx(defaultCaller), tx, test.drop, nil, nil)
			require.NoError(t, err)

			count := 0
			err = interp.Execute(newEngineCtx(defaultCaller), tx, test.verify, nil, func(r *common.Row) error {
				count++
				return nil
			})
			require.NoError(t, err)
			assert.Equal(t, 0, count)
		})

		// the second run through, we will test creating, mocking a restart, dropping, and verifying
		t.Run(test.name+"_2", func(t *testing.T) {
			ctx := context.Background()
			tx, err := db.BeginTx(ctx)
			require.NoError(t, err)
			defer tx.Rollback(ctx) // always rollback

			interp := newTestInterp(t, tx, nil)

			err = interp.Execute(newEngineCtx(defaultCaller), tx, test.create, nil, nil)
			require.NoError(t, err)

			interp = newTestInterp(t, tx, nil)

			err = interp.Execute(newEngineCtx(defaultCaller), tx, test.drop, nil, nil)
			require.NoError(t, err)

			count := 0
			err = interp.Execute(newEngineCtx(defaultCaller), tx, test.verify, nil, func(v *common.Row) error {
				count++
				return nil
			})
			require.NoError(t, err)
			assert.Equal(t, 0, count)
		})
	}
}

func newEngineCtx(caller string) *common.EngineContext {
	if caller == "" {
		caller = defaultCaller
	}
	return &common.EngineContext{
		TxContext: &common.TxContext{
			Ctx: context.Background(),
			BlockContext: &common.BlockContext{
				Height: 1,
				ChainContext: &common.ChainContext{
					NetworkParameters: &common.NetworkParameters{},
					MigrationParams:   &common.MigrationContext{},
				},
			},
			Caller:        caller,
			Signer:        []byte(caller),
			Authenticator: "test_authenticator",
		}}
}

func adminCtx() *common.EngineContext {
	return &common.EngineContext{
		TxContext: &common.TxContext{
			Ctx: context.Background(),
			BlockContext: &common.BlockContext{
				Height: 1,
				ChainContext: &common.ChainContext{
					NetworkParameters: &common.NetworkParameters{},
					MigrationParams:   &common.MigrationContext{},
				},
			},
			Caller:        "",
			Signer:        []byte(""),
			Authenticator: "test_authenticator",
		},
		OverrideAuthz: true,
	}
}

func Test_Actions(t *testing.T) {
	type testcase struct {
		name string // name of the test
		// array of sql statements, first element is the namespace, second is the sql statement
		// they can begin with {namespace}sql, or just sql
		stmt []string
		// namespace in which the action is defined
		namespace string
		// action to execute
		action string
		// values to pass to the action
		values []any
		// expected results
		results [][]any
		// expected error
		err error
		// caller to use for the action, will default to defaultCaller
		caller string
	}

	// rawTest is a helper that allows us to write test logic purely in Kuneiform.
	rawTest := func(name string, body string) testcase {
		return testcase{
			name:   name,
			stmt:   []string{`CREATE ACTION raw_test() public {` + body + `}`},
			action: "raw_test",
		}
	}

	tests := []testcase{
		{
			name: "insert and select",
			stmt: []string{`
			CREATE ACTION create_user($name text, $age int) public returns (count int) {
				INSERT INTO users (id, name, age)
				VALUES (1, $name, $age);

				for $row in SELECT count(*) as count FROM users WHERE name = $name {
					RETURN $row.count;
				};

				error('user not found');
			}
			`},
			action: "create_user",
			values: []any{"Alice", int64(30)},
			results: [][]any{
				{int64(1)},
			},
		},
		{
			name: "read values out of the db, perform arithmetic and conditionals",
			stmt: []string{`INSERT INTO users(id, name, age) VALUES (1, 'satoshi', 42), (2, 'hal finney', 50), (3, 'craig wright', 45)`, `
			CREATE ACTION do_something() public view returns (result int) {
				$total int8;
				$sum int8;
				for $row in SELECT count(*) as count, sum(age) as sum FROM users WHERE age >= 45 {
					$total := $row.count::int8;
					$sum := $row.sum::int8;
				}
				if $total is null {
					error('no users found');
				}

				-- random arithmetic
				$result := ($total * 2 + $sum)/3; -- (2 * 2 + 95)/3 = 33

				if $result < 33 {
					error('result is less than 33');
				}
				if $result > 33 {
					error('result is greater than 33');
				}

				RETURN $result;
			}
			`},
			action: "do_something",
			results: [][]any{
				{int64(33)},
			},
		},
		{
			name: "return next from another action",
			stmt: []string{
				`INSERT INTO users(id, name, age) VALUES (1, 'satoshi', 42), (2, 'hal finney', 50), (3, 'craig wright', 45)`,
				`CREATE NAMESPACE test;`,
				`{test}CREATE ACTION get_users() public view returns table(name text, age int) {
					return SELECT name, age FROM main.users ORDER BY id;
				}`,
				`CREATE ACTION get_next_user($d int) public view returns table(name text, age int) {
					for $row in test.get_users() {
						RETURN NEXT $row.name, $row.age/$d;
					}
				}`,
			},
			values: []any{int64(2)},
			action: "get_next_user",
			results: [][]any{
				{"satoshi", int64(21)},
				{"hal finney", int64(25)},
				{"craig wright", int64(22)},
			},
		},
		rawTest("loop over array", `
		$arr := array[1,2,3];
		$sum := 0;
		for $i in $arr {
			$sum := $sum + $i;
		};

		if $sum != 6{
			error('sum is not 6');
		};
		`),
		rawTest("loop over range", `
		$sum := 0;
		for $i in 1..4 {
			$sum := $sum + $i;
		}

		if $sum != 10 {
			error('sum is not 10');
		}
		`),
		rawTest("slice", `
		$arr := array[1,2,3,4,5];
		$slice := $arr[2:3];
		if $slice != array[2,3] {
			error('slice is not [2,3]');
		}
		`),
		rawTest("assign to array value", `
		$arr := array[1,2,3];
		$arr[2] := 5;
		if $arr != array[1,5,3] {
			error('array is not [1,5,3]');
		}
		if $arr[3] != 3 {
			error('array[3] is not 3');
		}
		`),
		{
			name: "call another action",
			stmt: []string{
				`CREATE NAMESPACE other;`,
				`{other}CREATE ACTION get_single_value() public view returns (value int) { return 1; }`,
				`{other}CREATE ACTION get_several_values() public view returns (value1 int, value2 int) { return 2, 3; }`,
				`{other}CREATE ACTION get_table($to int, $from int) public view returns table(value int) { 
					for $i in $to..$from {
						RETURN NEXT $i;
					};
				}`,
				`CREATE ACTION call_other() public view {
					$value1 := other.get_single_value();
					if $value1 != 1 {
						error('value1 is not 1');
					}

					$value2, $value3 := other.get_several_values();
					if $value2 != 2 {
						error('value2 is not 2');
					}
					if $value3 != 3 {
						error('value3 is not 3');
					}

					_, $value3Again := other.get_several_values();
					if $value3Again != 3 {
						error('value3Again is not 3');
					}

					$iter := 0;
					for $row in other.get_table(1, 4) {
						$iter := $iter + 1;
						if $row.value != $iter {
							error('value is not equal to iter');
						}
					}
					if $iter != 4 {
						error('iter is not 4');
					}
				}`,
			},
			action: "call_other",
		},
		rawTest("testing if, else if, else", `
		$a := 1;
		$b := 2;
		$total := 0;
		if $a < $b {
			$total := $total + 1;
		} else {
			error('a is not less than b'); 
		}

		if $a > $b {
			error('a is not greater than b');
		} else if $a == $b {
			error('a is not equal to b'); 
		} else {
			$total := $total + 1; 
		}

		if $a + $b == 4 {
			error('a + b is not 4');
		} elseif $a + $b == 3 { -- supports else if and elseif
			$total := $total + 1; 
		} else {
			error('a + b is not 3'); 
		}

		if $total != 3 {
			error('total is not 3');
		}
		`),
		rawTest("break", `
		$sum := 0;
		for $i in 1..10 {
			$sum := $sum + $i;
			if $i == 5 {
				break;
			}
		}

		if $sum != 15 {
			error('sum is not 15, but ' || $sum::text);
		}
		`),
		rawTest("continue", `
		$sum := 0;
		for $i in 1..10 {
			if $i == 5 {
				continue;
			}
			$sum := $sum + $i;
		}

		if $sum != 50 {
			error('sum is not 50, but ' || $sum::text);
		}
		`),
		rawTest("function call expression", `
		if abs(-1) != 1 {
			error('abs(-1) is not 1');
		}
		`),
		rawTest("logical expression", `
		if true and false {
			error('true and false is not false');
		}

		if true or false {
			-- pass
		} else {
			error('true or false is not true'); 
		}

		if (true or false) and true {
			-- pass
		} else {
			error('(true or false) and true is not true');
		}

		if true and null {
			error('true and null should be null');
		}
		if null and false {
			error('null and false should be null');
		}
		`),
		rawTest("arithmetic", `
		if 1 + 2 != 3 {
			error('1 + 2 is not 3');
		}
		if 1 - 2 != -1 {
			error('1 - 2 is not -1');
		}
		if 2 * 2 != 4 {
			error('2 * 2 is not 4');
		}
		if 4 / 2 != 2 {
			error('4 / 2 is not 2');
		}
		if 5 % 2 != 1 {
			error('5 % 2 is not 1');
		}
		if 5 + null is not null {
			error('5 + null is not null');
		}
		`),
	}

	db, err := newTestDB()
	require.NoError(t, err)
	defer db.Close()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			tx, err := db.BeginTx(ctx)
			require.NoError(t, err)
			defer tx.Rollback(ctx) // always rollback

			interp := newTestInterp(t, tx, test.stmt)

			var results [][]any
			// TODO: add expected logs
			_, err = interp.Call(newEngineCtx(test.caller), tx, test.namespace, test.action, test.values, func(v *common.Row) error {
				results = append(results, v.Values)
				return nil
			})
			if test.err != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, test.err)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, len(test.results), len(results))
			for i, row := range results {
				require.Equal(t, len(test.results[i]), len(row))
				for j, val := range row {
					require.EqualValues(t, test.results[i][j], val)
				}
			}
		})
	}
}

// This function tests precompiles
func Test_Extensions(t *testing.T) {
	type testExtension struct {
		alias string
		i     int
		meta  map[string]any
	}

	// notifications track in which order the extension is initialized, used, and unused
	var notifications []string

	// below we create a test precompile extension
	err := precompiles.RegisterPrecompile("test", precompiles.PrecompileExtension[testExtension]{
		Initialize: func(ctx context.Context, service *common.Service, db sql.DB, alias string, metadata map[string]any) (*testExtension, error) {
			te := &testExtension{
				alias: alias,
			}
			te.meta = metadata
			notifications = append(notifications, "initialize")

			return te, nil
		},
		OnUse: func(ctx *common.EngineContext, app *common.App, t *testExtension) error {
			notifications = append(notifications, "use")
			return nil
		},
		OnUnuse: func(ctx *common.EngineContext, app *common.App, t *testExtension) error {
			notifications = append(notifications, "unuse")
			return nil
		},
		Methods: []precompiles.Method[testExtension]{
			{
				Name: "concat",
				Handler: func(ctx *common.EngineContext, app *common.App, inputs []any, resultFn func([]any) error, te *testExtension) error {
					te.i++
					if len(inputs) != 2 {
						return fmt.Errorf("expected 2 inputs, got %d", len(inputs))
					}

					if _, ok := inputs[0].(string); !ok {
						return errors.New("first input is not a string")
					}

					if _, ok := inputs[1].(string); !ok {
						return errors.New("second input is not a string")
					}

					return resultFn([]any{inputs[0].(string) + inputs[1].(string)})
				},
				AccessModifiers: []precompiles.Modifier{precompiles.SYSTEM},
			},
			{
				Name: "get",
				Handler: func(ctx *common.EngineContext, app *common.App, inputs []any, resultFn func([]any) error, te *testExtension) error {
					count := 0
					ctx.OverrideAuthz = true
					_, err := app.Engine.Call(ctx, app.DB, te.alias, "internal", nil, func(v *common.Row) error {
						count++
						if v.Values[0].(string) != "internal" {
							return errors.New("expected 'internal' message")
						}
						return nil
					})
					if err != nil {
						return err
					}
					if count != 1 {
						return errors.New("expected 1 internal call")
					}
					ctx.OverrideAuthz = false

					if len(inputs) != 1 {
						return errors.New("expected 1 input")
					}
					str, ok := inputs[0].(string)
					if !ok {
						return errors.New("expected string input")
					}

					meta, ok := te.meta[str]
					if !ok {
						return errors.New("expected key in metadata")
					}

					te.i++
					return resultFn([]any{meta})
				},
				AccessModifiers: []precompiles.Modifier{precompiles.PUBLIC, precompiles.VIEW},
				ReturnColumns:   []string{"response"},
			},
			{
				Name: "owner_only",
				Handler: func(ctx *common.EngineContext, app *common.App, inputs []any, resultFn func([]any) error, te *testExtension) error {
					return resultFn([]any{"owner_only"})
				},
				AccessModifiers: []precompiles.Modifier{precompiles.OWNER, precompiles.PUBLIC},
			},
			{
				Name: "internal",
				Handler: func(ctx *common.EngineContext, app *common.App, inputs []any, resultFn func([]any) error, te *testExtension) error {
					if len(inputs) != 0 {
						return errors.New("expected 0 inputs")
					}
					te.i++
					return resultFn([]any{"internal"})
				},
				AccessModifiers: []precompiles.Modifier{precompiles.PRIVATE},
			},
		},
	})
	require.NoError(t, err)

	// now we can test calling the precompile
	// There are a few things we should test. All of these tests should be performed twice:
	// once with the extension initialized and deployed at the same time, and once with an interpreter
	// restarted, which simulates a node restart.
	//
	// The following tests should be performed:
	// 1. concat can only be called by other actions
	// 2. get can be called publicly or by other actions, and can be called in a view action
	// 3. owner_only can only be called by the owner
	// 4. internal can only be called by other methods, and not externally
	// 5. metadata of every type is passed correctly (verify using get)
	// 6. importing the extension creates a namespace that can cannot dropped by the admin
	// 7. the extension namespace can have privileges granted to roles
	// 8. the owner cannot add tables, actions, etc. to the extension namespace. The admin can.

	// finally, the extension is initialized, used, and unused in the correct order (notifications)

	db, err := newTestDB()
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()
	tx, err := db.BeginTx(ctx)
	require.NoError(t, err)
	defer tx.Rollback(ctx) // always rollback

	i := 0 // tracks iters, since some of these cannot be made idempotent (e.g. alter table)

	do := func(interp *interpreter.ThreadSafeInterpreter) {

		callFromUser := func(caller string, namespace, action string, values []any, fn func(*common.Row) error) error {
			_, err := interp.Call(newEngineCtx(caller), tx, namespace, action, values, fn)
			return err
		}
		adminCall := func(namespace, action string, values []any, fn func(*common.Row) error) error {
			_, err := interp.Call(adminCtx(), tx, namespace, action, values, fn)
			return err
		}
		execFromUser := func(caller string, sql string, fn func(*common.Row) error) error {
			return interp.Execute(newEngineCtx(caller), tx, sql, nil, fn)
		}
		adminExec := func(sql string, fn func(*common.Row) error) error {
			return interp.Execute(adminCtx(), tx, sql, nil, fn)
		}

		// initialize the extension
		err = interp.Execute(newEngineCtx(defaultCaller), tx, `
		USE IF NOT EXISTS test {
			text: 'text',
			int: 1+2,
			bool: true,
			blob: 0x010203, -- hex encoded []byte{1,2,3}
			uuid: 'f47ac10b-58cc-4372-a567-0e02b2c3d479'::uuid,
			decimal: 1.23
		} AS test_ext;
	`, nil, func(r *common.Row) error {
			return nil
		})
		require.NoError(t, err)

		// 1: concat can only be called by other actions
		err = callFromUser(defaultCaller, "test_ext", "concat", []any{"hello", "world"}, nil)
		assert.ErrorIs(t, err, interpreter.ErrSystemOnly)

		err = adminExec("CREATE ACTION IF NOT EXISTS call_concat() public view returns (text) { return test_ext.concat('hello', 'world'); }", nil)
		require.NoError(t, err)

		err = callFromUser(defaultCaller, "", "call_concat", nil, exact("helloworld"))
		require.NoError(t, err)

		// cannot be called in view action
		err = adminExec("CREATE ACTION IF NOT EXISTS call_concat_view() public view returns (text) { return test_ext.concat('hello', 'world'); }", nil)
		require.NoError(t, err)

		readTx, err := db.BeginReadTx(ctx)
		require.NoError(t, err)
		defer readTx.Rollback(ctx)

		_, err = interp.Call(newEngineCtx(defaultCaller), readTx, "", "call_concat_view", nil, exact("helloworld"))
		assert.ErrorIs(t, err, interpreter.ErrActionMutatesState)

		// 2: get can be called publicly or by other actions, and can be called in a view action
		err = callFromUser(defaultCaller, "test_ext", "get", []any{"text"}, exact("text"))
		require.NoError(t, err)

		err = adminExec("CREATE ACTION IF NOT EXISTS call_get() public view returns (text) { return test_ext.get('text'); }", nil)
		require.NoError(t, err)

		err = callFromUser(defaultCaller, "", "call_get", nil, exact("text"))
		require.NoError(t, err)

		// 3: owner_only can only be called by the owner
		err = adminCall("test_ext", "owner_only", nil, exact("owner_only"))
		require.NoError(t, err)

		err = callFromUser("user", "test_ext", "owner_only", nil, nil)
		assert.ErrorIs(t, err, interpreter.ErrActionOwnerOnly)

		// 4: internal can only be called by other methods, and not externally nor by
		// other actions in different namespaces

		// callable by other methods is tested implicitly with the get method
		err = callFromUser(defaultCaller, "test_ext", "internal", nil, exact("internal"))
		assert.ErrorIs(t, err, interpreter.ErrActionPrivate)

		// action in "main" namespace cannot call internal
		err = adminExec("CREATE ACTION IF NOT EXISTS call_internal() public view returns (text) { return test_ext.internal(); }", nil)
		require.NoError(t, err)

		err = callFromUser(defaultCaller, "", "call_internal", nil, exact("internal"))
		assert.ErrorIs(t, err, interpreter.ErrActionPrivate)

		// calling a private action with overrideAuthz should work
		err = adminCall("test_ext", "internal", nil, exact("internal"))
		require.NoError(t, err)

		// action in "test_ext" namespace can call internal
		err = adminExec("{test_ext}CREATE ACTION IF NOT EXISTS call_internal() public view returns (text) { return internal(); }", nil)
		require.NoError(t, err)

		err = callFromUser(defaultCaller, "test_ext", "call_internal", nil, exact("internal"))
		require.NoError(t, err)

		// 5. metadata of every type is passed correctly (verify using get)
		for _, get := range []struct {
			key   string
			value any
		}{
			{"text", "text"},
			{"int", int64(3)},
			{"bool", true},
			{"blob", []byte{1, 2, 3}},
			{"uuid", mustUUID("f47ac10b-58cc-4372-a567-0e02b2c3d479")},
			{"decimal", mustDecimal("1.23")},
		} {
			err = adminCall("test_ext", "get", []any{get.key}, exact(get.value))
			require.NoError(t, err)
		}

		// 6. The imported extension creates a namespace that cannot be dropped by the admin
		err = adminExec("DROP NAMESPACE IF EXISTS test_ext;", nil)
		require.ErrorIs(t, err, interpreter.ErrCannotMutateExtension)

		// 7. The extension namespace can have privileges granted to roles
		err = execFromUser(defaultCaller, "CREATE ROLE IF NOT EXISTS test_role;", nil)
		require.NoError(t, err)

		err = execFromUser(defaultCaller, "GRANT IF NOT GRANTED select ON test_ext TO test_role;", nil)
		require.NoError(t, err)

		// 8. The owner cannot create/alter/drop tables, actions, etc. to the extension namespace. The admin can.

		// tables
		err = execFromUser(defaultCaller, "{test_ext}CREATE TABLE IF NOT EXISTS test_table (id INT PRIMARY KEY);", nil)
		require.ErrorIs(t, err, interpreter.ErrCannotMutateExtension)

		err = adminExec("{test_ext}CREATE TABLE IF NOT EXISTS test_table (id INT PRIMARY KEY);", nil)
		require.NoError(t, err)

		err = execFromUser(defaultCaller, "{test_ext}ALTER TABLE test_table ADD COLUMN name TEXT;", nil)
		require.ErrorIs(t, err, interpreter.ErrCannotMutateExtension)

		if i == 0 {
			err = adminExec("{test_ext}ALTER TABLE test_table ADD COLUMN name TEXT;", nil)
			require.NoError(t, err)
		}

		// indexes
		err = execFromUser(defaultCaller, "{test_ext}CREATE INDEX IF NOT EXISTS test_index ON test_table (id);", nil)
		require.ErrorIs(t, err, interpreter.ErrCannotMutateExtension)

		err = adminExec("{test_ext}CREATE INDEX IF NOT EXISTS test_index ON test_table (id);", nil)
		require.NoError(t, err)

		err = execFromUser(defaultCaller, "{test_ext}DROP INDEX IF EXISTS test_index;", nil)
		require.ErrorIs(t, err, interpreter.ErrCannotMutateExtension)

		err = adminExec("{test_ext}DROP INDEX IF EXISTS test_index;", nil)
		require.NoError(t, err)

		// drop table

		err = execFromUser(defaultCaller, "{test_ext}DROP TABLE IF EXISTS test_table;", nil)
		require.ErrorIs(t, err, interpreter.ErrCannotMutateExtension)

		err = adminExec("{test_ext}DROP TABLE IF EXISTS test_table;", nil)
		require.NoError(t, err)

		// actions

		err = execFromUser(defaultCaller, "{test_ext}CREATE ACTION IF NOT EXISTS test_action() public view returns (text) { return 'test'; }", nil)
		require.ErrorIs(t, err, interpreter.ErrCannotMutateExtension)

		err = adminExec("{test_ext}CREATE ACTION IF NOT EXISTS test_action() public view returns (text) { return 'test'; }", nil)
		require.NoError(t, err)

		err = execFromUser(defaultCaller, "{test_ext}DROP ACTION IF EXISTS test_action;", nil)
		require.ErrorIs(t, err, interpreter.ErrCannotMutateExtension)

		err = adminExec("{test_ext}DROP ACTION IF EXISTS test_action;", nil)
		require.NoError(t, err)
	}

	// first run: new interpreter
	interp := newTestInterp(t, tx, nil)
	do(interp)

	// second run: restart interpreter
	// It will read in all previous data from the database.
	interp = newTestInterp(t, tx, nil)
	do(interp)
}

// This test tests that functions can be overwritten by extension methods, and extension
// methods can be overwritten by actions. Dropping actions should restore the correct previous
// behavior.
func Test_NamingOverwrites(t *testing.T) {
	db, err := newTestDB()
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()
	tx, err := db.BeginTx(ctx)
	require.NoError(t, err)
	defer tx.Rollback(ctx) // always rollback

	// we use an extension with one method "abs" to overwrite the built-in abs function
	absCalled := false
	err = precompiles.RegisterPrecompile("test2", precompiles.PrecompileExtension[struct{}]{
		Methods: []precompiles.Method[struct{}]{
			{
				Name: "abs",
				Handler: func(ctx *common.EngineContext, app *common.App, inputs []any, resultFn func([]any) error, _ *struct{}) error {
					absCalled = true
					return resultFn([]any{int64(math.Abs(float64(inputs[0].(int64))))})
				},
				AccessModifiers: []precompiles.Modifier{precompiles.PUBLIC, precompiles.VIEW},
			},
		},
	})
	require.NoError(t, err)

	interp := newTestInterp(t, tx, nil)

	err = interp.Execute(newEngineCtx(defaultCaller), tx, `USE IF NOT EXISTS test2 AS test_ext;`, nil, nil)
	require.NoError(t, err)

	// we will create an action in the extension schema and call abs
	err = interp.Execute(adminCtx(), tx, `{test_ext}CREATE ACTION OR REPLACE use_abs() public view returns (int) { return abs(-1); }`, nil, nil)
	require.NoError(t, err)

	_, err = interp.Call(newEngineCtx(defaultCaller), tx, "test_ext", "use_abs", nil, exact(int64(1)))
	require.NoError(t, err)

	assert.True(t, absCalled)

	// now we will create an action that overwrites the abs function
	absCalled = false
	err = interp.Execute(adminCtx(), tx, `{test_ext}CREATE ACTION OR REPLACE abs($a int) system view returns (int) { return 2; }`, nil, nil)
	require.NoError(t, err)

	_, err = interp.Call(newEngineCtx(defaultCaller), tx, "test_ext", "use_abs", nil, exact(int64(2)))
	require.NoError(t, err)
	assert.False(t, absCalled)

	// we will make a new interpreter to ensure that they are loaded correctly
	interp = newTestInterp(t, tx, nil)

	_, err = interp.Call(newEngineCtx(defaultCaller), tx, "test_ext", "use_abs", nil, exact(int64(2)))
	require.NoError(t, err)

	// dropping the action should restore the correct behavior
	err = interp.Execute(adminCtx(), tx, `{test_ext}DROP ACTION abs;`, nil, nil)
	require.NoError(t, err)

	_, err = interp.Call(newEngineCtx(defaultCaller), tx, "test_ext", "use_abs", nil, exact(int64(1)))
	require.NoError(t, err)
	assert.True(t, absCalled)

	// will now try it to make sure we can go from action -> function
	err = interp.Execute(adminCtx(), tx, `create action OR REPLACE length($a text) public view returns (int) { return 1; }`, nil, nil)
	require.NoError(t, err)

	err = interp.Execute(adminCtx(), tx, `create action OR REPLACE use_length() public view returns (int) { return length('hello'); }`, nil, nil)
	require.NoError(t, err)

	_, err = interp.Call(newEngineCtx(defaultCaller), tx, "", "use_length", nil, exact(int64(1)))
	require.NoError(t, err)

	// dropping the action should restore the correct behavior
	err = interp.Execute(adminCtx(), tx, `DROP ACTION length;`, nil, nil)
	require.NoError(t, err)

	_, err = interp.Call(newEngineCtx(defaultCaller), tx, "", "use_length", nil, exact(int64(5)))
	require.NoError(t, err)
}

// exact is a helper function that verifies that a result is called exactly once, and that the result is equal to the expected value.
func exact(val any) func(*common.Row) error {
	called := false
	return func(row *common.Row) error {
		if called {
			return errors.New("result called multiple times")
		}
		called = true

		if len(row.Values) != 1 {
			return fmt.Errorf("expected 1 value, got %d", len(row.Values))
		}

		switch t := val.(type) {
		case []byte:
			t2, ok := row.Values[0].([]byte)
			if !ok {
				return fmt.Errorf("expected []byte, got %T", row.Values[0])
			}

			if !bytes.Equal(t, t2) {
				return fmt.Errorf("expected %v, got %v", t, t2)
			}

		default:
			// if it is a pointer, we need to dereference it
			if reflect.TypeOf(val).Kind() == reflect.Ptr {
				val = reflect.ValueOf(val).Elem().Interface()
			}

			// if row.Values[0] is a pointer, we need to dereference it
			rowVal := row.Values[0]
			if reflect.TypeOf(rowVal).Kind() == reflect.Ptr {
				rowVal = reflect.ValueOf(rowVal).Elem().Interface()
			}
			if rowVal != val {
				return fmt.Errorf("expected %v, got %v", val, rowVal)
			}
		}

		return nil
	}
}

func mustDecimal(s string) *decimal.Decimal {
	d, err := decimal.NewFromString(s)
	if err != nil {
		panic(err)
	}
	return d
}

func mustUUID(s string) *types.UUID {
	u, err := types.ParseUUID(s)
	if err != nil {
		panic(err)
	}
	return u
}

func newTestDB() (*pg.DB, error) {
	cfg := &pg.DBConfig{
		PoolConfig: pg.PoolConfig{
			ConnConfig: pg.ConnConfig{
				Host:   "127.0.0.1",
				Port:   "5432",
				User:   "kwild",
				Pass:   "kwild", // would be ignored if pg_hba.conf set with trust
				DBName: "kwil_test_db",
			},
			MaxConns: 11,
		},
	}

	ctx := context.Background()

	return pg.NewDB(ctx, cfg)
}

// newTestInterp creates a new interpreter for testing purposes.
// It is seeded with the default tables.
func newTestInterp(t *testing.T, tx sql.DB, seeds []string) *interpreter.ThreadSafeInterpreter {
	interp, err := interpreter.NewInterpreter(context.Background(), tx, &common.Service{})
	require.NoError(t, err)

	engCtx := newEngineCtx(defaultCaller)
	engCtx.OverrideAuthz = true
	err = interp.Execute(engCtx, tx, "GRANT IF NOT GRANTED owner TO $user", map[string]any{
		"user": defaultCaller,
	}, nil)
	require.NoError(t, err)

	for _, stmt := range append([]string{createUsersTable, createPostsTable}, seeds...) {
		err := interp.Execute(newEngineCtx(defaultCaller), tx, stmt, nil, nil)
		require.NoError(t, err)
	}

	return interp
}
