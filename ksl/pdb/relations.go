package pdb

import (
	"fmt"
	"ksl"
	"ksl/syntax/ast"
)

type RelationsContext struct {
	Storage  []Relation
	Forward  map[Rel]struct{}
	Backward map[Rel]struct{}
}

func (r *RelationsContext) FromModel(modelA ModelID) []RelationID {
	var rels []RelationID
	for rel := range r.Forward {
		if rel.A == modelA {
			rels = append(rels, rel.ID)
		}
	}
	return rels
}

func (r *RelationsContext) ToModel(modelA ModelID) []RelationID {
	var rels []RelationID
	for rel := range r.Backward {
		if rel.A == modelA {
			rels = append(rels, rel.ID)
		}
	}
	return rels
}

func (r RelationsContext) Get(id RelationID) Relation {
	return r.Storage[id]
}

func (r RelationsContext) Relations() []RelationItem {
	var items []RelationItem
	for rel := range r.Forward {
		items = append(items, RelationItem{ModelA: rel.A, ModelB: rel.B, Relation: r.Get(rel.ID)})
	}
	return items
}

type Rel struct {
	A  ModelID
	B  ModelID
	ID RelationID
}

type RelationItem struct {
	ModelA   ModelID
	ModelB   ModelID
	Relation Relation
}

type Relation struct {
	Name   string
	Type   RelationType
	ModelA ModelID
	ModelB ModelID
}

func (r Relation) HasField(mfid ModelFieldID) bool {
	switch f := r.Type.(type) {
	case ImplicitManyToMany:
		return (r.ModelA == mfid.Model() && f.FieldA == mfid.Field()) || (r.ModelB == mfid.Model() && f.FieldB == mfid.Field())
	case OneToOneBoth:
		return (r.ModelA == mfid.Model() && f.FieldA == mfid.Field()) || (r.ModelB == mfid.Model() && f.FieldB == mfid.Field())
	case OneToManyBoth:
		return (r.ModelA == mfid.Model() && f.FieldA == mfid.Field()) || (r.ModelB == mfid.Model() && f.FieldB == mfid.Field())
	case OneToOneForward:
		return r.ModelA == mfid.Model() && f.Field == mfid.Field()
	case OneToManyForward:
		return r.ModelA == mfid.Model() && f.Field == mfid.Field()
	case OneToManyBack:
		return r.ModelB == mfid.Model() && f.Field == mfid.Field()
	default:
		return false
	}
}

func (r Relation) IsImplicitManyToMany() bool {
	_, ok := r.Type.(ImplicitManyToMany)
	return ok
}

func (r Relation) IsOneToOneForward() bool {
	_, ok := r.Type.(OneToOneForward)
	return ok
}

func (r Relation) IsOneToOneBoth() bool {
	_, ok := r.Type.(OneToOneBoth)
	return ok
}

func (r Relation) IsOneToManyForward() bool {
	_, ok := r.Type.(OneToManyForward)
	return ok
}

func (r Relation) IsOneToManyBack() bool {
	_, ok := r.Type.(OneToManyBack)
	return ok
}

func (r Relation) IsOneToManyBoth() bool {
	_, ok := r.Type.(OneToManyBoth)
	return ok
}

type RelationEvidence struct {
	Model                 *ast.Model
	Field                 *ast.Field
	RelationField         *RelationField
	IsSelfRelation        bool
	OppositeModel         *ast.Model
	OppositeField         *ast.Field
	OppositeRelationField *RelationField
}

type RefAction struct {
	Action ReferentialAction
	Span   ksl.Range
}

type ReferentialAction string

const (
	NoAction   ReferentialAction = "NoAction"
	Restrict   ReferentialAction = "Restrict"
	Cascade    ReferentialAction = "Cascade"
	SetNull    ReferentialAction = "SetNull"
	SetDefault ReferentialAction = "SetDefault"
)

type RelationField struct {
	Name         string
	ModelID      ModelID
	FieldID      FieldID
	RefModelID   ModelID
	OnUpdate     *RefAction
	OnDelete     *RefAction
	Fields       []FieldRef
	References   []FieldRef
	Ignore       bool
	MappedName   string
	AnnotationID AnnotID
}

func (ctx *context) InferRelations() {
	for mfid, rf := range ctx.Types.RelationFields {
		evidence := ctx.gatherRelationEvidence(mfid, rf)
		ctx.ingestRelation(evidence)
	}
}

func (ctx *context) ingestRelation(evidence *RelationEvidence) {
	hasOppositeRelation := evidence.OppositeRelationField != nil
	var relationType RelationType

	switch arity := evidence.Field.Type.Arity; {
	case evidence.Field.IsRepeated() && hasOppositeRelation && evidence.OppositeField.IsRepeated():
		// this is an implicit many-to-many relation

		// We will meet the relation twice when we walk over all relation
		// fields, so we only instantiate it when the relation field is that
		// of model A, and the opposite is model B.
		if evidence.Model.GetName() > evidence.OppositeModel.GetName() {
			return
		}

		// For self-relations, the ordering logic is different: model A and model B are the same.
		// The lexicographical order is on field names.
		if evidence.IsSelfRelation && evidence.Field.GetName() > evidence.OppositeField.GetName() {
			return
		}

		relationType = ImplicitManyToMany{
			FieldA: evidence.RelationField.FieldID,
			FieldB: evidence.OppositeRelationField.FieldID,
		}

	case evidence.Field.IsRequired() && hasOppositeRelation && evidence.OppositeField.IsOptional():
		// This is a required 1:1 relation, and we are on the required side.
		relationType = OneToOneBoth{
			FieldA: evidence.RelationField.FieldID,
			FieldB: evidence.OppositeRelationField.FieldID,
		}

	case evidence.Field.IsRequired() && hasOppositeRelation && evidence.OppositeField.IsRequired():
		// This is a 1:1 relation that is required on both sides. We are going to reject this later,
		// so which model is model A doesn't matter.
		if fmt.Sprintf("%s.%s", evidence.Model.GetName(), evidence.Field.GetName()) < fmt.Sprintf("%s.%s", evidence.OppositeModel.GetName(), evidence.OppositeField.GetName()) {
			return
		}

		relationType = OneToOneBoth{
			FieldA: evidence.RelationField.FieldID,
			FieldB: evidence.OppositeRelationField.FieldID,
		}

	case evidence.Field.IsOptional() && hasOppositeRelation && evidence.OppositeField.IsRequired():
		// This is a required 1:1 relation, and we are on the virtual side. Skip.
		return

	case evidence.Field.IsOptional() && hasOppositeRelation && evidence.OppositeField.IsOptional():
		// This is a 1:1 relation that is optional on both sides. We must infer which side is model A.
		if len(evidence.RelationField.Fields) > 0 {
			// If the relation field has fields, then model A is the one with the relation field.
			relationType = OneToOneBoth{
				FieldA: evidence.RelationField.FieldID,
				FieldB: evidence.OppositeRelationField.FieldID,
			}
		} else if len(evidence.OppositeRelationField.Fields) == 0 {
			// No fields defined, we have to break the tie: take the first model name / field name (self relations)
			// in lexicographic order.
			if fmt.Sprintf("%s.%s", evidence.Model.GetName(), evidence.Field.GetName()) > fmt.Sprintf("%s.%s", evidence.OppositeModel.GetName(), evidence.OppositeField.GetName()) {
				return
			}

			relationType = OneToOneBoth{
				FieldA: evidence.RelationField.FieldID,
				FieldB: evidence.OppositeRelationField.FieldID,
			}
		} else {
			// Opposite field has fields, it's the forward side. Return.
			return
		}

	case evidence.Field.IsRepeated() && hasOppositeRelation:
		// This is a 1:m relation defined on both sides. We skip the virtual side.
		return

	case evidence.Field.IsRepeated() && !hasOppositeRelation:
		// This is a 1:m relation defined on the virtual side only.
		relationType = OneToManyBack{Field: evidence.RelationField.FieldID}

	case arity.IsAny(ast.Required, ast.Optional) && hasOppositeRelation:
		// This is a 1:m relation defined on both sides.
		relationType = OneToManyBoth{
			FieldA: evidence.RelationField.FieldID,
			FieldB: evidence.OppositeRelationField.FieldID,
		}

	case arity.IsAny(ast.Required, ast.Optional) && !hasOppositeRelation:
		// This is a relation defined on both sides. We check whether the relation scalar fields are unique to
		// determine whether it is a 1:1 or a 1:m relation.

		switch {
		case len(evidence.RelationField.Fields) > 0:
			fieldsAreUnique := false
			for _, idx := range ctx.Types.ModelAnnotations[evidence.RelationField.ModelID].Indexes {
				fields := make([]FieldID, len(idx.Fields))
				for i, f := range evidence.RelationField.Fields {
					fields[i] = f.FieldID
				}

				if idx.IsUnique() && idx.HasFields(fields) {
					fieldsAreUnique = true
					break
				}
			}

			if fieldsAreUnique {
				relationType = OneToOneForward{Field: evidence.RelationField.FieldID}
			} else {
				relationType = OneToManyForward{Field: evidence.RelationField.FieldID}
			}

		default:
			relationType = OneToManyForward{Field: evidence.RelationField.FieldID}
		}
	}

	var relation Relation

	switch typ := relationType.(type) {
	case OneToManyBack:
		relation = Relation{
			Name:   evidence.RelationField.Name,
			Type:   typ,
			ModelA: evidence.RelationField.RefModelID,
			ModelB: evidence.RelationField.ModelID,
		}
	default:
		relation = Relation{
			Name:   evidence.RelationField.Name,
			Type:   typ,
			ModelA: evidence.RelationField.ModelID,
			ModelB: evidence.RelationField.RefModelID,
		}
	}

	relationID := RelationID(len(ctx.Relations.Storage))
	ctx.Relations.Storage = append(ctx.Relations.Storage, relation)
	ctx.Relations.Forward[Rel{evidence.RelationField.ModelID, evidence.RelationField.RefModelID, relationID}] = struct{}{}
	ctx.Relations.Backward[Rel{evidence.RelationField.RefModelID, evidence.RelationField.ModelID, relationID}] = struct{}{}
}

func (ctx *context) gatherRelationEvidence(mfid ModelFieldID, rf *RelationField) *RelationEvidence {
	model := ctx.Ast.GetModel(mfid.Model())
	field := ctx.Ast.GetModelField(mfid)
	refModel := ctx.Ast.GetModel(rf.RefModelID)
	isSelfRelation := mfid.Model() == rf.RefModelID

	var oppositeRelationField *RelationField
	var oppositeField *ast.Field

	for fid, field := range refModel.Fields {
		fid := FieldID(fid)
		orf, ok := ctx.Types.RelationFields[MakeModelFieldID(rf.RefModelID, fid)]
		// only consider relations between same models
		if !ok || mfid.Model() != orf.RefModelID {
			continue
		}
		// filter out the field itself, in case of self-relations
		if isSelfRelation && fid == mfid.Field() {
			continue
		}

		if orf.Name == rf.Name {
			oppositeRelationField = orf
			oppositeField = field
			break
		}
	}

	return &RelationEvidence{
		Model:                 model,
		Field:                 field,
		IsSelfRelation:        isSelfRelation,
		RelationField:         rf,
		OppositeModel:         refModel,
		OppositeField:         oppositeField,
		OppositeRelationField: oppositeRelationField,
	}
}

type RelationType interface{ reltyp() }

type ImplicitManyToMany struct {
	FieldA FieldID
	FieldB FieldID
}

type OneToOneForward struct {
	Field FieldID
}
type OneToOneBoth struct {
	FieldA FieldID
	FieldB FieldID
}

type OneToManyForward struct {
	Field FieldID
}
type OneToManyBack struct {
	Field FieldID
}
type OneToManyBoth struct {
	FieldA FieldID
	FieldB FieldID
}

func (ImplicitManyToMany) reltyp() {}
func (OneToOneForward) reltyp()    {}
func (OneToOneBoth) reltyp()       {}
func (OneToManyForward) reltyp()   {}
func (OneToManyBack) reltyp()      {}
func (OneToManyBoth) reltyp()      {}

type RelationName interface {
	relname()
	IsExplicit() bool
	IsGenerated() bool
	String() string
}
type ExplicitRelationName struct {
	Name string
}
type GeneratedRelationName struct {
	Name string
}

func (ExplicitRelationName) relname()           {}
func (ExplicitRelationName) IsExplicit() bool   { return true }
func (ExplicitRelationName) IsGenerated() bool  { return false }
func (n ExplicitRelationName) String() string   { return n.Name }
func (GeneratedRelationName) relname()          {}
func (GeneratedRelationName) IsExplicit() bool  { return false }
func (GeneratedRelationName) IsGenerated() bool { return true }
func (n GeneratedRelationName) String() string  { return n.Name }

type RelationIdentifier struct {
	ModelA ModelID
	ModelB ModelID
	Name   RelationName
}
