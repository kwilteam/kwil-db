package ddl_test

import (
	"fmt"
	"testing"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/internal/engine/ddl"
	"github.com/stretchr/testify/assert"
)

func Test_Procedure(t *testing.T) {
	name := "test_procedure"
	schema := "test_schema"
	body := "test_body"

	type testcase struct {
		name        string
		fields      []*types.NamedType
		returns     *types.ProcedureReturn
		decls       []*types.NamedType
		loopTargets []string
		want        string
	}

	tests := []testcase{
		{
			name: "basic usage",
			fields: []*types.NamedType{
				{
					Name: "field1",
					Type: &types.DataType{
						Name: "text",
					},
				},
			},
			returns: nil,
			decls:   nil,
			want:    "CREATE OR REPLACE FUNCTION test_schema.test_procedure(field1 text) \nRETURNS void AS $$\nBEGIN\ntest_body\nEND;\n$$ LANGUAGE plpgsql;",
		},
		{
			name: "multiple fields and return types",
			fields: []*types.NamedType{
				{
					Name: "field1",
					Type: &types.DataType{
						Name:    "text",
						IsArray: true,
					},
				},
				{
					Name: "field2",
					Type: types.TextType,
				},
			},
			returns: &types.ProcedureReturn{
				Types: []*types.DataType{
					{
						Name:    "text",
						IsArray: true,
					},
					{
						Name: "bool",
					},
				},
			},
			decls: nil,
			want:  "CREATE OR REPLACE FUNCTION test_schema.test_procedure(field1 text[], field2 text, OUT _out_0 text[], OUT _out_1 bool) AS $$\nBEGIN\ntest_body\nEND;\n$$ LANGUAGE plpgsql;",
		},
		{
			name:   "no fields, multiple return types",
			fields: nil,
			returns: &types.ProcedureReturn{
				Types: []*types.DataType{
					{
						Name:    "text",
						IsArray: true,
					},
					{
						Name: "bool",
					},
				},
			},
			decls: nil,
			want:  "CREATE OR REPLACE FUNCTION test_schema.test_procedure(OUT _out_0 text[], OUT _out_1 bool) AS $$\nBEGIN\ntest_body\nEND;\n$$ LANGUAGE plpgsql;",
		},
		{
			name: "single field, single return type",
			fields: []*types.NamedType{
				{
					Name: "field1",
					Type: types.TextType,
				},
			},
			returns: &types.ProcedureReturn{
				Types: []*types.DataType{
					types.TextType,
				},
			},
			decls: nil,
			want:  "CREATE OR REPLACE FUNCTION test_schema.test_procedure(field1 text, OUT _out_0 text) AS $$\nBEGIN\ntest_body\nEND;\n$$ LANGUAGE plpgsql;",
		},
		{
			name:   "return table",
			fields: nil,
			returns: &types.ProcedureReturn{
				Table: []*types.NamedType{
					{
						Name: "field1",
						Type: types.TextType,
					},
					{
						Name: "field2",
						Type: &types.DataType{
							Name:    "int",
							IsArray: true,
						},
					},
				},
			},
			decls: []*types.NamedType{
				{
					Name: "local_type",
					Type: types.TextType,
				},
				{
					Name: "cars",
					Type: &types.DataType{
						Name:    "int",
						IsArray: true,
					},
				},
			},
			want: "CREATE OR REPLACE FUNCTION test_schema.test_procedure() \nRETURNS TABLE(field1 text, field2 int[]) AS $$\nDECLARE\nlocal_type text;\ncars int[];\nBEGIN\ntest_body\nEND;\n$$ LANGUAGE plpgsql;",
		},
		{
			name: "variable is declared as parameter",
			fields: []*types.NamedType{
				{
					Name: "field1",
					Type: types.TextType,
				},
			},
			returns: nil,
			decls: []*types.NamedType{
				{
					Name: "field1",
					Type: types.TextType,
				},
			},
			want: "CREATE OR REPLACE FUNCTION test_schema.test_procedure(field1 text) \nRETURNS void AS $$\nBEGIN\ntest_body\nEND;\n$$ LANGUAGE plpgsql;",
		},
		{
			name:    "loops",
			fields:  nil,
			returns: nil,
			decls:   nil,
			loopTargets: []string{
				"loop1",
			},
			want: "CREATE OR REPLACE FUNCTION test_schema.test_procedure() \nRETURNS void AS $$\nDECLARE\nloop1 RECORD;\nBEGIN\ntest_body\nEND;\n$$ LANGUAGE plpgsql;",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var outParams []*types.NamedType
			if test.returns != nil && test.returns.Types != nil {
				for i, t := range test.returns.Types {
					outParams = append(outParams, &types.NamedType{
						Name: fmt.Sprintf("_out_%d", i),
						Type: t,
					})
				}
			}

			got, err := ddl.GenerateProcedure(test.fields, test.loopTargets, test.returns, test.decls, outParams, schema, name, body)
			if err != nil {
				t.Errorf("ddl.GeneratedProcedure() error = %v", err)
				return
			}
			assert.Equal(t, test.want, got)
		})
	}
}
