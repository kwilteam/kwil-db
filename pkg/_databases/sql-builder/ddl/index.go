package ddlbuilder

import (
	"github.com/kwilteam/kwil-db/pkg/databases/spec"
	"strings"
)

type index struct {
	name    string
	schema  string
	table   string
	columns []string
	using   spec.IndexType
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
	Name(string) indexSchemaPicker
}

type indexTablePicker interface {
	Table(string) indexColumnPicker
}

type indexColumnPicker interface {
	Columns(...string) indexUsingPicker
}

type indexUsingPicker interface {
	Using(using spec.IndexType) builder
}

type indexSchemaPicker interface {
	Schema(string) indexTablePicker
	Table(string) indexColumnPicker
}

func (b *indexBuilder) Name(name string) indexSchemaPicker {
	b.index.name = name
	return b
}

func (b *indexBuilder) Schema(schema string) indexTablePicker {
	b.index.schema = schema
	return b
}

func (b *indexBuilder) Table(table string) indexColumnPicker {
	b.index.table = table
	return b
}

func (b *indexBuilder) Columns(columns ...string) indexUsingPicker {
	b.index.columns = columns
	return b
}

func (b *indexBuilder) Using(using spec.IndexType) builder {
	b.index.using = using
	return b.index
}

func (b *index) Build() string {
	sb := &strings.Builder{}
	sb.WriteString("CREATE INDEX ")
	sb.WriteString(b.name)
	sb.WriteString(" ON ")
	sb.WriteString(b.nameWithSchema())
	sb.WriteString(" USING ")
	sb.WriteString(b.using.String())
	sb.WriteString(" (")
	sb.WriteString(strings.Join(b.columns, ", "))
	sb.WriteString(");")

	return sb.String()
}

func (b *index) nameWithSchema() string {
	sb := &strings.Builder{}
	sb.WriteString(`"`)
	if b.schema != "" {
		sb.WriteString(b.schema)
		sb.WriteString(`"."`)
	}
	sb.WriteString(b.table)
	sb.WriteString(`"`)
	return sb.String()
}
