package catalog

import (
	"kwil/x/sql/ast"
)

func (c *Catalog) createExtensionAST(stmt *ast.CreateExtensionStmt) error {
	if stmt.Extname == nil {
		return nil
	}
	// TODO: Implement IF NOT EXISTS
	if _, exists := c.extensions[*stmt.Extname]; exists {
		return nil
	}
	if c.LoadExtension == nil {
		return nil
	}
	ext := c.LoadExtension(*stmt.Extname)
	if ext == nil {
		return nil
	}
	s, err := c.GetSchema(c.DefaultSchema)
	if err != nil {
		return err
	}
	// TODO: Error on duplicate functions
	s.Funcs = append(s.Funcs, ext.Funcs...)
	return nil
}
