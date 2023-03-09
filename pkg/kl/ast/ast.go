package ast

import (
	"encoding/json"
	"kwil/pkg/kl/token"
)

type Node interface {
}

type Expr interface {
	Node
	exprNode()
}

type Stmt interface {
	Node
	stmtNode()
}

// TODO: is it better to separate Decl and Stmt?
//type Decl interface {
//	Node
//	declNode()
//}

// ----------------------------------------
// Expression
// only contains Ident,Boolean,Integer
type (
	BadExpr struct{}

	Ident struct {
		//Token token.Token
		Name string
	}

	BasicLit struct {
		Type  token.Token
		Value string
	}

	AttrExpr struct {
		Name   Ident
		Params []Expr
	}

	ParentExpr struct {
		X Expr
	}
)

func (x *BadExpr) exprNode()    {}
func (x *Ident) exprNode()      {}
func (x *BasicLit) exprNode()   {}
func (x *AttrExpr) exprNode()   {}
func (x *ParentExpr) exprNode() {}

// ----------------------------------------
// Statements
type (
	BadStmt struct{}

	ExprStmt struct {
		X Expr
	}

	BlockStmt struct {
		Token      token.Token
		Statements []Stmt
	}
)

func (x *BadStmt) stmtNode()   {}
func (x *ExprStmt) stmtNode()  {}
func (x *BlockStmt) stmtNode() {}

// ----------------------------------------
// Declarations

type FieldList struct {
	Names []*Ident
}

type (
	AttrDef struct {
		Name  *Ident
		Type  token.Token
		Param Expr
	}

	ColumnDef struct {
		Name  *Ident
		Type  *Ident
		Attrs []*AttrDef
	}

	TableDecl struct {
		Name *Ident
		Body []*ColumnDef
	}

	DatabaseDecl struct {
		Name *Ident
		Body *BlockStmt
	}

	ActionDecl struct {
		Name       *Ident
		Parameters *FieldList
		Body       *BlockStmt
	}
)

func (x *ColumnDef) stmtNode()    {}
func (x *AttrDef) stmtNode()      {}
func (x *TableDecl) stmtNode()    {}
func (x *DatabaseDecl) stmtNode() {}
func (x *ActionDecl) stmtNode()   {}

func (d *DatabaseDecl) Build() (def DBDefinition) {
	def.Name = d.Name.Name
	for _, stmt := range d.Body.Statements {
		switch stmt.(type) {
		case *TableDecl:
			def.Tables = append(def.Tables, stmt.(*TableDecl).Build())
		}
	}
	return
}

func (d *TableDecl) Build() (def TableDefinition) {
	def.Name = d.Name.Name
	def.Columns = []ColumnDefinition{}

	for _, column := range d.Body {
		def.Columns = append(def.Columns, column.Build())
	}
	return
}

func (d *ColumnDef) Build() (def ColumnDefinition) {
	def.Name = d.Name.Name
	def.Type = d.Type.Name
	def.Attrs = []Attribute{}
	for _, attr := range d.Attrs {
		//at := Attribute{
		//	AType: attr.Param.Type.ToInt(),
		//	Value: "MAGIC",
		//}

		at := Attribute{}

		switch attr.Param.(type) {
		case *BasicLit:
			at.Value = attr.Param.(*BasicLit).Value
			at.AType = attr.Param.(*BasicLit).Type.ToInt()
		case *Ident:
			at.Value = attr.Param.(*Ident).Name
			at.AType = token.IDENT.ToInt()
		}

		def.Attrs = append(def.Attrs, at)
	}
	return
}

type Ast struct {
	Statements []Stmt // top level is always a database declaration, and only one
}

// Generate generates JSON string from AST
func (a *Ast) Generate() []byte {
	db := DBDefinition{}

	for _, stmt := range a.Statements {
		switch stmt.(type) {
		case *DatabaseDecl:
			db = stmt.(*DatabaseDecl).Build()
		}
	}

	res, err := json.MarshalIndent(db, "", "  ")
	if err != nil {
		panic(err)
	}
	return res
}
