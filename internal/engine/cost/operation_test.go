package cost_test

import (
	"fmt"

	"testing"
)

func TestScan(t *testing.T) {
	testSchema := cost.Schema(
		cost.Field{
			Name: "id",
			Type: "int",
		},
		cost.Field{
			Name: "name",
			Type: "string",
		},
		cost.Field{
			Name: "age",
			Type: "int",
		})

	testData := [][]any{
		{1, "John", 20},
		{2, "Doe", 30},
		{3, "Jane", 25},
		{4, "Wu", 30},
	}
	dataSource := cost.NewMemDataSource(testSchema, testData)
	plan := cost.Scan(dataSource)

	fmt.Println(cost.Format(plan, 0))
}
