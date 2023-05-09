package tree_test

import (
	"kwil/pkg/engine/tree"
	"testing"
)

func TestGroupBy_ToSQL(t *testing.T) {
	type fields struct {
		Expressions []tree.Expression
		Having      tree.Expression
	}
	tests := []struct {
		name      string
		fields    fields
		want      string
		wantPanic bool
	}{
		{
			name: "valid group by",
			fields: fields{
				Expressions: []tree.Expression{
					&tree.ExpressionColumn{Column: "foo"},
				},
			},
			want: ` GROUP BY "foo"`,
		},
		{
			name: "valid group by with having",
			fields: fields{
				Expressions: []tree.Expression{
					&tree.ExpressionColumn{Column: "foo"},
					&tree.ExpressionColumn{Column: "bar"},
				},
				Having: &tree.ExpressionBinaryComparison{
					Left:     &tree.ExpressionColumn{Column: "foo"},
					Operator: tree.ComparisonOperatorGreaterThan,
					Right:    &tree.ExpressionBindParameter{Parameter: "$a"},
				},
			},
			want: ` GROUP BY "foo", "bar" HAVING "foo" > $a`,
		},
		{
			name: "no expressions",
			fields: fields{
				Expressions: []tree.Expression{},
			},
			wantPanic: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &tree.GroupBy{
				Expressions: tt.fields.Expressions,
				Having:      tt.fields.Having,
			}
			if tt.wantPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("GroupBy.ToSQL() should have panicked")
					}
				}()
			}

			got := g.ToSQL()
			if tt.wantPanic {
				return
			}

			if !compareIgnoringWhitespace(got, tt.want) {
				t.Errorf("GroupBy.ToSQL() = %v, want %v", got, tt.want)
			}
		})
	}
}
