package types_test

import (
	"testing"

	"github.com/kwilteam/kwil-db/pkg/types"
	"github.com/stretchr/testify/assert"
)

type serializable interface {
	Bytes() ([]byte, error)
	FromBytes([]byte) error
}

// this simply test that they all serialize and comply with RLP
func Test_Types(t *testing.T) {
	type testCase struct {
		name string
		obj  serializable
	}

	testCases := []testCase{
		{
			name: "schema",
			obj: &types.Schema{
				Owner: "user",
				Name:  "test_schema",
				Tables: []*types.Table{
					{
						Name: "users",
						Columns: []*types.Column{
							{
								Name: "id",
								Type: "integer",
								Attributes: []*types.Attribute{
									{
										Type:  "primary_key",
										Value: "true",
									},
								},
							},
						},
						ForeignKeys: []*types.ForeignKey{
							{
								ChildKeys:   []string{"child_id"},
								ParentKeys:  []string{"parent_id"},
								ParentTable: "parent_table",
								Actions: []*types.ForeignKeyAction{
									{
										On: "delete",
										Do: "cascade",
									},
								},
							},
						},
						Indexes: []*types.Index{
							{
								Name:    "index_name",
								Columns: []string{"id", "name"},
								Type:    "btree",
							},
						},
					},
				},
				Actions: []*types.Action{
					{
						Name:        "get_user",
						Inputs:      []string{"user_id"},
						Mutability:  types.MutabilityUpdate.String(),
						Auxiliaries: []string{types.AuxiliaryTypeMustSign.String()},
						Public:      true,
						Statements:  []string{"SELECT * FROM users WHERE id = $user_id"},
					},
				},
				Extensions: []*types.Extension{
					{
						Name: "auth",
						Config: []*types.ExtensionConfig{
							{
								Argument: "token",
								Value:    "abc123",
							},
						},
						Alias: "authentication",
					},
				},
			},
		},
		{
			name: "action_execution",
			obj: &types.ActionExecution{
				DBID:   "db_id",
				Action: "action",
				Arguments: [][]string{
					{
						"arg1",
						"arg2",
					},
					{
						"arg3",
						"arg4",
					},
				},
			},
		},
		{
			name: "action_call",
			obj: &types.ActionCall{
				DBID:   "db_id",
				Action: "action",
				Arguments: []string{
					"arg1",
					"arg2",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			bts, err := tc.obj.Bytes()
			if err != nil {
				t.Fatal(err)
			}

			var obj serializable
			switch tc.obj.(type) {
			case *types.Schema:
				obj = &types.Schema{}
			case *types.ActionExecution:
				obj = &types.ActionExecution{}
			case *types.ActionCall:
				obj = &types.ActionCall{}
			default:
				t.Fatal("unknown type")
			}

			if err := obj.FromBytes(bts); err != nil {
				t.Fatal(err)
			}

			// reflect
			assert.EqualValuesf(t, tc.obj, obj, "objects are not equal")
		})
	}
}
