package algebra_test

import (
	"fmt"

	"testing"

	"github.com/kwilteam/kwil-db/internal/engine/algebra"
)

func TestScan(t *testing.T) {
	testSchema := algebra.Schema(
		algebra.Field{
			Name: "id",
			Type: "int",
		},
		algebra.Field{
			Name: "name",
			Type: "string",
		},
		algebra.Field{
			Name: "age",
			Type: "int",
		})

	testData := [][]any{
		{1, "John", 20},
		{2, "Doe", 30},
		{3, "Jane", 25},
		{4, "Wu", 30},
	}
	dataSource := algebra.NewMemDataSource(testSchema, testData)
	plan := algebra.Scan(dataSource)


	fmt.Println(algebra.Format(plan, 0))
}
