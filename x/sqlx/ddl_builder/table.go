package ddlbuilder

import (
	"fmt"
	"strings"
)

type table struct {
	schema  string
	name    string
	columns map[string]ColumnBuilder // using a map to ensure uniqueness
}

type tableBuilder struct {
	table *table
}

type TableBuilder interface {
	Build() ([]string, error)
	AddColumn(ColumnBuilder) TableBuilder
}

type tableSchemaPicker interface {
	Schema(string) tableNamePicker
	Name(string) TableBuilder
}

type tableNamePicker interface {
	Name(string) TableBuilder
}

func NewTableBuilder() tableSchemaPicker {
	return &tableBuilder{
		table: &table{
			columns: make(map[string]ColumnBuilder),
		},
	}
}

func (b *tableBuilder) Schema(schema string) tableNamePicker {
	b.table.schema = schema
	return b
}

func (b *tableBuilder) Name(name string) TableBuilder {
	b.table.name = name
	return b
}

func (b *tableBuilder) AddColumn(col ColumnBuilder) TableBuilder {
	b.table.columns[col.GetName()] = col
	return b
}

func (b *tableBuilder) Build() ([]string, error) {
	var statements []string

	sb := &strings.Builder{}
	sb.WriteString("CREATE TABLE ")
	if b.table.schema != "" {
		sb.WriteString(b.table.schema)
		sb.WriteString(".")
	}
	sb.WriteString(b.table.name)
	sb.WriteString(" (")

	cols := len(b.table.columns)
	if cols == 0 {
		return statements, fmt.Errorf("table %s has no columns", b.table.name)
	}
	i := 0
	for _, col := range b.table.columns {
		column := col.Build()
		sb.WriteString(column)
		if i < cols-1 {
			sb.WriteString(", ")
		}
		i++

		attrs, err := col.BuildAttributes(b.table.schema, b.table.name)
		if err != nil {
			return statements, err
		}
		statements = append(statements, attrs...)
	}
	sb.WriteString(");")
	ret := append([]string{sb.String()}, statements...)

	return ret, nil
}
