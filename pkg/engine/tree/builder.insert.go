package tree

import (
	"fmt"

	"github.com/doug-martin/goqu/v9"
)

type insertBuilder struct {
	insertDS *goqu.InsertDataset
	table    string
}

func (b *builder) InsertInto(table string) *insertBuilder {
	return &insertBuilder{
		table:    table,
		insertDS: b.dialect.Insert(table),
	}
}

func (b *insertBuilder) As(alias string) *insertBuilder {
	b.insertDS = b.insertDS.Into(fmt.Sprintf(`%s" AS "%s`, b.table, alias)) // goqu insert aliasing does not work, so this is a workaround!
	return b
}

func (i *insertBuilder) Columns(columns []any) *insertBuilder {
	i.insertDS = i.insertDS.Cols(columns...)
	return i
}

func (i *insertBuilder) Values(values ...InsertExpression) *insertBuilder {
	vals := make([]any, len(values))
	for i, value := range values {
		vals[i] = value.ToSqlStruct()
	}
	i.insertDS = i.insertDS.Vals(vals)
	return i
}

func (i *insertBuilder) WithUpsert(upsert *Upsert) *insertBuilder {
	i.insertDS = i.insertDS.OnConflict(upsert.toGoqu())
	return i
}

func (i *insertBuilder) ToSql() (string, []any, error) {
	return i.insertDS.ToSQL()
}
