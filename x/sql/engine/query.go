package engine

import (
	"kwil/x/sql/ast"
)

type Function struct {
	Rel        *ast.FuncName
	ReturnType *ast.TypeName
}

type Table struct {
	Rel     *ast.TableName
	Columns []*Column
}

type Column struct {
	Name         string
	DataType     string
	NotNull      bool
	IsArray      bool
	Comment      string
	Length       *int
	IsNamedParam bool
	IsFuncCall   bool

	// XXX: Figure out what PostgreSQL calls `foo.id`
	Scope      string
	Table      *ast.TableName
	TableAlias string
	Type       *ast.TypeName

	skipTableRequiredCheck bool
}

type Query struct {
	SQL      string
	Columns  []*Column
	Params   []Parameter
	Comments []string
}

type Parameter struct {
	Number int
	Column *Column
}
