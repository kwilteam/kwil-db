package dto

func NewTable() *TableBuilder {
	return &TableBuilder{}
}

type TableBuilder struct {
	table *Table
}

func (t *TableBuilder) Name(name string) *TableBuilder {
	t.table.Name = name
	return t
}

func (t *TableBuilder) Cols(cols ...*Column) *TableBuilder {
	t.table.Columns = cols
	return t
}

func (t *TableBuilder) Build() *Table {
	return t.table
}

func NewCol() *ColumnBuilder {
	return &ColumnBuilder{}
}

type ColumnBuilder struct {
	col *Column
}

func (c *ColumnBuilder) Name(name string) *ColumnBuilder {
	c.col.Name = name
	return c
}

func (c *ColumnBuilder) Type(dataType DataType) *ColumnBuilder {
	c.col.Type = dataType
	return c
}

func (c *ColumnBuilder) Attribute(typ AttributeType, val ...any) *ColumnBuilder {
	if len(val) > 1 {
		panic("too many values")
	}

	attr := &Attribute{
		Type: typ,
	}

	if len(val) == 1 {
		attr.Value = val[0]
	}

	c.col.Attributes = append(c.col.Attributes, attr)
	return c
}

func (c *ColumnBuilder) Build() *Column {
	return c.col
}
