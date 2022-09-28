package catalog

import (
	"github.com/kwilteam/kwil-db/internal/schemadef/schema"
	"github.com/kwilteam/kwil-db/internal/sql/ast"
)

type Updater interface {
	AddSchema(*schema.Schema) error
}

// Catalog describes a database instance consisting of metadata in which database objects are defined
type Catalog struct {
	Comment       string
	DefaultSchema string
	Name          string
	Schemas       []*Schema
	SearchPath    []string
	LoadExtension func(string) *Schema

	extensions map[string]struct{}
}

// New creates a new catalog
func New(defaultSchema string) *Catalog {
	newCatalog := &Catalog{
		DefaultSchema: defaultSchema,
		Schemas:       make([]*Schema, 0),
		extensions:    make(map[string]struct{}),
	}

	if newCatalog.DefaultSchema != "" {
		newCatalog.Schemas = append(newCatalog.Schemas, &Schema{Name: defaultSchema})
	}

	return newCatalog
}

func (c *Catalog) BuildDDL(stmts []ast.Statement) error {
	for i := range stmts {
		if err := c.UpdateDDL(stmts[i], nil); err != nil {
			return err
		}
	}
	return nil
}

func (c *Catalog) UpdateDDL(stmt ast.Statement, colGen columnGenerator) error {
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
