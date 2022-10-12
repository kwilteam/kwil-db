package pgschema

import (
	"kwil/x/schemadef/sqlschema"
	"kwil/x/sql/catalog"
)

type catalogConverter struct{}

func (c *catalogConverter) ConvertColumn(tab *sqlschema.Table, col *sqlschema.Column) (*catalog.Column, error) {
	var com sqlschema.Comment
	sqlschema.Has(col.Attrs, &com)

	_, isArray := arrayType(col.Type.Raw)

	sqlschema.Has(col.Attrs, &com)

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
