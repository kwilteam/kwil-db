package ast

import (
	"encoding/json"
	"fmt"

	"github.com/kwilteam/kwil-db/pkg/kuneiform/schema"
	"github.com/kwilteam/kwil-db/pkg/kuneiform/token"
	sqlParser "github.com/kwilteam/kwil-db/pkg/sql_parser"
	"github.com/pkg/errors"
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
)

func (x *BadStmt) stmtNode()   {}
func (x *ExprStmt) stmtNode()  {}
func (x *BlockStmt) stmtNode() {}
func (x SQLStmt) stmtNode()    {}

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

func (a *ActionDecl) Validate(action string, ctx sqlParser.DatabaseContext) error {
	declaredParams := map[string]bool{}

	for _, param := range a.Params {
		p, ok := param.(*Ident)
		if !ok {
			return errors.Wrap(ErrInvalidActionParam, param.String())
		}
		if _, ok := declaredParams[p.Name]; ok {
			return errors.Wrap(ErrDuplicateActionParam, p.Name)
		}
		declaredParams[p.Name] = true
	}

	for _, stmt := range a.Body.Statements {
		switch st := stmt.(type) {
		case *SQLStmt:
			//fp := p.file.Position(pos)
			lineNum := 0 //int(fp.Line)
			if err := sqlParser.ParseRawSQL(st.SQL, lineNum, action, ctx, nil, false, true); err != nil {
				return errors.Wrap(err, action)
			}
		default:
			return ErrInvalidStatement // TODO: add more info(pos)
		}
	}

	return nil
}

func (a *ActionDecl) Build() (def schema.Action) {
	def = schema.Action{}
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

func (d *TableDecl) Validate(ctx sqlParser.TableContext) error {
	tableName := d.Name.Name
	if len(ctx.PrimaryKeys) > 1 {
		return errors.Wrap(ErrMultiplePrimaryKeys, tableName)
	}

	names := map[string]bool{}
	for _, name := range ctx.Columns {
		if _, ok := names[name]; ok {
			return errors.Wrap(ErrDuplicateColumnOrIndexName, fmt.Sprintf("%s.%s", tableName, name))
		}
		names[name] = true
	}

	for _, name := range ctx.IndexColumns {
		if _, ok := names[name]; !ok {
			return errors.Wrap(sqlParser.ErrColumnNotFound, fmt.Sprintf("%s.%s", tableName, name))
		}
	}

	for _, name := range ctx.Indexes {
		if _, ok := names[name]; ok {
			return errors.Wrap(ErrDuplicateColumnOrIndexName, fmt.Sprintf("%s.%s", tableName, name))
		}
		names[name] = true
	}

	return nil
}

func (d *TableDecl) Build() (def schema.Table) {
	def = schema.Table{}
	def.Name = d.Name.Name
	def.Columns = []schema.Column{}
	def.Indexes = []schema.Index{}

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

func (d *ColumnDef) Build() (def schema.Column) {
	def = schema.Column{}
	def.Name = d.Name.Name

	var err error
	def.Type, err = schema.GetColumnType(d.Type.Name)
	if err != nil {
		panic(err)
	}

	def.Attributes = []schema.Attribute{}
	for _, attr := range d.Attrs {
		def.Attributes = append(def.Attributes, attr.Build())
	}
	return
}

func (d *AttrDef) Build() (def schema.Attribute) {
	def = schema.Attribute{}

	var err error
	def.Type, err = schema.GetAttributeType(d.Name.Name)
	if err != nil {
		panic(err)
	}

	if d.Param == nil {
		return
	}

	switch a := d.Param.(type) {
	case *BasicLit:
		// TODO: a standalone types package
		def.Value = a.Value
	case *Ident:
		def.Value = a.Name
	}
	return
}

func (d *IndexDef) Build() (def schema.Index) {
	def = schema.Index{}
	// remove the prefix # of index name
	def.Name = d.Name.Name[1:]
	def.Type = schema.IdxBtree
	if d.Unique {
		def.Type = schema.IdxUniqueBtree
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

	schema *schema.Schema
}

// Schema returns the schema of the database
func (d *Database) Schema() *schema.Schema {
	return d.schema
}

// GenerateJson generates JSON from AST
func (d *Database) GenerateJson() []byte {
	res, err := json.MarshalIndent(d.schema, "", "  ")
	if err != nil {
		panic(err)
	}
	return res
}

func (d *Database) BuildSchemaCtx() (ctx sqlParser.DatabaseContext) {
	ctx = sqlParser.NewDatabaseContext()
	db := schema.Schema{}
	db.Name = d.Name.Name
	for _, decl := range d.Decls {
		switch a := decl.(type) {
		case *TableDecl:
			db.Tables = append(db.Tables, a.Build())
		case *ActionDecl:
			db.Actions = append(db.Actions, a.Build())
		}
	}
	d.schema = &db

	// same table/index name will be overwritten
	for _, table := range d.schema.Tables {
		tCtx := sqlParser.NewTableContext()
		for _, column := range table.Columns {
			tCtx.Columns = append(tCtx.Columns, column.Name)
			for _, attr := range column.Attributes {
				if attr.Type == schema.AttrPrimaryKey {
					tCtx.PrimaryKeys = append(tCtx.PrimaryKeys, column.Name)
				}
			}
		}
		for _, index := range table.Indexes {
			tCtx.Indexes = append(tCtx.Indexes, index.Name)
			tCtx.IndexColumns = index.Columns
		}

		ctx.Tables[table.Name] = tCtx
	}

	for _, action := range d.schema.Actions {
		aCtx := sqlParser.NewActionContext()
		for _, input := range action.Inputs {
			aCtx[input] = true
		}

		ctx.Actions[action.Name] = aCtx
	}

	return
}

func (d *Database) Validate() error {
	ctx := d.BuildSchemaCtx()

	actionNames := map[string]bool{}
	tableNames := map[string]bool{}

	for _, decl := range d.Decls {
		switch a := decl.(type) {
		case *ActionDecl:
			if _, ok := actionNames[a.Name.Name]; ok {
				return errors.Wrap(ErrDuplicateActionName, a.Name.Name)
			}
			actionNames[a.Name.Name] = true

			if err := a.Validate(a.Name.Name, ctx); err != nil {
				return err
			}
		case *TableDecl:
			if _, ok := tableNames[a.Name.Name]; ok {
				return errors.Wrap(ErrDuplicateTableName, a.Name.Name)
			}
			tableNames[a.Name.Name] = true

			if err := a.Validate(ctx.Tables[a.Name.Name]); err != nil {
				return err
			}
		}
	}

	return nil
}
