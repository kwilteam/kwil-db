package tree_test

import (
	"fmt"
	"kwil/pkg/engine/tree"
	"testing"
)

func Test_Insert(t *testing.T) {
	ins := tree.InsertStatement{
		Table: "foo",
		Columns: []string{
			"bar",
			"baz",
		},
		Values: [][]tree.InsertExpression{
			[]tree.InsertExpression{
				&tree.ExpressionLiteral{"barVal"},
				&tree.ExpressionBindParameter{"$a"},
			},
		},
		Upsert: &tree.Upsert{
			ConflictTargetColumn: "bar",
			Type:                 tree.UpsertTypeDoUpdate,
			Set: map[string]tree.Expression{
				"baz": &tree.ExpressionBindParameter{"$b"},
			},
			Where: &tree.WhereClause{
				Expression: &tree.ExpressionBinaryComparison{},
			},
		},
	}

	sql, args, err := ins.ToSql()
	if err != nil {
		t.Error(err)
	}

	fmt.Println(sql)
	fmt.Println(args)

	panic("")
}
