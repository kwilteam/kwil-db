package ddlbuilder

import (
	"fmt"
	"strings"
)

type table struct {
	schema  string
	name    string
	columns []ColumnBuilder
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
			columns: []ColumnBuilder{},
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
	b.table.columns = append(b.table.columns, col)
	return b
}

func (b *tableBuilder) Build() ([]string, error) {
	var statements []string

	sb := &strings.Builder{}
	sb.WriteString(`CREATE TABLE "`)
	if b.table.schema != "" {
		sb.WriteString(b.table.schema)
		sb.WriteString(`"."`)
	}
	sb.WriteString(b.table.name)
	sb.WriteString(`" (`)

	cols := len(b.table.columns)
	if cols == 0 {
		return statements, fmt.Errorf("table %s has no columns", b.table.name)
	}
	i := 0

	// map to guarantee uniqueness of column names
	colNames := make(map[string]struct{})
	for _, col := range b.table.columns {
		// check for duplicate column names
		if _, ok := colNames[col.GetName()]; ok {
			return statements, fmt.Errorf("duplicate column name %s", col.GetName())
		}
		colNames[col.GetName()] = struct{}{}

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
