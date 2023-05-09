package tree_test

import (
	"fmt"
	"kwil/pkg/engine/tree"
	"testing"
)

func Test_Insert(t *testing.T) {
	ins := tree.Insert{
		Table: "foo",
		Columns: []string{
			"bar",
			"baz",
		},
		Values: [][]tree.InsertExpression{
			{
				&tree.ExpressionLiteral{"barVal"},
				&tree.ExpressionBindParameter{"$a"},
			},
			{
				&tree.ExpressionLiteral{"bazVal"},
				&tree.ExpressionBindParameter{"$b"},
			},
		},
		Upsert: &tree.Upsert{
			ConflictTarget: &tree.ConflictTarget{
				IndexedColumns: []*tree.IndexedColumn{
					{
						Column: "bar",
					},
				},
				Where: &tree.ExpressionBinaryComparison{
					Left:     &tree.ExpressionColumn{Column: "baz"},
					Operator: tree.ComparisonOperatorEqual,
					Right:    &tree.ExpressionBindParameter{"$c"},
				},
			},
			Type: tree.UpsertTypeDoUpdate,
			Updates: []*tree.UpdateSetClause{
				{
					Columns:    []string{"bar"},
					Expression: &tree.ExpressionLiteral{"5"},
				},
				{
					Columns:    []string{"baz"},
					Expression: &tree.ExpressionBindParameter{"$d"},
				},
			},
			Where: &tree.ExpressionBinaryComparison{
				Left:     &tree.ExpressionColumn{Column: "baz"},
				Operator: tree.ComparisonOperatorEqual,
				Right:    &tree.ExpressionBindParameter{"$c"},
			},
		},
		ReturningClause: &tree.ReturningClause{
			Returned: []*tree.ReturningClauseColumn{
				{
					All: true,
				},
			},
		},
	}

	sql, err := ins.ToSql()
	if err != nil {
		t.Error(err)
	}

	fmt.Println(sql)
}
