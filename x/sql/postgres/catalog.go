package postgres

import (
	"kwil/x/schemadef/schema"
	"kwil/x/sql/ast"
	"kwil/x/sql/catalog"
)

func NewCatalog() *catalog.Catalog {
	c := catalog.New("public")
	c.Schemas = append(c.Schemas, pgTemp())
	c.Schemas = append(c.Schemas, genPGCatalog())
	c.Schemas = append(c.Schemas, genInformationSchema())
	c.SearchPath = []string{"pg_catalog"}
	c.LoadExtension = loadExtension
	return c
}

// The generated pg_catalog is very slow to compare because it has so
// many entries. For testing, don't include it.
func newTestCatalog() *catalog.Catalog {
	c := catalog.New("public")
	c.Schemas = append(c.Schemas, pgTemp())
	c.LoadExtension = loadExtension
	return c
}

type catalogConverter struct{}

func (c *catalogConverter) ConvertColumn(tab *schema.Table, col *schema.Column) (*catalog.Column, error) {
	var com schema.Comment
	schema.Has(col.Attrs, &com)

	_, isArray := arrayType(col.Type.Raw)
	tc := &catalog.Column{
		Name:      col.Name,
		Type:      ast.TypeName{Name: col.Type.Raw},
		IsNotNull: !col.Type.Nullable,
		IsArray:   isArray,
		Comment:   com.Text,
		Length:    nil,
	}

	return tc, nil
}
