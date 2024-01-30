package cost

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// testSchemaUsers is the same as first line of ./testdata/users.csv
var testSchemaUsers = Schema(
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

// testDataUsers is the same as ./testdata/users.csv
var testDataUsers = []row{
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

func checkRecords(t *testing.T, result *record, expectedSchema *schema, expectedData []row) {
	t.Helper()

	s := result.Schema()
	assert.EqualValues(t, expectedSchema, s)

	idx := 0
	for r := range result.rows {
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

func TestMemDataSource_scanWithProjection(t *testing.T) {
	ds := NewMemDataSource(testSchemaUsers, testDataUsers)

	// Test filtered result
	expectedSchema := Schema(
		Field{
			Name: "username",
			Type: "string",
		},
		Field{
			Name: "age",
			Type: "int",
		})
	expectedData := []row{
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

func TestCSVDataSource(t *testing.T) {
	dataFilePath := "./testdata/users.csv"
	ds, err := NewCSVDataSource(dataFilePath)
	assert.NoError(t, err)

	checkRecords(t, ds.Scan(), testSchemaUsers, testDataUsers)
}

func TestCSVDataSource_scanWithProjection(t *testing.T) {
	dataFilePath := "./testdata/users.csv"
	ds, err := NewCSVDataSource(dataFilePath)
	assert.NoError(t, err)

	// Test filtered result
	expectedSchema := Schema(
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
	expectedData := []row{
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
