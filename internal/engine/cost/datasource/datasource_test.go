package datasource

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// testSchemaUsers is the same as first line of ../../testdata/users.csv
var testSchemaUsers = NewSchema(
	Field{
		Name: "id",
		Type: "int",
	},
	Field{
		Name: "username",
		Type: "string",
	},
	Field{
		Name: "age",
		Type: "int",
	},
	Field{
		Name: "state",
		Type: "string",
	},
	Field{
		Name: "wallet",
		Type: "string",
	},
)

// testDataUsers is the same as ../../testdata/users.csv
var testDataUsers = []Row{
	{
		NewLiteralColumnValue(1),
		NewLiteralColumnValue("Adam"),
		NewLiteralColumnValue(20),
		NewLiteralColumnValue("CA"),
		NewLiteralColumnValue("x001"),
	},
	{
		NewLiteralColumnValue(2),
		NewLiteralColumnValue("Bob"),
		NewLiteralColumnValue(24),
		NewLiteralColumnValue("CA"),
		NewLiteralColumnValue("x002"),
	},
	{
		NewLiteralColumnValue(3),
		NewLiteralColumnValue("Cat"),
		NewLiteralColumnValue(27),
		NewLiteralColumnValue("CA"),
		NewLiteralColumnValue("x003"),
	},
	{
		NewLiteralColumnValue(4),
		NewLiteralColumnValue("Doe"),
		NewLiteralColumnValue(26),
		NewLiteralColumnValue("IL"),
		NewLiteralColumnValue("x004"),
	},
	{
		NewLiteralColumnValue(5),
		NewLiteralColumnValue("Eve"),
		NewLiteralColumnValue(29),
		NewLiteralColumnValue("TX"),
		NewLiteralColumnValue("x005"),
	},
}

func checkRecords(t *testing.T, result *Result, expectedSchema *Schema, expectedData []Row) {
	t.Helper()

	s := result.Schema()
	assert.EqualValues(t, expectedSchema, s)

	idx := 0
	for r := range result.stream {
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
	result := ds.Scan()

	checkRecords(t, result, testSchemaUsers, testDataUsers)
}

func Example_MemDataSource_ToCsv() {
	ds := NewMemDataSource(testSchemaUsers, testDataUsers)
	result := ds.Scan()
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
	expectedSchema := NewSchema(
		Field{
			Name: "username",
			Type: "string",
		},
		Field{
			Name: "age",
			Type: "int",
		})
	expectedData := []Row{
		{
			NewLiteralColumnValue("Adam"),
			NewLiteralColumnValue(20),
		},
		{
			NewLiteralColumnValue("Bob"),
			NewLiteralColumnValue(24),
		},
		{
			NewLiteralColumnValue("Cat"),
			NewLiteralColumnValue(27),
		},
		{
			NewLiteralColumnValue("Doe"),
			NewLiteralColumnValue(26),
		},
		{
			NewLiteralColumnValue("Eve"),
			NewLiteralColumnValue(29),
		},
	}

	filteredResult := ds.Scan("username", "age")

	checkRecords(t, filteredResult, expectedSchema, expectedData)
}

func Example_MemDataSource_scanWithProjection_ToCsv() {
	ds := NewMemDataSource(testSchemaUsers, testDataUsers)
	result := ds.Scan("username", "age")
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

	checkRecords(t, ds.Scan(), testSchemaUsers, testDataUsers)
}

func Example_CSVDataSource_ToCsv() {
	dataFilePath := "../testdata/users.csv"
	ds, _ := NewCSVDataSource(dataFilePath)
	result := ds.Scan()
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
	expectedSchema := NewSchema(
		Field{
			Name: "id",
			Type: "int",
		},
		Field{
			Name: "username",
			Type: "string",
		},
		Field{
			Name: "state",
			Type: "string",
		},
	)
	expectedData := []Row{
		{
			NewLiteralColumnValue(1),
			NewLiteralColumnValue("Adam"),
			NewLiteralColumnValue("CA"),
		},
		{
			NewLiteralColumnValue(2),
			NewLiteralColumnValue("Bob"),
			NewLiteralColumnValue("CA"),
		},
		{
			NewLiteralColumnValue(3),
			NewLiteralColumnValue("Cat"),
			NewLiteralColumnValue("CA"),
		},
		{
			NewLiteralColumnValue(4),
			NewLiteralColumnValue("Doe"),
			NewLiteralColumnValue("IL"),
		},
		{
			NewLiteralColumnValue(5),
			NewLiteralColumnValue("Eve"),
			NewLiteralColumnValue("TX"),
		},
	}

	filteredResult := ds.Scan("id", "username", "state")
	checkRecords(t, filteredResult, expectedSchema, expectedData)
}

func Example_CSVDataSource_scanWithProjection_ToCsv() {
	dataFilePath := "../testdata/users.csv"
	ds, _ := NewCSVDataSource(dataFilePath)
	result := ds.Scan("id", "username", "state")
	fmt.Println(result.ToCsv())
	//Output:
	// id,username,state
	// 1,Adam,CA
	// 2,Bob,CA
	// 3,Cat,CA
	// 4,Doe,IL
	// 5,Eve,TX
}
