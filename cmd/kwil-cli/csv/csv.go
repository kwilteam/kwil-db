package csv

import (
	"encoding/csv"
	"io"
	"os"
	"strings"
)

type CSVReaderFlag uint8

const (
	ContainsHeader CSVReaderFlag = 1 << iota
)

type CSV struct {
	Header  []string
	Records [][]string
}

func Read(csvFile *os.File, flags CSVReaderFlag) (*CSV, error) {
	reader := csv.NewReader(csvFile)
	reader.LazyQuotes = true
	csvStruct := &CSV{
		Records: [][]string{},
	}

	if flags&ContainsHeader == ContainsHeader {
		err := csvStruct.readAndTrimHeader(reader)
		if err != nil {
			return nil, err
		}
	}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		csvStruct.Records = append(csvStruct.Records, record)
	}

	return csvStruct, nil
}

func (c *CSV) readAndTrimHeader(reader *csv.Reader) error {
	header, err := reader.Read()
	if err != nil {
		return err
	}

	for i, singleHeader := range header {
		header[i] = cleanHeader(singleHeader)
	}

	c.Header = header

	return nil
}

// some headers are formatted as "\ufeff\"Date\"" instead of "Date"
// this function will remove the leading \ufeff
func cleanHeader(header string) string {
	str := header
	if strings.HasPrefix(header, "\ufeff") {
		str = strings.Replace(header, "\ufeff", "", 1)
	}

	// remove leading and trailing quotes
	str = strings.Trim(str, "\"")

	return str
}

// BuildInputs is the same as BuildInputs, but it takes a schema to ensure that the input
// is valid. If the input is invalid, it will return an error.
// The function takes an action, as well as a map mapping the CSV column name to the
// action input name.
func (c *CSV) BuildInputs(inputNames map[string]string) ([]map[string]string, error) {
	resultMap := make([]map[string]string, 0)
	err := c.ForEachRecord(func(record []CSVCell) error {
		input := make(map[string]string)

		for _, cell := range record {
			inputName, ok := inputNames[*cell.Column]
			if !ok {
				continue
			}

			input[inputName] = cell.Value
		}

		resultMap = append(resultMap, input)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return resultMap, nil
}

// GetColumnIndex returns the index of the column in the CSV. If the column doesn't exist,
// it will return -1.
func (c *CSV) GetColumnIndex(column string) int {
	for i, col := range c.Header {
		if col == column {
			return i
		}
	}
	return -1
}

// ForEachRecord will loop through each record in the CSV and call the function with the
// record as a map of the CSV column name to value.
func (c *CSV) ForEachRecord(fn func([]CSVCell) error) error {
	var err error
	for _, record := range c.Records {
		err = fn(c.buildCSVCells(record))
		if err != nil {
			return err
		}
	}

	return nil
}

type CSVCell struct {
	Column *string
	Value  string
}

// buildCSVCells will build a map of the CSV column name to value.
// The values are serialized strings.
// If for some reason it fails to serialize to string, it will panic.
func (c *CSV) buildCSVCells(record []string) []CSVCell {
	csvVals := make([]CSVCell, len(record))
	for i, column := range record {

		csvVals[i] = CSVCell{
			Column: &c.Header[i],
			Value:  column,
		}
	}

	return csvVals
}
