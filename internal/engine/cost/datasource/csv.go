package datasource

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"slices"

	"github.com/kwilteam/kwil-db/internal/engine/cost/datatypes"
)

// CsvDataSource is a data source that reads data from a CSV file.
type CsvDataSource struct {
	path    string
	records []Row
	schema  *datatypes.Schema
}

func NewCSVDataSource(path string) (*CsvDataSource, error) {
	ds := &CsvDataSource{path: path, schema: &datatypes.Schema{}}
	if err := ds.load(); err != nil {
		return nil, err
	}

	return ds, nil
}

func (ds *CsvDataSource) load() error {
	in, err := os.Open(ds.path)
	if err != nil {
		return err
	}

	r := csv.NewReader(in)

	header, err := r.Read()
	if err != nil {
		return err
	}

	columnTypes := make([]string, len(header))
	columnTypesInfered := false

	for {
		columns, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		newRow := make(Row, len(header))
		for i, col := range columns {
			colType, colValue := colTypeCast(col)
			if columnTypesInfered {
				// check if the column type is consistent
				if columnTypes[i] != colType {
					return fmt.Errorf("inconsistent column type at column %d, got %s, want %s",
						i, colType, columnTypes[i])
				}
			} else {
				// NOTE: use the first row of 'data' to infer column types
				columnTypes[i] = colType
			}
			newRow[i] = NewLiteralColumnValue(colValue)
		}

		ds.records = append(ds.records, newRow)

		columnTypesInfered = true
	}

	for i, name := range header {
		ds.schema.Fields = append(ds.schema.Fields,
			datatypes.Field{Name: name, Type: columnTypes[i]})
	}

	slices.Clip(ds.schema.Fields)

	return nil
}

func (ds *CsvDataSource) Schema() *datatypes.Schema {
	return ds.schema
}

func (ds *CsvDataSource) Statistics() *datatypes.Statistics {
	panic("not implemented")
}

func (ds *CsvDataSource) Scan(projection ...string) *Result {
	return dsScan(ds.schema, ds.records, projection)
}

func (ds *CsvDataSource) SourceType() SourceType {
	return "csv"
}
