package tree_test

import (
	"testing"

	"github.com/kwilteam/kwil-db/parse/sql/tree"
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
			name: "expression literal with type cast",
			fields: &tree.ExpressionLiteral{
				Value:    "'foo'",
				TypeCast: tree.TypeCastText,
			},
			want: "'foo' ::text",
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
			name: "expression literal with wrapped paren and type cast",
			fields: &tree.ExpressionLiteral{
				Value:    "'foo'",
				Wrapped:  true,
				TypeCast: tree.TypeCastText,
			},
			want: "( 'foo' ) ::text",
		},
		{
			name: "expression literal with int",
			fields: &tree.ExpressionLiteral{
				Value: "1",
			},
			want: "1",
		},
		{
			name: "expression literal with int and type cast",
			fields: &tree.ExpressionLiteral{
				Value:    "1",
				TypeCast: tree.TypeCastInt,
			},
			want: "1 ::int",
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
			name: "expression $ bind parameter with type cast",
			fields: &tree.ExpressionBindParameter{
				Parameter: "$foo",
				TypeCast:  tree.TypeCastText,
			},
			want: "$foo ::text",
		},
		{
			name: "expression @ bind parameter",
			fields: &tree.ExpressionBindParameter{
				Parameter: "@foo",
			},
			want: "@foo",
		},
		{
			name: "expression @ bind parameter with type cast",
			fields: &tree.ExpressionBindParameter{
				Parameter: "@foo",
				TypeCast:  tree.TypeCastText,
			},
			want: "@foo ::text",
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
			name: "expression column with type cast",
			fields: &tree.ExpressionColumn{
				Column:   "foo",
				TypeCast: tree.TypeCastText,
			},
			want: `"foo" ::text`,
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
			name: "expression column with table and type cast",
			fields: &tree.ExpressionColumn{
				Table:    "bar",
				Column:   "foo",
				TypeCast: tree.TypeCastText,
			},
			want: `"bar"."foo" ::text`,
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
			name: "expression unary operator with type cast",
			fields: &tree.ExpressionUnary{
				Operator: tree.UnaryOperatorNot,
				Operand: &tree.ExpressionColumn{
					Column: "foo",
				},
				TypeCast: tree.TypeCastInt,
				Wrapped:  true,
			},
			want: `(NOT "foo") ::int`,
		},
		{
			name: "expression unary operator with type cast without wrapped",
			fields: &tree.ExpressionUnary{
				Operator: tree.UnaryOperatorNot,
				Operand: &tree.ExpressionColumn{
					Column: "foo",
				},
				TypeCast: tree.TypeCastInt,
			},
			wantPanic: true,
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
			name: "expression binary comparison with type cast",
			fields: &tree.ExpressionBinaryComparison{
				Left: &tree.ExpressionColumn{
					Column: "foo",
				},
				Operator: tree.ComparisonOperatorEqual,
				Right: &tree.ExpressionLiteral{
					Value: "'bar'",
				},
				TypeCast: tree.TypeCastInt,
				Wrapped:  true,
			},
			want: `("foo" = 'bar') ::int`,
		},
		{
			name: "expression binary comparison with type cast without wrapped",
			fields: &tree.ExpressionBinaryComparison{
				Left: &tree.ExpressionColumn{
					Column: "foo",
				},
				Operator: tree.ComparisonOperatorEqual,
				Right: &tree.ExpressionLiteral{
					Value: "'bar'",
				},
				TypeCast: tree.TypeCastInt,
			},
			wantPanic: true,
		},
		{
			name: "expression abs function",
			fields: &tree.ExpressionFunction{
				Function: tree.FunctionABSGetter(nil),
				Inputs: []tree.Expression{
					&tree.ExpressionColumn{
						Column: "foo",
					},
				},
			},
			want: "abs(\"foo\")",
		},
		{
			name: "expression abs function with type cast",
			fields: &tree.ExpressionFunction{
				Function: &tree.FunctionABS,
				Inputs: []tree.Expression{
					&tree.ExpressionColumn{
						Column: "foo",
					},
				},
				TypeCast: tree.TypeCastInt,
			},
			want: "abs(\"foo\") ::int",
		},
		{
			name: "expression abs function with multiple inputs",
			fields: &tree.ExpressionFunction{
				Function: tree.FunctionABSGetter(nil),
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
			name: "expression list with element type cast",
			fields: &tree.ExpressionList{
				Expressions: []tree.Expression{
					&tree.ExpressionColumn{
						Column:   "foo",
						TypeCast: tree.TypeCastText,
					},
					&tree.ExpressionColumn{
						Column:   "bar",
						TypeCast: tree.TypeCastInt,
					},
				},
			},
			want: `("foo" ::text, "bar" ::int)`,
		},
		{
			name: "expression list with type cast",
			// NOTE Seems no point in having a type cast for a list?? as we can't have a list types?
			fields: &tree.ExpressionList{
				Expressions: []tree.Expression{
					&tree.ExpressionColumn{
						Column: "foo",
					},
					&tree.ExpressionColumn{
						Column: "bar",
					},
				},
				TypeCast: tree.TypeCastText,
			},
			want: "(\"foo\", \"bar\") ::text",
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
				Collation: tree.CollationTypeNoCase,
			},
			want: `"foo" = 'bar' COLLATE NOCASE`,
		},
		{
			name: "collate with type cast",
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
				Collation: tree.CollationTypeNoCase,
				Wrapped:   true,
				TypeCast:  tree.TypeCastInt,
			},
			want: `("foo" = 'bar' COLLATE NOCASE) ::int`,
		},
		{
			name: "collate with type cast without wrapped",
			fields: &tree.ExpressionCollate{
				Expression: &tree.ExpressionBinaryComparison{},
				Collation:  tree.CollationTypeNoCase,
				TypeCast:   tree.TypeCastText,
			},
			wantPanic: true,
		},
		{
			name: "collate with no expression",
			fields: &tree.ExpressionCollate{
				Collation: tree.CollationTypeNoCase,
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
			name: "string compare with escape and type cast",
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
				TypeCast: tree.TypeCastInt,
				Wrapped:  true,
			},
			want: `("foo" NOT LIKE 'bar' ESCAPE 'baz') ::int`,
		},
		{
			name: "string compare with escape and type cast without wrapped",
			fields: &tree.ExpressionStringCompare{
				TypeCast: tree.TypeCastInt,
			},
			wantPanic: true,
		},
		{
			name: "IsNull",
			fields: &tree.ExpressionIs{
				Left: &tree.ExpressionColumn{
					Column: "foo",
				},
				Right: &tree.ExpressionLiteral{
					Value: "NULL",
				},
			},
			want: `"foo" IS NULL`,
		},
		{
			name: "IsNull with type cast",
			fields: &tree.ExpressionIs{
				Left: &tree.ExpressionColumn{
					Column: "foo",
				},
				Right: &tree.ExpressionLiteral{
					Value: "NULL",
				},
				TypeCast: tree.TypeCastInt,
				Wrapped:  true,
			},
			want: `("foo" IS NULL) ::int`,
		},
		{
			name: "IsNull with type cast without wrapped",
			fields: &tree.ExpressionIs{
				Left: &tree.ExpressionColumn{
					Column: "foo",
				},
				Right: &tree.ExpressionLiteral{
					Value: "NULL",
				},
				TypeCast: tree.TypeCastInt,
			},
			wantPanic: true,
		},
		{
			name:      "IsNull with no expression",
			fields:    &tree.ExpressionIs{},
			wantPanic: true,
		},
		{
			name: "Is Not Null",
			fields: &tree.ExpressionIs{
				Left: &tree.ExpressionColumn{
					Column: "foo",
				},
				Right: &tree.ExpressionLiteral{
					Value: "NULL",
				},
				Not: true,
			},
			want: `"foo" IS NOT NULL`,
		},
		{
			name: "Is Not Null with type cast",
			fields: &tree.ExpressionIs{
				Left: &tree.ExpressionColumn{
					Column: "foo",
				},
				Right: &tree.ExpressionLiteral{
					Value: "NULL",
				},
				Not:      true,
				TypeCast: tree.TypeCastInt,
				Wrapped:  true,
			},
			want: `("foo" IS NOT NULL) ::int`,
		},
		{
			name: "is not distinct from",
			fields: &tree.ExpressionIs{
				Left: &tree.ExpressionColumn{
					Column: "foo",
				},
				Right: &tree.ExpressionLiteral{
					Value: "'bar'",
				},
				Distinct: true,
				Not:      true,
			},
			want: `"foo" IS NOT DISTINCT FROM 'bar'`,
		},
		{
			name: "is not distinct from with type cast",
			fields: &tree.ExpressionIs{
				Left: &tree.ExpressionColumn{
					Column: "foo",
				},
				Right: &tree.ExpressionLiteral{
					Value: "'bar'",
				},
				Distinct: true,
				Not:      true,
				TypeCast: tree.TypeCastInt,
				Wrapped:  true,
			},
			want: `("foo" IS NOT DISTINCT FROM 'bar') ::int`,
		},
		{
			name: "is not distinct from with type cast without wrapped",
			fields: &tree.ExpressionIs{
				Left: &tree.ExpressionColumn{
					Column: "foo",
				},
				Right: &tree.ExpressionLiteral{
					Value: "'bar'",
				},
				Distinct: true,
				Not:      true,
				TypeCast: tree.TypeCastInt,
			},
			wantPanic: true,
		},
		{
			name: "expr is expr",
			fields: &tree.ExpressionIs{
				Left: &tree.ExpressionColumn{
					Column: "foo",
				},
				Right: &tree.ExpressionLiteral{
					Value: "'bar'",
				},
			},
			want: `"foo" IS 'bar'`,
		},
		{
			name: "distinct with no left",
			fields: &tree.ExpressionIs{
				Right: &tree.ExpressionLiteral{
					Value: "'bar'",
				},
			},
			wantPanic: true,
		},
		{
			name: "distinct with no right",
			fields: &tree.ExpressionIs{
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
			name: "valid between with type cast",
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
				TypeCast: tree.TypeCastInt,
				Wrapped:  true,
			},
			want: `("foo" NOT BETWEEN 'bar' AND 'baz') ::int`,
		},
		{
			name: "valid between with type cast without wrapped",
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
				TypeCast: tree.TypeCastInt,
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
			name: "select subquery with type cast",
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
				TypeCast: tree.TypeCastInt,
				Wrapped:  true,
			},
			want: `(NOT EXISTS (SELECT * FROM "foo" AS "f" WHERE "f"."foo" = $a)) ::int`,
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
			name: "arithmetic expression with type cast",
			fields: &tree.ExpressionArithmetic{
				Left: &tree.ExpressionColumn{
					Column: "foo",
				},
				Operator: tree.ArithmeticOperatorAdd,
				Right: &tree.ExpressionLiteral{
					Value: "1",
				},
				TypeCast: tree.TypeCastInt,
				Wrapped:  true,
			},
			want: `("foo" + 1) ::int`,
		},
		{
			name: "arithmetic expression with type cast without wrapped",
			fields: &tree.ExpressionArithmetic{
				Left: &tree.ExpressionColumn{
					Column: "foo",
				},
				Operator: tree.ArithmeticOperatorAdd,
				Right: &tree.ExpressionLiteral{
					Value: "1",
				},
				TypeCast: tree.TypeCastInt,
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
