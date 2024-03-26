package datasource

import (
	"fmt"
	"strings"

	"github.com/kwilteam/kwil-db/internal/engine/cost/datatypes"
)

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
