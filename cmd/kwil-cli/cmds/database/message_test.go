package database

import (
	"encoding/hex"

	"github.com/kwilteam/kwil-db/app/shared/display"
	"github.com/kwilteam/kwil-db/core/types"
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
		Data: &types.QueryResult{
			ColumnNames: []string{"a", "b"},
			ColumnTypes: []*types.DataType{types.TextType, types.TextType},
			Values:      [][]any{{"1", "2"}, {"3", "4"}},
		},
	},
		nil, "text")
	// Output:
	// | a | b |
	// +---+---+
	// | 1 | 2 |
	// | 3 | 4 |
}

func Example_respRelations_json() {
	display.Print(&respRelations{
		Data: &types.QueryResult{
			ColumnNames: []string{"a", "b"},
			ColumnTypes: []*types.DataType{types.TextType, types.TextType},
			Values:      [][]any{{"1", "2"}, {"3", "4"}},
		},
	},
		nil, "json")
	// Output:
	// {
	//   "result": {
	//     "column_names": [
	//       "a",
	//       "b"
	//     ],
	//     "column_types": [
	//       {
	//         "name": "text",
	//         "is_array": false,
	//         "metadata": [
	//           0,
	//           0
	//         ]
	//       },
	//       {
	//         "name": "text",
	//         "is_array": false,
	//         "metadata": [
	//           0,
	//           0
	//         ]
	//       }
	//     ],
	//     "values": [
	//       [
	//         "1",
	//         "2"
	//       ],
	//       [
	//         "3",
	//         "4"
	//       ]
	//     ]
	//   },
	//   "error": ""
	// }
}
