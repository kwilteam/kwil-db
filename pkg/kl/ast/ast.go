package ast

import (
	"encoding/json"
	"fmt"
	"kwil/pkg/kl/token"
)

type Node interface {
}

type Expr interface {
	Node
	String() string
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
		Name   *Ident
		Params []Expr
	}

	ParentExpr struct {
		X Expr
	}

	SelectorExpr struct {
		Name *Ident
		Sel  *Ident
	}
)

func (x *BadExpr) exprNode()      {}
func (x *Ident) exprNode()        {}
func (x *BasicLit) exprNode()     {}
func (x *AttrExpr) exprNode()     {}
func (x *ParentExpr) exprNode()   {}
func (x *SelectorExpr) exprNode() {}

func (x *BadExpr) String() {}

func (x *Ident) String() string {
	return x.Name
}

func (x *BasicLit) String() string {
	return x.Value
}

func (x *SelectorExpr) String() string {
	return fmt.Sprintf("%s.%s", x.Name, x.Sel)
}

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

	InsertStmt struct {
		Table   *Ident
		Columns []Expr
		Values  []Expr
	}

	UpdateStmt struct {
		Table   *Ident
		Columns []Expr
		Values  []Expr
	}
)

func (x *BadStmt) stmtNode()    {}
func (x *ExprStmt) stmtNode()   {}
func (x *BlockStmt) stmtNode()  {}
func (x *InsertStmt) stmtNode() {}
func (x *UpdateStmt) stmtNode() {}

func (s *InsertStmt) Validate() error {
	if len(s.Columns) != len(s.Values) {
		return fmt.Errorf("number of columns and values are different")
	}

	return nil
}

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

	// IndexDef is a definition of an index, of table.
	IndexDef struct {
		Name    *Ident
		Columns []Expr
		Unique  bool
	}

	ColumnDef struct {
		Name  *Ident
		Type  *Ident
		Attrs []*AttrDef
	}

	TableDecl struct {
		Name *Ident
		Body []Stmt
		Idx  []Stmt
	}

	DatabaseDecl struct {
		Name *Ident
		Body *BlockStmt
	}

	ActionDecl struct {
		Name   *Ident
		Public bool
		Params []Expr
		Body   *BlockStmt
	}
)

func (x *ColumnDef) stmtNode()    {}
func (x *AttrDef) stmtNode()      {}
func (x *IndexDef) stmtNode()     {}
func (x *TableDecl) stmtNode()    {}
func (x *DatabaseDecl) stmtNode() {}
func (x *ActionDecl) stmtNode()   {}

func (a *ActionDecl) Validate() error {
	declaredParams := map[string]bool{}

	for _, param := range a.Params {
		p, ok := param.(*Ident)
		if !ok {
			return fmt.Errorf("unsupported action parameter, got %s", param)
		}
		declaredParams[p.Name] = true
	}

	for _, stmt := range a.Body.Statements {
		switch st := stmt.(type) {
		case *InsertStmt:
			if len(st.Columns) > 0 {
				for _, column := range st.Columns {
					_, ok := column.(*Ident)
					if !ok {
						return fmt.Errorf("unsupported column, got %s", column)
					}
				}
			}

			if len(st.Values) != len(st.Columns) {
				return fmt.Errorf("unmatched number of columns and values")
			}

			for _, value := range st.Values {
				switch v := value.(type) {
				case *BasicLit:
					continue
				case *Ident:
					_, paramExist := declaredParams[v.Name]
					if !paramExist {
						return fmt.Errorf("undefined parameter: %s", v.Name)
					}
				}
			}
		case *UpdateStmt:
			continue
		default:
			return fmt.Errorf("unsupported action statement, got %s", stmt)
		}
	}

	return nil
}

func (a *ActionDecl) Build() (def ActionDefinition) {
	// should be validated before build
	def.Name = a.Name.Name
	def.Public = a.Public
	def.Params = []string{}
	def.Ops = []SQLOP{}

	declaredParams := map[string]bool{}
	for _, param := range a.Params {
		if p, ok := param.(*Ident); ok {
			def.Params = append(def.Params, p.Name)
			declaredParams[p.Name] = true
		}
	}

	for _, stmt := range a.Body.Statements {
		switch st := stmt.(type) {
		case *InsertStmt:
			var columns []string
			if len(st.Columns) > 0 {
				for _, column := range st.Columns {
					cl, _ := column.(*Ident)
					columns = append(columns, cl.Name)
				}
			}
			def.Ops = append(def.Ops, SQLOP{Op: "insert", Args: []string{st.Table.Name}})
			def.Ops = append(def.Ops, SQLOP{Op: "columns", Args: columns})

			var values []string
			for _, value := range st.Values {
				switch v := value.(type) {
				case *BasicLit:
					values = append(values, v.Value)
				case *Ident:
					if _, ok := declaredParams[v.Name]; ok {
						values = append(values, v.Name)
					}
				}
			}
			def.Ops = append(def.Ops, SQLOP{Op: "values", Args: values})

		case *UpdateStmt:
			continue
		}
	}
	return
}

func (d *DatabaseDecl) Build() (def DBDefinition) {
	def.Name = d.Name.Name
	for _, stmt := range d.Body.Statements {
		switch s := stmt.(type) {
		case *TableDecl:
			def.Tables = append(def.Tables, s.Build())
		case *ActionDecl:
			def.Actions = append(def.Actions, s.Build())
		}
	}
	return
}

func (d *TableDecl) Build() (def TableDefinition) {
	def.Name = d.Name.Name
	def.Columns = []ColumnDefinition{}
	def.Indexes = []IndexDefinition{}

	for _, column := range d.Body {
		def.Columns = append(def.Columns, column.(*ColumnDef).Build())
	}

	for _, index := range d.Idx {
		def.Indexes = append(def.Indexes, index.(*IndexDef).Build())
	}
	return
}

func (d *ColumnDef) Build() (def ColumnDefinition) {
	def.Name = d.Name.Name
	def.Type = d.Type.Name
	def.Attrs = []Attribute{}
	for _, attr := range d.Attrs {
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

func (d *IndexDef) Build() (def IndexDefinition) {
	def.Name = d.Name.Name
	def.Type = "index"
	if d.Unique {
		def.Type = "unique"
	}
	def.Columns = []string{}
	for _, col := range d.Columns {
		switch col.(type) {
		case *Ident:
			def.Columns = append(def.Columns, col.(*Ident).Name)
		case *SelectorExpr:
			def.Columns = append(def.Columns, col.(*SelectorExpr).String())
		}
	}
	return
}

type Ast struct {
	Statements []Stmt // top level is always a database declaration, and only one
}

type File struct {
	Decl DatabaseDecl
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
