package parser_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/kwilteam/kwil-db/core/types"
	parser "github.com/kwilteam/kwil-db/parse/procedures/parser"
	parseTypes "github.com/kwilteam/kwil-db/parse/types"
)

func Test_Parser(t *testing.T) {
	type testcase struct {
		name string
		stmt string
		want []parser.Statement
	}

	tests := []testcase{
		{
			name: "basic declaration",
			stmt: "$a int;",
			want: []parser.Statement{
				&parser.StatementVariableDeclaration{
					Name: "$a",
					Type: &types.DataType{
						Name: "int",
					},
				},
			},
		},
		{
			name: "declaration with assignment",
			stmt: "$a int := 1;",
			want: []parser.Statement{
				&parser.StatementVariableAssignmentWithDeclaration{
					Name: "$a",
					Type: &types.DataType{
						Name: "int",
					},
					Value: &parser.ExpressionIntLiteral{
						Value: 1,
					},
				},
			},
		},
		{
			name: "declare, then assign",
			stmt: "$a int; $a := 1;",
			want: []parser.Statement{
				&parser.StatementVariableDeclaration{
					Name: "$a",
					Type: &types.DataType{
						Name: "int",
					},
				},
				&parser.StatementVariableAssignment{
					Name: "$a",
					Value: &parser.ExpressionIntLiteral{
						Value: 1,
					},
				},
			},
		},
		{
			name: "call procedure",
			stmt: "$res int := my_procedure();",
			want: []parser.Statement{
				&parser.StatementVariableAssignmentWithDeclaration{
					Name: "$res",
					Type: types.IntType,
					Value: &parser.ExpressionCall{
						Name: "my_procedure",
					},
				},
			},
		},
		{
			name: "foreign call",
			stmt: "other_procedure[$dbid, 'procedure']($arg);",
			want: []parser.Statement{
				&parser.StatementProcedureCall{
					Call: &parser.ExpressionForeignCall{
						Name: "other_procedure",
						ContextArgs: []parser.Expression{
							&parser.ExpressionVariable{
								Name: "$dbid",
							},
							&parser.ExpressionTextLiteral{
								Value: "procedure",
							},
						},
						Arguments: []parser.Expression{
							&parser.ExpressionVariable{
								Name: "$arg",
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		res, err := parser.Parse(tt.stmt)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(res) != len(tt.want) {
			t.Errorf("unexpected result length: got %d, want %d", len(res), len(tt.want))
			return
		}

		for i, r := range res {
			if !deepCompare(r, tt.want[i]) {
				t.Errorf("unexpected result: got %v, want %v", r, tt.want[i])
			}
		}
	}
}

// deepCompare deep compares the values of two nodes.
// It ignores the parseTypes.Node field.
func deepCompare(node1, node2 any) bool {
	// we return true for the parseTypes.Node field,
	// we also need to ignore the unexported "schema" fields
	return cmp.Equal(node1, node2, cmp.Comparer(func(x, y parseTypes.Node) bool {
		return true
	}), cmpopts.IgnoreUnexported(
		parser.StatementVariableDeclaration{},
		parser.StatementVariableAssignment{},
		parser.StatementVariableAssignmentWithDeclaration{},
		parser.StatementProcedureCall{},
		parser.StatementForLoop{},
		parser.StatementIf{},
		parser.StatementSQL{},
		parser.StatementReturn{},
		parser.StatementReturnNext{},
		parser.StatementBreak{},

		parser.ExpressionTextLiteral{},
		parser.ExpressionBooleanLiteral{},
		parser.ExpressionIntLiteral{},
		parser.ExpressionNullLiteral{},
		parser.ExpressionBlobLiteral{},
		parser.ExpressionMakeArray{},
		parser.ExpressionCall{},
		parser.ExpressionForeignCall{},
		parser.ExpressionVariable{},
		parser.ExpressionArrayAccess{},
		parser.ExpressionFieldAccess{},
		parser.ExpressionParenthesized{},
		parser.ExpressionComparison{},
		parser.ExpressionArithmetic{},

		parser.LoopTargetRange{},
		parser.LoopTargetCall{},
		parser.LoopTargetSQL{},
		parser.LoopTargetVariable{},

		parser.IfThen{},
	))
}
