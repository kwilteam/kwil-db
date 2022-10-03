package schemapb

import (
	"kwil/x/schemadef/schema"
	"kwil/x/sql/postgres"
)

func ToChanges(changes []*Change) []schema.SchemaChange {
	out := make([]schema.SchemaChange, len(changes))
	for i, c := range changes {
		out[i] = ToChange(c)
	}
	return out
}

func ToChange(c *Change) schema.SchemaChange {
	switch c.Value.(type) {
	case *Change_AddSchema:
		return &schema.AddSchema{}
	case *Change_DropSchema:
		return &schema.DropSchema{}
	case *Change_ModifySchema:
		return &schema.ModifySchema{}
	case *Change_AddTable:
		return &schema.AddTable{}
	case *Change_DropTable:
		return &schema.DropTable{}
	case *Change_ModifyTable:
		return &schema.ModifyTable{}
	case *Change_RenameTable:
		return &schema.RenameTable{}
	case *Change_AddColumn:
		return &schema.AddColumn{}
	case *Change_DropColumn:
		return &schema.DropColumn{}
	case *Change_ModifyColumn:
		return &schema.ModifyColumn{}
	case *Change_RenameColumn:
		return &schema.RenameColumn{}
	case *Change_AddIndex:
		return &schema.AddIndex{}
	case *Change_DropIndex:
		return &schema.DropIndex{}
	case *Change_ModifyIndex:
		return &schema.ModifyIndex{}
	case *Change_RenameIndex:
		return &schema.RenameIndex{}
	case *Change_AddForeignKey:
		return &schema.AddForeignKey{}
	case *Change_DropForeignKey:
		return &schema.DropForeignKey{}
	case *Change_ModifyForeignKey:
		return &schema.ModifyForeignKey{}
	case *Change_AddCheck:
		return &schema.AddCheck{}
	case *Change_DropCheck:
		return &schema.DropCheck{}
	case *Change_ModifyCheck:
		return &schema.ModifyCheck{}
	case *Change_AddAttr:
		return &schema.AddAttr{}
	case *Change_DropAttr:
		return &schema.DropAttr{}
	case *Change_ModifyAttr:
		return &schema.ModifyAttr{}
	default:
		return nil
	}
}

func ToRealm(pb *Realm) *schema.Realm {
	r := schema.NewRealm()
	for _, s := range pb.Schemas {
		loadSchema(r, s)
	}

	for _, q := range pb.Queries {
		loadQuery(r, q)
	}

	for _, rol := range pb.Roles {
		loadRole(r, rol)
	}

	r.AddAttrs(toAttrs(pb.Attrs)...)

	return r
}

func FromRealm(r *schema.Realm) *Realm {
	pb := &Realm{}
	for _, s := range r.Schemas {
		pb.Schemas = append(pb.Schemas, fromSchema(s))
	}

	for _, q := range r.Queries {
		pb.Queries = append(pb.Queries, fromQuery(q))
	}

	for _, rol := range r.Roles {
		pb.Roles = append(pb.Roles, fromRole(rol))
	}

	pb.Attrs = fromAttrs(r.Attrs)

	return pb
}

func fromQuery(q *schema.Query) *Query {
	return &Query{
		Name:      q.Name,
		Statement: q.Statement,
	}
}

func fromRole(r *schema.Role) *Role {
	queryNames := make([]string, len(r.Queries))
	for i, q := range r.Queries {
		queryNames[i] = q.Name
	}

	pb := &Role{
		Name:    r.Name,
		Default: r.Default,
		Queries: queryNames,
	}
	return pb
}

func fromSchema(s *schema.Schema) *Schema {
	pb := &Schema{Name: s.Name}
	for _, t := range s.Tables {
		pb.Tables = append(pb.Tables, fromTable(t))
	}

	for _, e := range s.Enums {
		pb.Enums = append(pb.Enums, fromEnum(e))
	}

	pb.Attrs = fromAttrs(s.Attrs)

	return pb
}

func fromEnum(e *schema.Enum) *Enum {
	pb := &Enum{
		QualName: &QualName{
			Schema: e.Schema.Name,
			Name:   e.Name,
		},
	}
	pb.Values = append(pb.Values, e.Values...)

	return pb
}

func fromTable(t *schema.Table) *Table {
	pb := &Table{
		QualName: &QualName{
			Schema: t.Schema.Name,
			Name:   t.Name,
		},
	}
	for _, c := range t.Columns {
		pb.Columns = append(pb.Columns, fromColumn(c))
	}

	for _, i := range t.Indexes {
		pb.Indexes = append(pb.Indexes, fromIndex(i))
	}

	if pk := t.PrimaryKey; pk != nil {
		pb.PrimaryKey = fromIndex(t.PrimaryKey)
	}

	pb.ForeignKeys = fromForeignKeys(t.ForeignKeys)
	pb.Attrs = fromAttrs(t.Attrs)

	return pb
}

func fromColumn(c *schema.Column) *Column {
	pb := &Column{
		Name: c.Name,
		Type: fromColumnType(c.Type),
	}
	if c.Default != nil {
		pb.Default = fromExpr(c.Default)
	}

	for _, idx := range c.Indexes {
		pb.Indexes = append(pb.Indexes, fromIndex(idx))
	}

	pb.ForeignKeys = fromForeignKeys(c.ForeignKeys)
	pb.Attrs = fromAttrs(c.Attrs)

	return pb
}

func fromColumnType(ct *schema.ColumnType) *ColumnType {
	return &ColumnType{
		Type:     fromType(ct.Type),
		Raw:      ct.Raw,
		Nullable: ct.Nullable,
	}
}

func fromIndex(i *schema.Index) *Index {
	pb := &Index{
		Name: i.Name,
	}
	for _, p := range i.Parts {
		pb.Parts = append(pb.Parts, fromIndexPart(p))
	}

	pb.Attrs = fromAttrs(i.Attrs)

	return pb
}

func fromIndexPart(p *schema.IndexPart) *IndexPart {
	pb := &IndexPart{
		Seq:        int64(p.Seq),
		Descending: p.Descending,
	}

	switch {
	case p.Column != nil:
		pb.Value = &IndexPart_Column{Column: p.Column.Name}
	case p.Expr != nil:
		pb.Value = &IndexPart_Expr{Expr: fromExpr(p.Expr)}
	}

	pb.Attrs = fromAttrs(p.Attrs)
	return pb
}

func fromExpr(e schema.Expr) *Expr {
	pb := &Expr{}
	switch e := e.(type) {
	case *schema.Literal:
		pb.Value = &Expr_Literal{Literal: &Literal{Value: e.V}}
	case *schema.RawExpr:
		pb.Value = &Expr_RawExpr{RawExpr: &RawExpr{Expr: e.X}}
	}
	return pb
}

func fromAttrs(attrs []schema.Attr) []*Attr {
	if attrs == nil {
		return nil
	}

	m := make([]*Attr, len(attrs))
	for i, a := range attrs {
		m[i] = fromAttr(a)
	}
	return m
}

func fromAttr(a schema.Attr) *Attr {
	switch a := a.(type) {
	case *schema.Comment:
		return &Attr{Value: &Attr_CommentAttr{CommentAttr: &Comment{Text: a.Text}}}
	case *schema.Charset:
		return &Attr{Value: &Attr_CharsetAttr{CharsetAttr: &Charset{Value: a.V}}}
	case *schema.Collation:
		return &Attr{Value: &Attr_CollationAttr{CollationAttr: &Collation{Value: a.V}}}
	case *schema.Check:
		return &Attr{Value: &Attr_CheckAttr{CheckAttr: &Check{Name: a.Name, Expr: a.Expr}}}
	case *schema.GeneratedExpr:
		return &Attr{Value: &Attr_GeneratedExprAttr{GeneratedExprAttr: &GeneratedExpr{Expr: a.Expr, Type: a.Type}}}
	case *postgres.ConstraintType:
		return &Attr{Value: &Attr_ConstraintTypeAttr{ConstraintTypeAttr: &ConstraintType{Type: a.T}}}
	case *postgres.Identity:
		return &Attr{Value: &Attr_IdentityAttr{IdentityAttr: &Identity{Generation: a.Generation, Sequence: &Sequence{Start: a.Sequence.Start, Last: a.Sequence.Last, Increment: a.Sequence.Increment}}}}
	case *postgres.IndexType:
		return &Attr{Value: &Attr_IndexTypeAttr{IndexTypeAttr: &IndexType{Type: a.T}}}
	case *postgres.IndexPredicate:
		return &Attr{Value: &Attr_IndexPredicateAttr{IndexPredicateAttr: &IndexPredicate{Predicate: a.Predicate}}}
	case *postgres.IndexColumnProperty:
		return &Attr{Value: &Attr_IndexColumnPropertyAttr{IndexColumnPropertyAttr: &IndexColumnProperty{NullsFirst: a.NullsFirst, NullsLast: a.NullsLast}}}
	case *postgres.IndexStorageParams:
		return &Attr{Value: &Attr_IndexStorageParamsAttr{IndexStorageParamsAttr: &IndexStorageParams{AutoSummarize: a.AutoSummarize, PagesPerRange: a.PagesPerRange}}}
	case *postgres.IndexInclude:
		return &Attr{Value: &Attr_IndexIncludeAttr{IndexIncludeAttr: &IndexInclude{Columns: a.Columns}}}
	case *postgres.NoInherit:
		return &Attr{Value: &Attr_NoInheritAttr{NoInheritAttr: &NoInherit{}}}
	case *postgres.CheckColumns:
		return &Attr{Value: &Attr_CheckColumnsAttr{CheckColumnsAttr: &CheckColumns{Columns: a.Columns}}}
	case *postgres.Partition:
		parts := make([]*PartitionPart, len(a.Parts))
		for i, p := range a.Parts {
			parts[i] = &PartitionPart{
				Expr:   fromExpr(p.Expr),
				Column: p.Column,
				Attrs:  fromAttrs(p.Attrs),
			}
		}
		return &Attr{Value: &Attr_PartitionAttr{PartitionAttr: &Partition{Type: a.T, Parts: parts}}}
	default:
		return nil
	}
}

func fromForeignKeys(fks []*schema.ForeignKey) []*ForeignKey {
	m := make([]*ForeignKey, len(fks))
	for i, fk := range fks {
		m[i] = fromForeignKey(fk)
	}
	return m
}

func columnNames(cols []*schema.Column) []string {
	names := make([]string, len(cols))
	for i, c := range cols {
		names[i] = c.Name
	}
	return names
}

func fromForeignKey(fk *schema.ForeignKey) *ForeignKey {
	pb := &ForeignKey{
		Symbol:   fk.Name,
		OnUpdate: fromForeignKeyAction(fk.OnUpdate),
		OnDelete: fromForeignKeyAction(fk.OnDelete),
	}

	if fk.Table != nil {
		pb.Table = &QualName{Schema: fk.Table.Schema.Name, Name: fk.Table.Name}
		pb.Columns = columnNames(fk.Columns)
	}

	if fk.RefTable != nil {
		pb.RefTable = &QualName{Schema: fk.RefTable.Schema.Name, Name: fk.RefTable.Name}
		pb.RefColumns = columnNames(fk.RefColumns)
	}

	return pb
}

func fromForeignKeyAction(a schema.ReferenceOption) ReferenceOption {
	switch a {
	case schema.NoAction:
		return ReferenceOption_NoAction
	case schema.Restrict:
		return ReferenceOption_Restrict
	case schema.Cascade:
		return ReferenceOption_Cascade
	case schema.SetNull:
		return ReferenceOption_SetNull
	case schema.SetDefault:
		return ReferenceOption_SetDefault
	}
	return ReferenceOption_NoAction
}

func loadSchema(r *schema.Realm, pb *Schema) {
	s := r.GetOrCreateSchema(pb.Name)

	for _, t := range pb.Tables {
		loadTable(r, t)
	}
	for _, e := range pb.Enums {
		loadEnum(r, e)
	}

	s.AddAttrs(toAttrs(pb.Attrs)...)
}

func loadTable(r *schema.Realm, pb *Table) {
	t := r.GetOrCreateTable(pb.QualName.Schema, pb.QualName.Name)
	for _, c := range pb.Columns {
		loadColumn(r, t, c)
	}

	for _, i := range pb.Indexes {
		loadIndex(t, i)
	}

	if pb.PrimaryKey != nil {
		t.SetPrimaryKey(toIndex(t, pb.PrimaryKey))
	}

	fks := make([]*schema.ForeignKey, len(pb.ForeignKeys))
	for i, fk := range pb.ForeignKeys {
		fks[i] = toForeignKey(r, fk)
	}
	t.AddForeignKeys(fks...)
	t.AddAttrs(toAttrs(pb.Attrs)...)
}

func loadColumn(r *schema.Realm, t *schema.Table, pb *Column) {
	c := t.GetOrCreateColumn(pb.Name)
	c.Type = toColumnType(pb.Type)
	if pb.Default != nil {
		c.Default = toExpr(pb.Default)
	}
	c.AddAttrs(toAttrs(pb.Attrs)...)

	for _, idx := range pb.Indexes {
		loadIndex(t, idx)
	}

	fks := make([]*schema.ForeignKey, len(pb.ForeignKeys))
	for i, fk := range pb.ForeignKeys {
		fks[i] = toForeignKey(r, fk)
	}
	c.AddForeignKeys(fks...)
}

func toColumnType(pb *ColumnType) *schema.ColumnType {
	ct := &schema.ColumnType{
		Raw:      pb.Raw,
		Nullable: pb.Nullable,
	}
	if typ, ok := toType(pb.Type); ok {
		ct.Type = typ
	}
	return ct
}

func toIndex(t *schema.Table, pb *Index) *schema.Index {
	pk := schema.NewIndex(pb.Name).SetUnique(pb.Unique).AddAttrs(toAttrs(pb.Attrs)...)
	for _, c := range pb.Parts {
		pk.AddParts(toIndexPart(t, c))
	}
	return pk
}

func loadIndex(t *schema.Table, pb *Index) {
	index := t.GetOrCreateIndex(pb.Name).SetUnique(pb.Unique).AddAttrs(toAttrs(pb.Attrs)...)
	for _, c := range pb.Parts {
		index.AddParts(toIndexPart(t, c))
	}
}

func toIndexPart(t *schema.Table, pb *IndexPart) *schema.IndexPart {
	p := &schema.IndexPart{
		Seq:        int(pb.Seq),
		Descending: pb.Descending,
	}
	switch part := pb.Value.(type) {
	case *IndexPart_Column:
		p.Column = t.GetOrCreateColumn(part.Column)
	case *IndexPart_Expr:
		p.Expr = toExpr(part.Expr)
	}
	p.AddAttrs(toAttrs(pb.Attrs)...)
	return p
}

func toForeignKey(r *schema.Realm, pb *ForeignKey) *schema.ForeignKey {
	t := r.GetOrCreateTable(pb.Table.Schema, pb.Table.Name)
	fk := schema.NewForeignKey(pb.Symbol)

	cols := make([]*schema.Column, len(pb.Columns))
	for i, c := range pb.Columns {
		cols[i] = t.GetOrCreateColumn(c)
	}
	fk.AddColumns(cols...)

	if pb.RefTable != nil {
		rt := r.GetOrCreateTable(pb.RefTable.Schema, pb.RefTable.Name)
		fk.SetRefTable(rt)

		rcols := make([]*schema.Column, len(pb.RefColumns))
		for i, c := range pb.RefColumns {
			rcols[i] = rt.GetOrCreateColumn(c)
		}
		fk.AddRefColumns(rcols...)
	}

	fk.SetOnUpdate(toForeignKeyAction(pb.OnUpdate))
	fk.SetOnDelete(toForeignKeyAction(pb.OnDelete))

	return fk
}

func toForeignKeyAction(pb ReferenceOption) schema.ReferenceOption {
	switch pb {
	case ReferenceOption_Cascade:
		return schema.Cascade
	case ReferenceOption_SetNull:
		return schema.SetNull
	case ReferenceOption_SetDefault:
		return schema.SetDefault
	case ReferenceOption_NoAction:
		return schema.NoAction
	case ReferenceOption_Restrict:
		return schema.Restrict
	}
	return schema.NoAction
}

func loadEnum(r *schema.Realm, pb *Enum) {
	r.GetOrCreateSchema(pb.QualName.Schema).
		GetOrCreateEnum(pb.QualName.Name).
		AddValues(pb.Values...)
}

func loadQuery(r *schema.Realm, pb *Query) {
	r.GetOrCreateQuery(pb.Name).
		SetStatement(pb.Statement)
}

func loadRole(r *schema.Realm, pb *Role) {
	rol := r.GetOrCreateRole(pb.Name).SetDefault(pb.Default)
	queries := make([]*schema.Query, len(pb.Queries))
	for i, q := range pb.Queries {
		queries[i] = r.GetOrCreateQuery(q)
	}
	rol.AddQueries(queries...)
}

func toAttr(pb *Attr) (schema.Attr, bool) {
	switch a := pb.Value.(type) {
	case *Attr_CharsetAttr:
		return &schema.Charset{V: a.CharsetAttr.Value}, true
	case *Attr_CollationAttr:
		return &schema.Collation{V: a.CollationAttr.Value}, true
	case *Attr_CommentAttr:
		return &schema.Comment{Text: a.CommentAttr.Text}, true
	case *Attr_CheckAttr:
		return &schema.Check{Name: a.CheckAttr.Name, Expr: a.CheckAttr.Expr, Attrs: toAttrs(a.CheckAttr.Attrs)}, true
	case *Attr_GeneratedExprAttr:
		return &schema.GeneratedExpr{Expr: a.GeneratedExprAttr.Expr, Type: a.GeneratedExprAttr.Type}, true
	case *Attr_ConstraintTypeAttr:
		return &postgres.ConstraintType{T: a.ConstraintTypeAttr.Type}, true
	case *Attr_IdentityAttr:
		seq := &postgres.Sequence{Start: a.IdentityAttr.Sequence.Start, Last: a.IdentityAttr.Sequence.Last, Increment: a.IdentityAttr.Sequence.Increment}
		return &postgres.Identity{Generation: a.IdentityAttr.Generation, Sequence: seq}, true
	case *Attr_IndexTypeAttr:
		return &postgres.IndexType{T: a.IndexTypeAttr.Type}, true
	case *Attr_IndexPredicateAttr:
		return &postgres.IndexPredicate{Predicate: a.IndexPredicateAttr.Predicate}, true
	case *Attr_IndexColumnPropertyAttr:
		return &postgres.IndexColumnProperty{NullsFirst: a.IndexColumnPropertyAttr.NullsFirst, NullsLast: a.IndexColumnPropertyAttr.NullsLast}, true
	case *Attr_IndexStorageParamsAttr:
		return &postgres.IndexStorageParams{AutoSummarize: a.IndexStorageParamsAttr.AutoSummarize, PagesPerRange: a.IndexStorageParamsAttr.PagesPerRange}, true
	case *Attr_IndexIncludeAttr:
		return &postgres.IndexInclude{Columns: a.IndexIncludeAttr.Columns}, true
	case *Attr_NoInheritAttr:
		return &postgres.NoInherit{}, true
	case *Attr_CheckColumnsAttr:
		return &postgres.CheckColumns{Columns: a.CheckColumnsAttr.Columns}, true
	case *Attr_PartitionAttr:
		parts := make([]*postgres.PartitionPart, len(a.PartitionAttr.Parts))
		for i, p := range a.PartitionAttr.Parts {
			parts[i] = &postgres.PartitionPart{Expr: toExpr(p.Expr), Column: p.Column, Attrs: toAttrs(p.Attrs)}
		}
		return &postgres.Partition{T: a.PartitionAttr.Type, Parts: parts}, true
	default:
		return nil, false
	}
}

func toType(pb *Type) (schema.Type, bool) {
	switch pb := pb.Value.(type) {
	case *Type_EnumType:
		return &schema.EnumType{T: pb.EnumType.Type, Values: pb.EnumType.Values}, true
	case *Type_BinaryType:
		sz := int(pb.BinaryType.GetSize())
		return &schema.BinaryType{T: pb.BinaryType.Type, Size: &sz}, true
	case *Type_StringType:
		sz := int(pb.StringType.GetSize())
		return &schema.StringType{T: pb.StringType.Type, Size: sz}, true
	case *Type_BoolType:
		return &schema.BoolType{T: pb.BoolType.Type}, true
	case *Type_IntegerType:
		return &schema.IntegerType{T: pb.IntegerType.Type, Unsigned: pb.IntegerType.Unsigned, Attrs: toAttrs(pb.IntegerType.Attrs)}, true
	case *Type_TimeType:
		prec := int(pb.TimeType.GetPrecision())
		return &schema.TimeType{T: pb.TimeType.Type, Precision: &prec}, true
	case *Type_SpatialType:
		return &schema.SpatialType{T: pb.SpatialType.Type}, true
	case *Type_DecimalType:
		return &schema.DecimalType{T: pb.DecimalType.Type, Precision: int(pb.DecimalType.Precision), Scale: int(pb.DecimalType.Scale), Unsigned: pb.DecimalType.Unsigned}, true
	case *Type_FloatType:
		return &schema.FloatType{T: pb.FloatType.Type, Unsigned: pb.FloatType.Unsigned, Precision: int(pb.FloatType.Precision)}, true
	case *Type_JsonType:
		return &schema.JSONType{T: pb.JsonType.Type}, true
	case *Type_UnsupportedType:
		return &schema.UnsupportedType{T: pb.UnsupportedType.Type}, true
	case *Type_UserDefinedType:
		return postgres.UserDefinedType{T: pb.UserDefinedType.Type}, true
	case *Type_ArrayType:
		return postgres.ArrayType{T: pb.ArrayType.Type}, true
	case *Type_BitType:
		return &postgres.BitType{T: pb.BitType.Type, Width: pb.BitType.Width}, true
	case *Type_IntervalType:
		return &postgres.IntervalType{T: pb.IntervalType.Type, F: pb.IntervalType.Field}, true
	case *Type_NetworkType:
		return &postgres.NetworkType{T: pb.NetworkType.Type, Width: pb.NetworkType.Width}, true
	case *Type_CurrencyType:
		return &postgres.CurrencyType{T: pb.CurrencyType.Type}, true
	case *Type_UuidType:
		return &postgres.UUIDType{T: pb.UuidType.Type}, true
	case *Type_XmlType:
		return &postgres.XMLType{T: pb.XmlType.Type}, true
	default:
		return nil, false
	}
}

func toi64ptr(v *int) *int64 {
	if v == nil {
		return nil
	}
	o := int64(*v)
	return &o
}

func fromType(t schema.Type) *Type {
	switch t := t.(type) {
	case *schema.EnumType:
		return &Type{Value: &Type_EnumType{EnumType: &EnumType{Type: t.T, Values: t.Values}}}
	case *schema.BinaryType:
		return &Type{Value: &Type_BinaryType{BinaryType: &BinaryType{Type: t.T, Size: toi64ptr(t.Size)}}}
	case *schema.StringType:
		return &Type{Value: &Type_StringType{StringType: &StringType{Type: t.T, Size: int64(t.Size)}}}
	case *schema.BoolType:
		return &Type{Value: &Type_BoolType{BoolType: &BoolType{Type: t.T}}}
	case *schema.IntegerType:
		return &Type{Value: &Type_IntegerType{IntegerType: &IntegerType{Type: t.T, Unsigned: t.Unsigned, Attrs: fromAttrs(t.Attrs)}}}
	case *schema.TimeType:
		return &Type{Value: &Type_TimeType{TimeType: &TimeType{Type: t.T, Precision: toi64ptr(t.Precision)}}}
	case *schema.SpatialType:
		return &Type{Value: &Type_SpatialType{SpatialType: &SpatialType{Type: t.T}}}
	case *schema.DecimalType:
		return &Type{Value: &Type_DecimalType{DecimalType: &DecimalType{Type: t.T, Precision: int64(t.Precision), Scale: int64(t.Scale), Unsigned: t.Unsigned}}}
	case *schema.FloatType:
		return &Type{Value: &Type_FloatType{FloatType: &FloatType{Type: t.T, Unsigned: t.Unsigned, Precision: int64(t.Precision)}}}
	case *schema.JSONType:
		return &Type{Value: &Type_JsonType{JsonType: &JsonType{Type: t.T}}}
	case *schema.UnsupportedType:
		return &Type{Value: &Type_UnsupportedType{UnsupportedType: &UnsupportedType{Type: t.T}}}
	case *postgres.UserDefinedType:
		return &Type{Value: &Type_UserDefinedType{UserDefinedType: &UserDefinedType{Type: t.T}}}
	case *postgres.ArrayType:
		return &Type{Value: &Type_ArrayType{ArrayType: &ArrayType{Type: t.T}}}
	case *postgres.BitType:
		return &Type{Value: &Type_BitType{BitType: &BitType{Type: t.T, Width: t.Width}}}
	case *postgres.IntervalType:
		var precision *int64
		if t.Precision != nil {
			prec := int64(*t.Precision)
			precision = &prec
		}
		return &Type{Value: &Type_IntervalType{IntervalType: &IntervalType{Type: t.T, Precision: precision}}}
	case *postgres.NetworkType:
		return &Type{Value: &Type_NetworkType{NetworkType: &NetworkType{Type: t.T, Width: t.Width}}}
	case *postgres.CurrencyType:
		return &Type{Value: &Type_CurrencyType{CurrencyType: &CurrencyType{Type: t.T}}}
	case *postgres.SerialType:
		return &Type{Value: &Type_SerialType{SerialType: &SerialType{Type: t.T, Precision: int64(t.Precision), SequenceName: t.SequenceName}}}
	case *postgres.UUIDType:
		return &Type{Value: &Type_UuidType{UuidType: &UUIDType{Type: t.T}}}
	case *postgres.XMLType:
		return &Type{Value: &Type_XmlType{XmlType: &XMLType{Type: t.T}}}
	default:
		return nil
	}
}

func toExpr(pb *Expr) schema.Expr {
	switch e := pb.Value.(type) {
	case *Expr_Literal:
		return &schema.Literal{V: e.Literal.Value}
	case *Expr_RawExpr:
		return &schema.RawExpr{X: e.RawExpr.Expr}
	}
	return nil
}

func toAttrs(pb []*Attr) []schema.Attr {
	if pb == nil {
		return nil
	}

	attrs := make([]schema.Attr, 0, len(pb))
	for _, a := range pb {
		if attr, ok := toAttr(a); ok {
			attrs = append(attrs, attr)
		}
	}
	return attrs
}
