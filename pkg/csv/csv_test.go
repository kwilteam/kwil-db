package csv_test

import (
	"github.com/kwilteam/kwil-db/pkg/csv"
	"github.com/kwilteam/kwil-db/pkg/engine/types"
	"os"
	"testing"
)

func Test_CSV(t *testing.T) {
	file, err := loadTestCSV(t)
	if err != nil {
		t.Fatal(err)
	}

	data, err := csv.Read(file, csv.ContainsHeader)
	if err != nil {
		t.Fatal(err)
	}

	if len(data.Header) != 3 {
		t.Fatal(`expected 3 columns, got: `, len(data.Header))
	}

	if data.ColumnTypes[0] != types.INT {
		t.Fatal(`expected first column to be an int, got: `, data.ColumnTypes[0])
	}

	if data.ColumnTypes[1] != types.TEXT {
		t.Fatal(`expected second column to be a string, got: `, data.ColumnTypes[1])
	}

	if data.ColumnTypes[2] != types.INT {
		t.Fatal(`expected third column to be an int, got: `, data.ColumnTypes[2])
	}
}

func Test_PrepareInputs(t *testing.T) {
	file, err := loadTestCSV(t)
	if err != nil {
		t.Fatal(err)
	}

	data, err := csv.Read(file, csv.ContainsHeader)
	if err != nil {
		t.Fatal(err)
	}

	inputNames := map[string]string{
		"id":        "$id",
		"full_name": "$name",
		"age":       "$age",
	}

	inputs, err := data.BuildInputs(inputNames)
	if err != nil {
		t.Fatal(err)
	}

	if len(inputs) != 100 {
		t.Fatal(`expected 100 records, got: `, len(inputs))
	}

	record := inputs[0]
	if len(record) != 3 {
		t.Fatal(`expected 3 columns, got: `, len(record))
	}

	row0col0 := types.NewFromSerialMust(record["$id"])
	row0col1 := types.NewFromSerialMust(record["$name"])
	row0col2 := types.NewFromSerialMust(record["$age"])

	if row0col0.String() != "1" {
		t.Fatal("expected row 0, column 0 to be 1, got: ", types.NewFromSerialMust(record[inputNames["$id"]]).String())
	}

	if row0col1.String() != "Theodore Berry" {
		t.Fatal("expected row 0, column 1 to be Theodore Berry, got: ", types.NewFromSerialMust(record[inputNames["$name"]]).String())
	}

	if row0col2.String() != "51" {
		t.Fatal("expected row 0, column 2 to be 51, got: ", types.NewFromSerialMust(record[inputNames["$age"]]).String())
	}
}

func loadTestCSV(t *testing.T) (*os.File, error) {
	path := "./test.csv"
	return os.Open(path)
}

func loadCornCSV(t *testing.T) (*os.File, error) {
	path := "./corn.csv"
	return os.Open(path)
}

func Test_ReadDoubleQuotes(t *testing.T) {
	file, err := loadCornCSV(t)
	if err != nil {
		t.Fatal(err)
	}

	data, err := csv.Read(file, csv.ContainsHeader)
	if err != nil {
		t.Fatal(err)
	}

	inputNames := map[string]string{
		"Date":  "$dt",
		"Price": "$value",
	}

	res, err := data.BuildInputs(inputNames)
	if err != nil {
		t.Fatal(err)
	}

	if len(res) != 23 {
		t.Fatal(`expected 23 records, got: `, len(res))
	}

	if len(res[0]) != 2 {
		t.Fatal(`expected 2 columns, got: `, len(res[0]))
	}
}
