package sqlanalyzer_test

import (
	"testing"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/testdata"
	"github.com/kwilteam/kwil-db/parse/sql/postgres"
	"github.com/kwilteam/kwil-db/parse/sql/sqlanalyzer"
	"github.com/stretchr/testify/assert"
)

func Test_Analyze(t *testing.T) {
	type testCase struct {
		name    string
		stmt    string
		want    string
		tables  []*types.Table
		wantErr bool
	}

	tests := []testCase{
		{
			name: "simple select",
			stmt: "SELECT * FROM users",
			want: `SELECT * FROM "dbid"."users" ORDER BY "users"."id";`,
			tables: []*types.Table{
				tblUsers,
			},
		},
		{
			name: "select with joins and subqueries",
			stmt: `SELECT p.id, p.title
				FROM posts AS p
				INNER JOIN followers AS f ON p.user_id = f.user_id
				INNER JOIN users AS u ON u.id = f.user_id
				WHERE f.follower_id = (
					SELECT id FROM users WHERE username = $username
				)
				ORDER BY p.post_date DESC NULLS LAST
				LIMIT 20 OFFSET $offset;`,
			want: `SELECT "p"."id", "p"."title"
				FROM "dbid"."posts" AS "p"
				INNER JOIN "dbid"."followers" AS "f" ON "p"."user_id" = "f"."user_id"
				INNER JOIN "dbid"."users" AS "u" ON "u"."id" = "f"."user_id"
				WHERE "f"."follower_id" = (
					SELECT "id" FROM "dbid"."users" WHERE "username" = $1 ORDER BY "users"."id"
				)
				ORDER BY "p"."post_date" DESC NULLS LAST,
				"f"."follower_id" , "f"."user_id" , "p"."id" , "u"."id"
				LIMIT 20 OFFSET $2;`,
			tables: []*types.Table{
				tblUsers,
				tblPosts,
				tblFollowers,
			},
		},
		{
			name: "table joined on self",
			stmt: `SELECT u1.id, u1.name, u2.name
				FROM users AS u1
				INNER JOIN users AS u2 ON u1.id = u2.id`,
			want: `SELECT "u1"."id", "u1"."name", "u2"."name"
				FROM "dbid"."users" AS "u1"
				INNER JOIN "dbid"."users" AS "u2" ON "u1"."id" = "u2"."id"
				ORDER BY "u1"."id" , "u2"."id";`,
			tables: []*types.Table{
				tblUsers,
			},
		},
		{
			name: "common table expression",
			stmt: `WITH
				users_aged_20 AS (
					SELECT id, username FROM users WHERE age = 20
				)
				SELECT * FROM users_aged_20`,
			want: `WITH
				"users_aged_20" AS (
					SELECT "id", "username" FROM "dbid"."users" WHERE "age" = 20 ORDER BY "users"."id"
				)
				SELECT * FROM "users_aged_20" ORDER BY "users_aged_20"."id" , "users_aged_20"."username";`,
			tables: []*types.Table{
				tblUsers,
			},
		},
		{
			name: "basic insert",
			stmt: `INSERT INTO users (id, username, age) VALUES (1, 'user1', 20)`,
			want: `INSERT INTO "dbid"."users" ("id", "username", "age") VALUES (1, 'user1', 20);`,
			tables: []*types.Table{
				tblUsers,
			},
		},
		{
			name: "select with count",
			stmt: `SELECT COUNT(*) FROM users`,
			want: `SELECT count(*) FROM "dbid"."users";`,
			tables: []*types.Table{
				tblUsers,
			},
		},
		{
			name: "select with aggregate and group by",
			stmt: `SELECT sum(age) FROM users GROUP BY username;`,
			want: `SELECT sum("age") FROM "dbid"."users" GROUP BY "username" ORDER BY "username";`,
			tables: []*types.Table{
				tblUsers,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sqlanalyzer.ApplyRules(tt.stmt, sqlanalyzer.AllRules, &types.Schema{
				Tables: tt.tables,
			}, "dbid")
			if (err != nil) != tt.wantErr {
				t.Errorf("ApplyRules() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Equal(t, removeSpaces(tt.want), removeSpaces(got.Statement))

			err = postgres.CheckSyntaxReplaceDollar(got.Statement)
			assert.NoErrorf(t, err, "postgres syntax check failed: %s", err)
		})
	}
}

var (
	tblUsers = testdata.TableUsers

	tblPosts = &types.Table{
		Name: "posts",
		Columns: []*types.Column{
			{
				Name: "id",
				Type: types.IntType,
				Attributes: []*types.Attribute{
					{
						Type: types.PRIMARY_KEY,
					},
				},
			},
			{
				Name: "user_id",
				Type: types.IntType,
				Attributes: []*types.Attribute{
					{
						Type: types.NOT_NULL,
					},
				},
			},
			{
				Name: "title",
				Type: types.TextType,
			},
			{
				Name: "post_date",
				Type: types.IntType,
				Attributes: []*types.Attribute{
					{
						Type: types.NOT_NULL,
					},
				},
			},
		},
	}

	tblFollowers = &types.Table{
		Name: "followers",
		Columns: []*types.Column{
			{
				Name: "user_id",
				Type: types.IntType,
			},
			{
				Name: "follower_id",
				Type: types.IntType,
			},
		},
		Indexes: []*types.Index{
			{
				Name: "primary_key",
				Columns: []string{
					"user_id",
					"follower_id",
				},
				Type: types.PRIMARY,
			},
		},
	}
)

// removeSpaces removes all spaces from a string.
// this is useful for comparing strings, where one is generated
func removeSpaces(s string) string {
	var result []rune
	for _, ch := range s {
		if ch != ' ' && ch != '\n' && ch != '\r' && ch != '\t' {
			result = append(result, ch)
		}
	}
	return string(result)
}
