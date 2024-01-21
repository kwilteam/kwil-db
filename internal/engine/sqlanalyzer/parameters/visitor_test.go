package parameters_test

import (
	"strings"
	"testing"
	"unicode"

	"github.com/kwilteam/kwil-db/internal/engine/sqlanalyzer/parameters"
	sqlparser "github.com/kwilteam/kwil-db/parse/sql"
	"github.com/kwilteam/kwil-db/parse/sql/postgres"
	"github.com/kwilteam/kwil-db/parse/sql/tree"
	"github.com/stretchr/testify/assert"
)

func Test_NumberedParameters(t *testing.T) {
	type testCase struct {
		name       string
		stmt       string
		wantParams []string
		wantStmt   string
	}

	tests := []testCase{
		{
			name:       "simple select",
			stmt:       `SELECT * FROM "table" WHERE "id" = $id;`,
			wantParams: []string{"$id"},
			wantStmt:   `SELECT * FROM "table" WHERE "id" = $1;`,
		},
		{
			name:       "simple select with multiple parameters",
			stmt:       `SELECT * FROM "table" WHERE "id" = $id AND "name" = $name;`,
			wantParams: []string{"$id", "$name"},
			wantStmt:   `SELECT * FROM "table" WHERE "id" = $1 AND "name" = $2;`,
		},
		{
			name: "repeating parameters",
			stmt: `SELECT * FROM "table" WHERE "id" = $id AND "name" = $id AND "age" = $name AND "address" = $id;`,
			wantParams: []string{
				"$id",
				"$name",
			},
			wantStmt: `SELECT * FROM "table" WHERE "id" = $1 AND "name" = $1 AND "age" = $2 AND "address" = $1;`,
		},
		{
			name: "@ binding",
			stmt: `SELECT * FROM "table" WHERE "id" = @id AND "name" = @caller;`,
			wantParams: []string{
				"@id",
				"@caller",
			},
			wantStmt: `SELECT * FROM "table" WHERE "id" = $1 AND "name" = $2;`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := sqlparser.Parse(tt.stmt)
			if err != nil {
				t.Errorf("Parameters() = %v, want %v", err, tt.wantParams)
			}

			v := parameters.NewParametersWalker()

			if err := ast.Walk(v); err != nil {
				t.Errorf("Parameters() = %v, want %v", err, tt.wantParams)
			}

			got := v.OrderedParameters

			if len(got) != len(tt.wantParams) {
				t.Errorf("Parameters() = %v, want %v", got, tt.wantParams)
			}

			for i := range got {
				if got[i] != tt.wantParams[i] {
					t.Errorf("Parameters() = %v, want %v", got, tt.wantParams)
				}
			}

			str, err := tree.SafeToSQL(ast)
			if err != nil {
				t.Errorf("Parameters() = %v, want %v", err, tt.wantParams)
			}
			trimmedRes := removeWhitespace(str)
			trimmedWant := removeWhitespace(tt.wantStmt)

			if trimmedRes != trimmedWant {
				t.Errorf("Parameters() = %v, want %v", trimmedRes, trimmedWant)
			}

			err = postgres.CheckSyntaxReplaceDollar(str)
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
