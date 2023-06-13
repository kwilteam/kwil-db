package tree_test

import (
	"testing"

	"github.com/kwilteam/kwil-db/pkg/engine/sqlparser/tree"
)

func TestUpsert_ToSQL(t *testing.T) {
	type fields struct {
		ConflictTarget *tree.ConflictTarget
		Type           tree.UpsertType
		Updates        []*tree.UpdateSetClause
		Where          tree.Expression
	}
	tests := []struct {
		name      string
		fields    fields
		want      string
		wantPanic bool
	}{
		{
			name: "valid upsert",
			fields: fields{
				ConflictTarget: &tree.ConflictTarget{
					IndexedColumns: []string{"barCol", "bazCol"},
					Where: &tree.ExpressionBinaryComparison{
						Left: &tree.ExpressionColumn{
							Column: "barCol",
						},
						Operator: tree.ComparisonOperatorEqual,
						Right: &tree.ExpressionBindParameter{
							Parameter: "$a",
						},
					},
				},
				Type: tree.UpsertTypeDoUpdate,
				Updates: []*tree.UpdateSetClause{
					{
						Columns: []string{"barCol", "bazCol"},
						Expression: &tree.ExpressionBindParameter{
							Parameter: "$b",
						},
					},
				},
				Where: &tree.ExpressionBinaryComparison{
					Left: &tree.ExpressionColumn{
						Column: "bazCol",
					},
					Operator: tree.ComparisonOperatorEqual,
					Right: &tree.ExpressionBindParameter{
						Parameter: "@caller",
					},
				},
			},
			want: `ON CONFLICT ("barCol", "bazCol") WHERE "barCol" = $a DO UPDATE SET ("barCol", "bazCol") = $b WHERE "bazCol" = @caller`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("Upsert.ToSQL() should have panicked")
					}
				}()
			}

			u := &tree.Upsert{
				ConflictTarget: tt.fields.ConflictTarget,
				Type:           tt.fields.Type,
				Updates:        tt.fields.Updates,
				Where:          tt.fields.Where,
			}

			got := u.ToSQL()
			if tt.wantPanic {
				return
			}

			if !compareIgnoringWhitespace(got, tt.want) {
				t.Errorf("Upsert.ToSQL() = %v, want %v", got, tt.want)
			}
		})
	}
}
