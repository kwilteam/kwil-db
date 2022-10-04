package postgres

import (
	"kwil/x/schemadef/schema"
	"kwil/x/sql/catalog"
)

type catalogConverter struct{}

func (c *catalogConverter) ConvertColumn(tab *schema.Table, col *schema.Column) (*catalog.Column, error) {
	var com schema.Comment
	schema.Has(col.Attrs, &com)

	_, isArray := arrayType(col.Type.Raw)

	schema.Has(col.Attrs, &com)

	var length *int
	if l, ok := GetLength(col.Type.Type); ok {
		length = &l
	}

	tc := &catalog.Column{
		Name:      col.Name,
		Type:      &catalog.QualName{Name: col.Type.Raw},
		IsNotNull: !col.Type.Nullable,
		IsArray:   isArray,
		Comment:   com.Text,
		Length:    length,
	}

	return tc, nil
}
