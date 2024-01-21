package tree_test

import (
	"testing"

	"github.com/kwilteam/kwil-db/parse/sql/tree"
)

func TestUpdate_ToSQL(t *testing.T) {
	type fields struct {
		CTE       []*tree.CTE
		Statement *tree.UpdateStmt
	}
	tests := []struct {
		name    string
		fields  fields
		wantStr string
		wantErr bool
	}{
		{
			name: "valid update",
			fields: fields{
				CTE: []*tree.CTE{
					mockCTE,
				},
				Statement: &tree.UpdateStmt{
					Or: tree.UpdateOrAbort,
					QualifiedTableName: &tree.QualifiedTableName{
						TableName:  "foo",
						TableAlias: "f",
						NotIndexed: true,
					},
					UpdateSetClause: []*tree.UpdateSetClause{
						{
							Columns: []string{"bar", "baz"},
							Expression: &tree.ExpressionSelect{
								IsNot:    true,
								IsExists: true,
								Select: &tree.SelectStmt{
									SelectCores: []*tree.SelectCore{
										{
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
										},
									},
								},
							},
						},
					},
					Where: &tree.ExpressionBinaryComparison{
						Left: &tree.ExpressionColumn{
							Column: "foo",
						},
						Operator: tree.ComparisonOperatorEqual,
						Right: &tree.ExpressionBindParameter{
							Parameter: "$a",
						},
					},
					Returning: &tree.ReturningClause{
						Returned: []*tree.ReturningClauseColumn{
							{
								Expression: &tree.ExpressionColumn{Column: "foo"},
								Alias:      "fu",
							},
						},
					},
				},
			},
			wantStr: `WITH ` + mockCTE.ToSQL() + ` UPDATE OR ABORT "foo" AS "f" NOT INDEXED SET ("bar", "baz") = NOT EXISTS (SELECT "foo", "bar" FROM "foo") WHERE "foo" = $a RETURNING "foo" AS "fu";`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &tree.Update{
				CTE:        tt.fields.CTE,
				UpdateStmt: tt.fields.Statement,
			}
			gotStr, err := tree.SafeToSQL(u)
			if (err != nil) != tt.wantErr {
				t.Errorf("Update.ToSQL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !compareIgnoringWhitespace(gotStr, tt.wantStr) {
				t.Errorf("Update.ToSQL() = %v, want %v", gotStr, tt.wantStr)
			}
		})
	}
}
