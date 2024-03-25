package datasource

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/kwilteam/kwil-db/internal/engine/cost/datatypes"
)

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

type Row []ColumnValue

func (r Row) String() string {
	var cols []string
	for _, c := range r {
		cols = append(cols, fmt.Sprintf("%v", c.Value()))
	}
	return fmt.Sprintf("[%s]", strings.Join(cols, ", "))
}

type RowPipeline chan Row

func newRowPipeline(rows []Row) RowPipeline {
	out := make(RowPipeline)
	go func() {
		defer close(out)

		for _, r := range rows {
			out <- r
		}
	}()
	return out

}

type Result struct {
	schema *datatypes.Schema
	stream RowPipeline
}

func (r *Result) ToCsv() string {
	var sb strings.Builder
	for _, f := range r.schema.Fields {
		sb.WriteString(fmt.Sprintf("%s", f.Name))
		if f != r.schema.Fields[len(r.schema.Fields)-1] {
			sb.WriteString(",")
		}
	}

	sb.WriteString("\n")

	for {
		row, ok := <-r.stream
		if !ok {
			break
		}
		for i, col := range row {
			sb.WriteString(fmt.Sprintf("%v", col.Value()))
			if i < len(row)-1 {
				sb.WriteString(",")
			}
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

func ResultFromStream(s *datatypes.Schema, rows RowPipeline) *Result {
	return &Result{schema: s, stream: rows}
}

func ResultFromRaw(s *datatypes.Schema, rows []Row) *Result {
	// TODO: use RowPipeline all the way
	return &Result{schema: s, stream: newRowPipeline(rows)}
}

func (r *Result) Schema() *datatypes.Schema {
	return r.schema
}

func (r *Result) Next() (Row, bool) {
	row, ok := <-r.stream
	return row, ok
}

type SourceType string

// DataSource represents a data source.
type DataSource interface {
	SchemaSource

	// SourceType returns the type of the data source.
	SourceType() SourceType

	// Scan scans the data source, return selected columns.
	// If projection field is not found in the schema, it will be ignored.
	// NOTE: should panic?
	Scan(projection ...string) *Result

	// TODO
	// Should this in DataSource?
	// Statistics returns the statistics of the data source.
	//Statistics() *Statistics
}

// dsScan read the data source, return selected columns.
// TODO: use channel to return the result, e.g. iterator model.
func dsScan(dsSchema *datatypes.Schema, dsRecords []Row, projection []string) *Result {
	if len(projection) == 0 {
		return ResultFromRaw(dsSchema, dsRecords)
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

	fieldIndex := dsSchema.MapProjection(projection)
	newSchema := dsSchema.Project(projection...)
	//newFields := make([]datatypes.Field, len(projection))
	//for i, idx := range fieldIndex {
	//	newFields[i] = dsSchema.Fields[idx]
	//}
	//newSchema := datatypes.NewSchema(newFields...)

	out := make(RowPipeline)
	go func() {
		defer close(out)

		for _, _row := range dsRecords {
			newRow := make(Row, len(projection))
			for j, idx := range fieldIndex {
				newRow[j] = _row[idx]
			}
			out <- newRow
		}
	}()

	return ResultFromStream(newSchema, out)
}

// memDataSource is a data source that reads data from memory.
type memDataSource struct {
	schema  *datatypes.Schema
	records []Row
}

func NewMemDataSource(s *datatypes.Schema, data []Row) *memDataSource {
	return &memDataSource{schema: s, records: data}
}

func (ds *memDataSource) Schema() *datatypes.Schema {
	return ds.schema
}

func (ds *memDataSource) Scan(projection ...string) *Result {
	return dsScan(ds.schema, ds.records, projection)
}

func (ds *memDataSource) SourceType() SourceType {
	return "memory"
}

// csvDataSource is a data source that reads data from a CSV file.
type csvDataSource struct {
	path    string
	records []Row
	schema  *datatypes.Schema
}

func NewCSVDataSource(path string) (*csvDataSource, error) {
	ds := &csvDataSource{path: path, schema: &datatypes.Schema{}}
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
			newRow[i] = &LiteralColumnValue{colValue}
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

func (ds *csvDataSource) Schema() *datatypes.Schema {
	return ds.schema
}

func (ds *csvDataSource) Scan(projection ...string) *Result {
	return dsScan(ds.schema, ds.records, projection)
}

func (ds *csvDataSource) SourceType() SourceType {
	return "csv"
}

type DefaultSchemaSource struct {
	datasource DataSource
}

func (s *DefaultSchemaSource) Schema() *datatypes.Schema {
	return s.datasource.Schema()
}

func (s *DefaultSchemaSource) Scan(projection ...string) *Result {
	return s.datasource.Scan(projection...)
}

func DataAsSchemaSource(ds DataSource) SchemaSource {
	return &DefaultSchemaSource{datasource: ds}
}
