package schemapb

import (
	"kwil/x/schemadef/pgschema"
	"kwil/x/schemadef/sqlschema"
)

func ToChanges(changes []*Change) []sqlschema.SchemaChange {
	out := make([]sqlschema.SchemaChange, len(changes))
	for i, c := range changes {
		out[i] = ToChange(c)
	}
	return out
}

func ToChange(c *Change) sqlschema.SchemaChange {
	switch c.Value.(type) {
	case *Change_AddSchema:
		return &sqlschema.AddSchema{}
	case *Change_DropSchema:
		return &sqlschema.DropSchema{}
	case *Change_ModifySchema:
		return &sqlschema.ModifySchema{}
	case *Change_AddTable:
		return &sqlschema.AddTable{}
	case *Change_DropTable:
		return &sqlschema.DropTable{}
	case *Change_ModifyTable:
		return &sqlschema.ModifyTable{}
	case *Change_RenameTable:
		return &sqlschema.RenameTable{}
	case *Change_AddColumn:
		return &sqlschema.AddColumn{}
	case *Change_DropColumn:
		return &sqlschema.DropColumn{}
	case *Change_ModifyColumn:
		return &sqlschema.ModifyColumn{}
	case *Change_RenameColumn:
		return &sqlschema.RenameColumn{}
	case *Change_AddIndex:
		return &sqlschema.AddIndex{}
	case *Change_DropIndex:
		return &sqlschema.DropIndex{}
	case *Change_ModifyIndex:
		return &sqlschema.ModifyIndex{}
	case *Change_RenameIndex:
		return &sqlschema.RenameIndex{}
	case *Change_AddForeignKey:
		return &sqlschema.AddForeignKey{}
	case *Change_DropForeignKey:
		return &sqlschema.DropForeignKey{}
	case *Change_ModifyForeignKey:
		return &sqlschema.ModifyForeignKey{}
	case *Change_AddCheck:
		return &sqlschema.AddCheck{}
	case *Change_DropCheck:
		return &sqlschema.DropCheck{}
	case *Change_ModifyCheck:
		return &sqlschema.ModifyCheck{}
	case *Change_AddAttr:
		return &sqlschema.AddAttr{}
	case *Change_DropAttr:
		return &sqlschema.DropAttr{}
	case *Change_ModifyAttr:
		return &sqlschema.ModifyAttr{}
	default:
		return nil
	}
}

func ToRealm(pb *Realm) *sqlschema.Realm {
	r := sqlschema.NewRealm()
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

func FromRealm(r *sqlschema.Realm) *Realm {
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

func fromQuery(q *sqlschema.Query) *Query {
	return &Query{
		Name:      q.Name,
		Statement: q.Statement,
	}
}

func fromRole(r *sqlschema.Role) *Role {
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

func fromSchema(s *sqlschema.Schema) *Schema {
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

func fromEnum(e *sqlschema.Enum) *Enum {
	pb := &Enum{
		QualName: &QualName{
			Schema: e.Schema.Name,
			Name:   e.Name,
		},
	}
	pb.Values = append(pb.Values, e.Values...)

	return pb
}

func fromTable(t *sqlschema.Table) *Table {
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

func fromColumn(c *sqlschema.Column) *Column {
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

func fromColumnType(ct *sqlschema.ColumnType) *ColumnType {
	return &ColumnType{
		Type:     fromType(ct.Type),
		Raw:      ct.Raw,
		Nullable: ct.Nullable,
	}
}

func fromIndex(i *sqlschema.Index) *Index {
	pb := &Index{
		Name: i.Name,
	}
	for _, p := range i.Parts {
		pb.Parts = append(pb.Parts, fromIndexPart(p))
	}

	pb.Attrs = fromAttrs(i.Attrs)

	return pb
}

func fromIndexPart(p *sqlschema.IndexPart) *IndexPart {
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

func fromExpr(e sqlschema.Expr) *Expr {
	pb := &Expr{}
	switch e := e.(type) {
	case *sqlschema.Literal:
		pb.Value = &Expr_Literal{Literal: &Literal{Value: e.V}}
	case *sqlschema.RawExpr:
		pb.Value = &Expr_RawExpr{RawExpr: &RawExpr{Expr: e.X}}
	}
	return pb
}

func fromAttrs(attrs []sqlschema.Attr) []*Attr {
	if attrs == nil {
		return nil
	}

	m := make([]*Attr, len(attrs))
	for i, a := range attrs {
		m[i] = fromAttr(a)
	}
	return m
}

func fromAttr(a sqlschema.Attr) *Attr {
	switch a := a.(type) {
	case *sqlschema.Comment:
		return &Attr{Value: &Attr_CommentAttr{CommentAttr: &Comment{Text: a.Text}}}
	case *sqlschema.Charset:
		return &Attr{Value: &Attr_CharsetAttr{CharsetAttr: &Charset{Value: a.V}}}
	case *sqlschema.Collation:
		return &Attr{Value: &Attr_CollationAttr{CollationAttr: &Collation{Value: a.V}}}
	case *sqlschema.Check:
		return &Attr{Value: &Attr_CheckAttr{CheckAttr: &Check{Name: a.Name, Expr: a.Expr}}}
	case *sqlschema.GeneratedExpr:
		return &Attr{Value: &Attr_GeneratedExprAttr{GeneratedExprAttr: &GeneratedExpr{Expr: a.Expr, Type: a.Type}}}
	case *pgschema.ConstraintType:
		return &Attr{Value: &Attr_ConstraintTypeAttr{ConstraintTypeAttr: &ConstraintType{Type: a.T}}}
	case *pgschema.Identity:
		return &Attr{Value: &Attr_IdentityAttr{IdentityAttr: &Identity{Generation: a.Generation, Sequence: &Sequence{Start: a.Sequence.Start, Last: a.Sequence.Last, Increment: a.Sequence.Increment}}}}
	case *pgschema.IndexType:
		return &Attr{Value: &Attr_IndexTypeAttr{IndexTypeAttr: &IndexType{Type: a.T}}}
	case *pgschema.IndexPredicate:
		return &Attr{Value: &Attr_IndexPredicateAttr{IndexPredicateAttr: &IndexPredicate{Predicate: a.Predicate}}}
	case *pgschema.IndexColumnProperty:
		return &Attr{Value: &Attr_IndexColumnPropertyAttr{IndexColumnPropertyAttr: &IndexColumnProperty{NullsFirst: a.NullsFirst, NullsLast: a.NullsLast}}}
	case *pgschema.IndexStorageParams:
		return &Attr{Value: &Attr_IndexStorageParamsAttr{IndexStorageParamsAttr: &IndexStorageParams{AutoSummarize: a.AutoSummarize, PagesPerRange: a.PagesPerRange}}}
	case *pgschema.IndexInclude:
		return &Attr{Value: &Attr_IndexIncludeAttr{IndexIncludeAttr: &IndexInclude{Columns: a.Columns}}}
	case *pgschema.NoInherit:
		return &Attr{Value: &Attr_NoInheritAttr{NoInheritAttr: &NoInherit{}}}
	case *pgschema.CheckColumns:
		return &Attr{Value: &Attr_CheckColumnsAttr{CheckColumnsAttr: &CheckColumns{Columns: a.Columns}}}
	case *pgschema.Partition:
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

func fromForeignKeys(fks []*sqlschema.ForeignKey) []*ForeignKey {
	m := make([]*ForeignKey, len(fks))
	for i, fk := range fks {
		m[i] = fromForeignKey(fk)
	}
	return m
}

func columnNames(cols []*sqlschema.Column) []string {
	names := make([]string, len(cols))
	for i, c := range cols {
		names[i] = c.Name
	}
	return names
}

func fromForeignKey(fk *sqlschema.ForeignKey) *ForeignKey {
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

func fromForeignKeyAction(a sqlschema.ReferenceOption) ReferenceOption {
	switch a {
	case sqlschema.NoAction:
		return ReferenceOption_NoAction
	case sqlschema.Restrict:
		return ReferenceOption_Restrict
	case sqlschema.Cascade:
		return ReferenceOption_Cascade
	case sqlschema.SetNull:
		return ReferenceOption_SetNull
	case sqlschema.SetDefault:
		return ReferenceOption_SetDefault
	}
	return ReferenceOption_NoAction
}

func loadSchema(r *sqlschema.Realm, pb *Schema) {
	s := r.GetOrCreateSchema(pb.Name)

	for _, t := range pb.Tables {
		loadTable(r, t)
	}
	for _, e := range pb.Enums {
		loadEnum(r, e)
	}

	s.AddAttrs(toAttrs(pb.Attrs)...)
}

func loadTable(r *sqlschema.Realm, pb *Table) {
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

	fks := make([]*sqlschema.ForeignKey, len(pb.ForeignKeys))
	for i, fk := range pb.ForeignKeys {
		fks[i] = toForeignKey(r, fk)
	}
	t.AddForeignKeys(fks...)
	t.AddAttrs(toAttrs(pb.Attrs)...)
}

func loadColumn(r *sqlschema.Realm, t *sqlschema.Table, pb *Column) {
	c := t.GetOrCreateColumn(pb.Name)
	c.Type = toColumnType(pb.Type)
	if pb.Default != nil {
		c.Default = toExpr(pb.Default)
	}
	c.AddAttrs(toAttrs(pb.Attrs)...)

	for _, idx := range pb.Indexes {
		loadIndex(t, idx)
	}

	fks := make([]*sqlschema.ForeignKey, len(pb.ForeignKeys))
	for i, fk := range pb.ForeignKeys {
		fks[i] = toForeignKey(r, fk)
	}
	c.AddForeignKeys(fks...)
}

func toColumnType(pb *ColumnType) *sqlschema.ColumnType {
	ct := &sqlschema.ColumnType{
		Raw:      pb.Raw,
		Nullable: pb.Nullable,
	}
	if typ, ok := toType(pb.Type); ok {
		ct.Type = typ
	}
	return ct
}

func toIndex(t *sqlschema.Table, pb *Index) *sqlschema.Index {
	pk := sqlschema.NewIndex(pb.Name).SetUnique(pb.Unique).AddAttrs(toAttrs(pb.Attrs)...)
	for _, c := range pb.Parts {
		pk.AddParts(toIndexPart(t, c))
	}
	return pk
}

func loadIndex(t *sqlschema.Table, pb *Index) {
	index := t.GetOrCreateIndex(pb.Name).SetUnique(pb.Unique).AddAttrs(toAttrs(pb.Attrs)...)
	for _, c := range pb.Parts {
		index.AddParts(toIndexPart(t, c))
	}
}

func toIndexPart(t *sqlschema.Table, pb *IndexPart) *sqlschema.IndexPart {
	p := &sqlschema.IndexPart{
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

func toForeignKey(r *sqlschema.Realm, pb *ForeignKey) *sqlschema.ForeignKey {
	t := r.GetOrCreateTable(pb.Table.Schema, pb.Table.Name)
	fk := sqlschema.NewForeignKey(pb.Symbol)

	cols := make([]*sqlschema.Column, len(pb.Columns))
	for i, c := range pb.Columns {
		cols[i] = t.GetOrCreateColumn(c)
	}
	fk.AddColumns(cols...)

	if pb.RefTable != nil {
		rt := r.GetOrCreateTable(pb.RefTable.Schema, pb.RefTable.Name)
		fk.SetRefTable(rt)

		rcols := make([]*sqlschema.Column, len(pb.RefColumns))
		for i, c := range pb.RefColumns {
			rcols[i] = rt.GetOrCreateColumn(c)
		}
		fk.AddRefColumns(rcols...)
	}

	fk.SetOnUpdate(toForeignKeyAction(pb.OnUpdate))
	fk.SetOnDelete(toForeignKeyAction(pb.OnDelete))

	return fk
}

func toForeignKeyAction(pb ReferenceOption) sqlschema.ReferenceOption {
	switch pb {
	case ReferenceOption_Cascade:
		return sqlschema.Cascade
	case ReferenceOption_SetNull:
		return sqlschema.SetNull
	case ReferenceOption_SetDefault:
		return sqlschema.SetDefault
	case ReferenceOption_NoAction:
		return sqlschema.NoAction
	case ReferenceOption_Restrict:
		return sqlschema.Restrict
	}
	return sqlschema.NoAction
}

func loadEnum(r *sqlschema.Realm, pb *Enum) {
	r.GetOrCreateSchema(pb.QualName.Schema).
		GetOrCreateEnum(pb.QualName.Name).
		AddValues(pb.Values...)
}

func loadQuery(r *sqlschema.Realm, pb *Query) {
	r.GetOrCreateQuery(pb.Name).
		SetStatement(pb.Statement)
}

func loadRole(r *sqlschema.Realm, pb *Role) {
	rol := r.GetOrCreateRole(pb.Name).SetDefault(pb.Default)
	queries := make([]*sqlschema.Query, len(pb.Queries))
	for i, q := range pb.Queries {
		queries[i] = r.GetOrCreateQuery(q)
	}
	rol.AddQueries(queries...)
}

func toAttr(pb *Attr) (sqlschema.Attr, bool) {
	switch a := pb.Value.(type) {
	case *Attr_CharsetAttr:
		return &sqlschema.Charset{V: a.CharsetAttr.Value}, true
	case *Attr_CollationAttr:
		return &sqlschema.Collation{V: a.CollationAttr.Value}, true
	case *Attr_CommentAttr:
		return &sqlschema.Comment{Text: a.CommentAttr.Text}, true
	case *Attr_CheckAttr:
		return &sqlschema.Check{Name: a.CheckAttr.Name, Expr: a.CheckAttr.Expr, Attrs: toAttrs(a.CheckAttr.Attrs)}, true
	case *Attr_GeneratedExprAttr:
		return &sqlschema.GeneratedExpr{Expr: a.GeneratedExprAttr.Expr, Type: a.GeneratedExprAttr.Type}, true
	case *Attr_ConstraintTypeAttr:
		return &pgschema.ConstraintType{T: a.ConstraintTypeAttr.Type}, true
	case *Attr_IdentityAttr:
		seq := &pgschema.Sequence{Start: a.IdentityAttr.Sequence.Start, Last: a.IdentityAttr.Sequence.Last, Increment: a.IdentityAttr.Sequence.Increment}
		return &pgschema.Identity{Generation: a.IdentityAttr.Generation, Sequence: seq}, true
	case *Attr_IndexTypeAttr:
		return &pgschema.IndexType{T: a.IndexTypeAttr.Type}, true
	case *Attr_IndexPredicateAttr:
		return &pgschema.IndexPredicate{Predicate: a.IndexPredicateAttr.Predicate}, true
	case *Attr_IndexColumnPropertyAttr:
		return &pgschema.IndexColumnProperty{NullsFirst: a.IndexColumnPropertyAttr.NullsFirst, NullsLast: a.IndexColumnPropertyAttr.NullsLast}, true
	case *Attr_IndexStorageParamsAttr:
		return &pgschema.IndexStorageParams{AutoSummarize: a.IndexStorageParamsAttr.AutoSummarize, PagesPerRange: a.IndexStorageParamsAttr.PagesPerRange}, true
	case *Attr_IndexIncludeAttr:
		return &pgschema.IndexInclude{Columns: a.IndexIncludeAttr.Columns}, true
	case *Attr_NoInheritAttr:
		return &pgschema.NoInherit{}, true
	case *Attr_CheckColumnsAttr:
		return &pgschema.CheckColumns{Columns: a.CheckColumnsAttr.Columns}, true
	case *Attr_PartitionAttr:
		parts := make([]*pgschema.PartitionPart, len(a.PartitionAttr.Parts))
		for i, p := range a.PartitionAttr.Parts {
			parts[i] = &pgschema.PartitionPart{Expr: toExpr(p.Expr), Column: p.Column, Attrs: toAttrs(p.Attrs)}
		}
		return &pgschema.Partition{T: a.PartitionAttr.Type, Parts: parts}, true
	default:
		return nil, false
	}
}

func toType(pb *Type) (sqlschema.Type, bool) {
	switch pb := pb.Value.(type) {
	case *Type_EnumType:
		return &sqlschema.EnumType{T: pb.EnumType.Type, Values: pb.EnumType.Values}, true
	case *Type_BinaryType:
		sz := int(pb.BinaryType.GetSize())
		return &sqlschema.BinaryType{T: pb.BinaryType.Type, Size: &sz}, true
	case *Type_StringType:
		sz := int(pb.StringType.GetSize())
		return &sqlschema.StringType{T: pb.StringType.Type, Size: sz}, true
	case *Type_BoolType:
		return &sqlschema.BoolType{T: pb.BoolType.Type}, true
	case *Type_IntegerType:
		return &sqlschema.IntegerType{T: pb.IntegerType.Type, Unsigned: pb.IntegerType.Unsigned, Attrs: toAttrs(pb.IntegerType.Attrs)}, true
	case *Type_TimeType:
		prec := int(pb.TimeType.GetPrecision())
		return &sqlschema.TimeType{T: pb.TimeType.Type, Precision: &prec}, true
	case *Type_SpatialType:
		return &sqlschema.SpatialType{T: pb.SpatialType.Type}, true
	case *Type_DecimalType:
		return &sqlschema.DecimalType{T: pb.DecimalType.Type, Precision: int(pb.DecimalType.Precision), Scale: int(pb.DecimalType.Scale), Unsigned: pb.DecimalType.Unsigned}, true
	case *Type_FloatType:
		return &sqlschema.FloatType{T: pb.FloatType.Type, Unsigned: pb.FloatType.Unsigned, Precision: int(pb.FloatType.Precision)}, true
	case *Type_JsonType:
		return &sqlschema.JSONType{T: pb.JsonType.Type}, true
	case *Type_UnsupportedType:
		return &sqlschema.UnsupportedType{T: pb.UnsupportedType.Type}, true
	case *Type_UserDefinedType:
		return pgschema.UserDefinedType{T: pb.UserDefinedType.Type}, true
	case *Type_ArrayType:
		return pgschema.ArrayType{T: pb.ArrayType.Type}, true
	case *Type_BitType:
		return &pgschema.BitType{T: pb.BitType.Type, Size: pb.BitType.Width}, true
	case *Type_IntervalType:
		return &pgschema.IntervalType{T: pb.IntervalType.Type, F: pb.IntervalType.Field}, true
	case *Type_NetworkType:
		return &pgschema.NetworkType{T: pb.NetworkType.Type, Size: pb.NetworkType.Width}, true
	case *Type_CurrencyType:
		return &pgschema.CurrencyType{T: pb.CurrencyType.Type}, true
	case *Type_UuidType:
		return &pgschema.UUIDType{T: pb.UuidType.Type}, true
	case *Type_XmlType:
		return &pgschema.XMLType{T: pb.XmlType.Type}, true
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

func fromType(t sqlschema.Type) *Type {
	switch t := t.(type) {
	case *sqlschema.EnumType:
		return &Type{Value: &Type_EnumType{EnumType: &EnumType{Type: t.T, Values: t.Values}}}
	case *sqlschema.BinaryType:
		return &Type{Value: &Type_BinaryType{BinaryType: &BinaryType{Type: t.T, Size: toi64ptr(t.Size)}}}
	case *sqlschema.StringType:
		return &Type{Value: &Type_StringType{StringType: &StringType{Type: t.T, Size: int64(t.Size)}}}
	case *sqlschema.BoolType:
		return &Type{Value: &Type_BoolType{BoolType: &BoolType{Type: t.T}}}
	case *sqlschema.IntegerType:
		return &Type{Value: &Type_IntegerType{IntegerType: &IntegerType{Type: t.T, Unsigned: t.Unsigned, Attrs: fromAttrs(t.Attrs)}}}
	case *sqlschema.TimeType:
		return &Type{Value: &Type_TimeType{TimeType: &TimeType{Type: t.T, Precision: toi64ptr(t.Precision)}}}
	case *sqlschema.SpatialType:
		return &Type{Value: &Type_SpatialType{SpatialType: &SpatialType{Type: t.T}}}
	case *sqlschema.DecimalType:
		return &Type{Value: &Type_DecimalType{DecimalType: &DecimalType{Type: t.T, Precision: int64(t.Precision), Scale: int64(t.Scale), Unsigned: t.Unsigned}}}
	case *sqlschema.FloatType:
		return &Type{Value: &Type_FloatType{FloatType: &FloatType{Type: t.T, Unsigned: t.Unsigned, Precision: int64(t.Precision)}}}
	case *sqlschema.JSONType:
		return &Type{Value: &Type_JsonType{JsonType: &JsonType{Type: t.T}}}
	case *sqlschema.UnsupportedType:
		return &Type{Value: &Type_UnsupportedType{UnsupportedType: &UnsupportedType{Type: t.T}}}
	case *pgschema.UserDefinedType:
		return &Type{Value: &Type_UserDefinedType{UserDefinedType: &UserDefinedType{Type: t.T}}}
	case *pgschema.ArrayType:
		return &Type{Value: &Type_ArrayType{ArrayType: &ArrayType{Type: t.T}}}
	case *pgschema.BitType:
		return &Type{Value: &Type_BitType{BitType: &BitType{Type: t.T, Width: t.Size}}}
	case *pgschema.IntervalType:
		var precision *int64
		if t.Precision != nil {
			prec := int64(*t.Precision)
			precision = &prec
		}
		return &Type{Value: &Type_IntervalType{IntervalType: &IntervalType{Type: t.T, Precision: precision}}}
	case *pgschema.NetworkType:
		return &Type{Value: &Type_NetworkType{NetworkType: &NetworkType{Type: t.T, Width: t.Size}}}
	case *pgschema.CurrencyType:
		return &Type{Value: &Type_CurrencyType{CurrencyType: &CurrencyType{Type: t.T}}}
	case *pgschema.SerialType:
		return &Type{Value: &Type_SerialType{SerialType: &SerialType{Type: t.T, Precision: int64(t.Precision), SequenceName: t.SequenceName}}}
	case *pgschema.UUIDType:
		return &Type{Value: &Type_UuidType{UuidType: &UUIDType{Type: t.T}}}
	case *pgschema.XMLType:
		return &Type{Value: &Type_XmlType{XmlType: &XMLType{Type: t.T}}}
	default:
		return nil
	}
}

func toExpr(pb *Expr) sqlschema.Expr {
	switch e := pb.Value.(type) {
	case *Expr_Literal:
		return &sqlschema.Literal{V: e.Literal.Value}
	case *Expr_RawExpr:
		return &sqlschema.RawExpr{X: e.RawExpr.Expr}
	}
	return nil
}

func toAttrs(pb []*Attr) []sqlschema.Attr {
	if pb == nil {
		return nil
	}

	attrs := make([]sqlschema.Attr, 0, len(pb))
	for _, a := range pb {
		if attr, ok := toAttr(a); ok {
			attrs = append(attrs, attr)
		}
	}
	return attrs
}
