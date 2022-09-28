package catalog

import (
	"kwil/x/sql/ast"
	"kwil/x/sql/sqlerr"
)

func (c *Catalog) commentOnColumnAST(stmt *ast.CommentOnColumnStmt) error {
	_, t, err := c.getTable(stmt.Table)
	if err != nil {
		return err
	}
	for i := range t.Columns {
		if t.Columns[i].Name == stmt.Col.Name {
			if stmt.Comment != nil {
				t.Columns[i].Comment = *stmt.Comment
			} else {
				t.Columns[i].Comment = ""
			}
			return nil
		}
	}
	return sqlerr.ColumnNotFound(stmt.Table.Name, stmt.Col.Name)
}

func (c *Catalog) commentOnSchemaAST(stmt *ast.CommentOnSchemaStmt) error {
	s, err := c.GetSchema(stmt.Schema.Str)
	if err != nil {
		return err
	}
	if stmt.Comment != nil {
		s.Comment = *stmt.Comment
	} else {
		s.Comment = ""
	}
	return nil
}

func (c *Catalog) commentOnTableAST(stmt *ast.CommentOnTableStmt) error {
	_, t, err := c.getTable(stmt.Table)
	if err != nil {
		return err
	}
	if stmt.Comment != nil {
		t.Comment = *stmt.Comment
	} else {
		t.Comment = ""
	}
	return nil
}

func (c *Catalog) commentOnTypeAST(stmt *ast.CommentOnTypeStmt) error {
	t, _, err := c.getType(stmt.Type)
	if err != nil {
		return err
	}
	if stmt.Comment != nil {
		t.SetComment(*stmt.Comment)
	} else {
		t.SetComment("")
	}
	return nil
}
