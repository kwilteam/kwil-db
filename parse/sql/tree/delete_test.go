package tree_test

import (
	"testing"

	"github.com/kwilteam/kwil-db/parse/sql/tree"
)

func TestDelete_ToSQL(t *testing.T) {
	type fields struct {
		CTE        []*tree.CTE
		DeleteStmt *tree.DeleteStmt
	}
	tests := []struct {
		name    string
		fields  fields
		wantStr string
		wantErr bool
	}{
		{
			name: "valid delete",
			fields: fields{
				CTE: []*tree.CTE{
					mockCTE,
				},
				DeleteStmt: &tree.DeleteStmt{
					QualifiedTableName: &tree.QualifiedTableName{
						TableName: "foo",
					},
					Where: &tree.ExpressionBinaryComparison{
						Left:     &tree.ExpressionColumn{Column: "foo"},
						Operator: tree.ComparisonOperatorEqual,
						Right:    &tree.ExpressionBindParameter{Parameter: "$a"},
					},
					Returning: &tree.ReturningClause{
						Returned: []*tree.ReturningClauseColumn{
							{
								All: true,
							},
						},
					},
				},
			},
			wantStr: `WITH ` + mockCTE.ToSQL() + ` DELETE FROM "foo" WHERE "foo" = $a RETURNING *;`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &tree.Delete{
				CTE:        tt.fields.CTE,
				DeleteStmt: tt.fields.DeleteStmt,
			}
			gotStr, err := tree.SafeToSQL(d)
			if (err != nil) != tt.wantErr {
				t.Errorf("Delete.ToSQL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !compareIgnoringWhitespace(gotStr, tt.wantStr) {
				t.Errorf("Delete.ToSQL() = %v, want %v", gotStr, tt.wantStr)
			}
		})
	}
}
