package engine

import (
	"kwil/x/sql/catalog"
)

type Function struct {
	Rel        *catalog.QualName
	ReturnType *catalog.QualName
}

type Table struct {
	Rel     *catalog.QualName
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
	Table      *catalog.QualName
	TableAlias string
	Type       *catalog.QualName

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
