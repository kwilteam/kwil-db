package ast

import (
	"encoding/json"
	"fmt"
	"kwil/pkg/engine/models"
	"kwil/pkg/engine/types"
	"kwil/pkg/kl/token"
	"strings"
)

type ActionContext map[string]any

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

type Decl interface {
	Node
	declNode()
}

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

	DeclStmt struct {
		Decl Decl
	}

	BlockStmt struct {
		Token      token.Token
		Statements []Stmt
	}

	SQLStmt struct {
		SQL string
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
func (x SQLStmt) stmtNode()     {}
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
	BadDecl struct{}

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

	ActionDecl struct {
		Name   *Ident
		Public bool
		Params []Expr
		Body   *BlockStmt
	}
)

func (x *BadDecl) declNode()    {}
func (x *ColumnDef) stmtNode()  {}
func (x *AttrDef) stmtNode()    {}
func (x *IndexDef) stmtNode()   {}
func (x *TableDecl) declNode()  {}
func (x *ActionDecl) declNode() {}

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

func (a *ActionDecl) Build() (def *models.Action) {
	def = &models.Action{}
	// should be validated before build
	def.Name = a.Name.Name
	def.Public = a.Public
	def.Inputs = []string{}
	declaredParams := map[string]bool{}
	for _, param := range a.Params {
		if p, ok := param.(*Ident); ok {
			def.Inputs = append(def.Inputs, p.Name)
			declaredParams[p.Name] = true
		}
	}

	for _, stmt := range a.Body.Statements {
		switch st := stmt.(type) {
		case *SQLStmt:
			def.Statements = append(def.Statements, st.SQL)
		default:
			panic("statement not supported")
		}
	}
	return
}

func (d *TableDecl) Build() (def *models.Table) {
	def = &models.Table{}
	def.Name = d.Name.Name
	def.Columns = []*models.Column{}
	def.Indexes = []*models.Index{}

	for _, column := range d.Body {
		c, ok := column.(*ColumnDef)
		if !ok {
			panic("invalid column")
		}
		def.Columns = append(def.Columns, c.Build())
	}

	for _, index := range d.Idx {
		i, ok := index.(*IndexDef)
		if !ok {
			panic("invalid index")
		}

		def.Indexes = append(def.Indexes, i.Build())
	}
	return
}

func (d *ColumnDef) Build() (def *models.Column) {
	def = &models.Column{}
	def.Name = d.Name.Name

	typeName := strings.ToLower(d.Type.Name)

	def.Type = GetMappedColumnType(typeName)

	def.Attributes = []*models.Attribute{}
	for _, attr := range d.Attrs {
		def.Attributes = append(def.Attributes, attr.Build())
	}
	return
}

func (d *AttrDef) Build() (def *models.Attribute) {
	def = &models.Attribute{}
	at := token.ILLEGAL
	switch a := d.Param.(type) {
	case *BasicLit:
		def.Value = []byte(a.Value)
		at = a.Type
	case *Ident:
		def.Value = []byte(a.Name)
		at = token.IDENT
	}

	def.Type = GetMappedAttributeType(at)
	return
}

func (d *IndexDef) Build() (def *models.Index) {
	def = &models.Index{}
	def.Name = d.Name.Name
	def.Type = types.BTREE
	if d.Unique {
		def.Type = types.UNIQUE_BTREE
	}
	def.Columns = []string{}
	for _, col := range d.Columns {
		switch c := col.(type) {
		case *Ident:
			def.Columns = append(def.Columns, c.Name)
		case *SelectorExpr:
			def.Columns = append(def.Columns, c.String())
		}
	}
	return
}

type Database struct {
	Name  *Ident
	Decls []Decl
}

// Generate generates JSON string from AST
func (d *Database) Generate() []byte {
	db := models.Dataset{}
	db.Name = d.Name.Name
	for _, decl := range d.Decls {
		switch a := decl.(type) {
		case *TableDecl:
			db.Tables = append(db.Tables, a.Build())
		case *ActionDecl:
			db.Actions = append(db.Actions, a.Build())
		}
	}

	res, err := json.MarshalIndent(db, "", "  ")
	if err != nil {
		panic(err)
	}
	return res
}
