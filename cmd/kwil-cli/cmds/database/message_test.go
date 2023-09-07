package database

import (
	"errors"
	"testing"

	"github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds/common/display"
	"github.com/kwilteam/kwil-db/pkg/client"
	"github.com/kwilteam/kwil-db/pkg/transactions"
	"github.com/stretchr/testify/assert"
)

// NOTE: could do this for all the other tests,
// but using Example* is more handy and obvious
func Test_respTxHash(t *testing.T) {
	resp := respTxHash("1024")
	expectJson := `{"tx_hash":"31303234"}`
	expectText := `TxHash: 31303234`

	outText, err := resp.MarshalText()
	assert.NoError(t, err, "MarshalText should not return error")
	assert.Equal(t, expectText, outText, "MarshalText should return expected text")

	outJson, err := resp.MarshalJSON()
	assert.NoError(t, err, "MarshalJSON should not return error")
	assert.Equal(t, expectJson, string(outJson), "MarshalJSON should return expected json")
}

func Example_respTxHash_text() {
	msg := display.WrapMsg(respTxHash("1024"), nil)
	display.Print(msg, nil, "text")
	// Output:
	// TxHash: 31303234

}

func Example_respTxHash_json() {
	msg := display.WrapMsg(respTxHash("1024"), nil)
	display.Print(msg, nil, "json")
	// Output:
	// {
	//   "result": {
	//     "tx_hash": "31303234"
	//   },
	//   "error": ""
	// }
}

func Example_respTxHash_json_withError() {
	err := errors.New("an error")
	msg := display.WrapMsg(respTxHash("1024"), err)
	display.Print(msg, err, "json")
	// Output:
	// {
	//   "result": {
	//     "tx_hash": "31303234"
	//   },
	//   "error": "an error"
	// }
}

func Example_respDBlist_text_0() {
	msg := display.WrapMsg(
		&respDBList{Databases: []string{}, Owner: []byte("owner")},
		nil)
	display.Print(msg, nil, "text")
	// Output:
	// No databases found for '6f776e6572'.
}

func Example_respDBlist_text() {
	msg := display.WrapMsg(
		&respDBList{Databases: []string{"db_a", "db_b"}, Owner: []byte("owner")}, nil)
	display.Print(msg, nil, "text")
	// Output:
	// Databases belonging to '6f776e6572':
	//  - db_a   (dbid:xf1a24857f73e3bbdeaae383338e8fb4bde364e959207bd2327e375ea)
	//  - db_b   (dbid:x63e828a14a11c00b84adc9fc1473c5104557cd857ca81588638bb1f3)
}

func Example_respDBlist_json() {
	msg := display.WrapMsg(
		&respDBList{Databases: []string{"db_a", "db_b"}, Owner: []byte("owner")}, nil)
	display.Print(msg, nil, "json")
	// Output:
	// {
	//   "result": {
	//     "databases": [
	//       {
	//         "name": "db_a",
	//         "id": "xf1a24857f73e3bbdeaae383338e8fb4bde364e959207bd2327e375ea"
	//       },
	//       {
	//         "name": "db_b",
	//         "id": "x63e828a14a11c00b84adc9fc1473c5104557cd857ca81588638bb1f3"
	//       }
	//     ],
	//     "owner": "6f776e6572"
	//   },
	//   "error": ""
	// }
}

func Example_respRelations_text() {
	msg := display.WrapMsg(
		&respRelations{
			Data: client.NewRecordsFromMaps([]map[string]any{{"a": "1", "b": "2"}, {"a": "3", "b": "4"}})},
		nil)
	display.Print(msg, nil, "text")
	// Output:
	// | a | b |
	// +---+---+
	// | 1 | 2 |
	// | 3 | 4 |
}

func Example_respRelations_json() {
	msg := display.WrapMsg(
		&respRelations{
			Data: client.NewRecordsFromMaps([]map[string]any{{"a": "1", "b": "2"}, {"a": "3", "b": "4"}})},
		nil)
	display.Print(msg, nil, "json")
	// Output:
	// {
	//   "result": [
	//     {
	//       "a": "1",
	//       "b": "2"
	//     },
	//     {
	//       "a": "3",
	//       "b": "4"
	//     }
	//   ],
	//   "error": ""
	// }
}

var demoSchema = &respSchema{
	Schema: &transactions.Schema{
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
}

func Example_respSchema_text() {
	msg := display.WrapMsg(demoSchema, nil)
	display.Print(msg, nil, "text")
	// Output:
	// Tables:
	//   users
	//     Columns:
	//     id
	//       Type: integer
	//       primary_key
	//         true
	// Actions:
	//   get_user
	//     Inputs: [user_id]
}

func Example_respSchema_json() {
	msg := display.WrapMsg(demoSchema, nil)
	display.Print(msg, nil, "json")
	// Output:
	// {
	//   "result": {
	//     "Owner": "dXNlcg==",
	//     "name": "test_schema",
	//     "tables": [
	//       {
	//         "name": "users",
	//         "columns": [
	//           {
	//             "name": "id",
	//             "type": "integer",
	//             "attributes": [
	//               {
	//                 "type": "primary_key",
	//                 "value": "true"
	//               }
	//             ]
	//           }
	//         ],
	//         "indexes": [
	//           {
	//             "name": "index_name",
	//             "columns": [
	//               "id",
	//               "name"
	//             ],
	//             "type": "btree"
	//           }
	//         ],
	//         "foreign_keys": [
	//           {
	//             "child_keys": [
	//               "child_id"
	//             ],
	//             "parent_keys": [
	//               "parent_id"
	//             ],
	//             "parent_table": "parent_table",
	//             "actions": [
	//               {
	//                 "on": "delete",
	//                 "do": "cascade"
	//               }
	//             ]
	//           }
	//         ]
	//       }
	//     ],
	//     "actions": [
	//       {
	//         "name": "get_user",
	//         "inputs": [
	//           "user_id"
	//         ],
	//         "mutability": "update",
	//         "auxiliaries": [
	//           "mustsign"
	//         ],
	//         "public": true,
	//         "statements": [
	//           "SELECT * FROM users WHERE id = $user_id"
	//         ]
	//       }
	//     ],
	//     "extensions": [
	//       {
	//         "name": "auth",
	//         "config": [
	//           {
	//             "argument": "token",
	//             "value": "abc123"
	//           }
	//         ],
	//         "alias": "authentication"
	//       }
	//     ]
	//   },
	//   "error": ""
	// }
}
