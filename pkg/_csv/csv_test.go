package csv_test

import (
	"fmt"
	"kwil/pkg/csv"
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

	fmt.Println(data)
}

func loadTestCSV(t *testing.T) (*os.File, error) {
	path := "./test.csv"
	return os.Open(path)
}
