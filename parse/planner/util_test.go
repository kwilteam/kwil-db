package planner

import (
	"fmt"
	"testing"
)

func Test_RewriteExpr(t *testing.T) {
	tree := &Scan{
		Source: &ProcedureScanSource{
			ProcedureName: "peer",
			Args: []LogicalExpr{
				&Literal{Value: 1},
				&AggregateFunctionCall{
					FunctionName: "sum",
					Args: []LogicalExpr{
						&ArithmeticOp{
							Left:  &Literal{Value: 1},
							Right: &Literal{Value: 0},
							Op:    Add,
						},
					},
				},
			},
		},
		RelationName: "peer",
	}
	_ = tree

	tree2 := &Subquery{
		Plan: &Subplan{
			Plan: &Scan{
				Source: &TableScanSource{
					TableName: "peer",
				},
			},
		},
		Correlated: []*ColumnRef{
			{
				ColumnName: "peer",
			},
		},
	}

	_ = tree2

	target := tree

	res, err := Rewrite(target, &RewriteConfig{
		ExprCallback: func(le LogicalExpr) (LogicalExpr, error) {
			switch le.(type) {
			case *ColumnRef:
				return &ArithmeticOp{
					Left:  &Literal{Value: 1},
					Right: &Literal{Value: 0},
					Op:    Add,
				}, nil
			case *AggregateFunctionCall:
				return &ExprRef{
					Identified: &IdentifiedExpr{
						ID:   "sum",
						Expr: le,
					},
				}, nil
			}

			return le, nil
		},
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(res)
	panic("a")
}
