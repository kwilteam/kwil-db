package generate

import (
	"fmt"
	"testing"

	"github.com/kwilteam/kwil-db/core/types"
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
			want:    "CREATE OR REPLACE FUNCTION test_schema.test_procedure(field1 TEXT) \nRETURNS void AS $$\n#variable_conflict use_column\nBEGIN\ntest_body\nEND;\n$$ LANGUAGE plpgsql;",
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
				Fields: []*types.NamedType{
					{
						Name: "field1", // gets ignored
						Type: &types.DataType{
							Name:    "text",
							IsArray: true,
						},
					},
					{
						Name: "field2", // gets ignored
						Type: &types.DataType{
							Name: "bool",
						},
					},
				},
			},
			decls: nil,
			want:  "CREATE OR REPLACE FUNCTION test_schema.test_procedure(field1 TEXT[], field2 TEXT, OUT _out_0 TEXT[], OUT _out_1 BOOL) AS $$\n#variable_conflict use_column\nBEGIN\ntest_body\nEND;\n$$ LANGUAGE plpgsql;",
		},
		{
			name:   "no fields, multiple return types",
			fields: nil,
			returns: &types.ProcedureReturn{
				Fields: []*types.NamedType{
					{
						Name: "field1", // gets ignored
						Type: &types.DataType{
							Name:    "text",
							IsArray: true,
						},
					},
					{
						Name: "field2", // gets ignored
						Type: &types.DataType{
							Name: "bool",
						},
					},
				},
			},
			decls: nil,
			want:  "CREATE OR REPLACE FUNCTION test_schema.test_procedure(OUT _out_0 TEXT[], OUT _out_1 BOOL) AS $$\n#variable_conflict use_column\nBEGIN\ntest_body\nEND;\n$$ LANGUAGE plpgsql;",
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
				Fields: []*types.NamedType{
					{
						Name: "field1", // gets ignored
						Type: types.TextType,
					},
				},
			},
			decls: nil,
			want:  "CREATE OR REPLACE FUNCTION test_schema.test_procedure(field1 TEXT, OUT _out_0 TEXT) AS $$\n#variable_conflict use_column\nBEGIN\ntest_body\nEND;\n$$ LANGUAGE plpgsql;",
		},
		{
			name:   "return table",
			fields: nil,
			returns: &types.ProcedureReturn{
				IsTable: true,
				Fields: []*types.NamedType{
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
			want: "CREATE OR REPLACE FUNCTION test_schema.test_procedure() \nRETURNS TABLE(field1 TEXT, field2 INT8[]) AS $$\n#variable_conflict use_column\nDECLARE\nlocal_type TEXT;\ncars INT8[];\nBEGIN\ntest_body\nEND;\n$$ LANGUAGE plpgsql;",
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
			want: "CREATE OR REPLACE FUNCTION test_schema.test_procedure(field1 TEXT) \nRETURNS void AS $$\n#variable_conflict use_column\nBEGIN\ntest_body\nEND;\n$$ LANGUAGE plpgsql;",
		},
		{
			name:    "loops",
			fields:  nil,
			returns: nil,
			decls:   nil,
			loopTargets: []string{
				"loop1",
			},
			want: "CREATE OR REPLACE FUNCTION test_schema.test_procedure() \nRETURNS void AS $$\n#variable_conflict use_column\nDECLARE\nloop1 RECORD;\nBEGIN\ntest_body\nEND;\n$$ LANGUAGE plpgsql;",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var outParams []*types.NamedType
			if test.returns != nil && !test.returns.IsTable {
				for i, t := range test.returns.Fields {
					outParams = append(outParams, &types.NamedType{
						Name: fmt.Sprintf("_out_%d", i),
						Type: t.Type,
					})
				}

				test.returns.Fields = outParams
			}

			ret := types.ProcedureReturn{}
			if test.returns != nil {
				ret = *test.returns
			}

			got, err := generateProcedureWrapper(&analyzedProcedure{
				Name:              name,
				Parameters:        test.fields,
				Returns:           ret,
				DeclaredVariables: test.decls,
				LoopTargets:       test.loopTargets,
				Body:              body,
				IsView:            false,
				OwnerOnly:         false,
			}, schema)
			if err != nil {
				t.Errorf("ddl.GeneratedProcedure() error = %v", err)
				return
			}
			assert.Equal(t, test.want, got)
		})
	}
}
