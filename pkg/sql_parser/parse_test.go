package sql_parser

import (
	"flag"
	"fmt"
	"strings"
	"testing"

	"github.com/kwilteam/kwil-db/pkg/engine/tree"
	"github.com/stretchr/testify/assert"
)

var trace = flag.Bool("trace", false, "run tests with tracing")

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

func genSimpleExprNullSelectTree(value string, isNull bool) *tree.Select {
	return &tree.Select{
		SelectStmt: &tree.SelectStmt{
			SelectCores: []*tree.SelectCore{
				{
					SelectType: tree.SelectTypeAll,
					Columns: []tree.ResultColumn{
						&tree.ResultColumnExpression{
							Expression: &tree.ExpressionIsNull{
								Expression: &tree.ExpressionLiteral{Value: value},
								IsNull:     isNull,
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

func genSimpleUpdateOrTree(qt *tree.QualifiedTableName, uo tree.UpdateOr, column, value string) *tree.Update {
	return &tree.Update{
		UpdateStmt: &tree.UpdateStmt{
			Or:                 uo,
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

func TestParseRawSQL_visitor_allowed(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect tree.Ast
	}{
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
		{"table or subquery nest tos", "select * from (t1 as tt, t2 as ttt)",
			&tree.Select{
				SelectStmt: &tree.SelectStmt{
					SelectCores: []*tree.SelectCore{
						{
							SelectType: tree.SelectTypeAll,
							Columns:    columnStar,
							From: &tree.FromClause{
								JoinClause: &tree.JoinClause{
									TableOrSubquery: &tree.TableOrSubqueryList{
										TableOrSubqueries: []tree.TableOrSubquery{
											&tree.TableOrSubqueryTable{Name: "t1", Alias: "tt"},
											&tree.TableOrSubqueryTable{Name: "t2", Alias: "ttt"},
										},
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
												Natural:  false,
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
		{"null", "select null", genSelectColumnLiteralTree("null")},
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
		{"expr names", "select table.column",
			&tree.Select{
				SelectStmt: &tree.SelectStmt{
					SelectCores: []*tree.SelectCore{
						{
							SelectType: tree.SelectTypeAll,
							Columns: []tree.ResultColumn{
								&tree.ResultColumnExpression{
									Expression: &tree.ExpressionColumn{
										Table:  "table",
										Column: "column",
									},
								},
							},
						},
					},
				}}},
		// unary op
		//{"expr unary op +", "select +1"},
		//{"expr unary op -", "select -1"},
		//{"expr unary op ~", "select ~1"},
		//{"expr unary op not", "select not 1"},
		// binary op
		//{"expr binary op ||", "select 1 || 2",
		{"expr binary op *", "select 1 * 2",
			genSimpleBinaryCompareSelectTree(tree.ArithmeticOperatorMultiply, "1", "2")},
		{"expr binary op /", "select 1 / 2",
			genSimpleBinaryCompareSelectTree(tree.ArithmeticOperatorDivide, "1", "2")},
		{"expr binary op %", "select 1 % 2",
			genSimpleBinaryCompareSelectTree(tree.ArithmeticOperatorModulus, "1", "2")},
		{"expr binary op +", "select 1 + 2",
			genSimpleBinaryCompareSelectTree(tree.ArithmeticOperatorAdd, "1", "2")},
		{"expr binary op -", "select 1 - 2",
			genSimpleBinaryCompareSelectTree(tree.ArithmeticOperatorSubtract, "1", "2")},
		{"expr binary op <<", "select 1 << 2",
			genSimpleBinaryCompareSelectTree(tree.BitwiseOperatorLeftShift, "1", "2")},
		{"expr binary op >>", "select 1 >> 2",
			genSimpleBinaryCompareSelectTree(tree.BitwiseOperatorRightShift, "1", "2")},
		{"expr binary op &", "select 1 & 2",
			genSimpleBinaryCompareSelectTree(tree.BitwiseOperatorAnd, "1", "2")},
		{"expr binary op |", "select 1 | 2",
			genSimpleBinaryCompareSelectTree(tree.BitwiseOperatorOr, "1", "2")},
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
		//{"expr binary op <>", "select 1 <> 2"},
		{"expr binary op is", "select 1 is 2",
			genSimpleBinaryCompareSelectTree(tree.ComparisonOperatorIs, "1", "2")},
		{"expr binary op is not", "select 1 is not 2",
			genSimpleBinaryCompareSelectTree(tree.ComparisonOperatorIsNot, "1", "2")},
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
		{"expr binary op match", "select 1 match 2",
			genSimpleStringCompareSelectTree(tree.StringOperatorMatch, "1", "2", "")},
		{"expr binary op match", "select 1 not match 2",
			genSimpleStringCompareSelectTree(tree.StringOperatorNotMatch, "1", "2", "")},
		{"expr binary op regexp", "select 1 regexp 2",
			genSimpleStringCompareSelectTree(tree.StringOperatorRegexp, "1", "2", "")},
		{"expr binary op regexp", "select 1 not regexp 2",
			genSimpleStringCompareSelectTree(tree.StringOperatorNotRegexp, "1", "2", "")},
		{"expr binary op glob", "select 1 glob 2",
			genSimpleStringCompareSelectTree(tree.StringOperatorGlob, "1", "2", "")},
		{"expr binary op glob", "select 1 not glob 2",
			genSimpleStringCompareSelectTree(tree.StringOperatorNotGlob, "1", "2", "")},
		// function
		// core functions
		{"expr function abs", "select abs(1)",
			genSimpleFunctionSelectTree(&tree.FunctionABS, genLiteralExpression("1"))},
		{"expr function coalesce", "select coalesce(1,2)",
			genSimpleFunctionSelectTree(&tree.FunctionCOALESCE,
				genLiteralExpression("1"), genLiteralExpression("2"))},
		{"expr function format", `select format('%d',2)`,
			genSimpleFunctionSelectTree(&tree.FunctionFORMAT,
				genLiteralExpression(`'%d'`), genLiteralExpression("2"))},
		{"expr function glob", `select glob('1','2')`,
			genSimpleFunctionSelectTree(&tree.FunctionGLOB,
				genLiteralExpression(`'1'`), genLiteralExpression(`'2'`))},
		{"expr function hex", `select hex(1)`,
			genSimpleFunctionSelectTree(&tree.FunctionHEX, genLiteralExpression("1"))},
		{"expr function ifnull", `select ifnull(1,2)`,
			genSimpleFunctionSelectTree(&tree.FunctionIFNULL,
				genLiteralExpression("1"), genLiteralExpression("2"))},
		{"expr function iif", `select iif(1,2,3)`,
			genSimpleFunctionSelectTree(&tree.FunctionIIF,
				genLiteralExpression("1"), genLiteralExpression("2"), genLiteralExpression("3"))},
		{"expr function instr", `select instr(1,2)`,
			genSimpleFunctionSelectTree(&tree.FunctionINSTR,
				genLiteralExpression("1"), genLiteralExpression("2"))},
		{"expr function length", `select length(1)`,
			genSimpleFunctionSelectTree(&tree.FunctionLENGTH, genLiteralExpression("1"))},
		{"expr function like", `select like('1','2')`,
			genSimpleFunctionSelectTree(&tree.FunctionLIKE,
				genLiteralExpression("'1'"), genLiteralExpression("'2'"))},
		{"expr function like 2", `select like('1','2', '3')`,
			genSimpleFunctionSelectTree(&tree.FunctionLIKE,
				genLiteralExpression("'1'"), genLiteralExpression("'2'"), genLiteralExpression("'3'"))},
		{"expr function lower", `select lower('Z')`,
			genSimpleFunctionSelectTree(&tree.FunctionLOWER, genLiteralExpression("'Z'"))},
		{"expr function ltrim", `select ltrim('12')`,
			genSimpleFunctionSelectTree(&tree.FunctionLTRIM, genLiteralExpression("'12'"))},
		{"expr function ltrim 2", `select ltrim('12', '1')`,
			genSimpleFunctionSelectTree(&tree.FunctionLTRIM, genLiteralExpression("'12'"), genLiteralExpression("'1'"))},
		{"expr function max", `select max(1)`,
			genSimpleFunctionSelectTree(&tree.FunctionMAX, genLiteralExpression("1"))},
		{"expr function max", `select max(1,2,3)`,
			genSimpleFunctionSelectTree(&tree.FunctionMAX,
				genLiteralExpression("1"), genLiteralExpression("2"), genLiteralExpression("3"))},
		{"expr function min", `select min(1)`,
			genSimpleFunctionSelectTree(&tree.FunctionMIN, genLiteralExpression("1"))},
		{"expr function min", `select min(1,2,3)`,
			genSimpleFunctionSelectTree(&tree.FunctionMIN,
				genLiteralExpression("1"), genLiteralExpression("2"), genLiteralExpression("3"))},
		{"expr function nullif", `select nullif(1,2)`,
			genSimpleFunctionSelectTree(&tree.FunctionNULLIF,
				genLiteralExpression("1"), genLiteralExpression("2"))},
		{"expr function quote", `select quote(1)`,
			genSimpleFunctionSelectTree(&tree.FunctionQUOTE, genLiteralExpression("1"))},
		{"expr function replace", `select replace('1','2','3')`,
			genSimpleFunctionSelectTree(&tree.FunctionREPLACE,
				genLiteralExpression("'1'"), genLiteralExpression("'2'"), genLiteralExpression("'3'"))},
		{"expr function rtrim", `select rtrim('12')`,
			genSimpleFunctionSelectTree(&tree.FunctionRTRIM, genLiteralExpression("'12'"))},
		{"expr function rtrim 2", `select rtrim('12', '1')`,
			genSimpleFunctionSelectTree(&tree.FunctionRTRIM, genLiteralExpression("'12'"), genLiteralExpression("'1'"))},
		{"expr function sign", `select sign(1)`,
			genSimpleFunctionSelectTree(&tree.FunctionSIGN, genLiteralExpression("1"))},
		{"expr function substr", `select substr('1',2)`,
			genSimpleFunctionSelectTree(&tree.FunctionSUBSTR,
				genLiteralExpression("'1'"), genLiteralExpression("2"))},
		{"expr function substr 2", `select substr('1',2,3)`,
			genSimpleFunctionSelectTree(&tree.FunctionSUBSTR,
				genLiteralExpression("'1'"), genLiteralExpression("2"), genLiteralExpression("3"))},
		{"expr function trim", `select trim('12')`,
			genSimpleFunctionSelectTree(&tree.FunctionTRIM, genLiteralExpression("'12'"))},
		{"expr function trim 2", `select trim('12', '1')`,
			genSimpleFunctionSelectTree(&tree.FunctionTRIM, genLiteralExpression("'12'"), genLiteralExpression("'1'"))},
		{"expr function typeof", `select typeof(1)`,
			genSimpleFunctionSelectTree(&tree.FunctionTYPEOF, genLiteralExpression("1"))},
		{"expr function unhex", `select unhex(1)`,
			genSimpleFunctionSelectTree(&tree.FunctionUNHEX, genLiteralExpression("1"))},
		{"expr function unicode", `select unicode('1')`,
			genSimpleFunctionSelectTree(&tree.FunctionUNICODE, genLiteralExpression("'1'"))},
		{"expr function upper", `select upper('z')`,
			genSimpleFunctionSelectTree(&tree.FunctionUPPER, genLiteralExpression("'z'"))},
		// expr datetime functions
		{"expr function date", `select date(1092941466)`,
			genSimpleFunctionSelectTree(&tree.FunctionDATE, genLiteralExpression("1092941466"))},
		{"expr function date 1 modifier", `select date(1092941466,'start of month')`,
			genSimpleFunctionSelectTree(&tree.FunctionDATE,
				genLiteralExpression("1092941466"), genLiteralExpression("'start of month'"))},
		{"expr function date 2 modifiers", `select date(1092941466,'start of month','+1 month')`,
			genSimpleFunctionSelectTree(&tree.FunctionDATE,
				genLiteralExpression("1092941466"), genLiteralExpression("'start of month'"), genLiteralExpression("'+1 month'"))},
		{"expr function time", `select time(1092941466)`,
			genSimpleFunctionSelectTree(&tree.FunctionTIME, genLiteralExpression("1092941466"))},
		{"expr function time 1 modifier", `select time(1092941466,'start of month')`,
			genSimpleFunctionSelectTree(&tree.FunctionTIME,
				genLiteralExpression("1092941466"), genLiteralExpression("'start of month'"))},
		{"expr function time 2 modifiers", `select time(1092941466,'start of month','+1 month')`,
			genSimpleFunctionSelectTree(&tree.FunctionTIME,
				genLiteralExpression("1092941466"), genLiteralExpression("'start of month'"), genLiteralExpression("'+1 month'"))},
		{"expr function datetime", `select datetime(1092941466)`,
			genSimpleFunctionSelectTree(&tree.FunctionDATETIME, genLiteralExpression("1092941466"))},
		{"expr function datetime 1 modifier", `select datetime(1092941466,'start of month')`,
			genSimpleFunctionSelectTree(&tree.FunctionDATETIME,
				genLiteralExpression("1092941466"), genLiteralExpression("'start of month'"))},
		{"expr function datetime 2 modifiers", `select datetime(1092941466,'start of month','+1 month')`,
			genSimpleFunctionSelectTree(&tree.FunctionDATETIME,
				genLiteralExpression("1092941466"), genLiteralExpression("'start of month'"), genLiteralExpression("'+1 month'"))},
		{"expr function strftime", `select strftime('%d',1092941466)`,
			genSimpleFunctionSelectTree(&tree.FunctionSTRFTIME,
				genLiteralExpression("'%d'"), genLiteralExpression("1092941466"))},
		{"expr function strftime 1 modifer", `select strftime('%d',1092941466,'start of month')`,
			genSimpleFunctionSelectTree(&tree.FunctionSTRFTIME,
				genLiteralExpression("'%d'"), genLiteralExpression("1092941466"), genLiteralExpression("'start of month'"))},
		{"expr function strftime 2 modifiers", `select strftime('%d',1092941466,'start of month','+1 month')`,
			genSimpleFunctionSelectTree(&tree.FunctionSTRFTIME,
				genLiteralExpression("'%d'"), genLiteralExpression("1092941466"), genLiteralExpression("'start of month'"), genLiteralExpression("'+1 month'"))},
		{"expr function unixepoch", `select unixepoch(1092941466)`,
			genSimpleFunctionSelectTree(&tree.FunctionUNIXEPOCH, genLiteralExpression("1092941466"))},
		{"expr function unixepoch 1 modifier", `select unixepoch(1092941466,'start of month')`,
			genSimpleFunctionSelectTree(&tree.FunctionUNIXEPOCH,
				genLiteralExpression("1092941466"), genLiteralExpression("'start of month'"))},
		{"expr function unixepoch 2 modifiers", `select unixepoch(1092941466,'start of month','+1 month')`,
			genSimpleFunctionSelectTree(&tree.FunctionUNIXEPOCH,
				genLiteralExpression("1092941466"), genLiteralExpression("'start of month'"), genLiteralExpression("'+1 month'"))},
		// expr list
		{"expr list", "select (1)",
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
		{"expr collate binary", "select 1 collate binary",
			genSimpleCollateSelectTree(tree.CollationTypeBinary, "1")},
		{"expr collate rtrim", "select 1 collate rtrim",
			genSimpleCollateSelectTree(tree.CollationTypeRTrim, "1")},
		// null
		{"expr isnull", "select 1 isnull", genSimpleExprNullSelectTree("1", true)},
		{"expr notnull", "select 1 notnull", genSimpleExprNullSelectTree("1", false)},
		{"expr not null", "select 1 not null", genSimpleExprNullSelectTree("1", false)},
		// is
		{"expr is", "select 1 is 2",
			genSimpleBinaryCompareSelectTree(tree.ComparisonOperatorIs, "1", "2")},
		{"expr is not", "select 1 is not 2",
			genSimpleBinaryCompareSelectTree(tree.ComparisonOperatorIsNot, "1", "2")},
		{"expr is distinct from", "select 1 is distinct from 2",
			&tree.Select{
				SelectStmt: &tree.SelectStmt{
					SelectCores: []*tree.SelectCore{
						{
							SelectType: tree.SelectTypeAll,
							Columns: []tree.ResultColumn{
								&tree.ResultColumnExpression{
									Expression: &tree.ExpressionDistinct{
										Left:  genLiteralExpression("1"),
										Right: genLiteralExpression("2"),
										IsNot: false,
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
									Expression: &tree.ExpressionDistinct{
										Left:  genLiteralExpression("1"),
										Right: genLiteralExpression("2"),
										IsNot: true,
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
		{"insert replace", "replace into t1 values (1)",
			&tree.Insert{
				InsertStmt: &tree.InsertStmt{
					Table:      "t1",
					InsertType: tree.InsertTypeReplace,
					Values:     [][]tree.Expression{{genLiteralExpression("1")}},
				}}},
		{"insert or replace", "insert or replace into t1 values (1)",
			&tree.Insert{
				InsertStmt: &tree.InsertStmt{
					Table:      "t1",
					InsertType: tree.InsertTypeInsertOrReplace,
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
							Where:          &tree.ExpressionLiteral{"1"},
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
								Expression:   &tree.ExpressionColumn{Column: "c1"},
								Collation:    tree.CollationTypeNoCase,
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
		{"select with limit comma", "select * from t1 limit 1,10",
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
						Expression:       genLiteralExpression("1"),
						SecondExpression: genLiteralExpression("10"),
					},
				}}},
		//// join
		//{"join implicit", "select * from t1,t2 on t1.c1=t2.c1"},
		{"join on", "select * from t1 join t2 on t1.c1=t2.c1",
			genSimpleJoinSelectTree(&tree.JoinOperator{
				JoinType: tree.JoinTypeJoin,
			}, "t1", "c1", "t2", "c1")},
		{"natural join on", "select * from t1 natural join t2 on t1.c1=t2.c1",
			genSimpleJoinSelectTree(&tree.JoinOperator{
				JoinType: tree.JoinTypeJoin,
				Natural:  true,
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
		{"natural left join on", "select * from t1 natural left join t2 on t1.c1=t2.c1",
			genSimpleJoinSelectTree(&tree.JoinOperator{
				JoinType: tree.JoinTypeLeft,
				Natural:  true,
			}, "t1", "c1", "t2", "c1")},
		{"natural left outer join on", "select * from t1 natural left outer join t2 on t1.c1=t2.c1",
			genSimpleJoinSelectTree(&tree.JoinOperator{
				JoinType: tree.JoinTypeLeft,
				Outer:    true,
				Natural:  true,
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
		{"natural right join on", "select * from t1 natural right join t2 on t1.c1=t2.c1",
			genSimpleJoinSelectTree(&tree.JoinOperator{
				JoinType: tree.JoinTypeRight,
				Natural:  true,
			}, "t1", "c1", "t2", "c1")},
		{"natural right outer join on", "select * from t1 natural right outer join t2 on t1.c1=t2.c1",
			genSimpleJoinSelectTree(&tree.JoinOperator{
				JoinType: tree.JoinTypeRight,
				Natural:  true,
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
		{"natural full join on", "select * from t1 natural full join t2 on t1.c1=t2.c1",
			genSimpleJoinSelectTree(&tree.JoinOperator{
				JoinType: tree.JoinTypeFull,
				Natural:  true,
			}, "t1", "c1", "t2", "c1")},
		{"natural full outer join on", "select * from t1 natural full outer join t2 on t1.c1=t2.c1",
			genSimpleJoinSelectTree(&tree.JoinOperator{
				JoinType: tree.JoinTypeFull,
				Outer:    true,
				Natural:  true,
			}, "t1", "c1", "t2", "c1")},
		{"inner join on", "select * from t1 inner join t2 on t1.c1=t2.c1",
			genSimpleJoinSelectTree(&tree.JoinOperator{
				JoinType: tree.JoinTypeInner,
			}, "t1", "c1", "t2", "c1")},
		{"natural inner join on", "select * from t1 natural inner join t2 on t1.c1=t2.c1",
			genSimpleJoinSelectTree(&tree.JoinOperator{
				JoinType: tree.JoinTypeInner,
				Natural:  true,
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
		{"update with table indexed", "update t1 indexed by i1 set c1=1",
			genSimpleUpdateTree(&tree.QualifiedTableName{
				TableName: "t1",
				IndexedBy: "i1",
			}, "c1", "1")},
		{"update with table not indexed", "update t1 not indexed set c1=1",
			genSimpleUpdateTree(&tree.QualifiedTableName{
				TableName:  "t1",
				NotIndexed: true,
			}, "c1", "1")},
		{"update or abort", "update or abort t1 set c1=1",
			genSimpleUpdateOrTree(&tree.QualifiedTableName{TableName: "t1"}, tree.UpdateOrAbort, "c1", "1")},
		{"update or fail", "update or fail t1 set c1=1",
			genSimpleUpdateOrTree(&tree.QualifiedTableName{TableName: "t1"}, tree.UpdateOrFail, "c1", "1")},
		{"update or ignore", "update or ignore t1 set c1=1",
			genSimpleUpdateOrTree(&tree.QualifiedTableName{TableName: "t1"}, tree.UpdateOrIgnore, "c1", "1")},
		{"update or replace", "update or replace t1 set c1=1",
			genSimpleUpdateOrTree(&tree.QualifiedTableName{TableName: "t1"}, tree.UpdateOrReplace, "c1", "1")},
		{"update or rollback", "update or rollback t1 set c1=1",
			genSimpleUpdateOrTree(&tree.QualifiedTableName{TableName: "t1"}, tree.UpdateOrRollback, "c1", "1")},
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
	}

	ctx := DatabaseContext{Actions: map[string]ActionContext{"action1": {}}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				tt.expect = nil
			}()

			eh := NewErrorHandler(1)
			el := newSqliteErrorListener(eh)
			ast, err := ParseRawSQLVisitor(tt.input, 1, "action1", ctx, el, *trace, false)
			if err != nil {
				t.Errorf("ParseRawSQL() got %s", err)
				return
			}

			astNodes := ast.(asts)
			node := astNodes[0]
			//fmt.Printf("AST: %+v\n", node.(*tree.Insert).InsertStmt)
			//fmt.Printf("exp: %+v\n", tt.expect.(*tree.Insert).InsertStmt)
			// use assert.Exactly?
			assert.EqualValues(t, tt.expect, node, "ParseRawSQL() got %s, want %s", node, tt.expect)

			sql, err := node.(tree.Ast).ToSQL()
			if err != nil {
				t.Errorf("ParseRawSQL() got %s", err)
			}
			fmt.Println("SQL from AST: ", sql)
			//if sql != tt.input {
			//	t.Errorf("ParseRawSQL() got %s, want %s", sql, tt.input)
			//}
		})
	}
}

func TestParseRawSQL_listener_allowed(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		// literal value
		{"number", "select 1"},
		{"string", "select 'a'"},
		{"null", "select null"},
		{"true", "select true"},
		{"false", "select false"},

		// common table stmt
		{"cte", "with t as (select 1) select * from t"},
		{"cte with column", "with t1(c1,c2) as (select 1) select * from t"},

		// compound operator
		{"union", "select 1 union select 2"},
		{"union all", "select 1 union all select 2"},
		{"intersect", "select 1 intersect select 2"},
		{"except", "select 1 except select 2"},

		// table or subquery
		{"table or subquery", "select * from t1 as tt"},
		{"table or subquery nest select", "select * from (select 1) as tt"},
		{"table or subquery nest tos", "select * from (t1 as tt, t2 as ttt)"},
		{"table or subquery join", "select * from t1 as tt join t2 as ttt on tt.a = ttt.a"},

		// expr
		{"expr bind parameter $", "select $a"},
		{"expr bind parameter @", "select @a"},
		{"expr names", "select table.column"},
		//
		{"expr unary op +", "select +1"},
		{"expr unary op -", "select -1"},
		{"expr unary op ~", "select ~1"},
		//
		{"expr unary op not", "select not 1"},
		{"expr binary op ||", "select 1 || 2"},
		{"expr binary op *", "select 1 * 2"},
		{"expr binary op /", "select 1 / 2"},
		{"expr binary op %", "select 1 % 2"},
		{"expr binary op +", "select 1 + 2"},
		{"expr binary op -", "select 1 - 2"},
		{"expr binary op <<", "select 1 << 2"},
		{"expr binary op >>", "select 1 >> 2"},
		{"expr binary op &", "select 1 & 2"},
		{"expr binary op |", "select 1 | 2"},
		{"expr binary op <", "select 1 < 2"},
		{"expr binary op <=", "select 1 <= 2"},
		{"expr binary op >", "select 1 > 2"},
		{"expr binary op >=", "select 1 >= 2"},
		{"expr binary op =", "select 1 = 2"},
		{"expr binary op !=", "select 1 != 2"},
		{"expr binary op <>", "select 1 <> 2"},
		{"expr binary op is", "select 1 is 2"},
		{"expr binary op is not", "select 1 is not 2"},
		{"expr binary op in", "select 1 in (1,2)"},
		{"expr binary op not in", "select 1 not in (1,2)"},
		{"expr binary op like", "select 1 like 2"},
		{"expr binary op match", "select 1 match 2"},
		{"expr binary op regexp", "select 1 regexp 2"},
		{"expr binary op and", "select 1 and 2"},
		{"expr binary op or", "select 1 or 2"},
		//
		{"expr function no param", "select f()"},
		{"expr function one param", "select f(1)"},
		{"expr function multi param", "select f(1,2)"},
		{"expr function * param", "select f(*)"},
		//
		{"expr in parentheses", "select (1)"},
		//
		{"expr with collate", "select 1 collate nocase"},
		//
		{"expr like", "select 1 like 2"},
		{"expr like escape", "select 1 like 2 escape 3"},
		{"expr not like", "select 1 not like 2"},
		{"expr match", "select 1 match 2"},
		{"expr not match", "select 1 not match 2"},
		{"expr regexp", "select 1 regexp 2"},
		{"expr not regexp", "select 1 not regexp 2"},
		// null
		{"expr isnull", "select 1 isnull"},
		{"expr notnull", "select 1 notnull"},
		{"expr not null", "select 1 not null"},
		//
		{"expr is", "select 1 is 2"},
		{"expr is not", "select 1 is not 2"},
		{"expr is distinct from", "select 1 is not distinct from 2"},
		//
		{"expr between", "select 1 between 2 and 3"},
		{"expr not between", "select 1 not between 2 and 3"},
		//
		{"expr in", "select 1 in (1,2)"},
		{"expr not in", "select 1 not in (1,2)"},
		{"expr in subquery", "select 1 in (select 1)"},
		//
		{"expr exists", "select exists (select 1)"},
		{"expr not exists", "select not exists (select 1)"},
		//
		{"expr case", "select case when 1 then 2 end"},
		{"expr case else", "select case when 1 then 2 else 3 end"},
		{"expr case multi when", "select case when 1 then 2 when 3 then 4 end"},
		{"expr case expr", "select case 1 when 2 then 3 end"},

		{"insert", "insert into t1 values ('1')"},
		{"insert", "insert into t1 values (1)"},
		{"insert replace", "replace into t1 values (1)"},
		{"insert or replace", "insert or replace into t1 values (1)"},
		{"insert with columns", "insert into t1 (a,b) values (1,2)"},
		{"insert with cte", "with t as (select 1) insert into t1 (a,b) values (1,2)"},
		{"insert with returning", "insert into t1 (a,b) values (1,2) returning a"},
		{"insert with returning *", "insert into t1 (a,b) values (1,2) returning *"},
		{"insert with or replace", "insert or replace into t1 (a,b) values (1,2)"},
		{"insert with table alias", "insert into t1 as t (a,b) values (1,2)"},

		{"insert with values upsert without target do nothing", "insert into t1 (a,b) values (1,2) on conflict do nothing"},
		{"insert with values upsert with target without where do nothing",
			"insert into t1 (a,b) values (1,2) on conflict (c1,c2) do nothing"},
		{"insert with values upsert with target and where do nothing",
			"insert into t1 (a,b) values (1,2) on conflict(c1,c2) where 1 do nothing"},
		{"insert with values upsert with update column name",
			"insert into t1 (a,b) values (1,2) on conflict do update set b=1"},
		{"insert with values upsert with update multi column name",
			"insert into t1 (a,b) values (1,2) on conflict do update set b=1,c=2"},
		{"insert with values upsert with update column name list",
			"insert into t1 (a,b) values (1,2) on conflict do update set (b,c)=(1,2)"},
		{"insert with values upsert with update multi column name list",
			"insert into t1 (a,b) values (1,2) on conflict do update set (b,c)=(1,2), (d,e)=(3,4)"},
		{"insert with values upsert with update and where",
			"insert into t1 (a,b) values (1,2) on conflict do update set b=1 where 1"},

		// join
		{"join on", "select * from t1 join t2 on t1.c1=t2.c1"},
		//{"join implicit", "select * from t1,t2 on t1.c1=t2.c1"},
		{"left join", "select * from t1 left join t2 on t1.c1=t2.c1"},
		{"left outer join", "select * from t1 left outer join t2 on t1.c1=t2.c1"},
		{"right join", "select * from t1 right join t2 on t1.c1=t2.c1"},
		{"right outer join", "select * from t1 right outer join t2 on t1.c1=t2.c1"},
		{"full join", "select * from t1 full join t2 on t1.c1=t2.c1"},
		{"full outer join", "select * from t1 full outer join t2 on t1.c1=t2.c1"},
		{"inner join", "select * from t1 inner join t2 on t1.c1=t2.c1"},

		// select
		{"select *", "select * from t1"},
		{"select with cte", "with t as (select 1) select * from t1"},
		{"select distinct", "select distinct * from t1"},
		{"select from join clause", "select * from t1 join t2 on t1.c1=t2.c1"},
		{"select with where", "select * from t1 where 1"},
		{"select with group by", "select * from t1 group by c1"},
		{"select with group by and having", "select * from t1 group by c1 having 1"},
		{"select with compound operator union", "select * from t1 union select * from t2"},
		{"select with compound operator union all", "select * from t1 union all select * from t2"},
		{"select with compound operator intersect", "select * from t1 intersect select * from t2"},
		{"select with compound operator except", "select * from t1 except select * from t2"},
		{"select with order by", "select * from t1 order by c1 collate collate_name asc"},
		{"select with limit", "select * from t1 limit 1"},
		{"select with limit offset", "select * from t1 limit 1 offset 2"},
		{"select with limit comma", "select * from t1 limit 1,10"},
	}

	ctx := DatabaseContext{Actions: map[string]ActionContext{"action1": {}}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eh := NewErrorHandler(1)
			el := newSqliteErrorListener(eh)
			err := ParseRawSQL(tt.input, 1, "action1", ctx, el, *trace, false)
			if err != nil {
				t.Errorf("ParseRawSQL() got %s", err)
			}
		})
	}
}

func TestParseRawSQL_syntax_not_allowed(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		causeSymbol string
	}{
		// literal value
		{"current_date", "select current_date", "current_date"},
		{"current_time", "select current_time", "current_time"},
		{"current_timestamp", "select current_timestamp", "current_timestamp"},
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
		{"expr function distinct param", "select f(distinct 1,2)", "("},
		{"expr function with filter", "select f(1) filter (where 1)", "filter"},

		// join
		{"cross join", "select * from t3 cross join t4", "cross"},
		{"join using", "select * from t3 join t4 using (c1)", "using"},
		{"join without condition", "select * from t3 join t4", "<EOF>"},
		{"comma cartesian join 1", "select * from t3, t4", ","},
	}

	ctx := DatabaseContext{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eh := NewErrorHandler(1)
			el := newSqliteErrorListener(eh)
			err := ParseRawSQL(tt.input, 1, "action1", ctx, el, *trace, false)

			if err == nil || !strings.Contains(err.Error(), ErrSyntax.Error()) {
				t.Errorf("ParseRawSQL() expected error: %s, got %s", ErrSyntax, err)
			}

			if el.symbol != tt.causeSymbol {
				t.Errorf("ParseRawSQL() expected cause symbol: %s, got: %s", tt.causeSymbol, el.symbol)
			}
		})
	}
}

func TestParseRawSQL_banRules(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError error
	}{
		// non-deterministic time functions
		{"select date function", "select date('2020-02-02')", nil},
		{"select date function 1", "select date('now')", ErrFunctionNotSupported},
		{"select date function 2", "select date('now', '+1 day')", ErrFunctionNotSupported},
		{"select time function", "select time('now')", ErrFunctionNotSupported},
		{"select datetime function", "select datetime('now')", ErrFunctionNotSupported},
		{"select julianday function", "select julianday('now')", ErrFunctionNotSupported},
		{"select unixepoch function", "select unixepoch('now')", ErrFunctionNotSupported},
		{"select strftime function", "select strftime('%Y%m%d', 'now')", ErrFunctionNotSupported},
		{"select strftime function 2", "select strftime('%Y%m%d')", ErrFunctionNotSupported},
		//
		{"upsert date function", "insert into t1 values (date('now'))", ErrFunctionNotSupported},
		{"upsert date function 2", "insert into t1 values ( date('now', '+1 day'))", ErrFunctionNotSupported},
		{"upsert time function", "insert into t1 values ( time('now'))", ErrFunctionNotSupported},
		{"upsert datetime function", "insert into t1 values ( datetime('now'))", ErrFunctionNotSupported},
		{"upsert julianday function", "insert into t1 values ( julianday('now'))", ErrFunctionNotSupported},
		{"upsert unixepoch function", "insert into t1 values ( unixepoch('now'))", ErrFunctionNotSupported},
		{"upsert strftime function", "insert into t1 values (strftime('%Y%m%d', 'now'))", ErrFunctionNotSupported},
		{"upsert strftime function 2", "insert into t1 values (strftime('%Y%m%d'))", ErrFunctionNotSupported},
		// non-deterministic random functions
		{"random function", "select random()", ErrFunctionNotSupported},
		{"randomblob function", "select randomblob(10)", ErrFunctionNotSupported},
		{"random function", "insert into t2 values ( random())", ErrFunctionNotSupported},
		{"randomblob function", "insert into t2 values ( randomblob(10))", ErrFunctionNotSupported},
		// non-deterministic math functions
		{"select acos function", "select acos(1)", ErrFunctionNotSupported},
		{"select acosh function", "select acosh(1)", ErrFunctionNotSupported},
		{"select asin function", "select asin(1)", ErrFunctionNotSupported},
		{"select asinh function", "select asinh(1)", ErrFunctionNotSupported},
		{"select atan function", "select atan(1)", ErrFunctionNotSupported},
		{"select atan2 function", "select atan2(1, 1)", ErrFunctionNotSupported},
		{"select atanh function", "select atanh(1)", ErrFunctionNotSupported},
		{"select ceil function", "select ceil(1)", ErrFunctionNotSupported},
		{"select ceiling function", "select ceiling(1)", ErrFunctionNotSupported},
		{"select cos function", "select cos(1)", ErrFunctionNotSupported},
		{"select cosh function", "select cosh(1)", ErrFunctionNotSupported},
		{"select degrees function", "select degrees(1)", ErrFunctionNotSupported},
		{"select exp function", "select exp(1)", ErrFunctionNotSupported},
		{"select ln function", "select ln(1)", ErrFunctionNotSupported},
		{"select log function", "select log(1)", ErrFunctionNotSupported},
		{"select log function 2", "select log(1, 1)", ErrFunctionNotSupported},
		{"select log10 function", "select log10(1)", ErrFunctionNotSupported},
		{"select log2 function", "select log2(1)", ErrFunctionNotSupported},
		{"select mod function", "select mod(1, 1)", ErrFunctionNotSupported},
		{"select pi function", "select pi()", ErrFunctionNotSupported},
		{"select pow function", "select pow(1, 1)", ErrFunctionNotSupported},
		{"select power function", "select power(1, 1)", ErrFunctionNotSupported},
		{"select radians function", "select radians(1)", ErrFunctionNotSupported},
		{"select sin function", "select sin(1)", ErrFunctionNotSupported},
		{"select sinh function", "select sinh(1)", ErrFunctionNotSupported},
		{"select sqrt function", "select sqrt(1)", ErrFunctionNotSupported},
		{"select tan function", "select tan(1)", ErrFunctionNotSupported},
		{"select tanh function", "select tanh(1)", ErrFunctionNotSupported},
		{"select trunc function", "select trunc(1)", ErrFunctionNotSupported},
		// non-exist table/column
		{"select from table", "select * from t1", nil},
		{"select with CTE", "with tt as (select * from t1) select * from tt", nil},
		{"select non-exist table", "select * from t10", ErrTableNotFound},
		// joins
		{"joins 1", "select * from t1 join t2 on (1+2)=2", ErrJoinConditionTooDeep},
		{"joins 2", "select * from t1 join t2 on t1.c1=t7.c1 ", ErrTableNotFound},
		{"joins 3", "select * from t1 join t2 on t1.c1=t2.c7", ErrColumnNotFound},

		{"multi joins", `select * from t1 join t2 on t1.c1=t2.c1 join t3 on t3.c2=t1.c1 join t4 on t4.c1=t1.c1`, nil},
		{"multi joins too many", "select * from t1 join t2 on t1.c1=t2.c1 join t3 on t3.c2=t1.c1 join t4 on t4.c1=t1.c1 join t5 on t5.c2=t1.c1", ErrMultiJoinNotSupported},
		// join with condition
		{"join with non = cons", "select * from t3 join t4 on a and b", ErrJoinConditionOpNotSupported},
		{"join with non = cons 2", "select * from t3 join t4 on a + b", ErrJoinConditionOpNotSupported},
		//{"join with multi level binary cons", "select * from t3 join t4 on a=(b=c)", nil}, // TODO
		//{"join with function cons", "select * from t3 join t4 on random()", ErrJoinWithTrueCondition}, /// TODO: support this
		// action parameters
		{"insert with bind parameter", "insert into t3 values ($this)", nil},
		{"insert with non exist bond parameter", "insert into t3 values ($a)", ErrBindParameterNotFound},
		// modifiers
		{"modifier", "select * from t3 where a = @caller", nil},
		{"modifier 2", "select * from t3 where a = @block_height", nil},
		{"modifier 3", "select * from t3 where a = @any", ErrModifierNotSupported},
	}

	ctx := DatabaseContext{
		Tables: map[string]TableContext{
			"t1": {
				Columns:     []string{"c1", "c2", "c3"},
				PrimaryKeys: []string{"c1"},
			},
			"t2": {
				Columns:     []string{"c1", "c2", "c3"},
				PrimaryKeys: []string{"c1"},
			},
			"t3": {
				Columns:     []string{"c1", "c2", "c3"},
				PrimaryKeys: []string{"c1"},
			},
			"t4": {
				Columns:     []string{"c1", "c2", "c3"},
				PrimaryKeys: []string{"c1"},
			},
			"t5": {
				Columns:     []string{"c1", "c2", "c3"},
				PrimaryKeys: []string{"c1"},
			},
		},
		Actions: map[string]ActionContext{
			"action1": {
				"$this": nil,
				"$that": nil,
			},
			"action2": {
				"$here":  nil,
				"$there": nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eh := NewErrorHandler(1)
			el := newSqliteErrorListener(eh)
			err := ParseRawSQL(tt.input, 1, "action1", ctx, el, *trace, true)

			if err == nil && tt.wantError == nil {
				return
			}

			if err != nil && tt.wantError != nil {
				// TODO: errors.Is?
				if strings.Contains(err.Error(), tt.wantError.Error()) {
					return
				}
				t.Errorf("ParseRawSQL() expected error: %s, got %s", tt.wantError, err)
				return
			}

			t.Errorf("ParseRawSQL() expected: %s, got %s", tt.wantError, err)
		})
	}
}

func TestParseRawSQL_semantic_invalid(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		// expr datetime functions
		{"expr function date now", `select date('now')`},
		{"expr function date 1 modifier", `select date('now','start of month')`},
		{"expr function date 2 modifiers", `select date('now','start of month','+1 month')`},
		{"expr function time now", `select time('now')`},
		{"expr function time 1 modifier", `select time('now','start of month')`},
		{"expr function time 2 modifiers", `select time('now','start of month','+1 month')`},
		{"expr function datetime now", `select datetime('now')`},
		{"expr function datetime 1 modifier", `select datetime('now','start of month')`},
		{"expr function datetime 2 modifiers", `select datetime('now','start of month','+1 month')`},
		{"expr function strftime now", `select strftime('%d','now')`},
		{"expr function strftime 1 modifer", `select strftime('%d','now','start of month')`},
		{"expr function strftime 2 modifiers", `select strftime('%d','now','start of month','+1 month')`},
		{"expr function unixepoch now", `select unixepoch('now')`},
		{"expr function unixepoch 1 modifier", `select unixepoch('now','start of month')`},
		{"expr function unixepoch 2 modifiers", `select unixepoch('now','start of month','+1 month')`},
	}

	ctx := DatabaseContext{Actions: map[string]ActionContext{"action1": {}}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eh := NewErrorHandler(1)
			el := newSqliteErrorListener(eh)
			ast, err := ParseRawSQLVisitor(tt.input, 1, "action1", ctx, el, *trace, false)
			if err != nil {
				t.Errorf("ParseRawSQL() got %s", err)
				return
			}

			astNodes := ast.(asts)
			node := astNodes[0]

			_, err = node.(tree.Ast).ToSQL()
			assert.Error(t, err, "ToSQL() should return error")
		})
	}
}
