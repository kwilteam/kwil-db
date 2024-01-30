package datasource

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kwilteam/kwil-db/internal/engine/cost/datatypes"
)

// testSchemaUsers is the same as first line of ../../testdata/users.csv
var testSchemaUsers = datatypes.NewSchema(
	datatypes.Field{
		Name: "id",
		Type: "int64",
	},
	datatypes.Field{
		Name: "username",
		Type: "string",
	},
	datatypes.Field{
		Name: "age",
		Type: "int64",
	},
	datatypes.Field{
		Name: "state",
		Type: "string",
	},
	datatypes.Field{
		Name: "wallet",
		Type: "string",
	},
)

// testDataUsers is the same as ../../testdata/users.csv
var testDataUsers = []Row{
	{
		NewLiteralColumnValue(int64(1)),
		NewLiteralColumnValue("Adam"),
		NewLiteralColumnValue(int64(20)),
		NewLiteralColumnValue("CA"),
		NewLiteralColumnValue("x001"),
	},
	{
		NewLiteralColumnValue(int64(2)),
		NewLiteralColumnValue("Bob"),
		NewLiteralColumnValue(int64(24)),
		NewLiteralColumnValue("CA"),
		NewLiteralColumnValue("x002"),
	},
	{
		NewLiteralColumnValue(int64(3)),
		NewLiteralColumnValue("Cat"),
		NewLiteralColumnValue(int64(27)),
		NewLiteralColumnValue("CA"),
		NewLiteralColumnValue("x003"),
	},
	{
		NewLiteralColumnValue(int64(4)),
		NewLiteralColumnValue("Doe"),
		NewLiteralColumnValue(int64(26)),
		NewLiteralColumnValue("IL"),
		NewLiteralColumnValue("x004"),
	},
	{
		NewLiteralColumnValue(int64(5)),
		NewLiteralColumnValue("Eve"),
		NewLiteralColumnValue(int64(29)),
		NewLiteralColumnValue("TX"),
		NewLiteralColumnValue("x005"),
	},
}

func checkRecords(t *testing.T, result *Result, expectedSchema *datatypes.Schema, expectedData []Row) {
	t.Helper()

	s := result.Schema
	assert.EqualValues(t, expectedSchema, s)

	idx := 0
	for r := range result.Stream {
		assert.Len(t, r, len(expectedSchema.Fields))

		for i, c := range r {
			assert.EqualValues(t, expectedData[idx][i].Value(), c.Value())
			assert.EqualValues(t, expectedData[idx][i].Type(), c.Type())
		}

		idx++
	}
}

func TestMemDataSource(t *testing.T) {
	ds := NewMemDataSource(testSchemaUsers, testDataUsers)
	result := ds.Scan(context.TODO())

	checkRecords(t, result, testSchemaUsers, testDataUsers)
}

func Example_MemDataSource_ToCsv() {
	ds := NewMemDataSource(testSchemaUsers, testDataUsers)
	result := ds.Scan(context.TODO())
	fmt.Println(result.ToCsv())
	//Output:
	// id,username,age,state,wallet
	// 1,Adam,20,CA,x001
	// 2,Bob,24,CA,x002
	// 3,Cat,27,CA,x003
	// 4,Doe,26,IL,x004
	// 5,Eve,29,TX,x005
}

func TestMemDataSource_scanWithProjection(t *testing.T) {
	ds := NewMemDataSource(testSchemaUsers, testDataUsers)

	// Test filtered result
	expectedSchema := datatypes.NewSchema(
		datatypes.Field{
			Name: "username",
			Type: "string",
		},
		datatypes.Field{
			Name: "age",
			Type: "int64",
		})
	expectedData := []Row{
		{
			NewLiteralColumnValue("Adam"),
			NewLiteralColumnValue(int64(20)),
		},
		{
			NewLiteralColumnValue("Bob"),
			NewLiteralColumnValue(int64(24)),
		},
		{
			NewLiteralColumnValue("Cat"),
			NewLiteralColumnValue(int64(27)),
		},
		{
			NewLiteralColumnValue("Doe"),
			NewLiteralColumnValue(int64(26)),
		},
		{
			NewLiteralColumnValue("Eve"),
			NewLiteralColumnValue(int64(29)),
		},
	}

	filteredResult := ds.Scan(context.TODO(), "username", "age")

	checkRecords(t, filteredResult, expectedSchema, expectedData)
}

func Example_MemDataSource_scanWithProjection_ToCsv() {
	ds := NewMemDataSource(testSchemaUsers, testDataUsers)
	result := ds.Scan(context.TODO(), "username", "age")
	fmt.Println(result.ToCsv())
	//Output:
	// username,age
	// Adam,20
	// Bob,24
	// Cat,27
	// Doe,26
	// Eve,29
}

func TestCSVDataSource(t *testing.T) {
	dataFilePath := "../testdata/users.csv"
	ds, err := NewCSVDataSource(dataFilePath)
	assert.NoError(t, err)

	checkRecords(t, ds.Scan(context.TODO()), testSchemaUsers, testDataUsers)
}

func Example_CSVDataSource_ToCsv() {
	dataFilePath := "../testdata/users.csv"
	ds, _ := NewCSVDataSource(dataFilePath)
	result := ds.Scan(context.TODO())
	fmt.Println(result.ToCsv())
	//Output:
	// id,username,age,state,wallet
	// 1,Adam,20,CA,x001
	// 2,Bob,24,CA,x002
	// 3,Cat,27,CA,x003
	// 4,Doe,26,IL,x004
	// 5,Eve,29,TX,x005
}

func TestCSVDataSource_scanWithProjection(t *testing.T) {
	dataFilePath := "../testdata/users.csv"
	ds, err := NewCSVDataSource(dataFilePath)
	assert.NoError(t, err)

	// Test filtered result
	expectedSchema := datatypes.NewSchema(
		datatypes.Field{
			Name: "id",
			Type: "int64",
		},
		datatypes.Field{
			Name: "username",
			Type: "string",
		},
		datatypes.Field{
			Name: "state",
			Type: "string",
		},
	)
	expectedData := []Row{
		{
			NewLiteralColumnValue(int64(1)),
			NewLiteralColumnValue("Adam"),
			NewLiteralColumnValue("CA"),
		},
		{
			NewLiteralColumnValue(int64(2)),
			NewLiteralColumnValue("Bob"),
			NewLiteralColumnValue("CA"),
		},
		{
			NewLiteralColumnValue(int64(3)),
			NewLiteralColumnValue("Cat"),
			NewLiteralColumnValue("CA"),
		},
		{
			NewLiteralColumnValue(int64(4)),
			NewLiteralColumnValue("Doe"),
			NewLiteralColumnValue("IL"),
		},
		{
			NewLiteralColumnValue(int64(5)),
			NewLiteralColumnValue("Eve"),
			NewLiteralColumnValue("TX"),
		},
	}

	filteredResult := ds.Scan(context.TODO(), "id", "username", "state")
	checkRecords(t, filteredResult, expectedSchema, expectedData)
}

func Example_CSVDataSource_scanWithProjection_ToCsv() {
	dataFilePath := "../testdata/users.csv"
	ds, _ := NewCSVDataSource(dataFilePath)
	result := ds.Scan(context.TODO(), "id", "username", "state")
	fmt.Println(result.ToCsv())
	//Output:
	// id,username,state
	// 1,Adam,CA
	// 2,Bob,CA
	// 3,Cat,CA
	// 4,Doe,IL
	// 5,Eve,TX
}
