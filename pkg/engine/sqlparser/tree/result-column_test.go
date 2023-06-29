package tree_test

import (
	"testing"

	"github.com/kwilteam/kwil-db/pkg/engine/sqlparser/tree"
)

func TestResultColumnStar_ToSQL(t *testing.T) {
	tests := []struct {
		name string
		r    tree.ResultColumn
		want string
	}{
		{
			name: "valid star",
			r:    &tree.ResultColumnStar{},
			want: "*",
		},
		{
			name: "expression with alias",
			r: &tree.ResultColumnExpression{
				Expression: &tree.ExpressionColumn{Column: "foo"},
				Alias:      "f",
			},
			want: `"foo" AS "f"`,
		},
		{
			name: "expression without alias",
			r: &tree.ResultColumnExpression{
				Expression: &tree.ExpressionColumn{Column: "foo"},
			},
			want: `"foo"`,
		},
		{
			name: "table",
			r: &tree.ResultColumnTable{
				TableName: "foo",
			},
			want: `"foo".*`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := tt.r

			got := r.ToSQL()

			if !compareIgnoringWhitespace(got, tt.want) {
				t.Errorf("ResultColumnStar.ToSQL() = %v, want %v", got, tt.want)
			}
		})
	}
}
