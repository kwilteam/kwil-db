package common_test

import (
	"fmt"

	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/core/types"
)

func ExampleRow_String() {
	row := common.Row{
		ColumnNames: []string{"id", "name", "scores"},
		ColumnTypes: []*types.DataType{
			{Name: "INT"},
			{Name: "TEXT"},
			{Name: "INT", IsArray: true},
		},
		Values: []any{
			42,
			"Alice",
			[]int{95, 87, 91},
		},
	}
	for _, ty := range row.ColumnTypes {
		if err := ty.Clean(); err != nil {
			panic(err)
		}
	}

	fmt.Println(row.String())
	fmt.Println(row.TypeStrings())
	fmt.Println(row.ValueStrings())
	// Output:
	// id[int8]: 42, name[text]: Alice, scores[int8[]]: [95 87 91]
	// [int8 text int8[]]
	// [42 Alice [95 87 91]]
}
