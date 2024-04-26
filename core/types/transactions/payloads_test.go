package transactions_test

import (
	"testing"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/transactions"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
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
								Type: &transactions.DataType{
									Name: "int",
								},
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
						Annotations: []string{"sql(engine=sqlite3)"},
						Parameters:  []string{"user_id"},
						Public:      true,
						Modifiers:   []string{"view"},
						Body:        "SELECT * FROM users WHERE id = $user_id",
					},
				},
				Extensions: []*transactions.Extension{
					{
						Name: "auth",
						Initialization: []*transactions.ExtensionConfig{
							{
								Key:   "token",
								Value: "abc123",
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
			name: "action_execution with nils",
			obj: &transactions.ActionExecution{
				DBID:   "db_id",
				Action: "action",
				Arguments: [][]string{
					{
						"",
						"arg2",
					},
					{
						"arg3",
						"",
					},
				},
				NilArg: [][]bool{
					{true, false},
					{false, true},
				},
			},
		},
		{
			name: "action_call no nils",
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
			name: "action_call with nil",
			obj: &transactions.ActionCall{
				DBID:   "db_id",
				Action: "action",
				Arguments: []string{
					"",
					"arg2",
				},
				NilArg: []bool{true, false},
			},
		},
		{
			name: "drop_schema",
			obj: &transactions.DropSchema{
				DBID: "db_id",
			},
		},
		{
			name: "transfer funds",
			obj: &transactions.Transfer{
				To:     []byte("to be a user identifier"),
				Amount: "1234123400000",
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
		{
			name: "validator_vote_approve",
			obj: &transactions.ValidatorVoteIDs{
				ResolutionIDs: []types.UUID{
					types.NewUUIDV5([]byte("asdfadsf")),
					types.NewUUIDV5([]byte("asdfad2sf")),
				},
			},
		},
		{
			name: "validator_vote_bodies",
			obj: &transactions.ValidatorVoteBodies{
				Events: []*transactions.VotableEvent{
					{
						Type: "asdfadsf",
						Body: []byte("asdfadsf"),
					},
					{
						Type: "asdfad2sf",
						Body: []byte("asdfad2sf"),
					},
				},
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
			case *transactions.Transfer:
				obj = &transactions.Transfer{}
			case *transactions.ValidatorApprove:
				obj = &transactions.ValidatorApprove{}
			case *transactions.ValidatorJoin:
				obj = &transactions.ValidatorJoin{}
			case *transactions.ValidatorLeave:
				obj = &transactions.ValidatorLeave{}
			case *transactions.ValidatorRemove:
				obj = &transactions.ValidatorRemove{}
			case *transactions.ValidatorVoteIDs:
				obj = &transactions.ValidatorVoteIDs{}
			case *transactions.ValidatorVoteBodies:
				obj = &transactions.ValidatorVoteBodies{}
			default:
				t.Fatal("unknown type")
			}

			if err := obj.UnmarshalBinary(bts); err != nil {
				t.Fatal(err)
			}

			// compare, considering empty and nil slices the same
			if !cmp.Equal(tc.obj, obj, cmpopts.EquateEmpty()) {
				t.Error("objects are not equal")
				assert.EqualValuesf(t, tc.obj, obj, "objects are not equal") // for the diff
			}
		})
	}
}

func TestUnmarshalPayload(t *testing.T) {
	// for each payload type, ensure UnmarshalPayload can recreate an instance
	// of the expected type from just []byte and PayloadType. Contents and RLP
	// quirks are not important, only that the type returned from
	// UnmarshalPayload is correct.
	tests := []transactions.Payload{
		&transactions.DropSchema{},
		&transactions.Schema{},
		&transactions.ActionCall{},
		&transactions.ActionExecution{},
		&transactions.Transfer{},
		&transactions.ValidatorApprove{},
		&transactions.ValidatorJoin{},
		&transactions.ValidatorRemove{},
		&transactions.ValidatorLeave{},
		&transactions.ValidatorVoteIDs{},
		&transactions.ValidatorVoteBodies{},
	}
	for _, tt := range tests {
		t.Run(tt.Type().String(), func(t *testing.T) {
			payloadType := tt.Type()
			payloadIn, err := tt.MarshalBinary() // serialize.Encode(tt.in)
			if err != nil {
				t.Errorf("failed to encode input payload object: %v", err)
			}

			got, err := transactions.UnmarshalPayload(payloadType, payloadIn)
			if err != nil {
				t.Errorf("failed to unmarshal payload: %v", err)
			}

			assert.IsType(t, tt, got)

			// compare, considering empty and nil slices the same since RLP
			// can't round-trip things :/
			if !cmp.Equal(got, tt, cmpopts.EquateEmpty()) {
				t.Error("objects are not equal")
				assert.EqualValuesf(t, got, tt, "objects are not equal") // for the diff
			}
		})
	}
}
