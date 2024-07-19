package datasource

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"slices"
	"unsafe"

	"github.com/kwilteam/kwil-db/internal/engine/cost/datatypes"
)

// CsvDataSource is a data source that reads data from a CSV file.
// NOTE: This is for internal use only, mostly for testing.
type CsvDataSource struct {
	path    string
	records []Row
	schema  *datatypes.Schema
	stat    *datatypes.Statistics
}

func NewCSVDataSource(path string) (*CsvDataSource, error) {
	ds := &CsvDataSource{path: path, schema: &datatypes.Schema{}}
	if err := ds.load(); err != nil {
		return nil, err
	}

	return ds, nil
}

func (ds *CsvDataSource) collectStats() {
	if ds.stat == nil {
		ds.stat = datatypes.NewEmptyStatistics()
	}

	type counter map[any]int
	distinct := make(map[string]counter)

	rowCount := int64(len(ds.records))

	for i, f := range ds.schema.Fields {
		colStat := &datatypes.ColumnStatistics{}

		for _, r := range ds.records {
			rv := r[i].Value()

			if f.Type == "string" {
				v := rv.(string)
				if colStat.Min == nil {
					colStat.Min = v
				}

				if colStat.Max == nil {
					colStat.Max = v
				}
				vMin := colStat.Min.(string)
				vMax := colStat.Max.(string)

				colStat.Min = min(v, vMin)
				colStat.Max = max(v, vMax)
			} else {
				v := rv.(int64)
				if colStat.Min == nil {
					colStat.Min = v
				}

				if colStat.Max == nil {
					colStat.Max = v
				}
				vMin := colStat.Min.(int64)
				vMax := colStat.Max.(int64)
				colStat.Min = min(v, vMin)
				colStat.Max = max(v, vMax)
			}

			if _, ok := distinct[f.Name]; !ok {
				distinct[f.Name] = make(counter)
			}

			distinct[f.Name][rv]++

			if rv == nil {
				colStat.NullCount++
			}

			// not exactly correct
			colStat.AvgSize += int64(unsafe.Sizeof(rv))
		}

		colStat.DistinctCount = int64(len(distinct[f.Name]))
		colStat.AvgSize = colStat.AvgSize / rowCount

		ds.stat.ColumnStatistics = append(ds.stat.ColumnStatistics, *colStat)
	}

	ds.stat.RowCount = rowCount
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
	columnTypesInferred := false

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
			if columnTypesInferred {
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

		columnTypesInferred = true
	}

	for i, name := range header {
		ds.schema.Fields = append(ds.schema.Fields,
			datatypes.Field{Name: name, Type: columnTypes[i]})
	}

	ds.records = slices.Clip(ds.records)
	ds.schema.Fields = slices.Clip(ds.schema.Fields)

	return nil
}

func (ds *CsvDataSource) Schema() *datatypes.Schema {
	return ds.schema
}

func (ds *CsvDataSource) Statistics() *datatypes.Statistics {
	if ds.stat == nil {
		ds.collectStats()
	}
	return ds.stat
}

func (ds *CsvDataSource) Scan(ctx context.Context, projection ...string) *Result {
	return ScanData(ctx, ds.schema, ds.records, projection)
}

func (ds *CsvDataSource) SourceType() SourceType {
	return "csv"
}
