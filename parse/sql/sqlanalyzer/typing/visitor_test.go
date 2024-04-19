package typing_test

import (
	"errors"
	"strings"
	"testing"
	"unicode"

	"github.com/kwilteam/kwil-db/core/types"
	sqlparser "github.com/kwilteam/kwil-db/parse/sql"
	"github.com/kwilteam/kwil-db/parse/sql/sqlanalyzer/typing"
	"github.com/kwilteam/kwil-db/parse/sql/tree"
	"github.com/stretchr/testify/require"
)

// Test_Qualification tests that the qualification of column references works as expected.
func Test_Qualification(t *testing.T) {
	type testcase struct {
		name    string
		stmt    string
		want    string
		wantErr bool
	}

	tests := []testcase{
		{
			name: "simple select",
			stmt: "SELECT id, name FROM users WHERE name = 'satoshi' ORDER BY id LIMIT 10 OFFSET 10;",
			want: `SELECT "users"."id", "users"."name" FROM "users" WHERE "users"."name" = 'satoshi' ORDER BY "users"."id" LIMIT 10 OFFSET 10;`,
		},
		{
			name: "joins",
			stmt: `SELECT u1.id, u2.name FROM users AS u1
			INNER JOIN users AS u2 ON u1.id = u2.id
			WHERE u1.id = $id AND u2.name = $name;`,
			want: `SELECT "u1"."id", "u2"."name" FROM "users" AS "u1"
			INNER JOIN "users" AS "u2" ON "u1"."id" = "u2"."id"
			WHERE "u1"."id" = $id AND "u2"."name" = $name;`,
		},
		{
			name: "joins against subquery",
			stmt: `SELECT u1.id, u2.username FROM users AS u1
			INNER JOIN (SELECT id, name as username FROM users WHERE id = $id) AS u2 ON u1.id = u2.id
			WHERE u1.id = $id AND u2.username = $name;`,
			want: `SELECT "u1"."id", "u2"."username" FROM "users" AS "u1"
			INNER JOIN (SELECT "users"."id", "users"."name" AS "username" FROM "users" WHERE "users"."id" = $id) AS "u2" ON "u1"."id" = "u2"."id"
			WHERE "u1"."id" = $id AND "u2"."username" = $name;`,
		},
		{
			name: "common table expression",
			stmt: `WITH cte AS (SELECT id, name FROM users) SELECT cte.id as userid, posts.title as title FROM cte
			INNER JOIN posts ON cte.id = posts.author_id;`,
			want: `WITH "cte" AS (SELECT "users"."id", "users"."name" FROM "users") SELECT "cte"."id" AS "userid", "posts"."title" AS "title" FROM "cte"
			INNER JOIN "posts" ON "cte"."id" = "posts"."author_id";`,
		},
		{
			name: "insert returning",
			stmt: `INSERT INTO users (id, name) VALUES ($id+1, (select name from users where id = $id))
			RETURNING *, id as userid;`,
			want: `INSERT INTO "users" ("id", "name") VALUES ($id+1, (SELECT "users"."name" FROM "users" WHERE "users"."id" = $id))
			RETURNING *, "users"."id" AS "userid";`,
		},
		{
			name: "insert on conflict",
			stmt: `INSERT INTO users as u (id, name) VALUES ($id, $name) ON CONFLICT (id) where id = 1 DO UPDATE SET name = $name WHERE u.id = $id RETURNING u.name;`,
			want: `INSERT INTO "users" AS "u" ("id", "name") VALUES ($id, $name) ON CONFLICT ("id") WHERE "u"."id" = 1 DO UPDATE SET "name" = $name WHERE "u"."id" = $id RETURNING "u"."name";`,
		},
		{
			name: "update returning with qualified table name",
			stmt: `UPDATE users as u SET name = u1.name FROM users AS u1 WHERE u.id = $id RETURNING u.name as username, u1.name as uname;`,
			want: `UPDATE "users" AS "u" SET "name" = "u1"."name" FROM "users" AS "u1" WHERE "u"."id" = $id RETURNING "u"."name" AS "username", "u1"."name" AS "uname";`,
		},
		{
			name: "delete returning with qualified table name",
			stmt: `DELETE FROM users as u WHERE id = $id RETURNING *, u.id as uid;`,
			want: `DELETE FROM "users" AS "u" WHERE "u"."id" = $id
			RETURNING *, "u"."id" AS "uid";`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ast, err := sqlparser.Parse(test.stmt)
			require.NoError(t, err)

			_, err = typing.AnalyzeTypes(ast, []*types.Table{usersTable, postsTable}, &typing.AnalyzeOptions{
				ArbitraryBinds: true,
				Qualify:        true,
			})
			if test.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			str, err := tree.SafeToSQL(ast)
			require.NoError(t, err)

			require.Equal(t, removeWhitespace(test.want), removeWhitespace(str))
		})
	}
}

func removeWhitespace(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1 // skip this rune
		}
		return r
	}, s)
}

// Test_Typing tests that the typing visitor properly
// analyzes the types of the given statement.
func Test_Typing(t *testing.T) {
	type testcase struct {
		name     string
		stmt     string
		relation map[string]*types.DataType
		err      error // can be nil if no error is expected
	}

	tests := []testcase{
		{
			name: "simple select",
			stmt: "SELECT id, name FROM users WHERE name = 'satoshi' ORDER BY id LIMIT 10 OFFSET 10;",
			relation: map[string]*types.DataType{
				"id":   types.IntType,
				"name": types.TextType,
			},
		},
		{
			name: "CTE",
			stmt: `WITH cte AS (SELECT id, name FROM users) SELECT cte.id as userid, posts.title as title FROM cte
			INNER JOIN posts ON cte.id = posts.author_id;`,
			relation: map[string]*types.DataType{
				"userid": types.IntType,
				"title":  types.TextType,
			},
		},
		{
			name: "select with where, aggregate",
			stmt: `SELECT id, name FROM users WHERE id = $id GROUP BY id, name HAVING count(*) > 1;`,
			relation: map[string]*types.DataType{
				"id":   types.IntType,
				"name": types.TextType,
			},
		},
		{
			name: "subquery",
			stmt: `SELECT id, name FROM users WHERE id IN (SELECT author_id FROM posts WHERE title = $title);`,
			relation: map[string]*types.DataType{
				"id":   types.IntType,
				"name": types.TextType,
			},
		},
		{
			name: "joins",
			stmt: `SELECT u1.id, u2.name FROM users AS u1
			INNER JOIN users AS u2 ON u1.id = u2.id
			WHERE u1.id = $id AND u2.name = $name;`,
			relation: map[string]*types.DataType{
				"id":   types.IntType,
				"name": types.TextType,
			},
		},
		{
			name: "joins against subquery",
			stmt: `SELECT u1.id, u2.username FROM users AS u1
			INNER JOIN (SELECT id, name as username FROM users WHERE id = $id) AS u2 ON u1.id = u2.id
			WHERE u1.id = $id AND u2.username = $name;`,
			relation: map[string]*types.DataType{
				"id":       types.IntType,
				"username": types.TextType,
			},
		},
		{
			name: "insert",
			stmt: `INSERT INTO users (id, name) VALUES ($id+1, (select name from users where id = $id))
			RETURNING *, id as userid;`,
			relation: map[string]*types.DataType{
				"id":     types.IntType,
				"name":   types.TextType,
				"userid": types.IntType,
			},
		},
		{
			name: "insert invalid literal",
			stmt: `INSERT INTO users (id, name) VALUES ($id, 1);`,
			err:  typing.ErrInvalidType,
		},
		{
			name: "upsert",
			stmt: `INSERT INTO users as u (id, name) VALUES ($id, $name) ON CONFLICT (id) where id = 1 DO UPDATE SET name = $name WHERE u.id = $id RETURNING u.name;`,
			relation: map[string]*types.DataType{
				"name": types.TextType,
			},
		},
		{
			name: "update",
			stmt: `UPDATE users as u SET name = u1.name FROM users AS u1 WHERE u.id = $id RETURNING u.name as username;`,
			relation: map[string]*types.DataType{
				"username": types.TextType,
			},
		},
		{
			name: "update return all",
			stmt: `UPDATE users as u SET name = $name WHERE id = $id RETURNING *, u.name as username;`,
			relation: map[string]*types.DataType{
				"id":       types.IntType,
				"name":     types.TextType,
				"username": types.TextType,
			},
		},
		{
			name: "delete",
			stmt: `DELETE FROM users as u WHERE id = $id RETURNING *, u.id as uid;`,
			relation: map[string]*types.DataType{
				"id":   types.IntType,
				"uid":  types.IntType,
				"name": types.TextType,
			},
		},
		{
			name: "compound select",
			stmt: `SELECT id, name FROM users UNION SELECT id, name FROM users;`,
			relation: map[string]*types.DataType{
				"id":   types.IntType,
				"name": types.TextType,
			},
		},
		{
			name: "invalid compound select",
			stmt: `SELECT name, id FROM users UNION SELECT id, name FROM users;`,
			err:  typing.ErrCompoundShape,
		},
		{
			name: "select table",
			stmt: `SELECT u.* FROM users as u inner join users as u1 on u.id=u1.id;`,
			relation: map[string]*types.DataType{
				"id":   types.IntType,
				"name": types.TextType,
			},
		},
		{
			name: "select *",
			stmt: `SELECT * FROM users;`,
			relation: map[string]*types.DataType{
				"id":   types.IntType,
				"name": types.TextType,
			},
		},
		{
			name: "return unnamed column",
			stmt: `SELECT * from (select 1+1);`,
			err:  errAny,
		},
		{
			name: "aliased column",
			stmt: `SELECT 1+1 as two, (1+3)::text as three;`,
			relation: map[string]*types.DataType{
				"two":   types.IntType,
				"three": types.TextType,
			},
		},
		{
			name: "between",
			stmt: `SELECT * FROM users WHERE id BETWEEN $id AND $id+1;`,
			relation: map[string]*types.DataType{
				"id":   types.IntType,
				"name": types.TextType,
			},
		},
		{
			name: "case",
			stmt: `SELECT CASE WHEN id = $id THEN name ELSE 'default' END as name FROM users;`,
			relation: map[string]*types.DataType{
				"name": types.TextType,
			},
		},
		{
			name: "case with type cast to different types (should fail)",
			stmt: `SELECT CASE name = 'satoshi' WHEN id = $id THEN 1 ELSE 2::text END as name FROM users;`,
			err:  errAny,
		},
		{
			name: "collate",
			stmt: `SELECT * FROM users WHERE name = 'sAtoshi' COLLATE NOCASE;`,
			relation: map[string]*types.DataType{
				"id":   types.IntType,
				"name": types.TextType,
			},
		},
		{
			name: "exists",
			stmt: `SELECT * FROM users WHERE EXISTS (SELECT * FROM posts WHERE author_id = $id);`,
			relation: map[string]*types.DataType{
				"id":   types.IntType,
				"name": types.TextType,
			},
		},
		{
			name: "function",
			stmt: `SELECT * FROM users WHERE id = $id AND name = 'satoshi' AND length(name) = 7;`,
			relation: map[string]*types.DataType{
				"id":   types.IntType,
				"name": types.TextType,
			},
		},
		{
			name: "is null",
			stmt: `SELECT * FROM users WHERE name IS NULL;`,
			relation: map[string]*types.DataType{
				"id":   types.IntType,
				"name": types.TextType,
			},
		},
		{
			name: "select in",
			stmt: `SELECT * FROM users WHERE id IN (1, 2, 3);`,
			relation: map[string]*types.DataType{
				"id":   types.IntType,
				"name": types.TextType,
			},
		},
		{
			name: "boolean literal",
			stmt: `SELECT * FROM users WHERE id = $id AND name = 'satoshi' AND TRUE;`,
			relation: map[string]*types.DataType{
				"id":   types.IntType,
				"name": types.TextType,
			},
		},
		{
			name: "blob literal",
			stmt: `SELECT * FROM users WHERE id = $id AND name = 0x01::text;`,
			relation: map[string]*types.DataType{
				"id":   types.IntType,
				"name": types.TextType,
			},
		},
		{
			name: "string comparison",
			stmt: `SELECT * FROM users WHERE name LIKE 'satoshi%';`,
			relation: map[string]*types.DataType{
				"id":   types.IntType,
				"name": types.TextType,
			},
		},
		{
			name: "string comparison with escape",
			stmt: `SELECT id FROM users WHERE name LIKE 'satoshi\%' ESCAPE '\';`,
			relation: map[string]*types.DataType{
				"id": types.IntType,
			},
		},
		{
			name: "unary operator",
			stmt: `SELECT id FROM users WHERE id = $id AND -id = 1;`,
			relation: map[string]*types.DataType{
				"id": types.IntType,
			},
		},
		{
			name: "extremely complex with CTEs and subqueries",
			stmt: `WITH cte AS (SELECT id, name FROM users WHERE id = $id) SELECT cte.id as userid, ps.title as title FROM cte
			INNER JOIN (SELECT id, title, author_id FROM posts WHERE author_id IN (SELECT users.id FROM users WHERE name = $name)) as ps ON cte.id = ps.author_id;`,
			relation: map[string]*types.DataType{
				"userid": types.IntType,
				"title":  types.TextType,
			},
		},
		{
			name: "correlated subquery",
			stmt: `SELECT id, name FROM users WHERE id = (SELECT id FROM posts WHERE author_id = users.id);`,
			relation: map[string]*types.DataType{
				"id":   types.IntType,
				"name": types.TextType,
			},
		},
		{
			name: "alias used in having clause",
			stmt: `SELECT id, name as username FROM users WHERE id = $id GROUP BY id, username HAVING count(username) > 1;`,
			relation: map[string]*types.DataType{
				"id":       types.IntType,
				"username": types.TextType,
			},
		},
		{
			name: "join with having",
			stmt: `SELECT u1.id, u2.name FROM users AS u1
			INNER JOIN users AS u2 ON u1.id = u2.id
			GROUP BY u1.id, u2.name
			HAVING u1.id = $id AND u2.name = $name;`,
			relation: map[string]*types.DataType{
				"id":   types.IntType,
				"name": types.TextType,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ast, err := sqlparser.Parse(test.stmt)
			require.NoError(t, err)

			rel, err := typing.AnalyzeTypes(ast, []*types.Table{usersTable, postsTable}, &typing.AnalyzeOptions{
				BindParams: bindParams,
			})
			if test.err != nil {
				if errors.Is(test.err, errAny) {
					require.Error(t, err)
					return
				}

				// we are expecting an error
				require.ErrorAs(t, err, &test.err)
				return
			} else {
				require.NoError(t, err)
			}

			returned := make(map[string]*types.DataType)
			err = rel.Loop(func(s string, a *typing.Attribute) error {
				returned[s] = a.Type
				return nil
			})
			require.NoError(t, err)

			require.Equal(t, len(test.relation), len(returned))

			for k, v := range test.relation {
				found, ok := returned[k]
				require.True(t, ok)
				require.True(t, v.Equals(found))
			}
		})
	}
}

// TODO: we should add tables with custom data types to the tests
// this is not supported at the time of writing this test,
// but it will be before release
var (
	bindParams = map[string]*types.DataType{
		"$id":    types.IntType,
		"$name":  types.TextType,
		"$title": types.TextType,
	}

	usersTable = &types.Table{
		Name: "users",
		Columns: []*types.Column{
			{
				Name: "id",
				Type: types.IntType,
			},
			{
				Name: "name",
				Type: types.TextType,
			},
		},
	}

	postsTable = &types.Table{
		Name: "posts",
		Columns: []*types.Column{
			{
				Name: "id",
				Type: types.IntType,
			},
			{
				Name: "title",
				Type: types.TextType,
			},
			{
				Name: "content",
				Type: types.TextType,
			},
			{
				Name: "author_id",
				Type: types.IntType,
			},
		},
	}
)

// special case error for testing
var errAny = errors.New("any error")
