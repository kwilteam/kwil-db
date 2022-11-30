package query

import (
	"ksl"
	"ksl/dml"

	"github.com/samber/mo"
)

type InternalDataModel struct {
	DbName         string
	Models         []Model
	Relations      []Relation
	RelationFields []RelationField
	Enums          []Enum
}

type Model struct {
	Name              string
	Manifestation     mo.Option[string]
	Fields            Fields
	Indexes           []Index
	PrimaryIdentifier FieldSelection
	DmlModel          dml.Model

	InternalModel *InternalDataModel
}

type FieldSelection struct {
	Selections []*ScalarField
}

type Relation struct {
	Name       string
	ModelAName string
	ModelBName string

	ModelA *Model
	ModelB *Model

	FieldA *RelationField
	FieldB *RelationField

	Manifestation RelationLinkManifestation
	InternalModel *InternalDataModel
}

type Field interface{ field() }

type ScalarField struct {
	Name         string
	Type         ksl.Type
	IsID         bool
	Arity        dml.FieldArity
	Enum         mo.Option[*Enum]
	DatabaseName mo.Option[string]
	DefaultValue mo.Option[string]
	NativeType   mo.Option[ksl.Type]
	IsUnique     bool
	ReadOnly     bool

	Model *Model
}

type RelationField struct {
	Name            string
	Arity           dml.FieldArity
	RelationName    string
	RelationSide    *RelationSide
	RelationInfo    *dml.RelationInfo
	OnDeleteDefault dml.ReferentialAction
	OnUpdateDefault dml.ReferentialAction

	Model  *Model
	Fields []*ScalarField
}

func (ScalarField) field()   {}
func (RelationField) field() {}

type Enum struct {
	Name   string
	Values []EnumValue
}

type EnumValue struct {
	Name         string
	DatabaseName mo.Option[string]
}

type Fields struct {
	All        []*Field
	PrimaryKey mo.Option[PrimaryKey]
	Scalar     []*ScalarField
	Relation   []*RelationField

	Model *Model
}

type PrimaryKey struct {
	Alias  mo.Option[string]
	Fields []*ScalarField
}

type Index struct {
	Name   mo.Option[string]
	Fields []*ScalarField
	Type   IndexType
}

type IndexType int

const (
	Normal IndexType = iota
	Unique
)

type RelationLinkManifestation interface{ rlink() }

type RelationLinkManifestationTable struct {
	TableName string
	ColumnA   string
	ColumnB   string
}

type InlineRelation struct {
	InTableOfModelName string
}

func (RelationLinkManifestationTable) rlink() {}
func (InlineRelation) rlink()                 {}

type RelationSide string

const (
	RelationSideA RelationSide = "A"
	RelationSideB RelationSide = "B"
)
