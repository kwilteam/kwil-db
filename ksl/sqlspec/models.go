package sqlspec

import (
	"fmt"

	"ksl/kslspec"
)

type (
	Realm struct {
		Schemas     []*Schema
		Roles       []*Role
		Attrs       []Attr
		Queries     []*Query
		DefaultRole *Role
	}

	Schema struct {
		Name   string
		Realm  *Realm
		Tables []*Table
		Enums  []*Enum
		Attrs  []Attr
	}

	Table struct {
		Name        string
		Schema      *Schema
		Columns     []*Column
		Indexes     []*Index
		PrimaryKey  *Index
		ForeignKeys []*ForeignKey
		Attrs       []Attr
	}

	Column struct {
		Table *Table

		Name        string
		Type        *ColumnType
		Default     Expr
		Attrs       []Attr
		Indexes     []*Index
		ForeignKeys []*ForeignKey
	}

	ColumnType struct {
		Type     kslspec.Type
		Raw      string
		Nullable bool
	}

	Index struct {
		Name   string
		Unique bool
		Table  *Table
		Attrs  []Attr
		Parts  []*IndexPart
	}

	IndexPart struct {
		Seq        int
		Descending bool
		Expr       Expr
		Column     *Column
		Attrs      []Attr
	}

	ForeignKey struct {
		Name       string
		Table      *Table
		Columns    []*Column
		RefTable   *Table
		RefColumns []*Column
		OnUpdate   string
		OnDelete   string
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

	QualifiedTypeName struct {
		Schema string
		Name   string
	}

	QualifiedColumnName struct {
		Schema string
		Table  string
		Column string
	}
)

func (q *QualifiedColumnName) Normalize(schemaName, tableName string) *QualifiedColumnName {
	if q.Schema == "" {
		q.Schema = schemaName
	}
	if q.Table == "" {
		q.Table = tableName
	}
	return q
}

func (r *Realm) Schema(name string) (*Schema, bool) {
	for _, s := range r.Schemas {
		if s.Name == name {
			return s, true
		}
	}
	return nil, false
}

func (r *Realm) Enum(schemaName, enumName string) (*Enum, bool) {
	if schema, ok := r.Schema(schemaName); ok {
		return schema.Enum(enumName)
	}
	return nil, false
}

func (r *Realm) GetOrCreateEnum(schemaName, enumName string) *Enum {
	return r.GetOrCreateSchema(schemaName).GetOrCreateEnum(enumName)
}

func (r *Realm) HasSchema(name string) bool {
	_, ok := r.Schema(name)
	return ok
}

func (r *Realm) GetOrCreateSchema(name string) *Schema {
	s, ok := r.Schema(name)
	if !ok {
		s = &Schema{Name: name, Realm: r}
		r.Schemas = append(r.Schemas, s)
	}
	return s
}

func (r *Realm) Table(schemaName, tableName string) (*Table, bool) {
	if schema, ok := r.Schema(schemaName); ok {
		return schema.Table(tableName)
	}
	return nil, false
}

func (r *Realm) GetOrCreateTable(schemaName, tableName string) *Table {
	return r.GetOrCreateSchema(schemaName).GetOrCreateTable(tableName)
}

func (r *Realm) Query(name string) (*Query, bool) {
	for _, q := range r.Queries {
		if q.Name == name {
			return q, true
		}
	}
	return nil, false
}

func (r *Realm) GetOrCreateQuery(name string) *Query {
	q, ok := r.Query(name)
	if !ok {
		q = &Query{Name: name, Realm: r}
		r.Queries = append(r.Queries, q)
	}
	return q
}

func (r *Realm) HasTable(schemaName, tableName string) bool {
	_, ok := r.Table(schemaName, tableName)
	return ok
}

func (r *Realm) Role(name string) (*Role, bool) {
	for _, q := range r.Roles {
		if q.Name == name {
			return q, true
		}
	}
	return nil, false
}

func (r *Realm) GetOrCreateRole(name string) *Role {
	role, ok := r.Role(name)
	if !ok {
		role = &Role{Name: name, Realm: r}
		r.Roles = append(r.Roles, role)
	}
	return role
}

func (s *Schema) Table(name string) (*Table, bool) {
	for _, t := range s.Tables {
		if t.Name == name {
			return t, true
		}
	}
	return nil, false
}

func (s *Schema) GetOrCreateTable(name string) *Table {
	t, ok := s.Table(name)
	if !ok {
		t = &Table{Name: name, Schema: s}
		s.Tables = append(s.Tables, t)
	}
	return t
}

func (s *Schema) Enum(name string) (*Enum, bool) {
	for _, e := range s.Enums {
		if e.Name == name {
			return e, true
		}
	}
	return nil, false
}

func (s *Schema) GetOrCreateEnum(name string) *Enum {
	e, ok := s.Enum(name)
	if !ok {
		e = &Enum{Name: name, Schema: s}
		s.Enums = append(s.Enums, e)
	}
	return e
}

func (t *Table) Column(name string) (*Column, bool) {
	for _, c := range t.Columns {
		if c.Name == name {
			return c, true
		}
	}
	return nil, false
}

func (t *Table) HasColumn(name string) bool {
	_, ok := t.Column(name)
	return ok
}

func (t *Table) GetOrCreateColumn(name string) *Column {
	c, ok := t.Column(name)
	if !ok {
		c = &Column{Name: name, Table: t}
		t.Columns = append(t.Columns, c)
	}
	return c
}

func (t *Table) Index(name string) (*Index, bool) {
	for _, i := range t.Indexes {
		if i.Name == name {
			return i, true
		}
	}
	return nil, false
}

func (t *Table) GetOrCreateIndex(name string) *Index {
	i, ok := t.Index(name)
	if !ok {
		i = &Index{Name: name, Table: t}
		t.Indexes = append(t.Indexes, i)
	}
	return i
}

func (t *Table) ForeignKey(name string) (*ForeignKey, bool) {
	for _, f := range t.ForeignKeys {
		if f.Name == name {
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

func (i *Index) Column(name string) (*Column, bool) {
	for _, p := range i.Parts {
		if p.Column != nil && p.Column.Name == name {
			return p.Column, true
		}
	}
	return nil, false
}

func (i *Index) HasColumn(name string) bool {
	_, ok := i.Column(name)
	return ok
}

const (
	NoAction   = "NO ACTION"
	Restrict   = "RESTRICT"
	Cascade    = "CASCADE"
	SetNull    = "SET NULL"
	SetDefault = "SET DEFAULT"
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
	EnumType struct {
		kslspec.Type
		T      string
		Values []string
		Schema *Schema
	}

	BinaryType struct {
		kslspec.Type
		T    string
		Size *int
	}

	StringType struct {
		kslspec.Type
		T    string
		Size int
	}

	BoolType struct {
		kslspec.Type
		T string
	}

	IntegerType struct {
		kslspec.Type
		T        string
		Unsigned bool
	}

	TimeType struct {
		kslspec.Type
		T         string
		Precision *int
	}

	SpatialType struct {
		kslspec.Type
		T string
	}

	DecimalType struct {
		kslspec.Type
		T         string
		Precision int
		Scale     int
		Unsigned  bool
	}

	FloatType struct {
		kslspec.Type
		T         string
		Unsigned  bool
		Precision int
	}

	JSONType struct {
		kslspec.Type
		T string
	}

	UserDefinedType struct {
		kslspec.Type
		T string
	}

	enumType struct {
		kslspec.Type
		T      string
		Schema string
		ID     int64
		Values []string
	}

	ArrayType struct {
		kslspec.Type
		T string // Formatted type (e.g. int[]).
	}

	BitType struct {
		kslspec.Type
		T    string
		Size int64
	}

	IntervalType struct {
		kslspec.Type
		T         string // Type name.
		F         string // Optional field. YEAR, MONTH, ..., MINUTE TO SECOND.
		Precision *int   // Optional precision.
	}

	NetworkType struct {
		kslspec.Type
		T    string
		Size int64
	}

	CurrencyType struct {
		kslspec.Type
		T string
	}

	SerialType struct {
		kslspec.Type
		T            string
		Precision    int
		SequenceName string
	}

	UUIDType struct {
		kslspec.Type
		T string
	}

	XMLType struct {
		kslspec.Type
		T string
	}
)

type UnsupportedTypeError struct {
	kslspec.Type
}

func (e UnsupportedTypeError) Error() string {
	return fmt.Sprintf("unsupported type %T", e.Type)
}

type (
	Expr interface{ expr() }

	LiteralExpr struct {
		Value string
	}

	RawExpr struct {
		Expr string
	}
)

type (
	Attr interface{ attr() }

	CType struct {
		Value string
	}

	Comment struct {
		Text string
	}

	Charset struct {
		Value string
	}

	Collation struct {
		Value string
	}

	Check struct {
		Name  string
		Expr  string
		Attrs []Attr
	}

	GeneratedExpr struct {
		Expr string
		Type string
	}

	ConstraintType struct {
		Type string
	}

	Sequence struct {
		Start, Increment int64
		Last             int64
	}

	Identity struct {
		Generation string
		Sequence   *Sequence
	}

	IndexType struct {
		T string
	}

	IndexPredicate struct {
		Predicate string
	}

	IndexColumnProperty struct {
		NullsFirst bool
		NullsLast  bool
	}

	IndexStorageParams struct {
		AutoSummarize bool
		PagesPerRange int64
	}

	IndexInclude struct {
		Columns []string
	}

	NoInherit struct{}

	CheckColumns struct {
		Columns []string
	}

	Partition struct {
		T                   string
		Parts               []*PartitionPart
		start, attrs, exprs string
	}

	PartitionPart struct {
		Expr   Expr
		Column string
		Attrs  []Attr
	}
)

// expressions.
func (LiteralExpr) expr() {}
func (RawExpr) expr()     {}

// attributes.
func (Check) attr()               {}
func (Comment) attr()             {}
func (Charset) attr()             {}
func (Collation) attr()           {}
func (GeneratedExpr) attr()       {}
func (CType) attr()               {}
func (ConstraintType) attr()      {}
func (Identity) attr()            {}
func (IndexType) attr()           {}
func (IndexPredicate) attr()      {}
func (IndexColumnProperty) attr() {}
func (IndexStorageParams) attr()  {}
func (IndexInclude) attr()        {}
func (NoInherit) attr()           {}
func (CheckColumns) attr()        {}
func (Partition) attr()           {}
