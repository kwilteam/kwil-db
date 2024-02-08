package database

import (
	"encoding/base64"
	"encoding/hex"

	"github.com/kwilteam/kwil-db/cmd/common/display"
	"github.com/kwilteam/kwil-db/core/types"
	clientType "github.com/kwilteam/kwil-db/core/types/client"
	"github.com/kwilteam/kwil-db/core/types/transactions"
)

func Example_respDBlist_text_0() {
	display.Print(
		&respDBList{Info: []*types.DatasetIdentifier{}, owner: mustDecodeHex("6f776e6572")},
		nil, "text")
	// Output:
	// No databases found for '6f776e6572'.
}

func Example_respDBlist_text() {
	display.Print(
		&respDBList{Info: []*types.DatasetIdentifier{
			{
				Name:  "db_a",
				Owner: mustDecodeHex("6f776e6572"),
				DBID:  "xabc",
			},
			{
				Name:  "db_b",
				Owner: mustDecodeHex("6f776e6572"),
				DBID:  "xdef",
			},
		},
			owner: mustDecodeHex("6f776e6572")},
		nil, "text")
	// Output:
	// Databases belonging to '6f776e6572':
	//   DBID: xabc
	//     Name: db_a
	//     Owner: 6f776e6572
	//   DBID: xdef
	//     Name: db_b
	//     Owner: 6f776e6572
}

func mustDecodeHex(s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return b
}

func mustDecodeBase64(s string) []byte {
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return b
}

func Example_respDBlist_json() {

	display.Print(
		&respDBList{Info: []*types.DatasetIdentifier{
			{
				Name:  "db_a",
				Owner: mustDecodeBase64("b3duZXI="),
				DBID:  "xabc",
			},
			{
				Name:  "db_b",
				Owner: mustDecodeBase64("b3duZXI="),
				DBID:  "xdef",
			},
		}},
		nil, "json")

	// Output:
	// {
	//   "result": [
	//     {
	//       "name": "db_a",
	//       "owner": "b3duZXI=",
	//       "dbid": "xabc"
	//     },
	//     {
	//       "name": "db_b",
	//       "owner": "b3duZXI=",
	//       "dbid": "xdef"
	//     }
	//   ],
	//   "error": ""
	// }
}

func Example_respRelations_text() {
	display.Print(&respRelations{
		Data: clientType.NewRecordsFromMaps([]map[string]any{{"a": "1", "b": "2"}, {"a": "3", "b": "4"}})},
		nil, "text")
	// Output:
	// | a | b |
	// +---+---+
	// | 1 | 2 |
	// | 3 | 4 |
}

func Example_respRelations_json() {
	display.Print(&respRelations{
		Data: clientType.NewRecordsFromMaps([]map[string]any{{"a": "1", "b": "2"}, {"a": "3", "b": "4"}})},
		nil, "json")
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
	display.Print(demoSchema, nil, "text")
	// Output:
	// Tables:
	//   users
	//     Columns:
	//     id
	//       Type: integer
	//       primary_key
	//         true
	// Actions:
	//   get_user (public)
	//     Inputs: [user_id]
}

func Example_respSchema_json() {
	display.Print(demoSchema, nil, "json")
	// Output:
	// {
	//   "result": {
	//     "owner": "dXNlcg==",
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
