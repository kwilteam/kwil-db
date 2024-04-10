package tree_test

import (
	"testing"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/internal/parse/sql/tree"
)

func TestExpressionTextLiteral_ToSQL(t *testing.T) {
	type fields tree.Expression
	tests := []struct {
		name      string
		fields    fields
		want      string
		wantPanic bool
	}{
		{
			name: "expression literal",
			fields: &tree.ExpressionTextLiteral{
				Value: "foo",
			},
			want: "'foo'",
		},
		{
			name: "expression literal with type cast",
			fields: &tree.ExpressionTextLiteral{
				Value:    "foo",
				TypeCast: types.TextType,
			},
			want: "'foo' ::TEXT",
			// not variable. this should be the result string, no interpolation
		},
		{
			name: "expression literal with wrapped paren",
			fields: &tree.ExpressionTextLiteral{
				Value:   "foo",
				Wrapped: true,
			},
			want: "( 'foo' )",
		},
		{
			name: "expression literal with wrapped paren and type cast",
			fields: &tree.ExpressionTextLiteral{
				Value:    "foo",
				Wrapped:  true,
				TypeCast: types.TextType,
			},
			want: "( 'foo' ) ::TEXT",
		},
		{
			name: "expression literal with int",
			fields: &tree.ExpressionNumericLiteral{
				Value: 1,
			},
			want: "1",
		},
		{
			name: "expression literal with int and type cast",
			fields: &tree.ExpressionNumericLiteral{
				Value:    1,
				TypeCast: types.IntType,
			},
			want: "1 ::INT8",
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
				TypeCast:  types.TextType,
			},
			want: "$foo ::TEXT",
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
				TypeCast:  types.TextType,
			},
			want: "@foo ::TEXT",
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
				TypeCast: types.TextType,
			},
			want: `"foo" ::TEXT`,
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
				TypeCast: types.TextType,
			},
			want: `"bar"."foo" ::TEXT`,
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
				TypeCast: types.IntType,
				Wrapped:  true,
			},
			want: `(NOT "foo") ::INT8`,
		},
		{
			name: "expression unary operator with type cast without wrapped",
			fields: &tree.ExpressionUnary{
				Operator: tree.UnaryOperatorNot,
				Operand: &tree.ExpressionColumn{
					Column: "foo",
				},
				TypeCast: types.IntType,
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
				Right: &tree.ExpressionTextLiteral{
					Value: "bar",
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
				Right: &tree.ExpressionTextLiteral{
					Value: "bar",
				},
				TypeCast: types.IntType,
				Wrapped:  true,
			},
			want: `("foo" = 'bar') ::INT8`,
		},
		{
			name: "expression binary comparison with type cast without wrapped",
			fields: &tree.ExpressionBinaryComparison{
				Left: &tree.ExpressionColumn{
					Column: "foo",
				},
				Operator: tree.ComparisonOperatorEqual,
				Right: &tree.ExpressionTextLiteral{
					Value: "bar",
				},
				TypeCast: types.IntType,
			},
			wantPanic: true,
		},
		{
			name: "expression abs function",
			fields: &tree.ExpressionFunction{
				Function: "abs",
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
				Function: "abs",
				Inputs: []tree.Expression{
					&tree.ExpressionColumn{
						Column: "foo",
					},
				},
				TypeCast: types.IntType,
			},
			want: "abs(\"foo\") ::INT8",
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
						TypeCast: types.TextType,
					},
					&tree.ExpressionColumn{
						Column:   "bar",
						TypeCast: types.IntType,
					},
				},
			},
			want: `("foo" ::TEXT, "bar" ::INT8)`,
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
				TypeCast: types.TextType,
			},
			want: "(\"foo\", \"bar\") ::TEXT",
		},
		{
			name: "collate",
			fields: &tree.ExpressionCollate{
				Expression: &tree.ExpressionBinaryComparison{
					Left: &tree.ExpressionColumn{
						Column: "foo",
					},
					Operator: tree.ComparisonOperatorEqual,
					Right: &tree.ExpressionTextLiteral{
						Value: "bar",
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
					Right: &tree.ExpressionTextLiteral{
						Value: "bar",
					},
				},
				Collation: tree.CollationTypeNoCase,
				Wrapped:   true,
				TypeCast:  types.IntType,
			},
			want: `("foo" = 'bar' COLLATE NOCASE) ::INT8`,
		},
		{
			name: "collate with type cast without wrapped",
			fields: &tree.ExpressionCollate{
				Expression: &tree.ExpressionBinaryComparison{},
				Collation:  tree.CollationTypeNoCase,
				TypeCast:   types.TextType,
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
					Right: &tree.ExpressionTextLiteral{
						Value: "bar",
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
				Right: &tree.ExpressionTextLiteral{
					Value: "bar",
				},
				Escape: &tree.ExpressionTextLiteral{
					Value: "baz",
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
				Right: &tree.ExpressionTextLiteral{
					Value: "bar",
				},
				Escape: &tree.ExpressionTextLiteral{
					Value: "baz",
				},
				TypeCast: types.IntType,
				Wrapped:  true,
			},
			want: `("foo" NOT LIKE 'bar' ESCAPE 'baz') ::INT8`,
		},
		{
			name: "string compare with escape and type cast without wrapped",
			fields: &tree.ExpressionStringCompare{
				TypeCast: types.IntType,
			},
			wantPanic: true,
		},
		{
			name: "IsNull",
			fields: &tree.ExpressionIs{
				Left: &tree.ExpressionColumn{
					Column: "foo",
				},
				Right: &tree.ExpressionNullLiteral{},
			},
			want: `"foo" IS NULL`,
		},
		{
			name: "IsNull with type cast",
			fields: &tree.ExpressionIs{
				Left: &tree.ExpressionColumn{
					Column: "foo",
				},
				Right:    &tree.ExpressionNullLiteral{},
				TypeCast: types.IntType,
				Wrapped:  true,
			},
			want: `("foo" IS NULL) ::INT8`,
		},
		{
			name: "IsNull with type cast without wrapped",
			fields: &tree.ExpressionIs{
				Left: &tree.ExpressionColumn{
					Column: "foo",
				},
				Right:    &tree.ExpressionNullLiteral{},
				TypeCast: types.IntType,
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
				Right: &tree.ExpressionNullLiteral{},
				Not:   true,
			},
			want: `"foo" IS NOT NULL`,
		},
		{
			name: "Is Not Null with type cast",
			fields: &tree.ExpressionIs{
				Left: &tree.ExpressionColumn{
					Column: "foo",
				},
				Right:    &tree.ExpressionNullLiteral{},
				Not:      true,
				TypeCast: types.IntType,
				Wrapped:  true,
			},
			want: `("foo" IS NOT NULL) ::INT8`,
		},
		{
			name: "is not distinct from",
			fields: &tree.ExpressionIs{
				Left: &tree.ExpressionColumn{
					Column: "foo",
				},
				Right: &tree.ExpressionTextLiteral{
					Value: "bar",
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
				Right: &tree.ExpressionTextLiteral{
					Value: "bar",
				},
				Distinct: true,
				Not:      true,
				TypeCast: types.IntType,
				Wrapped:  true,
			},
			want: `("foo" IS NOT DISTINCT FROM 'bar') ::INT8`,
		},
		{
			name: "is not distinct from with type cast without wrapped",
			fields: &tree.ExpressionIs{
				Left: &tree.ExpressionColumn{
					Column: "foo",
				},
				Right: &tree.ExpressionTextLiteral{
					Value: "bar",
				},
				Distinct: true,
				Not:      true,
				TypeCast: types.IntType,
			},
			wantPanic: true,
		},
		{
			name: "expr is expr",
			fields: &tree.ExpressionIs{
				Left: &tree.ExpressionColumn{
					Column: "foo",
				},
				Right: &tree.ExpressionTextLiteral{
					Value: "bar",
				},
			},
			want: `"foo" IS 'bar'`,
		},
		{
			name: "distinct with no left",
			fields: &tree.ExpressionIs{
				Right: &tree.ExpressionTextLiteral{
					Value: "bar",
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
				Left: &tree.ExpressionTextLiteral{
					Value: "bar",
				},
				Right: &tree.ExpressionTextLiteral{
					Value: "baz",
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
				Left: &tree.ExpressionTextLiteral{
					Value: "bar",
				},
				Right: &tree.ExpressionTextLiteral{
					Value: "baz",
				},
				TypeCast: types.IntType,
				Wrapped:  true,
			},
			want: `("foo" NOT BETWEEN 'bar' AND 'baz') ::INT8`,
		},
		{
			name: "valid between with type cast without wrapped",
			fields: &tree.ExpressionBetween{
				Expression: &tree.ExpressionColumn{
					Column: "foo",
				},
				NotBetween: true,
				Left: &tree.ExpressionTextLiteral{
					Value: "bar",
				},
				Right: &tree.ExpressionTextLiteral{
					Value: "baz",
				},
				TypeCast: types.IntType,
			},
			wantPanic: true,
		},
		{
			name: "between with no expression",
			fields: &tree.ExpressionBetween{
				Left: &tree.ExpressionTextLiteral{
					Value: "bar",
				},
				Right: &tree.ExpressionTextLiteral{
					Value: "baz",
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
				Right: &tree.ExpressionTextLiteral{
					Value: "baz",
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
				Left: &tree.ExpressionTextLiteral{
					Value: "bar",
				},
			},
			wantPanic: true,
		},
		{
			name: "select subquery",
			fields: &tree.ExpressionSelect{
				IsNot:    true,
				IsExists: true,
				Select: &tree.SelectCore{
					SimpleSelects: []*tree.SimpleSelect{
						{
							SelectType: tree.SelectTypeAll,
							From: &tree.RelationTable{
								Name:  "foo",
								Alias: "f",
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
				Select: &tree.SelectCore{
					SimpleSelects: []*tree.SimpleSelect{
						{
							SelectType: tree.SelectTypeAll,
							From: &tree.RelationTable{
								Name:  "foo",
								Alias: "f",
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
				TypeCast: types.IntType,
				Wrapped:  true,
			},
			want: `(NOT EXISTS (SELECT * FROM "foo" AS "f" WHERE "f"."foo" = $a)) ::INT8`,
		},
		{
			name: "case expression",
			fields: &tree.ExpressionCase{
				CaseExpression: &tree.ExpressionColumn{
					Column: "foo",
				},
				WhenThenPairs: [][2]tree.Expression{
					{
						&tree.ExpressionTextLiteral{
							Value: "bar",
						},
						&tree.ExpressionTextLiteral{
							Value: "baz",
						},
					},
				},
				ElseExpression: &tree.ExpressionTextLiteral{
					Value: "qux",
				},
			},
			want: `CASE "foo" WHEN 'bar' THEN 'baz' ELSE 'qux' END`,
		},
		{
			name: "case expression with no case expression",
			fields: &tree.ExpressionCase{
				WhenThenPairs: [][2]tree.Expression{
					{
						&tree.ExpressionTextLiteral{
							Value: "bar",
						},
						&tree.ExpressionTextLiteral{
							Value: "baz",
						},
					},
				},
				ElseExpression: &tree.ExpressionTextLiteral{
					Value: "qux",
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
				ElseExpression: &tree.ExpressionTextLiteral{
					Value: "qux",
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
						&tree.ExpressionTextLiteral{
							Value: "bar",
						},
						&tree.ExpressionTextLiteral{
							Value: "baz",
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
				Right: &tree.ExpressionNumericLiteral{
					Value: 1,
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
				Right: &tree.ExpressionNumericLiteral{
					Value: 1,
				},
				TypeCast: types.IntType,
				Wrapped:  true,
			},
			want: `("foo" + 1) ::INT8`,
		},
		{
			name: "arithmetic expression with type cast without wrapped",
			fields: &tree.ExpressionArithmetic{
				Left: &tree.ExpressionColumn{
					Column: "foo",
				},
				Operator: tree.ArithmeticOperatorAdd,
				Right: &tree.ExpressionTextLiteral{
					Value: "1",
				},
				TypeCast: types.IntType,
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
