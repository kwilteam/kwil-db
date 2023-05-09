package tree_test

import (
	"github.com/kwilteam/kwil-db/pkg/engine/tree"
	"testing"
)

func TestConflictTarget_ToSQL(t *testing.T) {
	type fields struct {
		ConflictTarget *tree.ConflictTarget
	}
	tests := []struct {
		name      string
		fields    fields
		want      string
		wantPanic bool
	}{
		{
			name: "valid conflict target",
			fields: fields{
				ConflictTarget: &tree.ConflictTarget{
					IndexedColumns: []*tree.IndexedColumn{
						{
							Column: "bar",
						},
						{
							Expression: &tree.ExpressionBinaryComparison{
								Left:     &tree.ExpressionColumn{Column: "baz"},
								Operator: tree.ComparisonOperatorEqual,
								Right:    &tree.ExpressionBindParameter{Parameter: "$c"},
							},
							Collation: "collation",
							OrderType: tree.OrderTypeDesc,
						},
					},
					Where: &tree.ExpressionBinaryComparison{
						Left: &tree.ExpressionExpressionList{
							Expressions: []tree.Expression{
								&tree.ExpressionBinaryComparison{
									Left:     &tree.ExpressionColumn{Column: "baz"},
									Operator: tree.ComparisonOperatorEqual,
									Right:    &tree.ExpressionBindParameter{Parameter: "$c"},
								},
							},
						},
						Operator: tree.ComparisonOperatorEqual,
						Right:    &tree.ExpressionBindParameter{Parameter: "$d"},
					},
				},
			},
			want: `("bar", ("baz" = $c) COLLATE collation DESC) WHERE ("baz" = $c) = $d`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := tt.fields.ConflictTarget

			if tt.wantPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("ConflictTarget.String() should have panicked")
					}
				}()
			}

			got := c.ToSQL()
			if tt.wantPanic {
				return
			}

			if !compareIgnoringWhitespace(got, tt.want) {
				t.Errorf("ConflictTarget.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

/*
func TestConflictTarget_ToSQL(t *testing.T) {
	type fields struct {
		IndexedColumns []*tree.IndexedColumn
		Where          tree.Expression
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "valid conflict target",
			fields: fields{
				IndexedColumns: []*tree.IndexedColumn{
					{
						Column: "bar",
					},
				},
				Where: &tree.ExpressionBinaryComparison{
					Left: &tree.ExpressionBinaryComparison{
						Left:     &tree.ExpressionColumn{Column: "baz"},
						Operator: tree.ComparisonOperatorEqual,
						Right:    &tree.ExpressionBindParameter{Parameter: "$c"},
					},
				},
			},
			want: `("bar") WHERE ("baz" = $c)`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &tree.ConflictTarget{
				IndexedColumns: tt.fields.IndexedColumns,
				Where:          tt.fields.Where,
			}
			if got := c.ToSQL(); got != tt.want {
				t.Errorf("ConflictTarget.ToSQL() = %v, want %v", got, tt.want)
			}
		})
	}
}
*/
