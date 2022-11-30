package dml

import (
	"ksl"

	"github.com/samber/mo"
)

type Field interface{ field() }

type ScalarField struct {
	Name          string
	Type          FieldType
	Arity         FieldArity
	DatabaseName  string
	DefaultValue  mo.Option[string]
	Documentation string
	IsGenerated   bool
	IsIgnored     bool
}

type RelationField struct {
	Name             string
	RelationInfo     *RelationInfo
	Arity            FieldArity
	ReferentialArity FieldArity
	Documentation    string
	IsGenerated      bool
	IsIgnored        bool
	SupportsRestrict mo.Option[bool]
}

func (ScalarField) field()   {}
func (RelationField) field() {}

type FieldType interface{ fieldtyp() }

type ScalarFieldType struct {
	Type       ksl.Type
	NativeType mo.Option[ksl.Type]
}

type EnumFieldType struct {
	Enum string
}

type RelationFieldType struct {
	Relation RelationInfo
}

func (ScalarFieldType) fieldtyp()   {}
func (EnumFieldType) fieldtyp()     {}
func (RelationFieldType) fieldtyp() {}
