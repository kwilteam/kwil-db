package tree_test

import (
	"testing"

	"github.com/kwilteam/kwil-db/parse/sql/tree"
)

func TestSelect_ToSQL(t *testing.T) {
	type fields struct {
		CTE        []*tree.CTE
		SelectStmt *tree.SelectStmt
	}
	tests := []struct {
		name    string
		fields  fields
		wantStr string
		wantErr bool
	}{
		{
			name: "valid select",
			fields: fields{
				CTE: []*tree.CTE{
					mockCTE,
				},
				SelectStmt: &tree.SelectStmt{
					SelectCores: []*tree.SelectCore{{
						SelectType: tree.SelectTypeAll,
						Columns: []tree.ResultColumn{
							&tree.ResultColumnExpression{Expression: &tree.ExpressionColumn{Column: "foo"}},
							&tree.ResultColumnExpression{Expression: &tree.ExpressionColumn{Column: "bar"}},
						},
						From: &tree.FromClause{
							JoinClause: &tree.JoinClause{
								TableOrSubquery: &tree.TableOrSubqueryTable{
									Name: "foo",
								},
							},
						},
						Where: &tree.ExpressionBinaryComparison{
							Left:     &tree.ExpressionColumn{Column: "foo"},
							Operator: tree.ComparisonOperatorEqual,
							Right:    &tree.ExpressionBindParameter{Parameter: "$a"},
						},
						GroupBy: &tree.GroupBy{
							Expressions: []tree.Expression{
								&tree.ExpressionColumn{Column: "foo"},
								&tree.ExpressionColumn{Column: "bar"},
							},
							Having: &tree.ExpressionBinaryComparison{
								Left:     &tree.ExpressionColumn{Column: "foo"},
								Operator: tree.ComparisonOperatorEqual,
								Right:    &tree.ExpressionBindParameter{Parameter: "$b"},
							},
						},
					}},
					OrderBy: &tree.OrderBy{
						OrderingTerms: []*tree.OrderingTerm{
							{
								Expression: &tree.ExpressionColumn{Column: "foo"},
								Collation:  tree.CollationTypeNoCase,
							},
						},
					},
					Limit: &tree.Limit{
						Expression: &tree.ExpressionBindParameter{Parameter: "$c"},
						Offset:     &tree.ExpressionBindParameter{Parameter: "$d"},
					},
				},
			},
			wantStr: `WITH ` + mockCTE.ToSQL() + ` SELECT "foo", "bar" FROM "foo" WHERE "foo" = $a GROUP BY "foo", "bar" HAVING "foo" = $b ORDER BY "foo" COLLATE NOCASE LIMIT $c OFFSET $d;`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &tree.Select{
				CTE:        tt.fields.CTE,
				SelectStmt: tt.fields.SelectStmt,
			}
			gotStr, err := tree.SafeToSQL(s)
			if (err != nil) != tt.wantErr {
				t.Errorf("Select.ToSQL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !compareIgnoringWhitespace(gotStr, tt.wantStr) {
				t.Errorf("Select.ToSQL() = %v, want %v", gotStr, tt.wantStr)
			}
		})
	}
}
