package tree_test

import (
	"testing"

	"github.com/kwilteam/kwil-db/pkg/engine/sqlparser/tree"
)

func TestReturningClause_ToSQL(t *testing.T) {
	type fields struct {
		Returned []*tree.ReturningClauseColumn
	}
	tests := []struct {
		name      string
		fields    fields
		want      string
		wantPanic bool
	}{
		{
			name: "valid returning clause",
			fields: fields{
				Returned: []*tree.ReturningClauseColumn{
					{
						Expression: &tree.ExpressionColumn{Column: "foo"},
						Alias:      "f",
					},
					{
						Expression: &tree.ExpressionBindParameter{Parameter: "$a"},
					},
				},
			},
			want: ` RETURNING "foo" AS "f", $a`,
		},
		{
			name: "returns all",
			fields: fields{
				Returned: []*tree.ReturningClauseColumn{
					{
						All: true,
					},
				},
			},
			want: ` RETURNING *`,
		},
		{
			name: "contains expression and all",
			fields: fields{
				Returned: []*tree.ReturningClauseColumn{
					{
						Expression: &tree.ExpressionColumn{Column: "foo"},
						Alias:      "f",
						All:        true,
					},
				},
			},
			wantPanic: true,
		},
		{
			name:      "contains no columns",
			fields:    fields{},
			wantPanic: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("ReturningClause.ToSQL() should have panicked")
					}
				}()
			}

			r := &tree.ReturningClause{
				Returned: tt.fields.Returned,
			}

			got := r.ToSQL()
			if tt.wantPanic {
				return
			}

			if !compareIgnoringWhitespace(got, tt.want) {
				t.Errorf("ReturningClause.ToSQL() = %v, want %v", got, tt.want)
			}
		})
	}
}
