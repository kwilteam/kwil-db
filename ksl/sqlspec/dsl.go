package sqlspec

import (
	"fmt"
	"strings"

	"ksl/kslspec"
)

func New(name string) *Schema {
	return &Schema{Name: name}
}

func (s *Schema) SetCharset(v string) *Schema {
	replaceOrAppend(&s.Attrs, &Charset{Value: v})
	return s
}

func (s *Schema) UnsetCharset() *Schema {
	del(&s.Attrs, &Charset{})
	return s
}

func (s *Schema) SetCollation(v string) *Schema {
	replaceOrAppend(&s.Attrs, &Collation{Value: v})
	return s
}

func (s *Schema) UnsetCollation() *Schema {
	del(&s.Attrs, &Collation{})
	return s
}

func (s *Schema) SetComment(v string) *Schema {
	replaceOrAppend(&s.Attrs, &Comment{Text: v})
	return s
}

func (s *Schema) UnsetComment() *Schema {
	del(&s.Attrs, &Comment{})
	return s
}

func (s *Schema) AddAttrs(attrs ...Attr) *Schema {
	s.Attrs = append(s.Attrs, attrs...)
	return s
}

func (s *Schema) SetRealm(r *Realm) *Schema {
	s.Realm = r
	return s
}

func (r *Schema) AddEnums(enums ...*Enum) *Schema {
	for _, s := range enums {
		s.Schema = r
	}
	r.Enums = append(r.Enums, enums...)
	return r
}

func (s *Schema) GetEnum(name string) (*Enum, bool) {
	for _, e := range s.Enums {
		if e.Name == name {
			return e, true
		}
	}
	return nil, false
}

func (s *Schema) AddTables(tables ...*Table) *Schema {
	for _, t := range tables {
		t.SetSchema(s)
	}
	s.Tables = append(s.Tables, tables...)
	return s
}

func (s *Schema) GetTable(name string) (*Table, bool) {
	for _, t := range s.Tables {
		if t.Name == name {
			return t, true
		}
	}
	return nil, false
}

func NewRealm(schemas ...*Schema) *Realm {
	r := &Realm{Schemas: schemas}
	for _, s := range schemas {
		s.Realm = r
	}
	return r
}

func (r *Realm) AddSchemas(schemas ...*Schema) *Realm {
	for _, s := range schemas {
		s.SetRealm(r)
	}
	r.Schemas = append(r.Schemas, schemas...)
	return r
}

func (r *Realm) SetCharset(v string) *Realm {
	replaceOrAppend(&r.Attrs, &Charset{Value: v})
	return r
}

func (r *Realm) UnsetCharset() *Realm {
	del(&r.Attrs, &Charset{})
	return r
}

func (r *Realm) SetCollation(v string) *Realm {
	replaceOrAppend(&r.Attrs, &Collation{Value: v})
	return r
}

func (r *Realm) UnsetCollation() *Realm {
	del(&r.Attrs, &Collation{})
	return r
}

func (r *Realm) Collation() string {
	if c := (Collation{}); has(r.Attrs, &c) {
		return c.Value
	}
	return ""
}
func (r *Realm) Charset() string {
	if c := (CType{}); has(r.Attrs, &c) {
		return c.Value
	}
	return ""
}
func (r *Realm) Comment() string {
	if c := (Comment{}); has(r.Attrs, &c) {
		return c.Text
	}
	return ""
}

func (s *Realm) AddQueries(queries ...*Query) *Realm {
	for _, q := range queries {
		q.Realm = s
	}
	s.Queries = append(s.Queries, queries...)
	return s
}

func (r *Realm) AddAttrs(attrs ...Attr) *Realm {
	r.Attrs = append(r.Attrs, attrs...)
	return r
}

func (r *Realm) AddRoles(roles ...*Role) *Realm {
	for _, s := range roles {
		s.Realm = r
	}
	r.Roles = append(r.Roles, roles...)
	return r
}

func NewTable(name string) *Table {
	return &Table{Name: name}
}

func (t *Table) SetCharset(v string) *Table {
	replaceOrAppend(&t.Attrs, &Charset{Value: v})
	return t
}

func (t *Table) UnsetCharset() *Table {
	del(&t.Attrs, &Charset{})
	return t
}

func (t *Table) SetCollation(v string) *Table {
	replaceOrAppend(&t.Attrs, &Collation{Value: v})
	return t
}

func (t *Table) UnsetCollation() *Table {
	del(&t.Attrs, &Collation{})
	return t
}

func (t *Table) SetComment(v string) *Table {
	replaceOrAppend(&t.Attrs, &Comment{Text: v})
	return t
}

func (t *Table) Collation() string {
	if c := (Collation{}); has(t.Attrs, &c) {
		return c.Value
	}
	return ""
}
func (t *Table) Charset() string {
	if c := (CType{}); has(t.Attrs, &c) {
		return c.Value
	}
	return ""
}
func (t *Table) Comment() string {
	if c := (Comment{}); has(t.Attrs, &c) {
		return c.Text
	}
	return ""
}

func (t *Table) SetSchema(s *Schema) *Table {
	t.Schema = s
	return t
}

func (t *Table) AddChecks(checks ...*Check) *Table {
	for _, c := range checks {
		t.Attrs = append(t.Attrs, c)
	}
	return t
}

func (t *Table) SetPrimaryKey(pk *Index) *Table {
	pk.Table = t
	t.PrimaryKey = pk
	for _, p := range pk.Parts {
		if p.Column == nil {
			continue
		}
		if _, ok := t.Column(p.Column.Name); !ok {
			t.AddColumns(p.Column)
		}
	}
	return t
}

func (t *Table) AddColumns(columns ...*Column) *Table {
	for _, c := range columns {
		c.Table = t
	}
	t.Columns = append(t.Columns, columns...)
	return t
}

func (t *Table) AddIndexes(indexes ...*Index) *Table {
	for _, idx := range indexes {
		idx.Table = t
	}
	t.Indexes = append(t.Indexes, indexes...)
	return t
}

func (t *Table) AddForeignKeys(fks ...*ForeignKey) *Table {
	for _, fk := range fks {
		fk.Table = t
	}
	t.ForeignKeys = append(t.ForeignKeys, fks...)
	return t
}

func (t *Table) AddAttrs(attrs ...Attr) *Table {
	t.Attrs = append(t.Attrs, attrs...)
	return t
}

func NewColumn(name string) *Column {
	return &Column{Name: name}
}

func NewNullColumn(name string) *Column {
	return NewColumn(name).
		SetNull(true)
}

func NewBoolColumn(name, typ string) *Column {
	return NewColumn(name).
		SetType(&BoolType{T: typ})
}

func NewNullBoolColumn(name, typ string) *Column {
	return NewBoolColumn(name, typ).
		SetNull(true)
}

func NewIntColumn(name, typ string) *Column {
	return NewColumn(name).
		SetType(&IntegerType{T: typ})
}

func NewNullIntColumn(name, typ string) *Column {
	return NewIntColumn(name, typ).
		SetNull(true)
}

func NewUintColumn(name, typ string) *Column {
	return NewColumn(name).
		SetType(&IntegerType{T: typ, Unsigned: true})
}

func NewNullUintColumn(name, typ string) *Column {
	return NewUintColumn(name, typ).
		SetNull(true)
}

type EnumOption func(*EnumType)

func EnumName(name string) EnumOption {
	return func(e *EnumType) {
		e.T = name
	}
}

func EnumValues(values ...string) EnumOption {
	return func(e *EnumType) {
		e.Values = values
	}
}

func EnumSchema(s *Schema) EnumOption {
	return func(e *EnumType) {
		e.Schema = s
	}
}

func NewEnumColumn(name string, opts ...EnumOption) *Column {
	t := &EnumType{}
	for _, opt := range opts {
		opt(t)
	}
	return NewColumn(name).SetType(t)
}

func NewNullEnumColumn(name string, opts ...EnumOption) *Column {
	return NewEnumColumn(name, opts...).
		SetNull(true)
}

type FloatOption func(*FloatType)

func FloatPrecision(precision int) FloatOption {
	return func(b *FloatType) {
		b.Precision = precision
	}
}

func FloatUnsigned(unsigned bool) FloatOption {
	return func(b *FloatType) {
		b.Unsigned = unsigned
	}
}

func NewFloatColumn(name, typ string, opts ...FloatOption) *Column {
	t := &FloatType{T: typ}
	for _, opt := range opts {
		opt(t)
	}
	return NewColumn(name).SetType(t)
}

func NewNullFloatColumn(name, typ string, opts ...FloatOption) *Column {
	return NewFloatColumn(name, typ, opts...).
		SetNull(true)
}

type TimeOption func(*TimeType)

func TimePrecision(precision int) TimeOption {
	return func(b *TimeType) {
		b.Precision = &precision
	}
}

func NewTimeColumn(name, typ string, opts ...TimeOption) *Column {
	t := &TimeType{T: typ}
	for _, opt := range opts {
		opt(t)
	}
	return NewColumn(name).SetType(t)
}

func NewNullTimeColumn(name, typ string) *Column {
	return NewTimeColumn(name, typ).
		SetNull(true)
}

type StringOption func(*StringType)

func StringSize(size int) StringOption {
	return func(b *StringType) {
		b.Size = size
	}
}

func NewStringColumn(name, typ string, opts ...StringOption) *Column {
	t := &StringType{T: typ}
	for _, opt := range opts {
		opt(t)
	}
	return NewColumn(name).SetType(t)
}

func NewNullStringColumn(name, typ string, opts ...StringOption) *Column {
	return NewStringColumn(name, typ, opts...).
		SetNull(true)
}

func (c *Column) SetNull(b bool) *Column {
	if c.Type == nil {
		c.Type = &ColumnType{}
	}
	c.Type.Nullable = b
	return c
}

func (c *Column) SetType(t kslspec.Type) *Column {
	if c.Type == nil {
		c.Type = &ColumnType{}
	}
	c.Type.Type = t
	return c
}

func (c *Column) SetColumnType(typ kslspec.Type, raw string, nullable bool) *Column {
	if c.Type == nil {
		c.Type = &ColumnType{}
	}
	c.Type.Type = typ
	c.Type.Raw = raw
	c.Type.Nullable = nullable
	return c
}

func (c *Column) SetDefault(val Expr) *Column {
	c.Default = val
	return c
}

func (c *Column) SetCharset(v string) *Column {
	replaceOrAppend(&c.Attrs, &Charset{Value: v})
	return c
}

func (c *Column) UnsetCharset() *Column {
	del(&c.Attrs, &Charset{})
	return c
}

func (c *Column) SetCollation(v string) *Column {
	replaceOrAppend(&c.Attrs, &Collation{Value: v})
	return c
}

func (c *Column) UnsetCollation() *Column {
	del(&c.Attrs, &Collation{})
	return c
}

func (c *Column) SetComment(v string) *Column {
	replaceOrAppend(&c.Attrs, &Comment{Text: v})
	return c
}

func (col *Column) Collation() string {
	if c := (Collation{}); has(col.Attrs, &c) {
		return c.Value
	}
	return ""
}
func (col *Column) Charset() string {
	if c := (CType{}); has(col.Attrs, &c) {
		return c.Value
	}
	return ""
}
func (col *Column) Comment() string {
	if c := (Comment{}); has(col.Attrs, &c) {
		return c.Text
	}
	return ""
}

func (c *Column) AddAttrs(attrs ...Attr) *Column {
	c.Attrs = append(c.Attrs, attrs...)
	return c
}

func (t *Column) AddIndexes(indexes ...*Index) *Column {
	t.Indexes = append(t.Indexes, indexes...)
	return t
}

func (t *Column) AddForeignKeys(fks ...*ForeignKey) *Column {
	t.ForeignKeys = append(t.ForeignKeys, fks...)
	return t
}

func NewIndex(name string) *Index {
	return &Index{Name: name}
}

func NewUniqueIndex(name string) *Index {
	return NewIndex(name).SetUnique(true)
}

func NewPrimaryKey(columns ...*Column) *Index {
	return new(Index).SetUnique(true).AddColumns(columns...)
}

func (i *Index) SetName(name string) *Index {
	i.Name = name
	return i
}

func (i *Index) SetUnique(b bool) *Index {
	i.Unique = b
	return i
}

func (i *Index) SetTable(t *Table) *Index {
	i.Table = t
	return i
}

func (i *Index) SetComment(v string) *Index {
	replaceOrAppend(&i.Attrs, &Comment{Text: v})
	return i
}

func (i *Index) Comment() string {
	if c := (Comment{}); has(i.Attrs, &c) {
		return c.Text
	}
	return ""
}

func (i *Index) AddAttrs(attrs ...Attr) *Index {
	i.Attrs = append(i.Attrs, attrs...)
	return i
}

func (i *Index) AddColumns(columns ...*Column) *Index {
	for _, c := range columns {
		if !c.hasIndex(i) {
			c.Indexes = append(c.Indexes, i)
		}
		i.Parts = append(i.Parts, &IndexPart{Seq: len(i.Parts), Column: c})
	}
	return i
}

func (c *Column) hasIndex(idx *Index) bool {
	for i := range c.Indexes {
		if c.Indexes[i] == idx {
			return true
		}
	}
	return false
}

func (i *Index) AddExprs(exprs ...Expr) *Index {
	for _, x := range exprs {
		i.Parts = append(i.Parts, &IndexPart{Seq: len(i.Parts), Expr: x})
	}
	return i
}

func (i *Index) AddParts(parts ...*IndexPart) *Index {
	for _, p := range parts {
		if p.Column != nil && !p.Column.hasIndex(i) {
			p.Column.Indexes = append(p.Column.Indexes, i)
		}
		p.Seq = len(i.Parts)
		i.Parts = append(i.Parts, p)
	}
	return i
}

func (c *Column) SetGeneratedExpr(x *GeneratedExpr) *Column {
	replaceOrAppend(&c.Attrs, x)
	return c
}

func NewIndexPart() *IndexPart           { return &IndexPart{} }
func NewColumnPart(c *Column) *IndexPart { return &IndexPart{Column: c} }
func NewExprPart(x Expr) *IndexPart      { return &IndexPart{Expr: x} }

func (p *IndexPart) SetDesc(b bool) *IndexPart {
	p.Descending = b
	return p
}

func (p *IndexPart) AddAttrs(attrs ...Attr) *IndexPart {
	p.Attrs = append(p.Attrs, attrs...)
	return p
}

func (p *IndexPart) SetColumn(c *Column) *IndexPart {
	p.Column = c
	return p
}

func (p *IndexPart) SetExpr(x Expr) *IndexPart {
	p.Expr = x
	return p
}

func NewQuery(name string) *Query {
	return &Query{Name: name}
}

func (q *Query) SetStatement(x string) *Query {
	q.Statement = x
	return q
}

func NewRole(name string) *Role {
	return &Role{Name: name}
}

func NewDefaultRole(name string) *Role {
	return &Role{Name: name, Default: true}
}

func (r *Role) AddQueries(queries ...*Query) *Role {
	r.Queries = append(r.Queries, queries...)
	return r
}

func (r *Role) SetDefault(b bool) *Role {
	r.Default = b
	return r
}

func NewForeignKey(symbol string) *ForeignKey {
	return &ForeignKey{Name: symbol}
}

func (f *ForeignKey) SetTable(t *Table) *ForeignKey {
	f.Table = t
	return f
}

func (f *ForeignKey) AddColumns(columns ...*Column) *ForeignKey {
	for _, c := range columns {
		if !c.hasForeignKey(f) {
			c.ForeignKeys = append(c.ForeignKeys, f)
		}
	}
	f.Columns = append(f.Columns, columns...)
	return f
}

func (c *Column) hasForeignKey(fk *ForeignKey) bool {
	for i := range c.ForeignKeys {
		if c.ForeignKeys[i] == fk {
			return true
		}
	}
	return false
}

func (f *ForeignKey) SetRefTable(t *Table) *ForeignKey {
	f.RefTable = t
	return f
}

func (f *ForeignKey) AddRefColumns(columns ...*Column) *ForeignKey {
	f.RefColumns = append(f.RefColumns, columns...)
	return f
}

func (f *ForeignKey) SetOnUpdate(o string) *ForeignKey {
	f.OnUpdate = o
	return f
}

func (f *ForeignKey) SetOnDelete(o string) *ForeignKey {
	f.OnDelete = o
	return f
}

func NewEnum(name string) *Enum {
	return &Enum{Name: name}
}

func (e *Enum) AddValues(values ...string) *Enum {
	e.Values = append(e.Values, values...)
	return e
}

func DefaultPrimaryKeyName(t *Table) string {
	return fmt.Sprintf("%s_pkey", t.Name)
}

func DefaultIndexName(t *Table, columns ...*Column) string {
	var names []string
	for _, c := range columns {
		names = append(names, c.Name)
	}
	return fmt.Sprintf("%s_%s_idx", t.Name, strings.Join(names, "_"))
}

func DefaultForeignKeyName(t *Table, columns ...*Column) string {
	var names []string
	for _, c := range columns {
		names = append(names, c.Name)
	}
	return fmt.Sprintf("%s_%s_fkey", t.Name, strings.Join(names, "_"))
}

func DefaultUniqueIndexName(t *Table, columns ...*Column) string {
	var names []string
	for _, c := range columns {
		names = append(names, c.Name)
	}
	return fmt.Sprintf("%s_%s_unique_idx", t.Name, strings.Join(names, "_"))
}
