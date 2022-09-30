package catalog

import (
	"kwil/x/sql/ast"
	"kwil/x/sql/sqlerr"
)

func (c *Catalog) createViewAST(stmt *ast.ViewStmt, colGen ColumnGenerator) error {
	cols, err := colGen.OutputColumns(stmt.Query)
	if err != nil {
		return err
	}

	catName := ""
	if stmt.View.Catalogname != nil {
		catName = *stmt.View.Catalogname
	}
	schemaName := ""
	if stmt.View.Schemaname != nil {
		schemaName = *stmt.View.Schemaname
	}

	tbl := Table{
		Rel: &ast.TableName{
			Catalog: catName,
			Schema:  schemaName,
			Name:    *stmt.View.Relname,
		},
		Columns: cols,
	}

	ns := tbl.Rel.Schema
	if ns == "" {
		ns = c.DefaultSchema
	}
	schema, err := c.GetSchema(ns)
	if err != nil {
		return err
	}
	_, existingIdx, err := schema.GetTable(tbl.Rel.Name)
	if err == nil && !stmt.Replace {
		return sqlerr.RelationExists(tbl.Rel.Name)
	}

	if stmt.Replace && err == nil {
		schema.Tables[existingIdx] = &tbl
	} else {
		schema.Tables = append(schema.Tables, &tbl)
	}

	return nil
}
