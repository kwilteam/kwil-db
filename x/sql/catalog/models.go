package catalog

import (
	"kwil/x/schemadef/schema"
	"kwil/x/sql/sqlparse/ast"
)

type (
	Schema struct {
		Name   string
		Tables []*Table
		Types  []Type
		Funcs  []*Function

		Comment string
	}

	Table struct {
		QualName *QualName
		Columns  []*Column
		Comment  string
	}

	Column struct {
		Name      string
		Type      *QualName
		IsNotNull bool
		IsArray   bool
		Comment   string
		Length    *int
	}
	Function struct {
		Name               string
		Args               []*Argument
		ReturnType         *QualName
		Comment            string
		Desc               string
		ReturnTypeNullable bool
	}

	Argument struct {
		Name       string
		Type       *QualName
		HasDefault bool
		Mode       FuncParamMode
	}

	QualName struct {
		Catalog string
		Schema  string
		Name    string
	}
)

type (
	Type interface {
		typ()
		SetComment(string)
	}

	Enum struct {
		Name    string
		Vals    []string
		Comment string
	}

	CompositeType struct {
		Name    string
		Comment string
	}
)

type (
	Updater interface {
		UpdateSchema(*schema.Schema, ColumnConverter) error
	}

	ColumnConverter interface {
		ConvertColumn(tab *schema.Table, col *schema.Column) (*Column, error)
	}

	ColumnGenerator interface {
		OutputColumns(node ast.Node) ([]*Column, error)
	}
)

type FuncParamMode int

const (
	FuncParamIn FuncParamMode = iota
	FuncParamOut
	FuncParamInOut
	FuncParamVariadic
	FuncParamTable
)

func (f *Function) InArgs() []*Argument {
	var args []*Argument
	for _, a := range f.Args {
		switch a.Mode {
		case FuncParamTable, FuncParamOut:
			continue
		default:
			args = append(args, a)
		}
	}
	return args
}
