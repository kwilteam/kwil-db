package database

import (
	"encoding/hex"

	"github.com/kwilteam/kwil-db/cmd/common/display"
	"github.com/kwilteam/kwil-db/core/types"
	clientType "github.com/kwilteam/kwil-db/core/types/client"
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

// func mustDecodeBase64(s string) []byte {
// 	b, err := base64.StdEncoding.DecodeString(s)
// 	if err != nil {
// 		panic(err)
// 	}
// 	return b
// }

func Example_respDBlist_json() {

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
		}},
		nil, "json")

	// Output:
	// {
	//   "result": [
	//     {
	//       "name": "db_a",
	//       "owner": "6f776e6572",
	//       "dbid": "xabc"
	//     },
	//     {
	//       "name": "db_b",
	//       "owner": "6f776e6572",
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
	Schema: &types.Schema{
		Owner: []byte("user"),
		Name:  "test_schema",
		Tables: []*types.Table{
			{
				Name: "users",
				Columns: []*types.Column{
					{
						Name: "id",
						Type: types.IntType,
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
				Name:       "get_user",
				Parameters: []string{"user_id"},
				Public:     true,
				Body:       "SELECT * FROM users WHERE id = $user_id",
			},
		},
		Extensions: []*types.Extension{
			{
				Name: "auth",
				Initialization: []*types.ExtensionConfig{
					{
						Key:   "token",
						Value: "abc123",
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
	//       Type: int
	//       primary_key
	//         true
	// Actions:
	//   get_user (public)
	//     Inputs: [user_id]
	// Procedures:
}
