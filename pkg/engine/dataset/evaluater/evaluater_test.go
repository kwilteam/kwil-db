package evaluater_test

import (
	"fmt"
	"testing"

	"github.com/kwilteam/kwil-db/pkg/engine/dataset/evaluater"
	"github.com/kwilteam/kwil-db/pkg/engine/sqlparser/tree"
)

func Test_Evaluater(t *testing.T) {
	type testCase struct {
		name    string
		expr    tree.Expression
		values  map[string]any
		want    any
		wantErr bool
	}

	for _, tt := range []testCase{
		{
			name: "bind parameter int",
			expr: &tree.ExpressionBindParameter{
				Parameter: "$a",
			},
			values: map[string]any{
				"$a": 1,
			},
			want: 1,
		},
		{
			name: "bind parameter text",
			expr: &tree.ExpressionBindParameter{
				Parameter: "$a",
			},
			values: map[string]any{
				"$a": "satoshi",
			},
			want: "satoshi",
		},
		{
			name: "string literal",
			expr: &tree.ExpressionLiteral{
				Value: "'satoshi'", //  testing that the quotes are removed
			},
			want: "satoshi",
		},
		{
			name: "function with math",
			expr: &tree.ExpressionFunction{
				Function: &tree.FunctionABS,
				Inputs: []tree.Expression{
					&tree.ExpressionArithmetic{
						Left:     &tree.ExpressionLiteral{Value: "1"},
						Operator: tree.ArithmeticOperatorAdd,
						Right:    &tree.ExpressionLiteral{Value: "2"},
					},
				},
			},
			want: 3,
		},
		{
			name: "unary",
			expr: &tree.ExpressionUnary{
				Operator: tree.UnaryOperatorMinus,
				Operand: &tree.ExpressionLiteral{
					Value: "1",
				},
			},
			want: -1,
		},
		{
			name: "binary comparison",
			expr: &tree.ExpressionBinaryComparison{
				Left:     &tree.ExpressionLiteral{Value: "1"},
				Operator: tree.ComparisonOperatorEqual,
				Right:    &tree.ExpressionLiteral{Value: "3"},
			},
			want: 0, // sqlite does not return a boolean, but rather 0 or 1
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			eval, err := evaluater.NewEvaluater()
			if err != nil {
				t.Fatalf("failed to create evaluater: %v", err)
			}

			prepped, err := evaluater.PrepareExpression(tt.expr)
			if err != nil {
				t.Fatalf("failed to prepare expression: %v", err)
			}

			got, err := eval.Evaluate(prepped, tt.values)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Evaluate() error = %v, wantErr %v", err, tt.wantErr)
			}

			if fmt.Sprint(got) != fmt.Sprint(tt.want) {
				t.Fatalf("Evaluate() got = %v, want %v", got, tt.want)
			}
		})
	}
}
