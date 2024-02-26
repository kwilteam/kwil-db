package tree_test

import (
	"testing"

	"github.com/kwilteam/kwil-db/pkg/engine/sqlparser/tree"
)

func TestOrderBy_ToSQL(t *testing.T) {
	type fields struct {
		OrderingTerms []*tree.OrderingTerm
	}
	tests := []struct {
		name      string
		fields    fields
		want      string
		wantPanic bool
	}{
		{
			name: "valid order by with multiple terms",
			fields: fields{
				OrderingTerms: []*tree.OrderingTerm{
					{
						Expression: &tree.ExpressionColumn{Column: "foo"},
					},
					{
						Expression: &tree.ExpressionBinaryComparison{
							Left:     &tree.ExpressionColumn{Column: "bar"},
							Operator: tree.ComparisonOperatorEqual,
							Right:    &tree.ExpressionBindParameter{Parameter: "$a"},
						},
						Collation:    tree.CollationTypeNoCase,
						OrderType:    tree.OrderTypeDesc,
						NullOrdering: tree.NullOrderingTypeFirst,
					},
				},
			},
			want: ` ORDER BY "foo", "bar" = $a COLLATE NOCASE DESC NULLS FIRST`,
		},
		{
			name:      "no ordering terms",
			fields:    fields{},
			wantPanic: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &tree.OrderBy{
				OrderingTerms: tt.fields.OrderingTerms,
			}

			if tt.wantPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("OrderBy.ToSQL() should have panicked")
					}
				}()
			}

			got := o.ToSQL()
			if tt.wantPanic {
				return
			}

			if !compareIgnoringWhitespace(got, tt.want) {
				t.Errorf("OrderBy.ToSQL() = %v, want %v", got, tt.want)
			}
		})
	}
}
