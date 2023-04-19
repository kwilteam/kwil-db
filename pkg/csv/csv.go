package csv

import (
	"encoding/csv"
	"io"
	"kwil/pkg/engine/types"
	"os"
)

type CSVReaderFlag uint8

const (
	ContainsHeader CSVReaderFlag = 1 << iota
)

type CSV struct {
	Header      []string
	Records     [][]string
	ColumnTypes []types.DataType
}

func Read(csvFile *os.File, flags CSVReaderFlag) (*CSV, error) {
	reader := csv.NewReader(csvFile)

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

	err := csvStruct.determineSchema()
	if err != nil {
		return nil, err
	}

	return csvStruct, nil
}

func (c *CSV) readAndTrimHeader(reader *csv.Reader) error {
	header, err := reader.Read()
	if err != nil {
		return err
	}

	c.Header = header

	return nil
}

// determineSchema will determine whether a column should be a string or a number.
// It will loop through each column and try to convert it to a number. If it can't,
// it will assume it's a string.
func (c *CSV) determineSchema() error {
	colTypes := make([]types.DataType, len(c.Records[0]))
	for _, record := range c.Records {
		for i, column := range record {
			// if we've already determined the type is a string, skip
			if colTypes[i] == types.TEXT {
				continue
			}

			// try to convert to number
			_, err := types.INT.CoerceAny(column)
			if err != nil {
				// if we can't, assume string
				colTypes[i] = types.TEXT
				continue
			}

			// else, assume int
			colTypes[i] = types.INT
		}
	}

	c.ColumnTypes = colTypes

	return nil
}

// BuildInputs is the same as BuildInputs, but it takes a schema to ensure that the input
// is valid. If the input is invalid, it will return an error.
// The function takes an action, as well as a map mapping the CSV column name to the
// action input name.
func (c *CSV) BuildInputs(inputNames map[string]string) ([]map[string][]byte, error) {
	resultMap := make([]map[string][]byte, 0)
	err := c.ForEachRecord(func(record []*CSVCell) error {
		input := make(map[string][]byte)

		for _, cell := range record {
			inputName, ok := inputNames[*cell.Column]
			if !ok {
				continue
			}

			input[inputName] = *cell.Value
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
func (c *CSV) ForEachRecord(fn func([]*CSVCell) error) error {
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
	Value  *[]byte
}

// buildCSVCells will build a map of the CSV column name to value.
// The values are serialized strings.
// If for some reason it fails to serialize to string, it will panic.
func (c *CSV) buildCSVCells(record []string) []*CSVCell {
	csvVals := make([]*CSVCell, len(record))
	for i, column := range record {
		serializedValue := types.NewExplicitMust(column, types.TEXT).Bytes()

		csvVals[i] = &CSVCell{
			Column: &c.Header[i],
			Value:  &serializedValue,
		}
	}

	return csvVals
}
