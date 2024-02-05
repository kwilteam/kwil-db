package algebra

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
)

type ColumnValue interface {
	Type() string
	Value() any
}

type literalColumnValue struct {
	value any
}

func (c *literalColumnValue) Type() string {
	return fmt.Sprintf("%T", c.value)
}

func (c *literalColumnValue) Value() any {
	return c.value
}

type row []ColumnValue

type record struct {
	schema *schema
	rows   []row

	idx int
}

func Record(s *schema, rows []row) *record {
	return &record{schema: s, rows: rows}
}

func (r *record) Schema() *schema {
	return r.schema
}

func (r *record) Next() (row, bool) {
	if r.idx < len(r.rows) {
		row := r.rows[r.idx]
		r.idx++
		return row, true
	}
	return nil, false
}

func (r *record) Len() int {
	return len(r.rows)
}

func (r *record) Left() int {
	return len(r.rows) - r.idx
}

type DataSource interface {
	Schema() *schema
	Scan(projection []string) *record
}

//func filter(ds DataSource, projection []string) *record {
//	if len(projection) == 0 {
//		return Record(ds.Schema(), ds.records)
//	}
//
//	fieldIndexMap := make(map[string]int)
//	for i, field := range ds.schema.Fields {
//		fieldIndexMap[field.Name] = i
//	}
//
//	newFieldsIndex := make([]int, len(projection))
//	for i, name := range projection {
//		newFieldsIndex[i] = fieldIndexMap[name]
//	}
//
//	newFields := make([]Field, len(projection))
//	for i, idx := range newFieldsIndex {
//		newFields[i] = ds.schema.Fields[idx]
//	}
//
//	newschema := Schema(newFields...)
//
//	newRecords := make([]row, len(ds.records))
//	for i, _row := range ds.records {
//		newRow := make(row, len(projection))
//		for j, idx := range newFieldsIndex {
//			newRow[j] = _row[idx]
//		}
//		newRecords[i] = newRow
//	}
//
//	return Record(newschema, newRecords)
//}

type memDataSource struct {
	schema  *schema
	records []row
}

func NewMemDataSource(s *schema, data [][]any) *memDataSource {
	records := make([]row, len(data))
	for i, _row := range data {
		records[i] = make(row, len(_row))
		for j, col := range _row {
			records[i][j] = &literalColumnValue{col}
		}
	}

	return &memDataSource{schema: s, records: records}
}

func (ds *memDataSource) Schema() *schema {
	return ds.schema
}

func (ds *memDataSource) Scan(projection []string) *record {
	if len(projection) == 0 {
		return Record(ds.schema, ds.records)
	}

	fieldIndexMap := make(map[string]int)
	for i, field := range ds.schema.Fields {
		fieldIndexMap[field.Name] = i
	}

	newFieldsIndex := make([]int, len(projection))
	for i, name := range projection {
		newFieldsIndex[i] = fieldIndexMap[name]
	}

	newFields := make([]Field, len(projection))
	for i, idx := range newFieldsIndex {
		newFields[i] = ds.schema.Fields[idx]
	}

	newschema := Schema(newFields...)

	newRecords := make([]row, len(ds.records))
	for i, _row := range ds.records {
		newRow := make(row, len(projection))
		for j, idx := range newFieldsIndex {
			newRow[j] = _row[idx]
		}
		newRecords[i] = newRow
	}

	return Record(newschema, newRecords)
}

type csvDataSource struct {
	path    string
	records []row
	schema  *schema
}

func NewCSVDataSource(path string) (*csvDataSource, error) {
	ds := &csvDataSource{path: path, schema: &schema{}}
	if err := ds.load(); err != nil {
		return nil, err
	}

	return ds, nil
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

	for {
		newRow := make(row, len(header))
		_record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		for i, col := range _record {
			// TODO: infer the type of the column
			newRow[i] = &literalColumnValue{col}
		}

		ds.records = append(ds.records, newRow)
	}

	for _, name := range header {
		// type?
		ds.schema.Fields = append(ds.schema.Fields, Field{Name: name, Type: "string"})
	}

	return nil
}

func (ds *csvDataSource) Schema() *schema {
	return ds.schema
}

func (ds *csvDataSource) Scan(projection []string) *record {
	if len(projection) == 0 {
		return Record(ds.schema, ds.records)
	}

	fieldIndexMap := make(map[string]int)
	for i, field := range ds.schema.Fields {
		fieldIndexMap[field.Name] = i
	}

	newFieldsIndex := make([]int, len(projection))
	for i, name := range projection {
		newFieldsIndex[i] = fieldIndexMap[name]
	}

	newFields := make([]Field, len(projection))
	for i, idx := range newFieldsIndex {
		newFields[i] = ds.schema.Fields[idx]
	}

	newschema := Schema(newFields...)

	newRecords := make([]row, len(ds.records))
	for i, _row := range ds.records {
		newRow := make(row, len(projection))
		for j, idx := range newFieldsIndex {
			newRow[j] = _row[idx]
		}
		newRecords[i] = newRow
	}

	return Record(newschema, newRecords)
}
