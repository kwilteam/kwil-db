package pgschemapb

import (
	"kwil/x/schemadef/schema"
	"kwil/x/sql/postgres"
)

func (pb *Attr) ToSchema() (schema.Attr, bool) {
	switch attr := pb.Value.(type) {
	case *Attr_ConstraintTypeAttr:
		return attr.ConstraintTypeAttr.ToSchema(), true
	case *Attr_IdentityAttr:
		return attr.IdentityAttr.ToSchema(), true
	case *Attr_IndexTypeAttr:
		return attr.IndexTypeAttr.ToSchema(), true
	case *Attr_IndexPredicateAttr:
		return attr.IndexPredicateAttr.ToSchema(), true
	case *Attr_IndexColumnPropertyAttr:
		return attr.IndexColumnPropertyAttr.ToSchema(), true
	case *Attr_IndexStorageParamsAttr:
		return attr.IndexStorageParamsAttr.ToSchema(), true
	case *Attr_IndexIncludeAttr:
		return attr.IndexIncludeAttr.ToSchema(), true
	case *Attr_NoInheritAttr:
		return attr.NoInheritAttr.ToSchema(), true
	case *Attr_CheckColumnsAttr:
		return attr.CheckColumnsAttr.ToSchema(), true
	case *Attr_PartitionAttr:
		return attr.PartitionAttr.ToSchema(), true
	default:
		return nil, false
	}
}

func (pb *ConstraintTypeAttr) ToSchema() schema.Attr {
	return &postgres.ConstraintType{T: pb.Type}
}
func (pb *IdentityAttr) ToSchema() schema.Attr {
	seq := &postgres.Sequence{Start: pb.Sequence.Start, Last: pb.Sequence.Last, Increment: pb.Sequence.Increment}
	return &postgres.Identity{Generation: pb.Generation, Sequence: seq}
}
func (pb *IndexTypeAttr) ToSchema() schema.Attr {
	return &postgres.IndexType{T: pb.Type}
}
func (pb *IndexPredicateAttr) ToSchema() schema.Attr {
	return &postgres.IndexPredicate{Predicate: pb.Predicate}
}
func (pb *IndexColumnPropertyAttr) ToSchema() schema.Attr {
	return &postgres.IndexColumnProperty{NullsFirst: pb.NullsFirst, NullsLast: pb.NullsLast}
}
func (pb *IndexStorageParamsAttr) ToSchema() schema.Attr {
	return &postgres.IndexStorageParams{AutoSummarize: pb.AutoSummarize, PagesPerRange: pb.PagesPerRange}
}
func (pb *IndexIncludeAttr) ToSchema() schema.Attr {
	return &postgres.IndexInclude{}
}
func (pb *NoInheritAttr) ToSchema() schema.Attr {
	return &postgres.NoInherit{}
}
func (pb *CheckColumnsAttr) ToSchema() schema.Attr {
	return &postgres.CheckColumns{}
}
func (pb *PartitionAttr) ToSchema() schema.Attr {
	return &postgres.Partition{}
}

func (pb *Type) ToSchema() (schema.Type, bool) {
	switch t := pb.Value.(type) {
	case *Type_CType:
		return t.CType.ToSchema(), true
	case *Type_UserDefinedType:
		return t.UserDefinedType.ToSchema(), true
	case *Type_ArrayType:
		return t.ArrayType.ToSchema(), true
	case *Type_BitType:
		return t.BitType.ToSchema(), true
	case *Type_IntervalType:
		return t.IntervalType.ToSchema(), true
	case *Type_NetworkType:
		return t.NetworkType.ToSchema(), true
	case *Type_CurrencyType:
		return t.CurrencyType.ToSchema(), true
	case *Type_UuidType:
		return t.UuidType.ToSchema(), true
	case *Type_XmlType:
		return t.XmlType.ToSchema(), true
	default:
		return nil, false
	}
}

func (pb *CType) ToSchema() schema.Type {
	return nil
}
func (pb *UserDefinedType) ToSchema() schema.Type {
	return nil
}
func (pb *ArrayType) ToSchema() schema.Type {
	return nil
}
func (pb *BitType) ToSchema() schema.Type {
	return nil
}
func (pb *IntervalType) ToSchema() schema.Type {
	return nil
}
func (pb *NetworkType) ToSchema() schema.Type {
	return nil
}
func (pb *CurrencyType) ToSchema() schema.Type {
	return nil
}
func (pb *UUIDType) ToSchema() schema.Type {
	return nil
}
func (pb *XMLType) ToSchema() schema.Type {
	return nil
}
