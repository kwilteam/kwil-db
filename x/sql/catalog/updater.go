package catalog

import (
	"kwil/x/schemadef/schema"
	"kwil/x/sql/ast"
	"kwil/x/sql/sqlerr"
)

type Updater interface {
	UpdateSchema(*schema.Schema, ColumnConverter) error
	UpdateDDL(ast.Statement, ColumnGenerator) error
}

type ColumnConverter interface {
	ConvertColumn(tab *schema.Table, col *schema.Column) (*Column, error)
}

type updater struct {
	*Catalog
}

func NewUpdater(c *Catalog) Updater {
	return &updater{Catalog: c}
}

func (c *updater) UpdateSchema(s *schema.Schema, conv ColumnConverter) error {
	if _, err := c.GetSchema(s.Name); err != nil {
		c.Schemas = append(c.Schemas, &Schema{Name: s.Name})
	}

	for _, t := range s.Tables {
		if err := c.addTable(t, conv); err != nil {
			return err
		}
	}

	return nil
}

func (c *updater) addTable(tab *schema.Table, conv ColumnConverter) error {
	ns := ""
	if tab.Schema != nil {
		ns = tab.Schema.Name
	}
	if ns == "" {
		ns = c.DefaultSchema
	}
	sch, err := c.GetSchema(ns)
	if err != nil {
		return err
	}
	_, _, err = sch.GetTable(tab.Name)
	if err == nil {
		return sqlerr.RelationExists(tab.Name)
	}

	var com schema.Comment
	schema.Has(tab.Attrs, &com)

	tbl := Table{Rel: &ast.TableName{Schema: ns, Name: tab.Name}, Comment: com.Text}
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

func (c *updater) UpdateDDL(stmt ast.Statement, colGen ColumnGenerator) error {
	if stmt.Raw == nil {
		return nil
	}

	var err error
	switch n := stmt.Raw.Stmt.(type) {

	case *ast.AlterTableStmt:
		err = c.alterTableAST(n)

	case *ast.AlterTableSetSchemaStmt:
		err = c.alterTableSetSchemaAST(n)

	case *ast.AlterTypeAddValueStmt:
		err = c.alterTypeAddValueAST(n)

	case *ast.AlterTypeRenameValueStmt:
		err = c.alterTypeRenameValueAST(n)

	case *ast.CommentOnColumnStmt:
		err = c.commentOnColumnAST(n)

	case *ast.CommentOnSchemaStmt:
		err = c.commentOnSchemaAST(n)

	case *ast.CommentOnTableStmt:
		err = c.commentOnTableAST(n)

	case *ast.CommentOnTypeStmt:
		err = c.commentOnTypeAST(n)

	case *ast.CompositeTypeStmt:
		err = c.createCompositeTypeAST(n)

	case *ast.CreateEnumStmt:
		err = c.createEnumAST(n)

	case *ast.CreateExtensionStmt:
		err = c.createExtensionAST(n)

	case *ast.CreateFunctionStmt:
		err = c.createFunctionAST(n)

	case *ast.CreateSchemaStmt:
		err = c.createSchemaAST(n)

	case *ast.CreateTableStmt:
		err = c.createTableAST(n)

	case *ast.CreateTableAsStmt:
		err = c.createTableAsAST(n, colGen)

	case *ast.ViewStmt:
		err = c.createViewAST(n, colGen)

	case *ast.DropFunctionStmt:
		err = c.dropFunctionAST(n)

	case *ast.DropSchemaStmt:
		err = c.dropSchemaAST(n)

	case *ast.DropTableStmt:
		err = c.dropTableAST(n)

	case *ast.DropTypeStmt:
		err = c.dropTypeAST(n)

	case *ast.RenameColumnStmt:
		err = c.renameColumnAST(n)

	case *ast.RenameTableStmt:
		err = c.renameTableAST(n)

	case *ast.RenameTypeStmt:
		err = c.renameTypeAST(n)

	case *ast.List:
		for _, nn := range n.Items {
			if err = c.UpdateDDL(ast.Statement{
				Raw: &ast.RawStmt{
					Stmt:         nn,
					StmtLocation: stmt.Raw.StmtLocation,
					StmtLen:      stmt.Raw.StmtLen,
				},
			}, colGen); err != nil {
				return err
			}
		}

	}
	return err
}
