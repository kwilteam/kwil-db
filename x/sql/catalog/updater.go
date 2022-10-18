package catalog

import (
	"fmt"
	"kwil/x/schemadef/sqlschema"
	"kwil/x/sql/sqlerr"
)

type updater struct {
	*Catalog
}

func NewUpdater(c *Catalog) Updater {
	return &updater{Catalog: c}
}

func (c *updater) UpdateSchema(s *sqlschema.Schema, conv ColumnConverter) error {
	if _, ok := c.Schema(s.Name); !ok {
		c.Schemas = append(c.Schemas, &Schema{Name: s.Name})
	}

	for _, t := range s.Tables {
		if err := c.addTable(t, conv); err != nil {
			return err
		}
	}

	return nil
}

func (c *updater) addTable(tab *sqlschema.Table, conv ColumnConverter) error {
	ns := ""
	if tab.Schema != nil {
		ns = tab.Schema.Name
	}
	if ns == "" {
		ns = c.DefaultSchema
	}
	sch, ok := c.Schema(ns)
	if !ok {
		return fmt.Errorf("schema %q not found", ns)
	}
	if _, ok = sch.Table(tab.Name); ok {
		return sqlerr.RelationExists(tab.Name)
	}

	var com sqlschema.Comment
	sqlschema.Has(tab.Attrs, &com)

	tbl := Table{QualName: &QualName{Schema: ns, Name: tab.Name}, Comment: com.Text}
	com.Text = ""

	for _, col := range tab.Columns {
		c, err := conv.ConvertColumn(tab, col)
		if err != nil {
			return err
		}
		tbl.Columns = append(tbl.Columns, c)
	}
	sch.Tables = append(sch.Tables, &tbl)
	return nil
}
