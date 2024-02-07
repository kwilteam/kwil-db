package datasource

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
)

// Field represents a field in a schema.
type Field struct {
	Name string
	Type string
}

type Schema struct {
	Fields []Field
}

func NewSchema(fields ...Field) *Schema {
	return &Schema{Fields: fields}
}

// ColumnValue
type ColumnValue interface {
	Type() string
	Value() any
}

type LiteralColumnValue struct {
	value any
}

func (c *LiteralColumnValue) Type() string {
	return fmt.Sprintf("%T", c.value)
}

func (c *LiteralColumnValue) Value() any {
	return c.value
}

func NewLiteralColumnValue(v any) *LiteralColumnValue {
	return &LiteralColumnValue{value: v}
}

type row []ColumnValue
type rowIterator <-chan row

func newRowIterator(rows []row) rowIterator {
	ch := make(chan row)
	go func() {
		for _, r := range rows {
			ch <- r
		}
		close(ch)
	}()
	return ch

}

type record struct {
	schema *Schema
	rows   rowIterator

	idx int
}

func Record(s *Schema, rows []row) *record {
	// TODO: use rowIterator all the way
	return &record{schema: s, rows: newRowIterator(rows)}
}

func (r *record) Schema() *Schema {
	return r.schema
}

type DataSourceType string

type DataSource interface {
	// Schema returns the schema for the underlying data source
	Schema() *Schema
	// Scan scans the data source, return selected columns.
	// If projection field is not found in the schema, it will be ignored.
	// NOTE: should panic?
	Scan(projection ...string) *record
	// SourceType returns the type of the data source.
	SourceType() DataSourceType
	// TODO
	// Statistics returns the statistics of the data source.
	//Statistics() *Statistics
}

// dsScan read the data source, return selected columns.
// TODO: use channel to return the result, e.g. iterator model.
func dsScan(dsSchema *Schema, dsRecords []row, projection []string) *record {
	if len(projection) == 0 {
		return Record(dsSchema, dsRecords)
	}

	// panic if projection field not found
	//for _, name := range projection {
	//	found := false
	//	for _, field := range ds.schema.Fields {
	//		if field.Name == name {
	//			found = true
	//			break
	//		}
	//	}
	//	if !found {
	//		panic(fmt.Sprintf("projection field %s not found", name))
	//	}
	//}

	fieldIndexMap := make(map[string]int)
	for i, field := range dsSchema.Fields {
		fieldIndexMap[field.Name] = i
	}

	newFieldsIndex := make([]int, len(projection))
	for i, name := range projection {
		newFieldsIndex[i] = fieldIndexMap[name]
	}

	newFields := make([]Field, len(projection))
	for i, idx := range newFieldsIndex {
		newFields[i] = dsSchema.Fields[idx]
	}

	newschema := NewSchema(newFields...)

	newRecords := make([]row, len(dsRecords))
	for i, _row := range dsRecords {
		newRow := make(row, len(projection))
		for j, idx := range newFieldsIndex {
			newRow[j] = _row[idx]
		}
		newRecords[i] = newRow
	}

	return Record(newschema, newRecords)
}

// memDataSource is a data source that reads data from memory.
type memDataSource struct {
	schema  *Schema
	records []row
}

func NewMemDataSource(s *Schema, data []row) *memDataSource {
	return &memDataSource{schema: s, records: data}
}

func (ds *memDataSource) Schema() *Schema {
	return ds.schema
}

func (ds *memDataSource) Scan(projection ...string) *record {
	return dsScan(ds.schema, ds.records, projection)
}

func (ds *memDataSource) SourceType() DataSourceType {
	return "memory"
}

// csvDataSource is a data source that reads data from a CSV file.
type csvDataSource struct {
	path    string
	records []row
	schema  *Schema
}

func NewCSVDataSource(path string) (*csvDataSource, error) {
	ds := &csvDataSource{path: path, schema: &Schema{}}
	if err := ds.load(); err != nil {
		return nil, err
	}

	return ds, nil
}

// colTypeCast try to cast the raw column string to int, if failed, return the raw string.
// NOTE: we only support int/string for simplicity.
func colTypeCast(raw string) (kind string, value any) {
	v, err := strconv.Atoi(raw)
	if err == nil {
		return "int", v
	} else {
		return "string", raw
	}
}

func (ds *csvDataSource) load() error {
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

		newRow := make(row, len(header))
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
			newRow[i] = &LiteralColumnValue{colValue}
		}

		ds.records = append(ds.records, newRow)

		columnTypesInfered = true
	}

	for i, name := range header {
		ds.schema.Fields = append(ds.schema.Fields,
			Field{Name: name, Type: columnTypes[i]})
	}

	return nil
}

func (ds *csvDataSource) Schema() *Schema {
	return ds.schema
}

func (ds *csvDataSource) Scan(projection ...string) *record {
	return dsScan(ds.schema, ds.records, projection)
}

func (ds *csvDataSource) SourceType() DataSourceType {
	return "csv"
}
