package transactions_test

import (
	"testing"

	"github.com/kwilteam/kwil-db/core/types/transactions"

	"github.com/stretchr/testify/assert"
)

// this simply test that they all serialize and comply with RLP
func Test_Types(t *testing.T) {
	type testCase struct {
		name string
		obj  transactions.Payload
	}

	testCases := []testCase{
		{
			name: "schema",
			obj: &transactions.Schema{
				Owner: []byte("user"),
				Name:  "test_schema",
				Tables: []*transactions.Table{
					{
						Name: "users",
						Columns: []*transactions.Column{
							{
								Name: "id",
								Type: "integer",
								Attributes: []*transactions.Attribute{
									{
										Type:  "primary_key",
										Value: "true",
									},
								},
							},
						},
						ForeignKeys: []*transactions.ForeignKey{
							{
								ChildKeys:   []string{"child_id"},
								ParentKeys:  []string{"parent_id"},
								ParentTable: "parent_table",
								Actions: []*transactions.ForeignKeyAction{
									{
										On: "delete",
										Do: "cascade",
									},
								},
							},
						},
						Indexes: []*transactions.Index{
							{
								Name:    "index_name",
								Columns: []string{"id", "name"},
								Type:    "btree",
							},
						},
					},
				},
				Actions: []*transactions.Action{
					{
						Name:        "get_user",
						Inputs:      []string{"user_id"},
						Mutability:  transactions.MutabilityUpdate.String(),
						Auxiliaries: []string{transactions.AuxiliaryTypeMustSign.String()},
						Public:      true,
						Statements:  []string{"SELECT * FROM users WHERE id = $user_id"},
					},
				},
				Extensions: []*transactions.Extension{
					{
						Name: "auth",
						Config: []*transactions.ExtensionConfig{
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
			obj: &transactions.ActionExecution{
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
			obj: &transactions.ActionCall{
				DBID:   "db_id",
				Action: "action",
				Arguments: []string{
					"arg1",
					"arg2",
				},
			},
		},
		{
			name: "drop_schema",
			obj: &transactions.DropSchema{
				DBID: "db_id",
			},
		},
		{
			name: "validator_approve",
			obj: &transactions.ValidatorApprove{
				Candidate: []byte("asdfadsf"),
			},
		},
		{
			name: "validator_join",
			obj: &transactions.ValidatorJoin{
				Power: 1,
			},
		},
		{
			name: "validator_leave",
			obj:  &transactions.ValidatorLeave{},
		},
		{
			name: "validator_remove",
			obj: &transactions.ValidatorRemove{
				Validator: []byte("asdfadsf"),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			bts, err := tc.obj.MarshalBinary()
			if err != nil {
				t.Fatal(err)
			}

			var obj transactions.Payload
			switch tc.obj.(type) {
			case *transactions.Schema:
				obj = &transactions.Schema{}
			case *transactions.ActionExecution:
				obj = &transactions.ActionExecution{}
			case *transactions.ActionCall:
				obj = &transactions.ActionCall{}
			case *transactions.DropSchema:
				obj = &transactions.DropSchema{}
			case *transactions.ValidatorApprove:
				obj = &transactions.ValidatorApprove{}
			case *transactions.ValidatorJoin:
				obj = &transactions.ValidatorJoin{}
			case *transactions.ValidatorLeave:
				obj = &transactions.ValidatorLeave{}
			case *transactions.ValidatorRemove:
				obj = &transactions.ValidatorRemove{}
			default:
				t.Fatal("unknown type")
			}

			if err := obj.UnmarshalBinary(bts); err != nil {
				t.Fatal(err)
			}

			// reflect
			assert.EqualValuesf(t, tc.obj, obj, "objects are not equal")
		})
	}
}
