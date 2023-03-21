package ddlbuilder

import (
	"kwil/pkg/engine/types"
	"strings"
)

type index struct {
	name      string
	table     string
	columns   []string
	indexType types.IndexType
}

type indexBuilder struct {
	index *index
}

func NewIndexBuilder() indexNamePicker {
	return &indexBuilder{
		index: &index{},
	}
}

type indexNamePicker interface {
	Name(string) indexTablePicker
}

type indexTablePicker interface {
	Table(string) indexColumnPicker
}

type indexColumnPicker interface {
	Columns(...string) indexTypePicker
}

type indexTypePicker interface {
	Type(using types.IndexType) builder
}

func (b *indexBuilder) Name(name string) indexTablePicker {
	b.index.name = name
	return b
}

func (b *indexBuilder) Table(table string) indexColumnPicker {
	b.index.table = table
	return b
}

func (b *indexBuilder) Columns(columns ...string) indexTypePicker {
	b.index.columns = columns
	return b
}

func (b *indexBuilder) Type(using types.IndexType) builder {
	b.index.indexType = using
	return b.index
}

func (b *index) Build() string {
	sb := &strings.Builder{}
	sb.WriteString("CREATE ")
	if b.indexType == types.UNIQUE_BTREE {
		sb.WriteString("UNIQUE")
	}
	sb.WriteString(" INDEX IF NOT EXISTS ")
	sb.WriteString(b.name)
	sb.WriteString(" ON ")
	sb.WriteString(b.table)
	sb.WriteString(" (")
	sb.WriteString(strings.Join(b.columns, ", "))
	sb.WriteString(");")

	return sb.String()
}
