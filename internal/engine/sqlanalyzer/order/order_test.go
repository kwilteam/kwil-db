package order_test

import (
	"errors"
	"testing"

	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/internal/engine/sqlanalyzer/order"
	sqlparser "github.com/kwilteam/kwil-db/parse/sql"
	"github.com/kwilteam/kwil-db/parse/sql/postgres"
	"github.com/kwilteam/kwil-db/parse/sql/tree"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Order(t *testing.T) {
	type testcase struct {
		name string
		stmt string
		want string
		err  error // can be nil
	}

	tests := []testcase{
		{
			name: "simple select",
			stmt: `SELECT * FROM users;`,
			want: `SELECT * FROM "users" ORDER BY "users"."id";`,
		},
		{
			name: "table joined on self",
			stmt: `SELECT u1.id, u1.name, u2.name
			FROM users AS u1
			INNER JOIN users AS u2 ON u1.id = u2.id`,
			want: `SELECT "u1"."id", "u1"."name", "u2"."name"
			FROM "users" AS "u1"
			INNER JOIN "users" AS "u2" ON "u1"."id" = "u2"."id"
			ORDER BY "u1"."id" , "u2"."id";`,
		},
		{
			name: "aliased columns",
			stmt: `SELECT u.id AS user_id, u.name AS user_name FROM users AS u;`,
			want: `SELECT "u"."id" AS "user_id", "u"."name" AS "user_name" FROM "users" AS "u" ORDER BY "u"."id";`,
		},
		{
			name: "select count",
			stmt: `SELECT COUNT(*) FROM users;`,
			want: `SELECT count(*) FROM "users";`,
		},
		{
			name: "select with joins and aliases",
			stmt: `SELECT * FROM users AS u INNER JOIN posts AS p ON u.id = p.user_id;`,
			want: `SELECT * FROM "users" AS "u" INNER JOIN "posts" AS "p" ON "u"."id" = "p"."user_id" ORDER BY "p"."id", "u"."id";`,
		},
		{
			name: "select distinct",
			stmt: `SELECT DISTINCT id, name FROM users;`,
			want: `SELECT DISTINCT "id", "name" FROM "users" ORDER BY "id", "name";`,
		},
		{
			name: "select distinct with join",
			stmt: `SELECT DISTINCT u.id, p.title FROM users AS u INNER JOIN posts AS p ON u.id = p.user_id;`,
			want: `SELECT DISTINCT "u"."id", "p"."title" FROM "users" AS "u" INNER JOIN "posts" AS "p" ON "u"."id" = "p"."user_id" ORDER BY "u"."id", "p"."title";`,
		},
		{
			name: "SELECT DISTINCT with group by",
			stmt: `SELECT DISTINCT id, name FROM users GROUP BY id;`,
			err:  order.ErrDistinctWithGroupBy,
		},
		{
			name: "select distinct with self join and aliases",
			stmt: `SELECT DISTINCT u1.* FROM users AS u1 INNER JOIN users AS u2 ON u1.id = u2.id;`,
			want: `SELECT DISTINCT "u1".* FROM "users" AS "u1" INNER JOIN "users" AS "u2" ON "u1"."id" = "u2"."id" ORDER BY "u1"."id", "u1"."name";`,
		},
		{
			name: "select distinct * with self join and aliases",
			stmt: `SELECT DISTINCT * FROM users AS u1 INNER JOIN users AS u2 ON u1.id = u2.id;`,
			want: `SELECT DISTINCT * FROM "users" AS "u1" INNER JOIN "users" AS "u2" ON "u1"."id" = "u2"."id" ORDER BY "u1"."id", "u1"."name", "u2"."id", "u2"."name";`,
		},
		{
			name: "select with joins and subqueries", // it should not register the subquery as a table
			stmt: `SELECT p.id, p.title, (SELECT COUNT(*) FROM likes WHERE likes.post_id = p.id) AS total_likes
			FROM posts AS p
			INNER JOIN followers AS f ON p.user_id = f.user_id
			INNER JOIN users ON users.id = f.user_id
			WHERE f.follower_id = (
				SELECT liker_id from likes WHERE likes.post_id = p.id
			)
			ORDER BY p.post_date DESC NULLS LAST
			LIMIT 20 OFFSET $offset;`,
			want: `SELECT "p"."id", "p"."title", (SELECT count(*) FROM "likes" WHERE "likes"."post_id" = "p"."id") AS "total_likes"
			FROM "posts" AS "p"
			INNER JOIN "followers" AS "f" ON "p"."user_id" = "f"."user_id"
			INNER JOIN "users" ON "users"."id" = "f"."user_id"
			WHERE "f"."follower_id" = (
				SELECT "liker_id" FROM "likes" WHERE "likes"."post_id" = "p"."id" ORDER BY "likes"."liker_id", "likes"."post_id"
			)
			ORDER BY "p"."post_date" DESC NULLS LAST, "f"."follower_id", "f"."user_id", "p"."id", "users"."id"
			LIMIT 20 OFFSET $offset;`,
		},
		{
			name: "compound select",
			stmt: `SELECT id, name FROM users UNION SELECT id, name FROM users;`,
			want: `SELECT "id", "name" FROM "users" UNION SELECT "id", "name" FROM "users" ORDER BY "id", "name";`,
		},
		{
			name: "compound select with incompatible tables",
			stmt: `SELECT id, name FROM users UNION SELECT * FROM posts;`,
			err:  order.ErrCompoundStatementDifferentNumberOfColumns,
		},
		{
			name: "compound with group by",
			stmt: `SELECT id, name FROM users GROUP BY id UNION SELECT id, name FROM users;`,
			err:  order.ErrGroupByWithCompoundStatement,
		},
		{
			name: "common table expression",
			stmt: `WITH
				user_likes_count AS (
					SELECT liker_id as user_id, COUNT(*) AS likes_count FROM likes GROUP BY liker_id
				)
				SELECT u.id, u.name, ulc.likes_count
				FROM users AS u
				LEFT JOIN user_likes_count AS ulc ON u.id = ulc.user_id;`,
			want: `WITH
				"user_likes_count" AS (
					SELECT "likes"."liker_id" AS "user_id", count(*) AS "likes_count" FROM "likes" GROUP BY "liker_id" ORDER BY "liker_id"
				)
				SELECT "u"."id", "u"."name", "ulc"."likes_count"
				FROM "users" AS "u"
				LEFT JOIN "user_likes_count" AS "ulc" ON "u"."id" = "ulc"."user_id" ORDER BY "u"."id", "ulc"."likes_count", "ulc"."user_id";`,
			// order u ahead of ulc because we order based on alias
		},
		{
			name: "raw select",
			stmt: `SELECT $id AS result`,
			want: `SELECT $id AS "result";`,
		},
		{
			name: "joined subquery",
			stmt: `SELECT u.id, subq.total_likes
			FROM users AS u
			INNER JOIN (
				SELECT post_id, COUNT(*) AS total_likes FROM likes GROUP BY post_id
			) AS subq ON u.id = subq.post_id;`,
			want: `SELECT "u"."id", "subq"."total_likes"
			FROM "users" AS "u"
			INNER JOIN (
				SELECT "likes"."post_id", count(*) AS "total_likes" FROM "likes" GROUP BY "post_id" ORDER BY "post_id"
			) AS "subq" ON "u"."id" = "subq"."post_id" ORDER BY "subq"."post_id", "subq"."total_likes", "u"."id";`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stmt, err := sqlparser.Parse(tt.stmt)
			require.NoError(t, err)

			walker := order.NewOrderWalker(defaultTables)
			err = stmt.Walk(walker)

			if err != nil {
				require.True(t, errors.Is(err, tt.err))
				return
			}
			require.Equal(t, tt.err, err)

			sql, err := tree.SafeToSQL(stmt)
			require.NoError(t, err)

			assert.Equal(t, removeSpaces(tt.want), removeSpaces(sql))

			err = postgres.CheckSyntaxReplaceDollar(sql)
			assert.NoErrorf(t, err, "postgres syntax check failed: %s", err)
		})
	}
}

var defaultTables = []*common.Table{
	{
		Name: "users",
		Columns: []*common.Column{
			{
				Name: "id",
				Type: common.INT,
				Attributes: []*common.Attribute{
					{
						Type: common.PRIMARY_KEY,
					},
				},
			},
			{
				Name: "name",
				Type: common.TEXT,
			},
		},
		Indexes:     []*common.Index{},
		ForeignKeys: []*common.ForeignKey{},
	},
	{
		Name: "posts",
		Columns: []*common.Column{
			{
				Name: "id",
				Type: common.INT,
				Attributes: []*common.Attribute{
					{
						Type: common.PRIMARY_KEY,
					},
				},
			},
			{
				Name: "user_id",
				Type: common.INT,
				Attributes: []*common.Attribute{
					{
						Type: common.NOT_NULL,
					},
				},
			},
			{
				Name: "title",
				Type: common.TEXT,
			},
		},
	},
	{
		Name: "followers",
		Columns: []*common.Column{
			{
				Name: "user_id",
				Type: common.INT,
			},
			{
				Name: "follower_id",
				Type: common.INT,
			},
		},
		Indexes: []*common.Index{
			{
				Name: "primary_key",
				Columns: []string{
					"user_id",
					"follower_id",
				},
				Type: common.PRIMARY,
			},
		},
	},
	{
		// likes is a join table for liker id and post id
		Name: "likes",
		Columns: []*common.Column{
			{
				Name: "liker_id",
				Type: common.INT,
			},
			{
				Name: "post_id",
				Type: common.INT,
			},
		},
		Indexes: []*common.Index{
			{
				Name: "primary_key",
				Columns: []string{
					"liker_id",
					"post_id",
				},
				Type: common.PRIMARY,
			},
		},
	},
}

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
