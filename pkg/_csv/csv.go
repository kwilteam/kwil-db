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
	columnTypes []types.DataType
	containsHeader bool
}

func Read(csvFile *os.File, flags CSVReaderFlag) (*CSV, error) {
	reader := csv.NewReader(csvFile)

	csvStruct := &CSV{
		Records: [][]string{},
		containsHeader: flags&ContainsHeader == ContainsHeader,
	}

	if csvStruct.containsHeader {
		header, err := reader.Read()
		if err != nil {
			return nil, err
		}
		csvStruct.Header = header
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

func (c *CSV)

// determineSchema will determine whether a column should be a string or a number.
// It will loop through each column and try to convert it to a number. If it can't,
// it will assume it's a string.
func (c *CSV) determineSchema() error {
	colTypes := make([]types.DataType, len(c.Header))
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

	c.columnTypes = colTypes

	return nil
}
