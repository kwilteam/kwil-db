package schema

import "fmt"

type (
	// A Realm or a database describes a domain of schema resources that are logically connected
	// and can be accessed and queried in the same connection (e.g. a physical database instance).
	Realm struct {
		Schemas []*Schema
		Roles   []*Role
		Attrs   []Attr
		Queries []*Query
	}

	// A Schema describes a database schema (i.e. named database).
	Schema struct {
		Name   string
		Realm  *Realm
		Tables []*Table
		Enums  []*Enum
		Attrs  []Attr
	}

	// A Table represents a table definition.
	Table struct {
		Name        string
		Schema      *Schema
		Columns     []*Column
		Indexes     []*Index
		PrimaryKey  *Index
		ForeignKeys []*ForeignKey
		Attrs       []Attr
	}

	// A Column represents a column definition.
	Column struct {
		Name        string
		Type        *ColumnType
		Default     Expr
		Attrs       []Attr
		Indexes     []*Index
		ForeignKeys []*ForeignKey
	}

	// ColumnType represents a column type that is implemented by the dialect.
	ColumnType struct {
		Type     Type
		Raw      string
		Nullable bool
	}

	// An Index represents an index definition.
	Index struct {
		Name   string
		Unique bool
		Table  *Table
		Attrs  []Attr
		Parts  []*IndexPart
	}

	// An IndexPart represents an index part that
	// can be either an expression or a column.
	IndexPart struct {
		// SeqNo represents the sequence number of the key part in the index.
		SeqNo int
		// Descending indicates if the key part is stored in descending
		// order. All databases use ascending order as default.
		Descending bool
		X          Expr
		C          *Column
		Attrs      []Attr
	}

	// A ForeignKey represents an index definition.
	ForeignKey struct {
		Symbol     string
		Table      *Table
		Columns    []*Column
		RefTable   *Table
		RefColumns []*Column
		OnUpdate   ReferenceOption
		OnDelete   ReferenceOption
	}

	Enum struct {
		Name   string
		Schema *Schema
		Values []string
	}

	Query struct {
		Name      string
		Realm     *Realm
		Statement string
	}

	Role struct {
		Name    string
		Queries []*Query
		Realm   *Realm
		Default bool
	}

	Function struct {
		Name               string
		Args               []*Argument
		ReturnType         Type
		Comment            string
		Description        string
		ReturnTypeNullable bool
	}

	Argument struct {
		Name       string
		Type       Type
		HasDefault bool
		Mode       FuncParamMode
	}
)

func (r *Realm) Schema(name string) (*Schema, bool) {
	for _, s := range r.Schemas {
		if s.Name == name {
			return s, true
		}
	}
	return nil, false
}

func (r *Realm) Query(name string) (*Query, bool) {
	for _, q := range r.Queries {
		if q.Name == name {
			return q, true
		}
	}
	return nil, false
}

func (r *Realm) Role(name string) (*Role, bool) {
	for _, q := range r.Roles {
		if q.Name == name {
			return q, true
		}
	}
	return nil, false
}

func (s *Schema) Table(name string) (*Table, bool) {
	for _, t := range s.Tables {
		if t.Name == name {
			return t, true
		}
	}
	return nil, false
}

func (s *Schema) Enum(name string) (*Enum, bool) {
	for _, e := range s.Enums {
		if e.Name == name {
			return e, true
		}
	}
	return nil, false
}

func (t *Table) Column(name string) (*Column, bool) {
	for _, c := range t.Columns {
		if c.Name == name {
			return c, true
		}
	}
	return nil, false
}

func (t *Table) Index(name string) (*Index, bool) {
	for _, i := range t.Indexes {
		if i.Name == name {
			return i, true
		}
	}
	return nil, false
}

func (t *Table) ForeignKey(symbol string) (*ForeignKey, bool) {
	for _, f := range t.ForeignKeys {
		if f.Symbol == symbol {
			return f, true
		}
	}
	return nil, false
}

func (f *ForeignKey) Column(name string) (*Column, bool) {
	for _, c := range f.Columns {
		if c.Name == name {
			return c, true
		}
	}
	return nil, false
}

func (f *ForeignKey) RefColumn(name string) (*Column, bool) {
	for _, c := range f.RefColumns {
		if c.Name == name {
			return c, true
		}
	}
	return nil, false
}

// ReferenceOption for constraint actions.
type ReferenceOption string

// Reference options (actions) specified by ON UPDATE and ON DELETE
// subclauses of the FOREIGN KEY clause.
const (
	NoAction   ReferenceOption = "NO ACTION"
	Restrict   ReferenceOption = "RESTRICT"
	Cascade    ReferenceOption = "CASCADE"
	SetNull    ReferenceOption = "SET NULL"
	SetDefault ReferenceOption = "SET DEFAULT"
)

type FuncParamMode int

const (
	FuncParamIn FuncParamMode = iota
	FuncParamOut
	FuncParamInOut
	FuncParamVariadic
	FuncParamTable
)

type (
	// A Type represents a database type. The types below implements this
	// interface and can be used for describing schemas.
	Type interface {
		typ()
	}

	// EnumType represents an enum type.
	EnumType struct {
		T      string   // Optional type.
		Values []string // Enum values.
		Schema *Schema  // Optional schema.
	}

	// BinaryType represents a type that stores a binary data.
	BinaryType struct {
		T    string
		Size *int
	}

	// StringType represents a string type.
	StringType struct {
		T    string
		Size int
	}

	// BoolType represents a boolean type.
	BoolType struct {
		T string
	}

	// IntegerType represents an int type.
	IntegerType struct {
		T        string
		Unsigned bool
		Attrs    []Attr
	}

	// TimeType represents a date/time type.
	TimeType struct {
		T         string
		Precision *int
	}

	// SpatialType represents a spatial/geometric type.
	SpatialType struct {
		T string
	}

	// DecimalType represents a fixed-point type that stores exact numeric values.
	DecimalType struct {
		T         string
		Precision int
		Scale     int
		Unsigned  bool
	}

	// FloatType represents a floating-point type that stores approximate numeric values.
	FloatType struct {
		T         string
		Unsigned  bool
		Precision int
	}

	// JSONType represents a JSON type.
	JSONType struct {
		T string
	}
	// UnsupportedType represents a type that is not supported by the drivers.
	UnsupportedType struct {
		T string
	}
	UnsupportedTypeError struct {
		Type
	}
)

func (e UnsupportedTypeError) Error() string {
	return fmt.Sprintf("unsupported type %T", e.Type)
}

type (
	// Expr defines an SQL expression in schema DDL.
	Expr interface {
		expr()
	}

	// Literal represents a basic literal expression like 1, or '1'.
	// String literals are usually quoted with single or double quotes.
	Literal struct {
		V string
	}

	// RawExpr represents a raw expression like "uuid()" or "current_timestamp()".
	RawExpr struct {
		X string
	}
)

type (
	// Attr represents the interface that all attributes implement.
	Attr interface {
		attr()
	}

	// Comment describes a schema element comment.
	Comment struct {
		Text string
	}

	// Charset describes a column or a table character-set setting.
	Charset struct {
		V string
	}

	// Collation describes a column or a table collation setting.
	Collation struct {
		V string
	}

	// Check describes a CHECK constraint.
	Check struct {
		Name  string // Optional constraint name.
		Expr  string // Actual CHECK.
		Attrs []Attr // Additional attributes (e.g. ENFORCED).
	}

	// GeneratedExpr describes the expression used for generating
	// the value of a generated/virtual column.
	GeneratedExpr struct {
		Expr string
		Type string // Optional type. e.g. STORED or VIRTUAL.
	}
)

// expressions.
func (*Literal) expr() {}
func (*RawExpr) expr() {}

// types.
func (*BoolType) typ()        {}
func (*EnumType) typ()        {}
func (*StringType) typ()      {}
func (*BinaryType) typ()      {}
func (*IntegerType) typ()     {}
func (*TimeType) typ()        {}
func (*SpatialType) typ()     {}
func (*FloatType) typ()       {}
func (*DecimalType) typ()     {}
func (*JSONType) typ()        {}
func (*UnsupportedType) typ() {}

// attributes.
func (*Check) attr()         {}
func (*Comment) attr()       {}
func (*Charset) attr()       {}
func (*Collation) attr()     {}
func (*GeneratedExpr) attr() {}
