//go:build pglive

package interpreter_test

import (
	"context"
	"testing"

	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/node/engine/interpreter"
	"github.com/kwilteam/kwil-db/node/pg"
	"github.com/kwilteam/kwil-db/node/types/sql"
	"github.com/stretchr/testify/require"
)

const (
	defaultCaller    = "owner"
	createUsersTable = `
CREATE TABLE users (
	id INT PRIMARY KEY,
	name TEXT,
	age INT
);
	`

	createPostsTable = `
CREATE TABLE posts (
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
		execSQL     string  // sql to return the results. Either this or execAction must be set
		results     [][]any // table of results
		err         error   // expected error, can be nil. Errors _MUST_ occur on the exec. This is a direct match
		errContains string  // expected error message, can be empty. Errors _MUST_ occur on the exec. This is a substring match
		caller      string  // caller to use for the action, will default to defaultCaller
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
			name:        "drop default namespace",
			execSQL:     "DROP NAMESPACE main;",
			errContains: "cannot drop built-in namespace",
		},
		{
			name:        "drop info namespace",
			execSQL:     "DROP NAMESPACE info;",
			errContains: "cannot drop built-in namespace",
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
		// testing info schema
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
			if test.execSQL != "" {
				err = interp.Execute(newTxCtx(test.caller), tx, test.execSQL, nil, func(v *common.Row) error {
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

func newTxCtx(caller string) *common.TxContext {
	if caller == "" {
		caller = defaultCaller
	}
	return &common.TxContext{
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
			_, err = interp.Call(newTxCtx(test.caller), tx, test.namespace, test.action, test.values, func(v *common.Row) error {
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

	err = interp.SetOwner(context.Background(), tx, defaultCaller)
	require.NoError(t, err)

	for _, stmt := range append([]string{createUsersTable, createPostsTable}, seeds...) {
		err := interp.Execute(newTxCtx(defaultCaller), tx, stmt, nil, nil)
		require.NoError(t, err)
	}

	return interp
}
