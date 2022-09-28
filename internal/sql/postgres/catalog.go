package postgres

import (
	"github.com/kwilteam/kwil-db/internal/schemadef/schema"
	"github.com/kwilteam/kwil-db/internal/sql/ast"
	"github.com/kwilteam/kwil-db/internal/sql/catalog"
	"github.com/kwilteam/kwil-db/internal/sql/sqlerr"
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

func NewCatalogUpdater(c *catalog.Catalog) catalog.Updater {
	return &catalogUpdater{catalog: c}
}

type catalogUpdater struct {
	catalog *catalog.Catalog
}

func (c *catalogUpdater) AddSchema(s *schema.Schema) error {
	if _, err := c.catalog.GetSchema(s.Name); err != nil {
		c.catalog.Schemas = append(c.catalog.Schemas, &catalog.Schema{Name: s.Name})
	}

	for _, t := range s.Tables {
		if err := c.addTable(t); err != nil {
			return err
		}
	}

	return nil
}

func (c *catalogUpdater) addTable(tab *schema.Table) error {
	ns := ""
	if tab.Schema != nil {
		ns = tab.Schema.Name
	}
	if ns == "" {
		ns = c.catalog.DefaultSchema
	}
	sch, err := c.catalog.GetSchema(ns)
	if err != nil {
		return err
	}
	_, _, err = sch.GetTable(tab.Name)
	if err == nil {
		return sqlerr.RelationExists(tab.Name)
	}

	var com schema.Comment
	schema.Has(tab.Attrs, &com)

	tbl := catalog.Table{Rel: &ast.TableName{Schema: ns, Name: tab.Name}, Comment: com.Text}
	com.Text = ""

	for _, col := range tab.Columns {
		c, err := c.convertColumn(tab, col)
		if err != nil {
			return err
		}
		tbl.Columns = append(tbl.Columns, c)
	}
	sch.Tables = append(sch.Tables, &tbl)
	return nil
}

func (c *catalogUpdater) convertColumn(tab *schema.Table, col *schema.Column) (*catalog.Column, error) {
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
