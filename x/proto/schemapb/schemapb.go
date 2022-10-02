package schemapb

import (
	"kwil/x/schemadef/schema"
)

func ToRealm(pb *Realm) *schema.Realm {
	schemas := make([]*schema.Schema, len(pb.Schemas))
	for i, s := range pb.Schemas {
		sch := schema.New(s.Name)
		for _, t := range s.Tables {
			tab := schema.NewTable(t.Name)
			for _, c := range t.Columns {
				col := schema.NewColumn(c.Name)
				tab.AddColumns(col)
			}
			sch.AddTables(tab)
		}
		schemas[i] = sch
	}
	r := schema.NewRealm(schemas...)

	r.AddAttrs(attrs(pb.Attrs)...)
	for _, q := range pb.Queries {
		convertToQuery(r, q)
	}

	for _, rol := range pb.Roles {
		convertFomRole(r, rol)
	}

	return r
}

func convertToSchema(r *schema.Realm, pb *Schema) {
	s, ok := r.Schema(pb.Name)
	if !ok {
		s = schema.New(pb.Name)
		r.AddSchemas(s)
	}

	for _, t := range pb.Tables {
		convertToTable(s, t)
	}
	for _, e := range pb.Enums {
		convertEnum(s, e)
	}

	s.AddAttrs(attrs(pb.Attrs)...)
}

func convertToTable(s *schema.Schema, pb *Table) {
	t, ok := s.Table(pb.Name)
	if !ok {
		t = schema.NewTable(pb.Name)
		s.AddTables(t)
	}

	for _, c := range pb.Columns {
		convertToColumn(t, c)
	}

	for _, i := range pb.Indexes {
		convertToIndex(t, i)
	}

	if pb.PrimaryKey != nil {
		t.SetPrimaryKey(convertToPrimaryKey(t, pb.PrimaryKey))
	}

	for _, fk := range pb.ForeignKeys {
		t.AddForeignKeys(convertToForeignKey(t, fk))
	}
	t.AddAttrs(attrs(pb.Attrs)...)
}

func convertToColumn(t *schema.Table, pb *Column) {
	c, ok := t.Column(pb.Name)
	if !ok {
		c = schema.NewColumn(pb.Name)
		t.AddColumns(c)
	}

	c.Type = convertFomColumnType(pb.Type)
	c.Default, _ = convertExpr(pb.Default)
	c.AddAttrs(attrs(pb.Attrs)...)

	for _, idx := range pb.Indexes {
		convertToIndex(t, idx)
	}
	for _, fk := range pb.ForeignKeys {
		convertToForeignKey(t, fk)
	}
}

func convertFomColumnType(pb *ColumnType) *schema.ColumnType {
	ct := &schema.ColumnType{
		Raw:      pb.Raw,
		Nullable: pb.Nullable,
	}
	if typ, ok := convertType(pb.Type); ok {
		ct.Type = typ
	}
	return ct
}

func convertToPrimaryKey(t *schema.Table, pb *Index) *schema.Index {
	pk := schema.NewIndex(pb.Name)
	pk.Unique = pb.Unique
	pk.AddAttrs(attrs(pb.Attrs)...)
	for _, c := range pb.Parts {
		pk.AddParts(convertToIndexPart(t, c))
	}
	return pk
}

func convertToIndex(t *schema.Table, pb *Index) {
	i, ok := t.Index(pb.Name)
	if !ok {
		i = schema.NewIndex(pb.Name)
		t.AddIndexes(i)
	}

	i.Unique = pb.Unique
	i.AddAttrs(attrs(pb.Attrs)...)
	for _, c := range pb.Parts {
		i.AddParts(convertToIndexPart(t, c))
	}
}

func convertToIndexPart(t *schema.Table, pb *IndexPart) *schema.IndexPart {
	p := &schema.IndexPart{
		Seq:        int(pb.Seq),
		Descending: pb.Descending,
	}
	switch part := pb.Value.(type) {
	case *IndexPart_Column:
		if col, ok := t.Column(part.Column); ok {
			p.Column = col
		} else {
			p.Column = schema.NewColumn(part.Column)
		}
	case *IndexPart_Expr:
		p.Expr, _ = convertExpr(part.Expr)
	}
	p.AddAttrs(attrs(pb.Attrs)...)
	return p
}

func convertToForeignKey(t *schema.Table, pb *ForeignKey) *schema.ForeignKey {
	fk := schema.NewForeignKey(pb.Symbol)
	for _, c := range pb.Columns {
		if col, ok := t.Column(c); ok {
			fk.AddColumns(col)
		} else {
			fk.AddColumns(schema.NewColumn(c))
		}
	}

	return fk
}

func convertEnum(s *schema.Schema, pb *Enum) {
	e, ok := s.Enum(pb.Name)
	if !ok {
		e = schema.NewEnum(pb.Name)
		s.AddEnums(e)
	}
	e.AddValues(pb.Values...)
}

func convertToQuery(r *schema.Realm, pb *Query) {
	q, ok := r.Query(pb.Name)
	if !ok {
		q = schema.NewQuery(pb.Name)
		r.AddQueries(q)
	}

	q.SetStatement(pb.Statement)
}

func convertFomRole(r *schema.Realm, pb *Role) {
	rol, ok := r.Role(pb.Name)
	if !ok {
		rol = schema.NewRole(pb.Name)
		r.AddRoles(rol)
	}

	rol.Default = pb.Default
	for _, q := range pb.Queries {
		query, ok := r.Query(q)
		if !ok {
			query = schema.NewQuery(q)
			r.AddQueries(query)
		}
	}
}

func convertToAttr(pb *Attr) (schema.Attr, bool) {
	switch attr := pb.Value.(type) {
	case *Attr_CharsetAttr:
		return ToCharsetAttr(attr.CharsetAttr), true
	case *Attr_CollationAttr:
		return ToCollationAttr(attr.CollationAttr), true
	case *Attr_CommentAttr:
		return ToCommentAttr(attr.CommentAttr), true
	case *Attr_CheckAttr:
		return ToCheckAttr(attr.CheckAttr), true
	case *Attr_GeneratedExprAttr:
		return ToGeneratedExprAttr(attr.GeneratedExprAttr), true
	case *Attr_PgAttr:
		return attr.PgAttr.ToSchema()
	default:
		return nil, false
	}
}

func ToCommentAttr(a *CommentAttr) *schema.Comment {
	return &schema.Comment{Text: a.Text}
}
func ToCharsetAttr(a *CharsetAttr) *schema.Charset {
	return &schema.Charset{V: a.Value}
}
func ToCollationAttr(a *CollationAttr) *schema.Collation {
	return &schema.Collation{V: a.Value}
}
func ToCheckAttr(a *CheckAttr) *schema.Check {
	return &schema.Check{Name: a.Name, Expr: a.Expr, Attrs: attrs(a.Attrs)}
}
func ToGeneratedExprAttr(a *GeneratedExprAttr) *schema.GeneratedExpr {
	return &schema.GeneratedExpr{Expr: a.Expr, Type: a.Type}
}

func convertType(pb *Type) (schema.Type, bool) {
	switch t := pb.Value.(type) {
	case *Type_EnumType:
		return ToEnumType(t.EnumType), true
	case *Type_BinaryType:
		return ToBinaryType(t.BinaryType), true
	case *Type_StringType:
		return ToStringType(t.StringType), true
	case *Type_BoolType:
		return ToBoolType(t.BoolType), true
	case *Type_IntegerType:
		return ToIntegerType(t.IntegerType), true
	case *Type_TimeType:
		return ToTimeType(t.TimeType), true
	case *Type_SpatialType:
		return ToSpatialType(t.SpatialType), true
	case *Type_DecimalType:
		return ToDecimalType(t.DecimalType), true
	case *Type_FloatType:
		return ToFloatType(t.FloatType), true
	case *Type_JsonType:
		return ToJsonType(t.JsonType), true
	case *Type_PgType:
		return t.PgType.ToSchema()
	case *Type_UnsupportedType:
		return ToUnsupportedType(t.UnsupportedType), true
	default:
		return nil, false
	}
}

func ToEnumType(t *EnumType) *schema.EnumType {
	return &schema.EnumType{T: t.Type, Values: t.Values}
}
func ToBinaryType(t *BinaryType) *schema.BinaryType {
	sz := int(t.GetSize())
	return &schema.BinaryType{T: t.Type, Size: &sz}
}
func ToStringType(t *StringType) *schema.StringType {
	sz := int(t.GetSize())
	return &schema.StringType{T: t.Type, Size: sz}
}
func ToBoolType(t *BoolType) *schema.BoolType {
	return &schema.BoolType{T: t.Type}
}
func ToIntegerType(t *IntegerType) *schema.IntegerType {
	return &schema.IntegerType{T: t.Type, Unsigned: t.Unsigned, Attrs: attrs(t.Attrs)}
}
func ToTimeType(t *TimeType) *schema.TimeType {
	prec := int(t.GetPrecision())
	return &schema.TimeType{T: t.Type, Precision: &prec}
}
func ToSpatialType(t *SpatialType) *schema.SpatialType {
	return &schema.SpatialType{T: t.Type}
}
func ToDecimalType(t *DecimalType) *schema.DecimalType {
	return &schema.DecimalType{T: t.Type, Precision: int(t.Precision), Scale: int(t.Scale), Unsigned: t.Unsigned}
}
func ToFloatType(t *FloatType) *schema.FloatType {
	return &schema.FloatType{T: t.Type, Unsigned: t.Unsigned, Precision: int(t.Precision)}
}
func ToJsonType(t *JsonType) *schema.JSONType {
	return &schema.JSONType{T: t.Type}
}
func ToUnsupportedType(t *UnsupportedType) *schema.UnsupportedType {
	return &schema.UnsupportedType{T: t.Type}
}

func convertExpr(pb *Expr) (schema.Expr, bool) {
	switch e := pb.Value.(type) {
	case *Expr_Literal:
		return &schema.Literal{V: e.Literal.Value}, true
	case *Expr_RawExpr:
		return &schema.RawExpr{X: e.RawExpr.Expr}, true
	}
	return nil, false
}

func attrs(pb []*Attr) []schema.Attr {
	attrs := make([]schema.Attr, 0, len(pb))
	for _, a := range pb {
		if attr, ok := convertToAttr(a); ok {
			attrs = append(attrs, attr)
		}
	}
	return attrs
}
