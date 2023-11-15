package database

import (
	"github.com/kwilteam/kwil-db/cmd/common/display"
	"github.com/kwilteam/kwil-db/core/client"
	"github.com/kwilteam/kwil-db/core/types/transactions"
)

func Example_respDBlist_text_0() {
	display.Print(
		&respDBList{Databases: []string{}, Owner: []byte("owner")},
		nil, "text")
	// Output:
	// No databases found for '6f776e6572'.
}

func Example_respDBlist_text() {
	display.Print(
		&respDBList{Databases: []string{"db_a", "db_b"}, Owner: []byte("owner")},
		nil, "text")
	// Output:
	// Databases belonging to '6f776e6572':
	//  - db_a   (dbid:xf1a24857f73e3bbdeaae383338e8fb4bde364e959207bd2327e375ea)
	//  - db_b   (dbid:x63e828a14a11c00b84adc9fc1473c5104557cd857ca81588638bb1f3)
}

func Example_respDBlist_json() {
	display.Print(
		&respDBList{Databases: []string{"db_a", "db_b"}, Owner: []byte("owner")},
		nil, "json")
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
	display.Print(&respRelations{
		Data: client.NewRecordsFromMaps([]map[string]any{{"a": "1", "b": "2"}, {"a": "3", "b": "4"}})},
		nil, "text")
	// Output:
	// | a | b |
	// +---+---+
	// | 1 | 2 |
	// | 3 | 4 |
}

func Example_respRelations_json() {
	display.Print(&respRelations{
		Data: client.NewRecordsFromMaps([]map[string]any{{"a": "1", "b": "2"}, {"a": "3", "b": "4"}})},
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
	//   get_user
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
