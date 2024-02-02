package tree_test

import (
	"fmt"
	"testing"

	"github.com/kwilteam/kwil-db/parse/sql/tree"
)

var (
	// TODO: use real types
	typeString    = "string"
	typeStringSql = " ::string"
	typeInt       = "int"
	typeIntSql    = " ::int"
)

func TestExpressionLiteral_ToSQL(t *testing.T) {
	type fields tree.Expression
	tests := []struct {
		name      string
		fields    fields
		want      string
		wantPanic bool
	}{
		{
			name: "expression literal",
			fields: &tree.ExpressionLiteral{
				Value: "'foo'",
			},
			want: "'foo'",
		},
		{
			name: "expression literal with type hint",
			fields: &tree.ExpressionLiteral{
				Value:    "'foo'",
				TypeHint: typeString,
			},
			want: "'foo'" + typeStringSql, // TODO: replace with just string,
			// not variable. this should be the result string, no interpolation
		},
		{
			name: "expression literal with wrapped paren",
			fields: &tree.ExpressionLiteral{
				Value:   "'foo'",
				Wrapped: true,
			},
			want: "( 'foo' )",
		},
		{
			name: "expression literal with wrapped paren and type hint",
			fields: &tree.ExpressionLiteral{
				Value:    "'foo'",
				Wrapped:  true,
				TypeHint: typeString,
			},
			want: "( 'foo' )" + typeStringSql,
		},
		{
			name: "expression literal with int",
			fields: &tree.ExpressionLiteral{
				Value: "1",
			},
			want: "1",
		},
		{
			name: "expression literal witth int and type hint",
			fields: &tree.ExpressionLiteral{
				Value:    "1",
				TypeHint: typeInt,
			},
			want: "1" + typeIntSql,
		},
		{
			name: "expression literal with float",
			fields: &tree.ExpressionLiteral{
				Value: "1.1",
			},
			wantPanic: true,
		},
		{
			name: "expression $ bind parameter",
			fields: &tree.ExpressionBindParameter{
				Parameter: "$foo",
			},
			want: "$foo",
		},
		{
			name: "expression $ bind parameter with type hint",
			fields: &tree.ExpressionBindParameter{
				Parameter: "$foo",
				TypeHint:  typeString,
			},
			want: "$foo" + typeStringSql,
		},
		{
			name: "expression @ bind parameter",
			fields: &tree.ExpressionBindParameter{
				Parameter: "@foo",
			},
			want: "@foo",
		},
		{
			name: "expression @ bind parameter with type hint",
			fields: &tree.ExpressionBindParameter{
				Parameter: "@foo",
				TypeHint:  typeString,
			},
			want: "@foo" + typeStringSql,
		},
		{
			name: "expression parameter without $ or @",
			fields: &tree.ExpressionBindParameter{
				Parameter: "foo",
			},
			wantPanic: true,
		},
		{
			name: "expression parameter with empty string",
			fields: &tree.ExpressionBindParameter{
				Parameter: "",
			},
			wantPanic: true,
		},
		{
			name: "expression column",
			fields: &tree.ExpressionColumn{
				Column: "foo",
			},
			want: `"foo"`,
		},
		{
			name: "expression column with type hint",
			fields: &tree.ExpressionColumn{
				Column:   "foo",
				TypeHint: typeString,
			},
			want: `"foo"` + typeStringSql,
		},
		{
			name: "expression column with table",
			fields: &tree.ExpressionColumn{
				Table:  "bar",
				Column: "foo",
			},
			want: `"bar"."foo"`,
		},
		{
			name: "expression column with table and type hint",
			fields: &tree.ExpressionColumn{
				Table:    "bar",
				Column:   "foo",
				TypeHint: typeString,
			},
			want: `"bar"."foo"` + typeStringSql,
		},
		{
			name: "expression column with only table",
			fields: &tree.ExpressionColumn{
				Table: "bar",
			},
			wantPanic: true,
		},
		{
			name: "expression unary operator",
			fields: &tree.ExpressionUnary{
				Operator: tree.UnaryOperatorNot,
				Operand: &tree.ExpressionColumn{
					Column: "foo",
				},
			},
			want: `NOT "foo"`,
		},
		{
			name: "expression unary operator with type hint",
			fields: &tree.ExpressionUnary{
				Operator: tree.UnaryOperatorNot,
				Operand: &tree.ExpressionColumn{
					Column:   "foo",
					TypeHint: typeInt,
				},
			},
			want: `NOT "foo"` + typeIntSql,
		},
		{
			name: "expression binary comparison",
			fields: &tree.ExpressionBinaryComparison{
				Left: &tree.ExpressionColumn{
					Column: "foo",
				},
				Operator: tree.ComparisonOperatorEqual,
				Right: &tree.ExpressionLiteral{
					Value: "'bar'",
				},
			},
			want: `"foo" = 'bar'`,
		},
		{
			name: "expression binary comparison with type hint",
			fields: &tree.ExpressionBinaryComparison{
				Left: &tree.ExpressionColumn{
					Column: "foo",
				},
				Operator: tree.ComparisonOperatorEqual,
				Right: &tree.ExpressionLiteral{
					Value: "'bar'",
				},
				TypeHint: typeInt,
				Wrapped:  true,
			},
			want: `("foo" = 'bar')` + typeIntSql,
		},
		{
			name: "expression binary comparison with type hint without wrapped",
			fields: &tree.ExpressionBinaryComparison{
				Left: &tree.ExpressionColumn{
					Column: "foo",
				},
				Operator: tree.ComparisonOperatorEqual,
				Right: &tree.ExpressionLiteral{
					Value: "'bar'",
				},
				TypeHint: typeInt,
			},
			wantPanic: true,
		},
		{
			name: "expression abs function",
			fields: &tree.ExpressionFunction{
				Function: &tree.FunctionABS,
				Inputs: []tree.Expression{
					&tree.ExpressionColumn{
						Column: "foo",
					},
				},
			},
			want: "abs(\"foo\")",
		},
		{
			name: "expression abs function with type hint",
			fields: &tree.ExpressionFunction{
				Function: &tree.FunctionABS,
				Inputs: []tree.Expression{
					&tree.ExpressionColumn{
						Column: "foo",
					},
				},
				TypeHint: typeInt,
			},
			want: "abs(\"foo\")" + typeIntSql,
		},
		{
			name: "expression abs function with multiple inputs",
			fields: &tree.ExpressionFunction{
				Function: &tree.FunctionABS,
				Inputs: []tree.Expression{
					&tree.ExpressionColumn{
						Column: "foo",
					},
					&tree.ExpressionColumn{
						Column: "bar",
					},
				},
			},
			wantPanic: true,
		},
		{
			name: "expression date function with no inputs (uses current date)",
			fields: &tree.ExpressionFunction{
				Function: &tree.FunctionDATE,
			},
			wantPanic: true,
		},
		{
			name: "expression list",
			fields: &tree.ExpressionList{
				Expressions: []tree.Expression{
					&tree.ExpressionColumn{
						Column: "foo",
					},
					&tree.ExpressionColumn{
						Column: "bar",
					},
				},
			},
			want: "(\"foo\", \"bar\")",
		},
		{
			name: "expression list with element type hint",
			fields: &tree.ExpressionList{
				Expressions: []tree.Expression{
					&tree.ExpressionColumn{
						Column:   "foo",
						TypeHint: typeString,
					},
					&tree.ExpressionColumn{
						Column:   "bar",
						TypeHint: typeInt,
					},
				},
			},
			want: fmt.Sprintf(`("foo"%s, "bar"%s)`, typeStringSql, typeIntSql),
		},
		{
			name: "expression list with type hint",
			// NOTE Seems no point in having a type hint for a list?? as we can't have a list types?
			fields: &tree.ExpressionList{
				Expressions: []tree.Expression{
					&tree.ExpressionColumn{
						Column: "foo",
					},
					&tree.ExpressionColumn{
						Column: "bar",
					},
				},
				TypeHint: typeString,
			},
			want: "(\"foo\", \"bar\")" + typeStringSql,
		},
		{
			name: "collate",
			fields: &tree.ExpressionCollate{
				Expression: &tree.ExpressionBinaryComparison{
					Left: &tree.ExpressionColumn{
						Column: "foo",
					},
					Operator: tree.ComparisonOperatorEqual,
					Right: &tree.ExpressionLiteral{
						Value: "'bar'",
					},
				},
				Collation: tree.CollationTypeBinary,
			},
			want: `"foo" = 'bar' COLLATE BINARY`,
		},
		{
			name: "collate with type hint",
			fields: &tree.ExpressionCollate{
				Expression: &tree.ExpressionBinaryComparison{
					Left: &tree.ExpressionColumn{
						Column: "foo",
					},
					Operator: tree.ComparisonOperatorEqual,
					Right: &tree.ExpressionLiteral{
						Value: "'bar'",
					},
				},
				Collation: tree.CollationTypeBinary,
				Wrapped:   true,
				TypeHint:  typeInt,
			},
			want: `("foo" = 'bar' COLLATE BINARY)` + typeIntSql,
		},
		{
			name: "collate with type hint without wrapped",
			fields: &tree.ExpressionCollate{
				Expression: &tree.ExpressionBinaryComparison{},
				Collation:  tree.CollationTypeBinary,
				TypeHint:   typeString,
			},
			wantPanic: true,
		},
		{
			name: "collate with no expression",
			fields: &tree.ExpressionCollate{
				Collation: tree.CollationTypeBinary,
			},
			wantPanic: true,
		},
		{
			name: "collate with no collation",
			fields: &tree.ExpressionCollate{
				Expression: &tree.ExpressionBinaryComparison{
					Left: &tree.ExpressionColumn{
						Column: "foo",
					},
					Operator: tree.ComparisonOperatorEqual,
					Right: &tree.ExpressionLiteral{
						Value: "'bar'",
					},
				},
			},
			wantPanic: true,
		},
		{
			name: "string compare with escape",
			fields: &tree.ExpressionStringCompare{
				Left: &tree.ExpressionColumn{
					Column: "foo",
				},
				Operator: tree.StringOperatorNotLike,
				Right: &tree.ExpressionLiteral{
					Value: "'bar'",
				},
				Escape: &tree.ExpressionLiteral{
					Value: "'baz'",
				},
			},
			want: `"foo" NOT LIKE 'bar' ESCAPE 'baz'`,
		},
		{
			name: "string compare with escape and type hint",
			fields: &tree.ExpressionStringCompare{
				Left: &tree.ExpressionColumn{
					Column: "foo",
				},
				Operator: tree.StringOperatorNotLike,
				Right: &tree.ExpressionLiteral{
					Value: "'bar'",
				},
				Escape: &tree.ExpressionLiteral{
					Value: "'baz'",
				},
				TypeHint: typeInt,
				Wrapped:  true,
			},
			want: `("foo" NOT LIKE 'bar' ESCAPE 'baz')` + typeIntSql,
		},
		{
			name: "string compare with escape and type hint without wrapped",
			fields: &tree.ExpressionStringCompare{
				TypeHint: typeInt,
			},
			wantPanic: true,
		},
		{
			name: "string compare with escape and GLOB operator",
			fields: &tree.ExpressionStringCompare{
				Left: &tree.ExpressionColumn{
					Column: "foo",
				},
				Operator: tree.StringOperatorGlob,
				Right: &tree.ExpressionLiteral{
					Value: "'bar'",
				},
				Escape: &tree.ExpressionLiteral{
					Value: "'baz'",
				},
			},
			wantPanic: true,
		},
		{
			name: "string compare without escape and NOT GLOB operator",
			fields: &tree.ExpressionStringCompare{
				Left: &tree.ExpressionColumn{
					Column: "foo",
				},
				Operator: tree.StringOperatorNotGlob,
				Right: &tree.ExpressionLiteral{
					Value: "'bar'",
				},
			},
			want: `"foo" NOT GLOB 'bar'`,
		},
		{
			name: "IsNull",
			fields: &tree.ExpressionIsNull{
				Expression: &tree.ExpressionColumn{
					Column: "foo",
				},
				IsNull: true,
			},
			want: `"foo" IS NULL`,
		},
		{
			name: "IsNull with type hint",
			fields: &tree.ExpressionIsNull{
				Expression: &tree.ExpressionColumn{
					Column: "foo",
				},
				IsNull:   true,
				TypeHint: typeInt,
				Wrapped:  true,
			},
			want: `("foo" IS NULL)` + typeIntSql,
		},
		{
			name: "IsNull with type hint without wrapped",
			fields: &tree.ExpressionIsNull{
				Expression: &tree.ExpressionColumn{
					Column: "foo",
				},
				IsNull:   true,
				TypeHint: typeInt,
			},
			wantPanic: true,
		},
		{
			name: "IsNull with no expression",
			fields: &tree.ExpressionIsNull{
				IsNull: true,
			},
			wantPanic: true,
		},
		{
			name: "Is Not Null",
			fields: &tree.ExpressionIsNull{
				Expression: &tree.ExpressionColumn{
					Column: "foo",
				},
			},
			want: `"foo" NOT NULL`,
		},
		{
			name: "Is Not Null with type hint",
			fields: &tree.ExpressionIsNull{
				Expression: &tree.ExpressionColumn{
					Column: "foo",
				},
				TypeHint: typeInt,
				Wrapped:  true,
			},
			want: `("foo" NOT NULL)` + typeIntSql,
		},
		{
			name: "is not distinct from",
			fields: &tree.ExpressionDistinct{
				Left: &tree.ExpressionColumn{
					Column: "foo",
				},
				Right: &tree.ExpressionLiteral{
					Value: "'bar'",
				},
				IsNot: true,
			},
			want: `"foo" IS NOT DISTINCT FROM 'bar'`,
		},
		{
			name: "is not distinct from with type hint",
			fields: &tree.ExpressionDistinct{
				Left: &tree.ExpressionColumn{
					Column: "foo",
				},
				Right: &tree.ExpressionLiteral{
					Value: "'bar'",
				},
				IsNot:    true,
				TypeHint: typeInt,
				Wrapped:  true,
			},
			want: `("foo" IS NOT DISTINCT FROM 'bar')` + typeIntSql,
		},
		{
			name: "is not distinct from with type hint without wrapped",
			fields: &tree.ExpressionDistinct{
				Left: &tree.ExpressionColumn{
					Column: "foo",
				},
				Right: &tree.ExpressionLiteral{
					Value: "'bar'",
				},
				IsNot:    true,
				TypeHint: typeInt,
			},
			wantPanic: true,
		},
		{
			name: "expr is expr",
			fields: &tree.ExpressionBinaryComparison{
				Left: &tree.ExpressionColumn{
					Column: "foo",
				},
				Right: &tree.ExpressionLiteral{
					Value: "'bar'",
				},
				Operator: tree.ComparisonOperatorIs,
			},
			want: `"foo" IS 'bar'`,
		},
		{
			name: "distinct with no left",
			fields: &tree.ExpressionDistinct{
				Right: &tree.ExpressionLiteral{
					Value: "'bar'",
				},
			},
			wantPanic: true,
		},
		{
			name: "distinct with no right",
			fields: &tree.ExpressionDistinct{
				Left: &tree.ExpressionColumn{
					Column: "foo",
				},
			},
			wantPanic: true,
		},
		{
			name: "valid between",
			fields: &tree.ExpressionBetween{
				Expression: &tree.ExpressionColumn{
					Column: "foo",
				},
				NotBetween: true,
				Left: &tree.ExpressionLiteral{
					Value: "'bar'",
				},
				Right: &tree.ExpressionLiteral{
					Value: "'baz'",
				},
			},
			want: `"foo" NOT BETWEEN 'bar' AND 'baz'`,
		},
		{
			name: "valid between with type hint",
			fields: &tree.ExpressionBetween{
				Expression: &tree.ExpressionColumn{
					Column: "foo",
				},
				NotBetween: true,
				Left: &tree.ExpressionLiteral{
					Value: "'bar'",
				},
				Right: &tree.ExpressionLiteral{
					Value: "'baz'",
				},
				TypeHint: typeInt,
				Wrapped:  true,
			},
			want: `("foo" NOT BETWEEN 'bar' AND 'baz')` + typeIntSql,
		},
		{
			name: "valid between with type hint without wrapped",
			fields: &tree.ExpressionBetween{
				Expression: &tree.ExpressionColumn{
					Column: "foo",
				},
				NotBetween: true,
				Left: &tree.ExpressionLiteral{
					Value: "'bar'",
				},
				Right: &tree.ExpressionLiteral{
					Value: "'baz'",
				},
				TypeHint: typeInt,
			},
			wantPanic: true,
		},
		{
			name: "between with no expression",
			fields: &tree.ExpressionBetween{
				Left: &tree.ExpressionLiteral{
					Value: "'bar'",
				},
				Right: &tree.ExpressionLiteral{
					Value: "'baz'",
				},
			},
			wantPanic: true,
		},
		{
			name: "between with no left",
			fields: &tree.ExpressionBetween{
				Expression: &tree.ExpressionColumn{
					Column: "foo",
				},
				Right: &tree.ExpressionLiteral{
					Value: "'baz'",
				},
			},
			wantPanic: true,
		},
		{
			name: "between with no right",
			fields: &tree.ExpressionBetween{
				Expression: &tree.ExpressionColumn{
					Column: "foo",
				},
				Left: &tree.ExpressionLiteral{
					Value: "'bar'",
				},
			},
			wantPanic: true,
		},
		{
			name: "select subquery",
			fields: &tree.ExpressionSelect{
				IsNot:    true,
				IsExists: true,
				Select: &tree.SelectStmt{
					SelectCores: []*tree.SelectCore{
						{
							SelectType: tree.SelectTypeAll,
							From: &tree.FromClause{
								JoinClause: &tree.JoinClause{
									TableOrSubquery: &tree.TableOrSubqueryTable{
										Name:  "foo",
										Alias: "f",
									},
								},
							},
							Where: &tree.ExpressionBinaryComparison{
								Left: &tree.ExpressionColumn{
									Table:  "f",
									Column: "foo",
								},
								Operator: tree.ComparisonOperatorEqual,
								Right: &tree.ExpressionBindParameter{
									Parameter: "$a",
								},
							},
						},
					},
				},
			},
			want: `NOT EXISTS (SELECT * FROM "foo" AS "f" WHERE "f"."foo" = $a)`,
		},
		{
			name: "select subquery with type hint",
			fields: &tree.ExpressionSelect{
				IsNot:    true,
				IsExists: true,
				Select: &tree.SelectStmt{
					SelectCores: []*tree.SelectCore{
						{
							SelectType: tree.SelectTypeAll,
							From: &tree.FromClause{
								JoinClause: &tree.JoinClause{
									TableOrSubquery: &tree.TableOrSubqueryTable{
										Name:  "foo",
										Alias: "f",
									},
								},
							},
							Where: &tree.ExpressionBinaryComparison{
								Left: &tree.ExpressionColumn{
									Table:  "f",
									Column: "foo",
								},
								Operator: tree.ComparisonOperatorEqual,
								Right: &tree.ExpressionBindParameter{
									Parameter: "$a",
								},
							},
						},
					},
				},
				TypeHint: typeInt,
				Wrapped:  true,
			},
			want: `(NOT EXISTS (SELECT * FROM "foo" AS "f" WHERE "f"."foo" = $a))` + typeIntSql,
		},
		{
			name: "case expression",
			fields: &tree.ExpressionCase{
				CaseExpression: &tree.ExpressionColumn{
					Column: "foo",
				},
				WhenThenPairs: [][2]tree.Expression{
					{
						&tree.ExpressionLiteral{
							Value: "'bar'",
						},
						&tree.ExpressionLiteral{
							Value: "'baz'",
						},
					},
				},
				ElseExpression: &tree.ExpressionLiteral{
					Value: "'qux'",
				},
			},
			want: `CASE "foo" WHEN 'bar' THEN 'baz' ELSE 'qux' END`,
		},
		{
			name: "case expression with no case expression",
			fields: &tree.ExpressionCase{
				WhenThenPairs: [][2]tree.Expression{
					{
						&tree.ExpressionLiteral{
							Value: "'bar'",
						},
						&tree.ExpressionLiteral{
							Value: "'baz'",
						},
					},
				},
				ElseExpression: &tree.ExpressionLiteral{
					Value: "'qux'",
				},
			},
			want: `CASE WHEN 'bar' THEN 'baz' ELSE 'qux' END`,
		},
		{
			name: "case expression with no when then pairs",
			fields: &tree.ExpressionCase{
				CaseExpression: &tree.ExpressionColumn{
					Column: "foo",
				},
				ElseExpression: &tree.ExpressionLiteral{
					Value: "'qux'",
				},
			},
			wantPanic: true,
		},
		{
			name: "case expression with no else expression",
			fields: &tree.ExpressionCase{
				CaseExpression: &tree.ExpressionColumn{
					Column: "foo",
				},
				WhenThenPairs: [][2]tree.Expression{
					{
						&tree.ExpressionLiteral{
							Value: "'bar'",
						},
						&tree.ExpressionLiteral{
							Value: "'baz'",
						},
					},
				},
			},
			want: `CASE "foo" WHEN 'bar' THEN 'baz' END`,
		},
		{
			name: "arithmetic expression",
			fields: &tree.ExpressionArithmetic{
				Left: &tree.ExpressionColumn{
					Column: "foo",
				},
				Operator: tree.ArithmeticOperatorAdd,
				Right: &tree.ExpressionLiteral{
					Value: "1",
				},
			},
			want: `"foo" + 1`,
		},
		{
			name: "arithmetic expression with type hint",
			fields: &tree.ExpressionArithmetic{
				Left: &tree.ExpressionColumn{
					Column: "foo",
				},
				Operator: tree.ArithmeticOperatorAdd,
				Right: &tree.ExpressionLiteral{
					Value: "1",
				},
				TypeHint: typeInt,
				Wrapped:  true,
			},
			want: `("foo" + 1)` + typeIntSql,
		},
		{
			name: "arithmetic expression with type hint without wrapped",
			fields: &tree.ExpressionArithmetic{
				Left: &tree.ExpressionColumn{
					Column: "foo",
				},
				Operator: tree.ArithmeticOperatorAdd,
				Right: &tree.ExpressionLiteral{
					Value: "1",
				},
				TypeHint: typeInt,
			},
			wantPanic: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("Expression.ToSQL() should have panicked")
					}
				}()
			}

			got := tt.fields.ToSQL()
			if tt.wantPanic {
				return
			}

			if !compareIgnoringWhitespace(got, tt.want) {
				t.Errorf("Expression.ToSQL() = %v, want %v", got, tt.want)
			}
		})
	}
}
