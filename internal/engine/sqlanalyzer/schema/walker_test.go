package schema_test

import (
	"strings"
	"testing"
	"unicode"

	"github.com/kwilteam/kwil-db/internal/engine/sqlanalyzer/schema"
	sqlparser "github.com/kwilteam/kwil-db/parse/sql"
	"github.com/kwilteam/kwil-db/parse/sql/postgres"
	"github.com/kwilteam/kwil-db/parse/sql/tree"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_PGSchemas(t *testing.T) {
	type testcase struct {
		name   string
		stmt   string
		schema string
		want   string
	}

	tests := []testcase{
		{
			name:   "basic select",
			stmt:   `SELECT * FROM "foo";`,
			want:   `SELECT * FROM "baz"."foo";`,
			schema: "baz",
		},
		{
			name:   "insert",
			stmt:   `INSERT INTO "foo" ("bar", "baz") VALUES ('barVal', $a);`,
			want:   `INSERT INTO "baz"."foo" ("bar", "baz") VALUES ('barVal', $a);`,
			schema: "baz",
		},
		{
			name:   "update",
			stmt:   `UPDATE "foo" SET "bar" = 'barVal' WHERE "baz" = $a;`,
			want:   `UPDATE "baz"."foo" SET "bar" = 'barVal' WHERE "baz" = $a;`,
			schema: "baz",
		},
		{
			name:   "delete",
			stmt:   `DELETE FROM "foo" WHERE "bar" = 'barVal';`,
			want:   `DELETE FROM "baz"."foo" WHERE "bar" = 'barVal';`,
			schema: "baz",
		},
		{
			name:   "common table expression",
			stmt:   `WITH "cte" AS (SELECT * FROM "foo") SELECT * FROM "cte";`,
			want:   `WITH "cte" AS (SELECT * FROM "baz"."foo") SELECT * FROM "cte";`,
			schema: "baz",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := sqlparser.Parse(tt.stmt)
			require.NoError(t, err)

			w := schema.NewSchemaWalker(tt.schema)

			err = ast.Walk(w)
			require.NoError(t, err)

			got, err := tree.SafeToSQL(ast)
			require.NoError(t, err)

			require.Equal(t, removeWhitespace(tt.want), removeWhitespace(got))

			err = postgres.CheckSyntaxReplaceDollar(got)
			assert.NoErrorf(t, err, "postgres syntax check failed: %s", err)
		})
	}
}

// removeWhitespace removes all whitespace characters from a string.
func removeWhitespace(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1 // skip this rune
		}
		return r
	}, s)
}
