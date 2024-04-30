package parser_test

import (
	"testing"

	"github.com/kwilteam/kwil-db/core/types"
	parser "github.com/kwilteam/kwil-db/parse/procedures/parser"
	"github.com/stretchr/testify/require"
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
			require.EqualValues(t, tt.want[i], r, "unexpected result")
		}
	}
}
