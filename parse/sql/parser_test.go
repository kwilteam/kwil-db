package sqlparser

import (
	"errors"
	"flag"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/kwilteam/kwil-db/parse/sql/postgres"
	"github.com/kwilteam/kwil-db/parse/sql/tree"
	"github.com/stretchr/testify/assert"
)

var traceMode = flag.Bool("trace-mode", false, "run tests with tracing")

var columnStar = []tree.ResultColumn{
	&tree.ResultColumnStar{},
}

func genLiteralExpression(value string) tree.Expression {
	return &tree.ExpressionLiteral{Value: value}
}

func getResultColumnExprs(values ...string) []tree.ResultColumn {
	t := make([]tree.ResultColumn, len(values))
	for i, v := range values {
		t[i] = &tree.ResultColumnExpression{
			Expression: genLiteralExpression(v),
		}
	}
	return t
}

func genSelectUnaryExprTree(op tree.UnaryOperator, value string) *tree.Select {
	return &tree.Select{
		SelectStmt: &tree.SelectStmt{
			SelectCores: []*tree.SelectCore{
				{
					SelectType: tree.SelectTypeAll,
					Columns: []tree.ResultColumn{
						&tree.ResultColumnExpression{
							Expression: &tree.ExpressionUnary{
								Operator: op,
								Operand:  genLiteralExpression(value),
							},
						},
					},
				},
			},
		},
	}
}

func genSelectColumnLiteralTree(value string) *tree.Select {
	t := tree.Select{
		SelectStmt: &tree.SelectStmt{
			SelectCores: []*tree.SelectCore{
				{
					SelectType: tree.SelectTypeAll,
					Columns: []tree.ResultColumn{
						&tree.ResultColumnExpression{
							Expression: genLiteralExpression(value),
						},
					},
				},
			},
		}}
	return &t
}

func genSelectColumnStarTree() *tree.Select {
	t := tree.Select{
		SelectStmt: &tree.SelectStmt{
			SelectCores: []*tree.SelectCore{
				{
					SelectType: tree.SelectTypeAll,
					Columns: []tree.ResultColumn{
						&tree.ResultColumnStar{},
					},
				},
			},
		}}
	return &t
}

func genSelectColumnTableTree(table string) *tree.Select {
	t := tree.Select{
		SelectStmt: &tree.SelectStmt{
			SelectCores: []*tree.SelectCore{
				{
					SelectType: tree.SelectTypeAll,
					Columns: []tree.ResultColumn{
						&tree.ResultColumnTable{TableName: table},
					},
				},
			},
		}}
	return &t
}

func genSimpleCompoundSelectTree(op tree.CompoundOperatorType) *tree.Select {
	return &tree.Select{
		SelectStmt: &tree.SelectStmt{
			SelectCores: []*tree.SelectCore{
				{
					SelectType: tree.SelectTypeAll,
					Columns:    getResultColumnExprs("1"),
				},
				{
					SelectType: tree.SelectTypeAll,
					Columns:    getResultColumnExprs("2"),
					Compound:   &tree.CompoundOperator{Operator: op},
				},
			},
		},
	}
}

func genSimpleCollateSelectTree(collateType tree.CollationType, value string) *tree.Select {
	return &tree.Select{
		SelectStmt: &tree.SelectStmt{
			SelectCores: []*tree.SelectCore{
				{
					SelectType: tree.SelectTypeAll,
					Columns: []tree.ResultColumn{
						&tree.ResultColumnExpression{
							Expression: &tree.ExpressionCollate{
								Expression: &tree.ExpressionLiteral{Value: value},
								Collation:  collateType,
							},
						},
					},
				},
			},
		},
	}
}

func genSimpleBinaryCompareSelectTree(op tree.BinaryOperator, leftValue, rightValue string) *tree.Select {
	return &tree.Select{
		SelectStmt: &tree.SelectStmt{
			SelectCores: []*tree.SelectCore{
				{
					SelectType: tree.SelectTypeAll,
					Columns: []tree.ResultColumn{
						&tree.ResultColumnExpression{
							Expression: &tree.ExpressionBinaryComparison{
								Left:     &tree.ExpressionLiteral{Value: leftValue},
								Operator: op,
								Right:    &tree.ExpressionLiteral{Value: rightValue},
							},
						},
					},
				},
			},
		},
	}
}

func genSimplyArithmeticSelectTree(op tree.ArithmeticOperator, leftValue, rightValue string) *tree.Select {
	return &tree.Select{
		SelectStmt: &tree.SelectStmt{
			SelectCores: []*tree.SelectCore{
				{
					SelectType: tree.SelectTypeAll,
					Columns: []tree.ResultColumn{
						&tree.ResultColumnExpression{
							Expression: &tree.ExpressionArithmetic{
								Left:     &tree.ExpressionLiteral{Value: leftValue},
								Operator: op,
								Right:    &tree.ExpressionLiteral{Value: rightValue},
							},
						},
					},
				},
			},
		},
	}
}

func genSimpleStringCompareSelectTree(op tree.StringOperator, leftValue, rightValue, escape string) *tree.Select {
	escapeExpr := tree.Expression(&tree.ExpressionLiteral{Value: escape})
	if escape == "" {
		escapeExpr = nil
	}
	return &tree.Select{
		SelectStmt: &tree.SelectStmt{
			SelectCores: []*tree.SelectCore{
				{
					SelectType: tree.SelectTypeAll,
					Columns: []tree.ResultColumn{
						&tree.ResultColumnExpression{
							Expression: &tree.ExpressionStringCompare{
								Left:     tree.Expression(&tree.ExpressionLiteral{Value: leftValue}),
								Operator: op,
								Right:    tree.Expression(&tree.ExpressionLiteral{Value: rightValue}),
								Escape:   escapeExpr,
							},
						},
					},
				},
			},
		},
	}
}

func genSimpleCTETree(table, value string) *tree.CTE {
	return &tree.CTE{
		Table:  table,
		Select: genSelectColumnLiteralTree(value).SelectStmt,
	}
}

func genSimpleExprNullSelectTree(value string, not bool) *tree.Select {
	return &tree.Select{
		SelectStmt: &tree.SelectStmt{
			SelectCores: []*tree.SelectCore{
				{
					SelectType: tree.SelectTypeAll,
					Columns: []tree.ResultColumn{
						&tree.ResultColumnExpression{
							Expression: &tree.ExpressionIs{
								Left:     &tree.ExpressionLiteral{Value: value},
								Distinct: false,
								Not:      not,
								Right:    &tree.ExpressionLiteral{Value: "NULL"},
							},
						},
					},
				},
			},
		},
	}
}

func genSimpleExprIsSelectTree(left string, right string, not bool) *tree.Select {
	return &tree.Select{
		SelectStmt: &tree.SelectStmt{
			SelectCores: []*tree.SelectCore{
				{
					SelectType: tree.SelectTypeAll,
					Columns: []tree.ResultColumn{
						&tree.ResultColumnExpression{
							Expression: &tree.ExpressionIs{
								Left:     &tree.ExpressionLiteral{Value: left},
								Distinct: false,
								Not:      not,
								Right:    &tree.ExpressionLiteral{Value: right},
							},
						},
					},
				},
			},
		},
	}
}

func genSimpleFunctionSelectTree(f tree.SQLFunction, inputs ...tree.Expression) *tree.Select {
	return &tree.Select{
		SelectStmt: &tree.SelectStmt{
			SelectCores: []*tree.SelectCore{
				{
					SelectType: tree.SelectTypeAll,
					Columns: []tree.ResultColumn{
						&tree.ResultColumnExpression{
							Expression: &tree.ExpressionFunction{
								Function: f,
								Inputs:   inputs,
							},
						},
					},
				},
			},
		},
	}
}

func genDistinctFunctionSelectTree(f tree.SQLFunction, inputs ...tree.Expression) *tree.Select {
	return &tree.Select{
		SelectStmt: &tree.SelectStmt{
			SelectCores: []*tree.SelectCore{
				{
					SelectType: tree.SelectTypeAll,
					Columns: []tree.ResultColumn{
						&tree.ResultColumnExpression{
							Expression: &tree.ExpressionFunction{
								Function: f,
								Inputs:   inputs,
								Distinct: true,
							},
						},
					},
				},
			},
		},
	}
}

func genSimpleJoinSelectTree(joinOP *tree.JoinOperator, t1, t1Column, t2, t2Column string) *tree.Select {
	return &tree.Select{
		SelectStmt: &tree.SelectStmt{
			SelectCores: []*tree.SelectCore{
				{
					SelectType: tree.SelectTypeAll,
					Columns: []tree.ResultColumn{
						&tree.ResultColumnStar{},
					},
					From: &tree.FromClause{JoinClause: &tree.JoinClause{
						TableOrSubquery: &tree.TableOrSubqueryTable{
							Name: t1,
						},
						Joins: []*tree.JoinPredicate{
							{
								JoinOperator: joinOP,
								Table:        &tree.TableOrSubqueryTable{Name: t2},
								Constraint: &tree.ExpressionBinaryComparison{
									Operator: tree.ComparisonOperatorEqual,
									Left:     &tree.ExpressionColumn{Table: t1, Column: t1Column},
									Right:    &tree.ExpressionColumn{Table: t2, Column: t2Column},
								},
							},
						},
					}},
				},
			},
		},
	}
}

func genSimpleUpdateTree(qt *tree.QualifiedTableName, column, value string) *tree.Update {
	return &tree.Update{
		UpdateStmt: &tree.UpdateStmt{
			QualifiedTableName: qt,
			UpdateSetClause: []*tree.UpdateSetClause{
				{
					Columns:    []string{column},
					Expression: genLiteralExpression(value),
				},
			},
		},
	}
}

func TestParseRawSQL_syntax_valid(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect tree.AstNode
	}{
		// with semicolon at the end
		{"with semicolon", "select *;", genSelectColumnStarTree()},
		// common table stmt
		{"cte", "with t as (select 1) select *",
			&tree.Select{
				CTE:        []*tree.CTE{genSimpleCTETree("t", "1")},
				SelectStmt: genSelectColumnStarTree().SelectStmt,
			},
		},
		{"cte with column", "with t(c1,c2) as (select 1) select *",
			&tree.Select{
				CTE: []*tree.CTE{
					{
						Table:   "t",
						Columns: []string{"c1", "c2"},
						Select:  genSelectColumnLiteralTree("1").SelectStmt,
					},
				},
				SelectStmt: genSelectColumnStarTree().SelectStmt,
			},
		},
		//// compound operator
		{"compound union", "select 1 union select 2",
			genSimpleCompoundSelectTree(tree.CompoundOperatorTypeUnion),
		},
		{"compound union all", "select 1 union all select 2",
			genSimpleCompoundSelectTree(tree.CompoundOperatorTypeUnionAll),
		},
		{"compound intersect", "select 1 intersect select 2",
			genSimpleCompoundSelectTree(tree.CompoundOperatorTypeIntersect),
		},
		{"compound except", "select 1 except select 2",
			genSimpleCompoundSelectTree(tree.CompoundOperatorTypeExcept),
		},
		// result column
		{"*", "select *", genSelectColumnStarTree()},
		{"table.*", "select t.*", genSelectColumnTableTree("t")},
		//// table or subquery
		{"table or subquery", "select * from t1 as tt",
			&tree.Select{
				SelectStmt: &tree.SelectStmt{
					SelectCores: []*tree.SelectCore{
						{
							SelectType: tree.SelectTypeAll,
							Columns:    columnStar,
							From: &tree.FromClause{
								JoinClause: &tree.JoinClause{
									TableOrSubquery: &tree.TableOrSubqueryTable{
										Name:  "t1",
										Alias: "tt",
									},
								},
							},
						},
					},
				},
			},
		},
		{"table or subquery nest select", "select * from (select 1) as tt",
			&tree.Select{
				SelectStmt: &tree.SelectStmt{
					SelectCores: []*tree.SelectCore{
						{
							SelectType: tree.SelectTypeAll,
							Columns:    columnStar,
							From: &tree.FromClause{
								JoinClause: &tree.JoinClause{
									TableOrSubquery: &tree.TableOrSubquerySelect{
										Select: genSelectColumnLiteralTree("1").SelectStmt,
										Alias:  "tt",
									},
								},
							},
						},
					},
				},
			},
		},
		{"table or subquery join", "select * from t1 as tt join t2 as ttt on tt.a = ttt.a",
			&tree.Select{
				SelectStmt: &tree.SelectStmt{
					SelectCores: []*tree.SelectCore{
						{
							SelectType: tree.SelectTypeAll,
							Columns:    columnStar,
							From: &tree.FromClause{
								JoinClause: &tree.JoinClause{
									TableOrSubquery: &tree.TableOrSubqueryTable{Name: "t1", Alias: "tt"},
									Joins: []*tree.JoinPredicate{
										{
											JoinOperator: &tree.JoinOperator{
												JoinType: tree.JoinTypeJoin,
												Outer:    false,
											},
											Table: &tree.TableOrSubqueryTable{Name: "t2", Alias: "ttt"},
											Constraint: &tree.ExpressionBinaryComparison{
												Left: &tree.ExpressionColumn{
													Table:  "tt",
													Column: "a",
												},
												Operator: tree.ComparisonOperatorEqual,
												Right: &tree.ExpressionColumn{
													Table:  "ttt",
													Column: "a",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		//// expr
		// literal value,,
		{"number", "select 1", genSelectColumnLiteralTree("1")},
		{"string", "select 'a'", genSelectColumnLiteralTree("'a'")},
		{"null", "select null", genSelectColumnLiteralTree("NULL")},
		{"true", "select true", genSelectColumnLiteralTree("true")},
		{"false", "select false", genSelectColumnLiteralTree("false")},
		// bind parameter
		{"expr bind parameter $", "select $a",
			&tree.Select{
				SelectStmt: &tree.SelectStmt{
					SelectCores: []*tree.SelectCore{
						{
							SelectType: tree.SelectTypeAll,
							Columns: []tree.ResultColumn{
								&tree.ResultColumnExpression{
									Expression: &tree.ExpressionBindParameter{
										Parameter: "$a",
									},
								},
							},
						},
					},
				}}},
		{"expr bind parameter @", "select @a",
			&tree.Select{
				SelectStmt: &tree.SelectStmt{
					SelectCores: []*tree.SelectCore{
						{
							SelectType: tree.SelectTypeAll,
							Columns: []tree.ResultColumn{
								&tree.ResultColumnExpression{
									Expression: &tree.ExpressionBindParameter{
										Parameter: "@a",
									},
								},
							},
						},
					},
				}}},
		{"expr names", "select t1.c1",
			&tree.Select{
				SelectStmt: &tree.SelectStmt{
					SelectCores: []*tree.SelectCore{
						{
							SelectType: tree.SelectTypeAll,
							Columns: []tree.ResultColumn{
								&tree.ResultColumnExpression{
									Expression: &tree.ExpressionColumn{
										Table:  "t1",
										Column: "c1",
									},
								},
							},
						},
					},
				}}},
		// unary op
		{"expr unary op +", "select +1", genSelectUnaryExprTree(tree.UnaryOperatorPlus, "1")},
		{"expr unary op -", "select -1", genSelectUnaryExprTree(tree.UnaryOperatorMinus, "1")},
		{"expr unary op - twice, right associative", "select - -1",
			&tree.Select{
				SelectStmt: &tree.SelectStmt{
					SelectCores: []*tree.SelectCore{
						{
							SelectType: tree.SelectTypeAll,
							Columns: []tree.ResultColumn{
								&tree.ResultColumnExpression{
									Expression: &tree.ExpressionUnary{
										Operator: tree.UnaryOperatorMinus,
										Operand: &tree.ExpressionUnary{
											Operator: tree.UnaryOperatorMinus,
											Operand:  genLiteralExpression("1"),
										},
									},
								},
							},
						},
					},
				},
			},
		},
		//{"expr unary op ~", "select ~1", genSelectUnaryExprTree(tree.UnaryOperatorBitNot, "1")},
		{"expr unary op not", "select not 1", genSelectUnaryExprTree(tree.UnaryOperatorNot, "1")},
		{"expr unary op not twice, right associative", "select not not true",
			&tree.Select{
				SelectStmt: &tree.SelectStmt{
					SelectCores: []*tree.SelectCore{
						{
							SelectType: tree.SelectTypeAll,
							Columns: []tree.ResultColumn{
								&tree.ResultColumnExpression{
									Expression: &tree.ExpressionUnary{
										Operator: tree.UnaryOperatorNot,
										Operand: &tree.ExpressionUnary{
											Operator: tree.UnaryOperatorNot,
											Operand:  genLiteralExpression("true"),
										},
									},
								},
							},
						},
					},
				},
			},
		},
		// binary op
		//{"expr binary op ||", "select 1 || 2",
		//	genSimplyArithmeticSelectTree(tree.ArithmeticConcat, "1", "2")},
		{"expr binary op *", "select 1 * 2",
			genSimplyArithmeticSelectTree(tree.ArithmeticOperatorMultiply, "1", "2")},
		{"expr binary op /", "select 1 / 2",
			genSimplyArithmeticSelectTree(tree.ArithmeticOperatorDivide, "1", "2")},
		{"expr binary op %", "select 1 % 2",
			genSimplyArithmeticSelectTree(tree.ArithmeticOperatorModulus, "1", "2")},
		{"expr binary op +", "select 1 + 2",
			genSimplyArithmeticSelectTree(tree.ArithmeticOperatorAdd, "1", "2")},
		{"expr binary op -", "select 1 - 2",
			genSimplyArithmeticSelectTree(tree.ArithmeticOperatorSubtract, "1", "2")},
		//{"expr binary op <<", "select 1 << 2",
		//	genSimplyArithmeticSelectTree(tree.ArithmeticOperatorBitwiseLeftShift, "1", "2")},
		//{"expr binary op >>", "select 1 >> 2",
		//	genSimplyArithmeticSelectTree(tree.ArithmeticOperatorBitwiseRightShift, "1", "2")},
		//{"expr binary op &", "select 1 & 2",
		//	genSimplyArithmeticSelectTree(tree.ArithmeticOperatorBitwiseAnd, "1", "2")},
		//{"expr binary op |", "select 1 | 2",
		//	genSimplyArithmeticSelectTree(tree.ArithmeticOperatorBitwiseOr, "1", "2")},
		{"expr binary op <", "select 1 < 2",
			genSimpleBinaryCompareSelectTree(tree.ComparisonOperatorLessThan, "1", "2")},
		{"expr binary op <=", "select 1 <= 2",
			genSimpleBinaryCompareSelectTree(tree.ComparisonOperatorLessThanOrEqual, "1", "2")},
		{"expr binary op >", "select 1 > 2",
			genSimpleBinaryCompareSelectTree(tree.ComparisonOperatorGreaterThan, "1", "2")},
		{"expr binary op >=", "select 1 >= 2",
			genSimpleBinaryCompareSelectTree(tree.ComparisonOperatorGreaterThanOrEqual, "1", "2")},
		{"expr binary op =", "select 1 = 2",
			genSimpleBinaryCompareSelectTree(tree.ComparisonOperatorEqual, "1", "2")},
		{"expr binary op !=", "select 1 != 2",
			genSimpleBinaryCompareSelectTree(tree.ComparisonOperatorNotEqual, "1", "2")},
		{"expr binary op <>", "select 1 <> 2",
			genSimpleBinaryCompareSelectTree(tree.ComparisonOperatorNotEqualDiamond, "1", "2")},
		{"expr binary op and", "select 1 and 2",
			genSimpleBinaryCompareSelectTree(tree.LogicalOperatorAnd, "1", "2")},
		{"expr binary op or", "select 1 or 2",
			genSimpleBinaryCompareSelectTree(tree.LogicalOperatorOr, "1", "2")},
		// in
		{"expr binary op in", "select 1 in (1,2)",
			&tree.Select{
				SelectStmt: &tree.SelectStmt{
					SelectCores: []*tree.SelectCore{
						{
							SelectType: tree.SelectTypeAll,
							Columns: []tree.ResultColumn{
								&tree.ResultColumnExpression{
									Expression: &tree.ExpressionBinaryComparison{
										Left:     genLiteralExpression("1"),
										Operator: tree.ComparisonOperatorIn,
										Right: &tree.ExpressionList{
											Expressions: []tree.Expression{
												genLiteralExpression("1"),
												genLiteralExpression("2"),
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{"expr binary op not in", "select 1 not in (1,2)",
			&tree.Select{
				SelectStmt: &tree.SelectStmt{
					SelectCores: []*tree.SelectCore{
						{
							SelectType: tree.SelectTypeAll,
							Columns: []tree.ResultColumn{
								&tree.ResultColumnExpression{
									Expression: &tree.ExpressionBinaryComparison{
										Left:     genLiteralExpression("1"),
										Operator: tree.ComparisonOperatorNotIn,
										Right: &tree.ExpressionList{
											Expressions: []tree.Expression{
												genLiteralExpression("1"),
												genLiteralExpression("2"),
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{"expr binary op in with select", "select 1 in (select 1)",
			&tree.Select{
				SelectStmt: &tree.SelectStmt{
					SelectCores: []*tree.SelectCore{
						{
							SelectType: tree.SelectTypeAll,
							Columns: []tree.ResultColumn{
								&tree.ResultColumnExpression{
									Expression: &tree.ExpressionBinaryComparison{
										Left:     genLiteralExpression("1"),
										Operator: tree.ComparisonOperatorIn,
										Right: &tree.ExpressionSelect{
											Select: genSelectColumnLiteralTree("1").SelectStmt,
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{"expr binary op not in with select", "select 1 not in (select 1)",
			&tree.Select{
				SelectStmt: &tree.SelectStmt{
					SelectCores: []*tree.SelectCore{
						{
							SelectType: tree.SelectTypeAll,
							Columns: []tree.ResultColumn{
								&tree.ResultColumnExpression{
									Expression: &tree.ExpressionBinaryComparison{
										Left:     genLiteralExpression("1"),
										Operator: tree.ComparisonOperatorNotIn,
										Right: &tree.ExpressionSelect{
											Select: genSelectColumnLiteralTree("1").SelectStmt,
										},
									},
								},
							},
						},
					},
				},
			},
		},
		// string compare
		{"expr binary op like", "select 1 like 2",
			genSimpleStringCompareSelectTree(tree.StringOperatorLike, "1", "2", "")},
		{"expr binary op not like", "select 1 not like 2",
			genSimpleStringCompareSelectTree(tree.StringOperatorNotLike, "1", "2", "")},
		{"expr binary op like escape", "select 1 like 2 escape 3",
			genSimpleStringCompareSelectTree(tree.StringOperatorLike, "1", "2", "3")},
		// function
		// core functions
		{"expr function abs", "select abs(1)",
			genSimpleFunctionSelectTree(&tree.FunctionABS, genLiteralExpression("1"))},
		{"expr function error", "select error('error message')",
			genSimpleFunctionSelectTree(&tree.FunctionERROR, genLiteralExpression(`'error message'`))},
		{"expr function format", `select format('%d',2)`,
			genSimpleFunctionSelectTree(&tree.FunctionFORMAT,
				genLiteralExpression(`'%d'`), genLiteralExpression("2"))},
		{"expr function length", `select length(1)`,
			genSimpleFunctionSelectTree(&tree.FunctionLENGTH, genLiteralExpression("1"))},
		{"expr function lower", `select lower('Z')`,
			genSimpleFunctionSelectTree(&tree.FunctionLOWER, genLiteralExpression("'Z'"))},
		{"expr function upper", `select upper('z')`,
			genSimpleFunctionSelectTree(&tree.FunctionUPPER, genLiteralExpression("'z'"))},

		// aggregate functions
		{"expr function count", `select count(1)`,
			genSimpleFunctionSelectTree(&tree.FunctionCOUNT, genLiteralExpression("1"))},
		{"expr function count distinct", `select count(distinct 1)`,
			genDistinctFunctionSelectTree(&tree.FunctionCOUNT, genLiteralExpression("1"))},
		{"expr function sum", `select sum(1)`,
			genSimpleFunctionSelectTree(&tree.FunctionSUM, genLiteralExpression("1"))},
		{"expr function sum distinct", `select sum(distinct 1)`,
			genDistinctFunctionSelectTree(&tree.FunctionSUM, genLiteralExpression("1"))},

		// expr list
		{"expr list", "select (1,2)",
			&tree.Select{
				SelectStmt: &tree.SelectStmt{
					SelectCores: []*tree.SelectCore{
						{
							SelectType: tree.SelectTypeAll,
							Columns: []tree.ResultColumn{
								&tree.ResultColumnExpression{
									Expression: &tree.ExpressionList{
										Expressions: []tree.Expression{
											genLiteralExpression("1"),
											genLiteralExpression("2"),
										},
									},
								},
							},
						},
					},
				},
			},
		},
		// expr precedence
		{"expr precedence 1", "select -1 > 2",
			&tree.Select{
				SelectStmt: &tree.SelectStmt{
					SelectCores: []*tree.SelectCore{
						{
							SelectType: tree.SelectTypeAll,
							Columns: []tree.ResultColumn{
								&tree.ResultColumnExpression{
									Expression: &tree.ExpressionBinaryComparison{
										Operator: tree.ComparisonOperatorGreaterThan,
										Left: &tree.ExpressionUnary{
											Operator: tree.UnaryOperatorMinus,
											Operand:  genLiteralExpression("1"),
										},
										Right: genLiteralExpression("2"),
									},
								},
							},
						},
					},
				},
			},
		},
		{"expr precedence 2", "SELECT NOT (-1 = 1) AND 1 notnull OR 3 < 2",
			&tree.Select{
				SelectStmt: &tree.SelectStmt{
					SelectCores: []*tree.SelectCore{
						{
							SelectType: tree.SelectTypeAll,
							Columns: []tree.ResultColumn{
								&tree.ResultColumnExpression{
									Expression: &tree.ExpressionBinaryComparison{
										Operator: tree.LogicalOperatorOr,
										Left: &tree.ExpressionBinaryComparison{
											Operator: tree.LogicalOperatorAnd,
											Left: &tree.ExpressionUnary{
												Operator: tree.UnaryOperatorNot,
												Operand: &tree.ExpressionBinaryComparison{
													Wrapped:  true,
													Operator: tree.ComparisonOperatorEqual,
													Left: &tree.ExpressionUnary{
														Operator: tree.UnaryOperatorMinus,
														Operand:  genLiteralExpression("1"),
													},
													Right: genLiteralExpression("1"),
												},
											},
											Right: &tree.ExpressionIs{
												Left:  genLiteralExpression("1"),
												Not:   true,
												Right: genLiteralExpression("NULL"),
											},
										},
										Right: &tree.ExpressionBinaryComparison{
											Operator: tree.ComparisonOperatorLessThan,
											Left:     genLiteralExpression("3"),
											Right:    genLiteralExpression("2"),
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{"expr precedence 3", "SELECT NOT (-1 = 1) AND (1 notnull OR 3 < 2)",
			&tree.Select{
				SelectStmt: &tree.SelectStmt{
					SelectCores: []*tree.SelectCore{
						{
							SelectType: tree.SelectTypeAll,
							Columns: []tree.ResultColumn{
								&tree.ResultColumnExpression{
									Expression: &tree.ExpressionBinaryComparison{
										Operator: tree.LogicalOperatorAnd,
										Left: &tree.ExpressionUnary{
											Operator: tree.UnaryOperatorNot,
											Operand: &tree.ExpressionBinaryComparison{
												Operator: tree.ComparisonOperatorEqual,
												Left: &tree.ExpressionUnary{
													Operator: tree.UnaryOperatorMinus,
													Operand:  genLiteralExpression("1"),
												},
												Right:   genLiteralExpression("1"),
												Wrapped: true,
											},
										},
										Right: &tree.ExpressionBinaryComparison{
											Wrapped:  true,
											Operator: tree.LogicalOperatorOr,
											Left: &tree.ExpressionIs{
												Left:  genLiteralExpression("1"),
												Not:   true,
												Right: genLiteralExpression("NULL"),
											},
											Right: &tree.ExpressionBinaryComparison{
												Operator: tree.ComparisonOperatorLessThan,
												Left:     genLiteralExpression("3"),
												Right:    genLiteralExpression("2"),
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{"expr precedence 4", "select not 3 + 4 * 5 - 2 = 2 + -1",
			&tree.Select{
				SelectStmt: &tree.SelectStmt{
					SelectCores: []*tree.SelectCore{
						{
							SelectType: tree.SelectTypeAll,
							Columns: []tree.ResultColumn{
								&tree.ResultColumnExpression{
									Expression: &tree.ExpressionUnary{
										Operator: tree.UnaryOperatorNot,
										Operand: &tree.ExpressionBinaryComparison{
											Operator: tree.ComparisonOperatorEqual,
											Left: &tree.ExpressionArithmetic{
												Operator: tree.ArithmeticOperatorSubtract,
												Left: &tree.ExpressionArithmetic{
													Operator: tree.ArithmeticOperatorAdd,
													Left:     genLiteralExpression("3"),
													Right: &tree.ExpressionArithmetic{
														Operator: tree.ArithmeticOperatorMultiply,
														Left:     genLiteralExpression("4"),
														Right:    genLiteralExpression("5"),
													},
												},
												Right: genLiteralExpression("2"),
											},
											Right: &tree.ExpressionArithmetic{
												Left:     genLiteralExpression("2"),
												Operator: tree.ArithmeticOperatorAdd,
												Right: &tree.ExpressionUnary{
													Operator: tree.UnaryOperatorMinus,
													Operand:  genLiteralExpression("1"),
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},

		// collate
		{"expr collate nocase", "select 1 collate nocase",
			genSimpleCollateSelectTree(tree.CollationTypeNoCase, "1")},
		// is
		{"expr isnull", "select 1 isnull", genSimpleExprNullSelectTree("1", false)},
		{"expr notnull", "select 1 notnull", genSimpleExprNullSelectTree("1", true)},
		{"expr not null", "select 1 is not null", genSimpleExprNullSelectTree("1", true)},
		{"expr is", "select true is true",
			genSimpleExprIsSelectTree("true", "true", false)},
		{"expr is not", "select true is not true",
			genSimpleExprIsSelectTree("true", "true", true)},
		{"expr is distinct from", "select 1 is distinct from 2",
			&tree.Select{
				SelectStmt: &tree.SelectStmt{
					SelectCores: []*tree.SelectCore{
						{
							SelectType: tree.SelectTypeAll,
							Columns: []tree.ResultColumn{
								&tree.ResultColumnExpression{
									Expression: &tree.ExpressionIs{
										Left:     genLiteralExpression("1"),
										Right:    genLiteralExpression("2"),
										Distinct: true,
									},
								},
							},
						},
					},
				},
			},
		},
		{"expr is not distinct from", "select 1 is not distinct from 2",
			&tree.Select{
				SelectStmt: &tree.SelectStmt{
					SelectCores: []*tree.SelectCore{
						{
							SelectType: tree.SelectTypeAll,
							Columns: []tree.ResultColumn{
								&tree.ResultColumnExpression{
									Expression: &tree.ExpressionIs{
										Left:     genLiteralExpression("1"),
										Right:    genLiteralExpression("2"),
										Distinct: true,
										Not:      true,
									},
								},
							},
						},
					},
				},
			},
		},
		// between
		{"expr between", "select 1 between 2 and 3",
			&tree.Select{
				SelectStmt: &tree.SelectStmt{
					SelectCores: []*tree.SelectCore{
						{
							SelectType: tree.SelectTypeAll,
							Columns: []tree.ResultColumn{
								&tree.ResultColumnExpression{
									Expression: &tree.ExpressionBetween{
										Expression: genLiteralExpression("1"),
										Left:       genLiteralExpression("2"),
										Right:      &tree.ExpressionLiteral{Value: "3"},
									},
								},
							},
						},
					},
				},
			},
		},
		{"expr not between", "select 1 not between 2 and 3",
			&tree.Select{
				SelectStmt: &tree.SelectStmt{
					SelectCores: []*tree.SelectCore{
						{
							SelectType: tree.SelectTypeAll,
							Columns: []tree.ResultColumn{
								&tree.ResultColumnExpression{
									Expression: &tree.ExpressionBetween{
										Expression: genLiteralExpression("1"),
										Left:       genLiteralExpression("2"),
										Right:      &tree.ExpressionLiteral{Value: "3"},
										NotBetween: true,
									},
								},
							},
						},
					},
				},
			},
		},
		//
		{"expr exists", "select (select 1)",
			&tree.Select{
				SelectStmt: &tree.SelectStmt{
					SelectCores: []*tree.SelectCore{
						{
							SelectType: tree.SelectTypeAll,
							Columns: []tree.ResultColumn{
								&tree.ResultColumnExpression{
									Expression: &tree.ExpressionSelect{
										IsNot:    false,
										IsExists: false,
										Select:   genSelectColumnLiteralTree("1").SelectStmt,
									},
								},
							},
						},
					},
				},
			},
		},
		{"expr exists", "select exists (select 1)",
			&tree.Select{
				SelectStmt: &tree.SelectStmt{
					SelectCores: []*tree.SelectCore{
						{
							SelectType: tree.SelectTypeAll,
							Columns: []tree.ResultColumn{
								&tree.ResultColumnExpression{
									Expression: &tree.ExpressionSelect{
										IsNot:    false,
										IsExists: true,
										Select:   genSelectColumnLiteralTree("1").SelectStmt,
									},
								},
							},
						},
					},
				},
			},
		},
		{"expr not exists", "select not exists (select 1)",
			&tree.Select{
				SelectStmt: &tree.SelectStmt{
					SelectCores: []*tree.SelectCore{
						{
							SelectType: tree.SelectTypeAll,
							Columns: []tree.ResultColumn{
								&tree.ResultColumnExpression{
									Expression: &tree.ExpressionSelect{
										IsNot:    true,
										IsExists: true,
										Select:   genSelectColumnLiteralTree("1").SelectStmt,
									},
								},
							},
						},
					},
				},
			},
		},
		// case
		{"expr case", "select case when 1 then 2 end",
			&tree.Select{
				SelectStmt: &tree.SelectStmt{
					SelectCores: []*tree.SelectCore{
						{
							SelectType: tree.SelectTypeAll,
							Columns: []tree.ResultColumn{
								&tree.ResultColumnExpression{
									Expression: &tree.ExpressionCase{
										WhenThenPairs: [][2]tree.Expression{
											{
												genLiteralExpression("1"),
												genLiteralExpression("2"),
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{"expr case else", "select case when 1 then 2 else 3 end",
			&tree.Select{
				SelectStmt: &tree.SelectStmt{
					SelectCores: []*tree.SelectCore{
						{
							SelectType: tree.SelectTypeAll,
							Columns: []tree.ResultColumn{
								&tree.ResultColumnExpression{
									Expression: &tree.ExpressionCase{
										WhenThenPairs: [][2]tree.Expression{
											{
												genLiteralExpression("1"),
												genLiteralExpression("2"),
											},
										},
										ElseExpression: &tree.ExpressionLiteral{Value: "3"},
									},
								},
							},
						},
					},
				},
			},
		},
		{"expr case multi when", "select case when 1 then 2 when 3 then 4 end",
			&tree.Select{
				SelectStmt: &tree.SelectStmt{
					SelectCores: []*tree.SelectCore{
						{
							SelectType: tree.SelectTypeAll,
							Columns: []tree.ResultColumn{
								&tree.ResultColumnExpression{
									Expression: &tree.ExpressionCase{
										WhenThenPairs: [][2]tree.Expression{
											{
												genLiteralExpression("1"),
												genLiteralExpression("2"),
											},
											{
												&tree.ExpressionLiteral{Value: "3"},
												&tree.ExpressionLiteral{Value: "4"},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{"expr case expr", "select case 1 when 2 then 3 end",
			&tree.Select{
				SelectStmt: &tree.SelectStmt{
					SelectCores: []*tree.SelectCore{
						{
							SelectType: tree.SelectTypeAll,
							Columns: []tree.ResultColumn{
								&tree.ResultColumnExpression{
									Expression: &tree.ExpressionCase{
										CaseExpression: genLiteralExpression("1"),
										WhenThenPairs: [][2]tree.Expression{
											{
												genLiteralExpression("2"),
												&tree.ExpressionLiteral{Value: "3"},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		//// insert stmt
		{"insert", "insert into t1 values (1)",
			&tree.Insert{
				InsertStmt: &tree.InsertStmt{
					Table:      "t1",
					InsertType: tree.InsertTypeInsert,
					Values:     [][]tree.Expression{{genLiteralExpression("1")}},
				}}},
		{"insert with columns", "insert into t1 (a,b) values (1,2)",
			&tree.Insert{
				InsertStmt: &tree.InsertStmt{
					Table:      "t1",
					Columns:    []string{"a", "b"},
					InsertType: tree.InsertTypeInsert,
					Values: [][]tree.Expression{
						{
							genLiteralExpression("1"),
							genLiteralExpression("2"),
						},
					}}}},
		{"insert with columns with table alias", "insert into t1 as t (a,b) values (1,2)",
			&tree.Insert{
				InsertStmt: &tree.InsertStmt{
					Table:      "t1",
					TableAlias: "t",
					Columns:    []string{"a", "b"},
					InsertType: tree.InsertTypeInsert,
					Values: [][]tree.Expression{
						{
							genLiteralExpression("1"),
							genLiteralExpression("2"),
						},
					}}}},
		{"insert with columns with bind parameter", "insert into t1 (a,b) values ($a, $b)",
			&tree.Insert{
				InsertStmt: &tree.InsertStmt{
					Table:      "t1",
					Columns:    []string{"a", "b"},
					InsertType: tree.InsertTypeInsert,
					Values: [][]tree.Expression{
						{
							&tree.ExpressionBindParameter{Parameter: "$a"},
							&tree.ExpressionBindParameter{Parameter: "$b"},
						},
					}}}},
		{"insert with cte", "with t as (select 1) insert into t1 (a,b) values (1,2)",
			&tree.Insert{
				CTE: []*tree.CTE{genSimpleCTETree("t", "1")},
				InsertStmt: &tree.InsertStmt{
					Table:      "t1",
					Columns:    []string{"a", "b"},
					InsertType: tree.InsertTypeInsert,
					Values: [][]tree.Expression{
						{genLiteralExpression("1"), genLiteralExpression("2")},
					}}}},
		{"insert with returning", "insert into t1 (a,b) values (1,2) returning a as c",
			&tree.Insert{
				InsertStmt: &tree.InsertStmt{
					Table:      "t1",
					Columns:    []string{"a", "b"},
					InsertType: tree.InsertTypeInsert,
					Values: [][]tree.Expression{
						{genLiteralExpression("1"), genLiteralExpression("2")},
					},
					ReturningClause: &tree.ReturningClause{
						Returned: []*tree.ReturningClauseColumn{
							{
								Expression: &tree.ExpressionColumn{
									Column: "a",
								},
								Alias: "c",
							},
						},
					}}}},
		{"insert with returning literal", "insert into t1 (a,b) values (1,2) returning 1",
			&tree.Insert{
				InsertStmt: &tree.InsertStmt{
					Table:      "t1",
					Columns:    []string{"a", "b"},
					InsertType: tree.InsertTypeInsert,
					Values: [][]tree.Expression{
						{genLiteralExpression("1"), genLiteralExpression("2")},
					},
					ReturningClause: &tree.ReturningClause{
						Returned: []*tree.ReturningClauseColumn{
							{
								Expression: genLiteralExpression("1"),
							},
						},
					}}}},
		{"insert with returning *", "insert into t1 (a,b) values (1,2) returning *",
			&tree.Insert{
				InsertStmt: &tree.InsertStmt{
					Table:      "t1",
					Columns:    []string{"a", "b"},
					InsertType: tree.InsertTypeInsert,
					Values: [][]tree.Expression{
						{genLiteralExpression("1"), genLiteralExpression("2")},
					},
					ReturningClause: &tree.ReturningClause{
						Returned: []*tree.ReturningClauseColumn{
							{
								All: true,
							},
						},
					}}}},
		{"insert with values upsert without target do nothing", "insert into t1 (a,b) values (1,2) on conflict do nothing",
			&tree.Insert{
				InsertStmt: &tree.InsertStmt{
					Table:      "t1",
					Columns:    []string{"a", "b"},
					InsertType: tree.InsertTypeInsert,
					Upsert: &tree.Upsert{
						Type: tree.UpsertTypeDoNothing,
					},
					Values: [][]tree.Expression{
						{genLiteralExpression("1"), genLiteralExpression("2")},
					}}}},
		{"insert with values upsert with target without where do nothing",
			"insert into t1 (a,b) values (1,2) on conflict (c1,c2) do nothing",
			&tree.Insert{
				InsertStmt: &tree.InsertStmt{
					Table:      "t1",
					Columns:    []string{"a", "b"},
					InsertType: tree.InsertTypeInsert,
					Upsert: &tree.Upsert{
						ConflictTarget: &tree.ConflictTarget{
							IndexedColumns: []string{"c1", "c2"},
						},
						Type: tree.UpsertTypeDoNothing,
					},
					Values: [][]tree.Expression{
						{genLiteralExpression("1"), genLiteralExpression("2")},
					}}}},
		{"insert with values upsert with target and where do nothing",
			"insert into t1 (a,b) values (1,2) on conflict(c1,c2) where 1 do nothing",
			&tree.Insert{
				InsertStmt: &tree.InsertStmt{
					Table:      "t1",
					Columns:    []string{"a", "b"},
					InsertType: tree.InsertTypeInsert,
					Upsert: &tree.Upsert{
						ConflictTarget: &tree.ConflictTarget{
							IndexedColumns: []string{"c1", "c2"},
							Where:          &tree.ExpressionLiteral{Value: "1"},
						},
						Type: tree.UpsertTypeDoNothing,
					},
					Values: [][]tree.Expression{
						{genLiteralExpression("1"), genLiteralExpression("2")},
					}}}},
		{"insert with values upsert with update column name",
			"insert into t1 (a,b) values (1,2) on conflict do update set b=1",
			&tree.Insert{
				InsertStmt: &tree.InsertStmt{
					Table:      "t1",
					Columns:    []string{"a", "b"},
					InsertType: tree.InsertTypeInsert,
					Upsert: &tree.Upsert{
						Type: tree.UpsertTypeDoUpdate,
						Updates: []*tree.UpdateSetClause{
							{Columns: []string{"b"},
								Expression: genLiteralExpression("1")},
						},
					},
					Values: [][]tree.Expression{
						{genLiteralExpression("1"), genLiteralExpression("2")},
					}}}},
		{"insert with values upsert with update multi column name",
			"insert into t1 (a,b) values (1,2) on conflict do update set b=1,c=2",
			&tree.Insert{
				InsertStmt: &tree.InsertStmt{
					Table:      "t1",
					Columns:    []string{"a", "b"},
					InsertType: tree.InsertTypeInsert,
					Upsert: &tree.Upsert{
						Type: tree.UpsertTypeDoUpdate,
						Updates: []*tree.UpdateSetClause{
							{
								Columns:    []string{"b"},
								Expression: genLiteralExpression("1"),
							},
							{
								Columns:    []string{"c"},
								Expression: genLiteralExpression("2"),
							},
						},
					},
					Values: [][]tree.Expression{
						{genLiteralExpression("1"), genLiteralExpression("2")},
					}}}},
		{"insert with values upsert with update column name list",
			"insert into t1 (a,b) values (1,2) on conflict do update set (b,c)=(1,2)",
			&tree.Insert{
				InsertStmt: &tree.InsertStmt{
					Table:      "t1",
					Columns:    []string{"a", "b"},
					InsertType: tree.InsertTypeInsert,
					Upsert: &tree.Upsert{
						Type: tree.UpsertTypeDoUpdate,
						Updates: []*tree.UpdateSetClause{
							{
								Columns: []string{"b", "c"},
								Expression: &tree.ExpressionList{
									Expressions: []tree.Expression{
										genLiteralExpression("1"),
										genLiteralExpression("2"),
									}},
							},
						},
					},
					Values: [][]tree.Expression{
						{genLiteralExpression("1"), genLiteralExpression("2")},
					}}}},
		{"insert with values upsert with update multi column name list",
			"insert into t1 (a,b) values (1,2) on conflict do update set (b,c)=(1,2), (d,e)=(3,4)",
			&tree.Insert{
				InsertStmt: &tree.InsertStmt{
					Table:      "t1",
					Columns:    []string{"a", "b"},
					InsertType: tree.InsertTypeInsert,
					Upsert: &tree.Upsert{
						Type: tree.UpsertTypeDoUpdate,
						Updates: []*tree.UpdateSetClause{
							{
								Columns: []string{"b", "c"},
								Expression: &tree.ExpressionList{
									Expressions: []tree.Expression{
										genLiteralExpression("1"),
										genLiteralExpression("2"),
									}},
							},
							{
								Columns: []string{"d", "e"},
								Expression: &tree.ExpressionList{
									Expressions: []tree.Expression{
										&tree.ExpressionLiteral{Value: "3"},
										&tree.ExpressionLiteral{Value: "4"},
									}},
							},
						},
					},
					Values: [][]tree.Expression{
						{genLiteralExpression("1"), genLiteralExpression("2")},
					}}}},
		{"insert with values upsert with update and where",
			"insert into t1 (a,b) values (1,2) on conflict do update set b=1 where 1",
			&tree.Insert{
				InsertStmt: &tree.InsertStmt{
					Table:      "t1",
					Columns:    []string{"a", "b"},
					InsertType: tree.InsertTypeInsert,
					Upsert: &tree.Upsert{
						Type: tree.UpsertTypeDoUpdate,
						Updates: []*tree.UpdateSetClause{
							{Columns: []string{"b"},
								Expression: genLiteralExpression("1")},
						},
						Where: genLiteralExpression("1"),
					},
					Values: [][]tree.Expression{
						{genLiteralExpression("1"), genLiteralExpression("2")},
					}}}},
		//// select
		{"select *", "select * from t1",
			&tree.Select{
				SelectStmt: &tree.SelectStmt{
					SelectCores: []*tree.SelectCore{
						{
							SelectType: tree.SelectTypeAll,
							Columns:    columnStar,
							From: &tree.FromClause{JoinClause: &tree.JoinClause{
								TableOrSubquery: &tree.TableOrSubqueryTable{
									Name: "t1",
								},
							}},
						},
					},
				}}},
		{"select with cte", "with t as (select 1) select * from t1",
			&tree.Select{
				CTE: []*tree.CTE{genSimpleCTETree("t", "1")},
				SelectStmt: &tree.SelectStmt{
					SelectCores: []*tree.SelectCore{
						{
							SelectType: tree.SelectTypeAll,
							Columns:    columnStar,
							From: &tree.FromClause{JoinClause: &tree.JoinClause{
								TableOrSubquery: &tree.TableOrSubqueryTable{
									Name: "t1",
								},
							}},
						},
					},
				}}},
		{"select distinct", "select distinct * from t1",
			&tree.Select{
				SelectStmt: &tree.SelectStmt{
					SelectCores: []*tree.SelectCore{
						{
							SelectType: tree.SelectTypeDistinct,
							Columns:    columnStar,
							From: &tree.FromClause{JoinClause: &tree.JoinClause{
								TableOrSubquery: &tree.TableOrSubqueryTable{
									Name: "t1",
								},
							}},
						},
					},
				}}},
		{"select with where", "select * from t1 where c1=1",
			&tree.Select{
				SelectStmt: &tree.SelectStmt{
					SelectCores: []*tree.SelectCore{
						{
							SelectType: tree.SelectTypeAll,
							Columns:    columnStar,
							From: &tree.FromClause{JoinClause: &tree.JoinClause{
								TableOrSubquery: &tree.TableOrSubqueryTable{
									Name: "t1",
								},
							}},
							Where: &tree.ExpressionBinaryComparison{
								Left:     &tree.ExpressionColumn{Column: "c1"},
								Operator: tree.ComparisonOperatorEqual,
								Right:    genLiteralExpression("1"),
							},
						},
					},
				}}},
		{"select with where and", "select * from t1 where c1=1 and c2=2",
			&tree.Select{
				SelectStmt: &tree.SelectStmt{
					SelectCores: []*tree.SelectCore{
						{
							SelectType: tree.SelectTypeAll,
							Columns:    columnStar,
							From: &tree.FromClause{JoinClause: &tree.JoinClause{
								TableOrSubquery: &tree.TableOrSubqueryTable{
									Name: "t1",
								},
							}},
							Where: &tree.ExpressionBinaryComparison{
								Left: &tree.ExpressionBinaryComparison{
									Left:     &tree.ExpressionColumn{Column: "c1"},
									Operator: tree.ComparisonOperatorEqual,
									Right:    genLiteralExpression("1"),
								},
								Operator: tree.LogicalOperatorAnd,
								Right: &tree.ExpressionBinaryComparison{
									Left:     &tree.ExpressionColumn{Column: "c2"},
									Operator: tree.ComparisonOperatorEqual,
									Right:    genLiteralExpression("2"),
								},
							},
						},
					},
				}}},
		{"select with where or", "select * from t1 where c1=1 or c2=2",
			&tree.Select{
				SelectStmt: &tree.SelectStmt{
					SelectCores: []*tree.SelectCore{
						{
							SelectType: tree.SelectTypeAll,
							Columns:    columnStar,
							From: &tree.FromClause{JoinClause: &tree.JoinClause{
								TableOrSubquery: &tree.TableOrSubqueryTable{
									Name: "t1",
								},
							}},
							Where: &tree.ExpressionBinaryComparison{
								Left: &tree.ExpressionBinaryComparison{
									Left:     &tree.ExpressionColumn{Column: "c1"},
									Operator: tree.ComparisonOperatorEqual,
									Right:    genLiteralExpression("1"),
								},
								Operator: tree.LogicalOperatorOr,
								Right: &tree.ExpressionBinaryComparison{
									Left:     &tree.ExpressionColumn{Column: "c2"},
									Operator: tree.ComparisonOperatorEqual,
									Right:    genLiteralExpression("2"),
								},
							},
						},
					},
				}}},
		{"select with group by", "select * from t1 group by c1",
			&tree.Select{
				SelectStmt: &tree.SelectStmt{
					SelectCores: []*tree.SelectCore{
						{
							SelectType: tree.SelectTypeAll,
							Columns:    columnStar,
							From: &tree.FromClause{JoinClause: &tree.JoinClause{
								TableOrSubquery: &tree.TableOrSubqueryTable{
									Name: "t1",
								},
							}},
							GroupBy: &tree.GroupBy{
								Expressions: []tree.Expression{
									&tree.ExpressionColumn{Column: "c1"},
								},
							},
						},
					},
				}}},
		{"select with group by and having", "select * from t1 group by c1 having 1",
			&tree.Select{
				SelectStmt: &tree.SelectStmt{
					SelectCores: []*tree.SelectCore{
						{
							SelectType: tree.SelectTypeAll,
							Columns:    columnStar,
							From: &tree.FromClause{JoinClause: &tree.JoinClause{
								TableOrSubquery: &tree.TableOrSubqueryTable{
									Name: "t1",
								},
							}},
							GroupBy: &tree.GroupBy{
								Expressions: []tree.Expression{
									&tree.ExpressionColumn{Column: "c1"},
								},
								Having: genLiteralExpression("1"),
							},
						},
					},
				}}},
		{"select with order by", "select * from t1 order by c1",
			&tree.Select{
				SelectStmt: &tree.SelectStmt{
					SelectCores: []*tree.SelectCore{
						{
							SelectType: tree.SelectTypeAll,
							Columns:    columnStar,
							From: &tree.FromClause{JoinClause: &tree.JoinClause{
								TableOrSubquery: &tree.TableOrSubqueryTable{
									Name: "t1",
								},
							}},
						},
					},
					OrderBy: &tree.OrderBy{
						OrderingTerms: []*tree.OrderingTerm{
							{
								Expression: &tree.ExpressionColumn{Column: "c1"},
							},
						},
					},
				}}},

		{"select with order by all", "select * from t1 order by c1 collate nocase asc nulls first",
			&tree.Select{
				SelectStmt: &tree.SelectStmt{
					SelectCores: []*tree.SelectCore{
						{
							SelectType: tree.SelectTypeAll,
							Columns:    columnStar,
							From: &tree.FromClause{JoinClause: &tree.JoinClause{
								TableOrSubquery: &tree.TableOrSubqueryTable{
									Name: "t1",
								},
							}},
						},
					},
					OrderBy: &tree.OrderBy{
						OrderingTerms: []*tree.OrderingTerm{
							{
								Expression: &tree.ExpressionCollate{
									Expression: &tree.ExpressionColumn{Column: "c1"},
									Collation:  tree.CollationTypeNoCase,
								},
								OrderType:    tree.OrderTypeAsc,
								NullOrdering: tree.NullOrderingTypeFirst,
							},
						},
					},
				}}},
		{"select with limit", "select * from t1 limit 1",
			&tree.Select{
				SelectStmt: &tree.SelectStmt{
					SelectCores: []*tree.SelectCore{
						{
							SelectType: tree.SelectTypeAll,
							Columns:    columnStar,
							From: &tree.FromClause{JoinClause: &tree.JoinClause{
								TableOrSubquery: &tree.TableOrSubqueryTable{
									Name: "t1",
								},
							}},
						},
					},
					Limit: &tree.Limit{Expression: genLiteralExpression("1")},
				}}},
		{"select with limit offset", "select * from t1 limit 1 offset 2",
			&tree.Select{
				SelectStmt: &tree.SelectStmt{
					SelectCores: []*tree.SelectCore{
						{
							SelectType: tree.SelectTypeAll,
							Columns:    columnStar,
							From: &tree.FromClause{JoinClause: &tree.JoinClause{
								TableOrSubquery: &tree.TableOrSubqueryTable{
									Name: "t1",
								},
							}},
						},
					},
					Limit: &tree.Limit{
						Expression: genLiteralExpression("1"),
						Offset:     genLiteralExpression("2"),
					},
				}}},
		//// join
		{"join on", "select * from t1 join t2 on t1.c1=t2.c1",
			genSimpleJoinSelectTree(&tree.JoinOperator{
				JoinType: tree.JoinTypeJoin,
			}, "t1", "c1", "t2", "c1")},
		{"left join on", "select * from t1 left join t2 on t1.c1=t2.c1",
			genSimpleJoinSelectTree(&tree.JoinOperator{
				JoinType: tree.JoinTypeLeft,
			}, "t1", "c1", "t2", "c1")},
		{"left outer join on", "select * from t1 left outer join t2 on t1.c1=t2.c1",
			genSimpleJoinSelectTree(&tree.JoinOperator{
				JoinType: tree.JoinTypeLeft,
				Outer:    true,
			}, "t1", "c1", "t2", "c1")},
		{"right join on", "select * from t1 right join t2 on t1.c1=t2.c1",
			genSimpleJoinSelectTree(&tree.JoinOperator{
				JoinType: tree.JoinTypeRight,
			}, "t1", "c1", "t2", "c1")},
		{"right outer join on", "select * from t1 right outer join t2 on t1.c1=t2.c1",
			genSimpleJoinSelectTree(&tree.JoinOperator{
				JoinType: tree.JoinTypeRight,
				Outer:    true,
			}, "t1", "c1", "t2", "c1")},
		{"full join on", "select * from t1 full join t2 on t1.c1=t2.c1",
			genSimpleJoinSelectTree(&tree.JoinOperator{
				JoinType: tree.JoinTypeFull,
			}, "t1", "c1", "t2", "c1")},
		{"full outer join on", "select * from t1 full outer join t2 on t1.c1=t2.c1",
			genSimpleJoinSelectTree(&tree.JoinOperator{
				JoinType: tree.JoinTypeFull,
				Outer:    true,
			}, "t1", "c1", "t2", "c1")},
		{"inner join on", "select * from t1 inner join t2 on t1.c1=t2.c1",
			genSimpleJoinSelectTree(&tree.JoinOperator{
				JoinType: tree.JoinTypeInner,
			}, "t1", "c1", "t2", "c1")},
		{"join multi", "select * from t1 join t2 on t1.c1=t2.c1 left join t3 on t1.c1=t3.c1",
			&tree.Select{
				SelectStmt: &tree.SelectStmt{
					SelectCores: []*tree.SelectCore{
						{
							SelectType: tree.SelectTypeAll,
							Columns:    columnStar,
							From: &tree.FromClause{JoinClause: &tree.JoinClause{
								TableOrSubquery: &tree.TableOrSubqueryTable{
									Name: "t1",
								},
								Joins: []*tree.JoinPredicate{
									{
										JoinOperator: &tree.JoinOperator{
											JoinType: tree.JoinTypeJoin,
										},
										Table: &tree.TableOrSubqueryTable{Name: "t2"},
										Constraint: &tree.ExpressionBinaryComparison{
											Operator: tree.ComparisonOperatorEqual,
											Left:     &tree.ExpressionColumn{Table: "t1", Column: "c1"},
											Right:    &tree.ExpressionColumn{Table: "t2", Column: "c1"},
										},
									},
									{
										JoinOperator: &tree.JoinOperator{
											JoinType: tree.JoinTypeLeft,
										},
										Table: &tree.TableOrSubqueryTable{Name: "t3"},
										Constraint: &tree.ExpressionBinaryComparison{
											Operator: tree.ComparisonOperatorEqual,
											Left:     &tree.ExpressionColumn{Table: "t1", Column: "c1"},
											Right:    &tree.ExpressionColumn{Table: "t3", Column: "c1"},
										},
									},
								},
							},
							},
						},
					},
				}}},

		//// update stmt
		{"update", "update t1 set c1=1",
			genSimpleUpdateTree(&tree.QualifiedTableName{TableName: "t1"}, "c1", "1")},
		{"update with table alias", "update t1 as t set c1=1",
			genSimpleUpdateTree(&tree.QualifiedTableName{
				TableName:  "t1",
				TableAlias: "t",
			}, "c1", "1")},
		{"update with multi set", "update t1 set c1=1, c2=2",
			&tree.Update{
				UpdateStmt: &tree.UpdateStmt{
					QualifiedTableName: &tree.QualifiedTableName{TableName: "t1"},
					UpdateSetClause: []*tree.UpdateSetClause{
						{
							Columns:    []string{"c1"},
							Expression: genLiteralExpression("1"),
						},
						{
							Columns:    []string{"c2"},
							Expression: genLiteralExpression("2"),
						},
					},
				},
			},
		},
		{"update with column list set", "update t1 set (c1, c2)=(1,2)",
			&tree.Update{
				UpdateStmt: &tree.UpdateStmt{
					QualifiedTableName: &tree.QualifiedTableName{TableName: "t1"},
					UpdateSetClause: []*tree.UpdateSetClause{
						{
							Columns: []string{"c1", "c2"},
							Expression: &tree.ExpressionList{
								Expressions: []tree.Expression{
									genLiteralExpression("1"),
									genLiteralExpression("2"),
								},
							},
						},
					},
				},
			},
		},
		{"update from table", "update t1 set c1=1 from t2",
			&tree.Update{
				UpdateStmt: &tree.UpdateStmt{
					QualifiedTableName: &tree.QualifiedTableName{TableName: "t1"},
					UpdateSetClause: []*tree.UpdateSetClause{
						{
							Columns:    []string{"c1"},
							Expression: genLiteralExpression("1"),
						},
					},
					From: &tree.FromClause{JoinClause: &tree.JoinClause{
						TableOrSubquery: &tree.TableOrSubqueryTable{Name: "t2"},
					}},
				},
			},
		},
		{"update from join", "update t1 set c1=1 from t2 join t3 on t2.c1=t3.c1",
			&tree.Update{
				UpdateStmt: &tree.UpdateStmt{
					QualifiedTableName: &tree.QualifiedTableName{TableName: "t1"},
					UpdateSetClause: []*tree.UpdateSetClause{
						{
							Columns:    []string{"c1"},
							Expression: genLiteralExpression("1"),
						},
					},
					From: &tree.FromClause{
						JoinClause: &tree.JoinClause{
							TableOrSubquery: &tree.TableOrSubqueryTable{Name: "t2"},
							Joins: []*tree.JoinPredicate{
								{
									JoinOperator: &tree.JoinOperator{
										JoinType: tree.JoinTypeJoin,
									},
									Table: &tree.TableOrSubqueryTable{Name: "t3"},
									Constraint: &tree.ExpressionBinaryComparison{
										Operator: tree.ComparisonOperatorEqual,
										Left:     &tree.ExpressionColumn{Table: "t2", Column: "c1"},
										Right:    &tree.ExpressionColumn{Table: "t3", Column: "c1"},
									},
								},
							},
						},
					},
				},
			},
		},
		{"update where", "update t1 set c1=1 where c2=1",
			&tree.Update{
				UpdateStmt: &tree.UpdateStmt{
					QualifiedTableName: &tree.QualifiedTableName{TableName: "t1"},
					UpdateSetClause: []*tree.UpdateSetClause{
						{
							Columns:    []string{"c1"},
							Expression: genLiteralExpression("1"),
						},
					},
					Where: &tree.ExpressionBinaryComparison{
						Operator: tree.ComparisonOperatorEqual,
						Left:     &tree.ExpressionColumn{Column: "c2"},
						Right:    genLiteralExpression("1"),
					},
				},
			},
		},
		{"update returning", "update t1 set c1=1 returning *",
			&tree.Update{
				UpdateStmt: &tree.UpdateStmt{
					QualifiedTableName: &tree.QualifiedTableName{TableName: "t1"},
					UpdateSetClause: []*tree.UpdateSetClause{
						{
							Columns:    []string{"c1"},
							Expression: genLiteralExpression("1"),
						},
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
		},
		{"update with cte", "with t as (select 1) update t1 set c1=1",
			&tree.Update{
				CTE: []*tree.CTE{
					genSimpleCTETree("t", "1"),
				},
				UpdateStmt: &tree.UpdateStmt{
					QualifiedTableName: &tree.QualifiedTableName{TableName: "t1"},
					UpdateSetClause: []*tree.UpdateSetClause{
						{
							Columns:    []string{"c1"},
							Expression: genLiteralExpression("1"),
						},
					},
				},
			},
		},

		//// delete
		{"delete all", "delete from t1",
			&tree.Delete{
				DeleteStmt: &tree.DeleteStmt{
					QualifiedTableName: &tree.QualifiedTableName{
						TableName: "t1",
					},
				},
			},
		},
		{"delete with where", "delete from t1 where c1='1'",
			&tree.Delete{
				DeleteStmt: &tree.DeleteStmt{
					QualifiedTableName: &tree.QualifiedTableName{
						TableName: "t1",
					},
					Where: &tree.ExpressionBinaryComparison{
						Operator: tree.ComparisonOperatorEqual,
						Left:     &tree.ExpressionColumn{Column: "c1"},
						Right:    genLiteralExpression("'1'"),
					},
				},
			},
		},
		{"delete with returning", "delete from t1 returning *",
			&tree.Delete{
				DeleteStmt: &tree.DeleteStmt{
					QualifiedTableName: &tree.QualifiedTableName{
						TableName: "t1",
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
		},
		{"delete with cte", "with t as (select 1) delete from t1",
			&tree.Delete{
				CTE: []*tree.CTE{
					genSimpleCTETree("t", "1"),
				},
				DeleteStmt: &tree.DeleteStmt{
					QualifiedTableName: &tree.QualifiedTableName{
						TableName: "t1",
					},
				},
			},
		},
		//// identifier quotes, `"` and `[]` and "`"
		{"table name with double quote", `select * from "t1"`,
			&tree.Select{
				SelectStmt: &tree.SelectStmt{
					SelectCores: []*tree.SelectCore{
						{
							SelectType: tree.SelectTypeAll,
							Columns:    columnStar,
							From: &tree.FromClause{
								JoinClause: &tree.JoinClause{
									TableOrSubquery: &tree.TableOrSubqueryTable{
										Name: "t1",
									},
								},
							},
						},
					},
				},
			},
		},
		{"table name alias with double quote", `select * from "t1" as "t"`,
			&tree.Select{
				SelectStmt: &tree.SelectStmt{
					SelectCores: []*tree.SelectCore{
						{
							SelectType: tree.SelectTypeAll,
							Columns:    columnStar,
							From: &tree.FromClause{
								JoinClause: &tree.JoinClause{
									TableOrSubquery: &tree.TableOrSubqueryTable{
										Name:  "t1",
										Alias: "t",
									},
								},
							},
						},
					},
				},
			},
		},
		{"column name with bracket quote", `select [col1] from "t1"`,
			&tree.Select{
				SelectStmt: &tree.SelectStmt{
					SelectCores: []*tree.SelectCore{
						{
							SelectType: tree.SelectTypeAll,
							Columns: []tree.ResultColumn{
								&tree.ResultColumnExpression{
									Expression: &tree.ExpressionColumn{
										Column: "col1",
									},
								},
							},
							From: &tree.FromClause{
								JoinClause: &tree.JoinClause{
									TableOrSubquery: &tree.TableOrSubqueryTable{
										Name: "t1",
									},
								},
							},
						},
					},
				},
			},
		},
		{"column name alias with bracket quote", `select [col1] as [col] from t1`,
			&tree.Select{
				SelectStmt: &tree.SelectStmt{
					SelectCores: []*tree.SelectCore{
						{
							SelectType: tree.SelectTypeAll,
							Columns: []tree.ResultColumn{
								&tree.ResultColumnExpression{
									Expression: &tree.ExpressionColumn{
										Table:  "",
										Column: "col1",
									},
									Alias: "col",
								},
							}, From: &tree.FromClause{
								JoinClause: &tree.JoinClause{
									TableOrSubquery: &tree.TableOrSubqueryTable{
										Name: "t1",
									},
								},
							},
						},
					},
				},
			},
		},
		{"collation name with back tick quote", "select `col1` COLLATE `nocase` from `t1`; ",
			&tree.Select{
				SelectStmt: &tree.SelectStmt{
					SelectCores: []*tree.SelectCore{
						{
							SelectType: tree.SelectTypeAll,
							Columns: []tree.ResultColumn{
								&tree.ResultColumnExpression{
									Expression: &tree.ExpressionCollate{
										Expression: &tree.ExpressionColumn{
											Table:  "",
											Column: "col1",
										},
										Collation: tree.CollationTypeNoCase,
									},
								},
							},
							From: &tree.FromClause{
								JoinClause: &tree.JoinClause{
									TableOrSubquery: &tree.TableOrSubqueryTable{
										Name: "t1",
									},
								},
							},
						},
					},
				},
			},
		},
		{"function name with back tick quote", "select `abs`(1)",
			genSimpleFunctionSelectTree(&tree.FunctionABS, genLiteralExpression("1"))},

		//// type cast
		{"type cast",
			"select 1::int as x, @caller::text, t1.c1::text, (t1.c2::int * 3)::int, " +
				"(t1.c3 isnull)::int, abs(2)::int from t1;",
			&tree.Select{
				SelectStmt: &tree.SelectStmt{
					SelectCores: []*tree.SelectCore{
						{
							SelectType: tree.SelectTypeAll,
							Columns: []tree.ResultColumn{
								&tree.ResultColumnExpression{
									Expression: &tree.ExpressionLiteral{
										Value:    "1",
										TypeCast: tree.TypeCastInt,
									},
									Alias: "x",
								},
								&tree.ResultColumnExpression{
									Expression: &tree.ExpressionBindParameter{
										Parameter: "@caller",
										TypeCast:  tree.TypeCastText,
									},
								},
								&tree.ResultColumnExpression{
									Expression: &tree.ExpressionColumn{
										Table:    "t1",
										Column:   "c1",
										TypeCast: tree.TypeCastText,
									},
								},
								&tree.ResultColumnExpression{
									Expression: &tree.ExpressionArithmetic{
										Wrapped:  true,
										TypeCast: tree.TypeCastInt,
										Left: &tree.ExpressionColumn{
											Table:    "t1",
											Column:   "c2",
											TypeCast: tree.TypeCastInt,
										},
										Operator: tree.ArithmeticOperatorMultiply,
										Right:    genLiteralExpression("3"),
									},
								},
								&tree.ResultColumnExpression{
									Expression: &tree.ExpressionIs{
										Left: &tree.ExpressionColumn{
											Table:  "t1",
											Column: "c3",
										},
										Right:    genLiteralExpression("NULL"),
										Wrapped:  true,
										TypeCast: tree.TypeCastInt,
									},
								},
								&tree.ResultColumnExpression{
									Expression: &tree.ExpressionFunction{
										Function: &tree.FunctionABS,
										Inputs:   []tree.Expression{genLiteralExpression("2")},
										Distinct: false,
										TypeCast: tree.TypeCastInt,
									},
								},
							},
							From: &tree.FromClause{
								JoinClause: &tree.JoinClause{
									TableOrSubquery: &tree.TableOrSubqueryTable{
										Name: "t1",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// Replace multiple spaces with a single space
	re := regexp.MustCompile(`\s+`)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				tt.expect = nil
			}()

			astTree, err := ParseSql(tt.input, 1, nil, *traceMode, false)
			if err != nil {
				t.Errorf("ParseRawSQL() got %s", err)
				return
			}

			// use assert.Exactly?
			assert.EqualValues(t, tt.expect, astTree, "ParseRawSQL() got %+v, want %+v", astTree, tt.expect)

			sql, err := tree.SafeToSQL(astTree)
			if err != nil {
				t.Errorf("ParseRawSQL() got %s", err)
			}

			if *traceMode {
				fmt.Println("SQL from AST: ", sql)
			}

			singleSpaceSql := re.ReplaceAllString(sql, " ")
			t.Logf("%s \n=> %s\n", tt.input, singleSpaceSql)

			if !(strings.Contains(tt.input, "isnull") || strings.Contains(tt.input, "notnull")) {
				// assert original sql and sql from ast are equal, WITHOUT format
				assert.True(t,
					strings.EqualFold(unFormatSql(sql), unFormatSql(tt.input)),
					"ParseRawSQL() got %s, origin %s", sql, tt.input)
			}

			err = postgres.CheckSyntaxReplaceDollar(sql)
			assert.NoErrorf(t, err, "postgres syntax check failed: %s", err)
		})
	}
}

// unFormatSql remove unnecessary spaces, quotes, etc.
func unFormatSql(sql string) string {
	sql = strings.ReplaceAll(sql, " ", "")
	sql = strings.ReplaceAll(sql, ";", "")
	//// double quotes are for table/column names, astTree.ToSQL() will add them
	sql = strings.ReplaceAll(sql, `"`, "")
	// those markers are not supported by astTree.ToSQL(), but they are supported by sqlite(from original input)
	sql = strings.ReplaceAll(sql, `[`, "")
	sql = strings.ReplaceAll(sql, `]`, "")
	sql = strings.ReplaceAll(sql, "`", "")

	return sql
}

func TestParseRawSQL_syntax_invalid(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		causeSymbol string
	}{
		// literal value
		{"blob", "select x'01'", "x'01'"},

		// bind parameter
		{"expr bind parameter ?", "select ?", "?"},
		{"expr bind parameter ?1", "select ?1", "?"},
		{"expr bind parameter :a", "select :a", ":"},

		// common table stmt
		{"cte recursive", "with recursive t1(c1,c2) as (select 1) select * from t1", "recursive"},
		// common table expression
		{"cte not", "with t1(c1,c2) as (select 1) not select * from t1", "not"},
		{"cte materialized", "with t1(c1,c2) as (select 1) materialized select * from t1", "materialized"},

		// table or subquery
		{"table or subquery indexed", "select * from t1 indexed by index_a", "indexed"},
		{"table or subquery not indexed", "select * from t1 not indexed", "not"},
		// NOTE: what is table function??
		{"table or subquery table function", "SELECT value FROM f(1)", "("},
		{"table or subquery nest tos", "select * from (t1, t2)", ","},

		// expr
		{"expr names", "select schema.table.column", "."}, // no schema
		{"expr cast", "select cast(true as aaa)", "cast"},
		{"expr function with over", "select abs(1) over (partition by 1)", "over"},
		//{"expr raise", "select raise(fail, 'dsd')", "raise"},

		// insert
		{"insert or abort", "insert or abort into t1 values (1)", "abort"},
		{"insert or fail", "insert or fail into t1 values (1)", "fail"},
		{"insert or ignore", "insert or ignore into t1 values (1)", "ignore"},
		{"insert or rollback", "insert or rollback into t1 values (1)", "rollback"},
		{"insert schema_name", "insert or replace into schema.t1 values (1)", "."},
		{"insert into with select", "insert into t1 as tt with t1 as (select 1) select * from t2", "with"},
		{"insert into select", "insert into t1 as tt select * from t2", "select"},
		{"insert into default values", "insert into t1 default values", "default"},
		//wrong indexed_column syntax
		//"insert into t1 (a,b) values (1,2) on conflict (c1 collate collate_name asc) do nothing",

		// select
		{"select all", "select all c1 from t1", "c1"},
		{"select with window", "select * from t1 window w1 as (partition by 1)", "window"},
		{"select with not supported null ordering", "select * from t1 order by c1 nulls firstss", "firstss"},
		{"select values", "values (1)", "values"},
		{"select values with cte", "with t as (select 1) values (1)", "values"},
		{"select with compound operator and values", "select * from t1 union values (1)", "values"},

		// function
		{"expr function with filter", "select f(1) filter (where 1)", "filter"},

		// join
		{"cross join", "select * from t3 cross join t4", "cross"},
		{"join using", "select * from t3 join t4 using (c1)", "using"},
		{"join without condition", "select * from t3 join t4", "<EOF>"},
		{"comma cartesian join 1", "select * from t3, t4", ","},

		// other statement
		{"explain", "explain select * from t1", "explain"},
		{"explain query plan", "explain query plan select * from t1", "explain"},
		{"create table", "create table t1 (c1 int)", "create"},
		{"drop table", "drop table t1", "drop"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//eh := NewErrorHandler(1)
			//el := newSqliteErrorListener(eh)
			_, err := ParseSql(tt.input, 1, nil, *traceMode, false)
			assert.Errorf(t, err, "Parser should complain abould invalid syntax")

			if !errors.Is(err, ErrInvalidSyntax) {
				t.Fatalf("ParseRawSQL() expected error: %s, got %s", ErrInvalidSyntax, err)
			}

			//if el.symbol != tt.causeSymbol {
			//	t.Errorf("ParseRawSQL() expected cause symbol: %s, got: %s", tt.causeSymbol, el.symbol)
			//}
		})
	}
}

func TestParseRawSQL_semantic_invalid(t *testing.T) {
	// TODO: we probably should move all semantic checks to analysis phase
	// but some semantic checks need catalog, some don't, like these
	// let's keep it here for now
	tests := []struct {
		name   string
		input  string
		reason string
	}{
		// type cast
		{"type cast not supported type", "select 1::random", "panic: unknown type cast random"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseSql(tt.input, 1, nil, *traceMode, false)
			assert.Errorf(t, err, "Parser should complain abould invalid syntax")

			// should panic, which is caught by ParseRawSQL
			assert.Contains(t, err.Error(), tt.reason)
		})
	}
}
