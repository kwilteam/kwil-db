package planner3

import (
	"testing"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/stretchr/testify/require"
)

// this test tests the logic in aggregate.go
// aggregate.go is responsible for enforcing aggregation rules in the logical plan,
// and for analyzing trees of expressions to determine if they are valid.
func Test_Aggregate(t *testing.T) {
	// helper funcs
	colExpr := func(parent, name string) LogicalExpr {
		return &ColumnRef{
			Parent:     parent,
			ColumnName: name,
		}
	}

	arithExpr := func(left, right LogicalExpr) LogicalExpr {
		return &ArithmeticOp{
			Left:  left,
			Right: right,
			Op:    Add,
		}
	}

	litInt := func(val int) LogicalExpr {
		return &Literal{
			Value: val,
			Type:  types.IntType,
		}
	}

	type testCase struct {
		name       string
		exprs      []LogicalExpr
		groupExprs []LogicalExpr
		err        bool
	}

	testCases := []testCase{
		{
			name: "simple case",
			exprs: []LogicalExpr{
				colExpr("a", "b"),
			},
			groupExprs: []LogicalExpr{
				colExpr("a", "b"),
			},
		},
		{
			name: "aggregate function",
			exprs: []LogicalExpr{
				&ArithmeticOp{
					Left: colExpr("a", "c"),
					Right: &AggregateFunctionCall{
						FunctionName: "sum",
						Args:         []LogicalExpr{colExpr("a", "b")},
					},
					Op: Add,
				},
			},
			// only a.c must be in the group by
			groupExprs: []LogicalExpr{
				colExpr("a", "c"),
			},
		},
		{
			name: "aggregate function - negative",
			exprs: []LogicalExpr{
				&ArithmeticOp{
					Left: colExpr("a", "c"),
					Right: &AggregateFunctionCall{
						FunctionName: "sum",
						Args:         []LogicalExpr{colExpr("a", "b")},
					},
					Op: Add,
				},
			},
			// missing a.c
			groupExprs: []LogicalExpr{
				colExpr("a", "b"),
			},
			err: true,
		},
		{
			name: "grouped column has arithmetic",
			exprs: []LogicalExpr{
				colExpr("a", "b"),
			},
			groupExprs: []LogicalExpr{
				arithExpr(colExpr("a", "b"), litInt(1)),
			},
			err: true,
		},
		{
			name: "same column used twice, grouped once",
			exprs: []LogicalExpr{
				&ArithmeticOp{
					Left: colExpr("a", "c"),
					Right: &ArithmeticOp{
						Left: colExpr("a", "c"),
						Right: &Literal{
							Value: 1,
							Type:  types.IntType,
						},
						Op: Add,
					},
					Op: Add,
				},
			},
			groupExprs: []LogicalExpr{
				colExpr("a", "c"),
			},
		},
		{
			// fails because only a.c + 1 is grouped
			name: "referenced twice, only one expression grouped",
			exprs: []LogicalExpr{
				arithExpr(colExpr("a", "c"), litInt(1)),
				arithExpr(colExpr("a", "c"), litInt(2)),
			},
			groupExprs: []LogicalExpr{
				arithExpr(colExpr("a", "c"), litInt(1)),
			},
			err: true,
		},
		{
			name: "several group expressions",
			exprs: []LogicalExpr{
				colExpr("a", "b"),
				arithExpr(colExpr("a", "b"), litInt(1)),
				arithExpr(colExpr("a", "c"), &AggregateFunctionCall{
					FunctionName: "sum",
					Args:         []LogicalExpr{colExpr("a", "d")},
				}),
			},
			groupExprs: []LogicalExpr{
				colExpr("a", "b"),
				colExpr("a", "c"),
			},
		},
		{
			name: "subquery",
			exprs: []LogicalExpr{
				&Subquery{
					SubqueryType: ScalarSubquery,
					Query: &Project{
						Expressions: []LogicalExpr{
							arithExpr(colExpr("a", "b"), litInt(1)),
						},
						Child: &Noop{},
					},
				},
			},
			err: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			aggCheck, err := newAggregateChecker(tc.groupExprs)
			require.NoError(t, err)

			err = aggCheck.checkMany(tc.exprs)
			if tc.err {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
