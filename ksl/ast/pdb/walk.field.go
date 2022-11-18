package pdb

import (
	"fmt"
	"ksl"
	"ksl/syntax/nodes"

	"github.com/samber/mo"
)

type FieldWalker interface {
	ID() FieldID
	AstField() *nodes.Field
	Model() ModelWalker
	ModelID() ModelID
	Name() string
}

type ScalarFieldWalker struct {
	model       ModelID
	field       FieldID
	db          *Db
	scalarField *ScalarField
}

func (w ScalarFieldWalker) Db() *Db                 { return w.db }
func (w ScalarFieldWalker) ID() FieldID             { return w.field }
func (w ScalarFieldWalker) Model() ModelWalker      { return ModelWalker{db: w.db, id: w.model} }
func (w ScalarFieldWalker) ModelID() ModelID        { return w.model }
func (w ScalarFieldWalker) Get() ScalarField        { return *w.scalarField }
func (w ScalarFieldWalker) Documentation() string   { return w.AstField().Documentation() }
func (w ScalarFieldWalker) IsIgnored() bool         { return w.Get().Ignored }
func (w ScalarFieldWalker) Arity() nodes.FieldArity { return w.AstField().Type.Arity }
func (w ScalarFieldWalker) AstField() *nodes.Field {
	return w.db.Ast.GetModelField(MakeModelFieldID(w.model, w.field))
}

func (w ScalarFieldWalker) Name() string { return w.AstField().GetName() }
func (w ScalarFieldWalker) DatabaseName() string {
	if name := w.Get().MappedName; name != "" {
		return name
	}
	return w.Name()
}
func (w ScalarFieldWalker) IsOptional() bool { return w.AstField().IsOptional() }
func (w ScalarFieldWalker) IsRequired() bool { return w.AstField().IsRequired() }
func (w ScalarFieldWalker) IsRepeated() bool { return w.AstField().IsRepeated() }
func (w ScalarFieldWalker) NativeType() mo.Option[NativeTypeAnnotation] {
	if nt := w.Get().NativeType; nt != nil {
		return mo.Some(*nt)
	}
	return mo.None[NativeTypeAnnotation]()
}

func (w ScalarFieldWalker) NativeTypeAnnotation() *nodes.Annotation {
	if nativeType, ok := w.NativeType().Get(); ok {
		return w.db.Ast.GetAnnotation(nativeType.SourceAnnotation)
	}
	return nil
}

func (w ScalarFieldWalker) DefaultValue() *DefaultValueWalker {
	if w.scalarField.Default != nil {
		return &DefaultValueWalker{w.model, w.field, w.db, w.scalarField.Default}
	}
	return nil
}

func (w ScalarFieldWalker) ScalarFieldType() ScalarFieldType {
	return w.Get().FieldType
}

type RelationFieldWalker struct {
	model         ModelID
	field         FieldID
	db            *Db
	relationField *RelationField
}

func (w RelationFieldWalker) ID() FieldID             { return w.field }
func (w RelationFieldWalker) Model() ModelWalker      { return ModelWalker{db: w.db, id: w.model} }
func (w RelationFieldWalker) ModelID() ModelID        { return w.model }
func (w RelationFieldWalker) IsIgnored() bool         { return w.relationField.Ignore }
func (w RelationFieldWalker) Arity() nodes.FieldArity { return w.AstField().Type.Arity }
func (w RelationFieldWalker) Documentation() string   { return w.AstField().Documentation() }
func (w RelationFieldWalker) IsRequired() bool        { return w.AstField().IsRequired() }
func (w RelationFieldWalker) IsOptional() bool        { return w.AstField().IsOptional() }
func (w RelationFieldWalker) IsRepeated() bool        { return w.AstField().IsRepeated() }

func (w RelationFieldWalker) ReferencesSingularIDField() bool {
	refs := w.Get().References
	if len(refs) != 1 {
		return len(refs) == 0
	}

	fieldID := refs[0].FieldID

	if pk, ok := w.RelatedModel().PrimaryKey().Get(); ok {
		return pk.ContainsExactlyFieldsByID(fieldID)
	}
	return false
}

func (w RelationFieldWalker) AstField() *nodes.Field {
	return w.db.Ast.GetModelField(MakeModelFieldID(w.model, w.field))
}
func (w RelationFieldWalker) AstAnnotation() *nodes.Annotation {
	return w.db.Ast.GetAnnotation(w.relationField.AnnotationID)
}
func (w RelationFieldWalker) MappedName() string { return w.Get().MappedName }
func (w RelationFieldWalker) Name() string       { return w.AstField().GetName() }
func (w RelationFieldWalker) RelatedModel() ModelWalker {
	fld := w.Get()
	if fld == nil {
		panic("nil relation field")
	}
	return ModelWalker{db: w.db, id: w.Get().RefModelID}
}
func (w RelationFieldWalker) Relation() RelationWalker {
	model := w.Model()
	for _, relation := range append(model.RelationsFrom(), model.RelationsTo()...) {
		if relation.HasField(MakeModelFieldID(w.model, w.field)) {
			return relation
		}
	}
	panic("relation not found")
}

func (w RelationFieldWalker) RelationName() RelationName {
	name := w.relationField.Name
	if name != "" {
		return ExplicitRelationName{Name: name}
	}
	modelName := w.Model().Name()
	relatedModelName := w.RelatedModel().Name()
	if modelName < relatedModelName {
		return GeneratedRelationName{Name: fmt.Sprintf("%sTo%s", modelName, relatedModelName)}
	} else {
		return GeneratedRelationName{Name: fmt.Sprintf("%sTo%s", relatedModelName, modelName)}
	}
}

func (w RelationFieldWalker) ReferencedFields() []ScalarFieldWalker {
	var fields []ScalarFieldWalker
	for _, field := range w.Get().References {
		fields = append(fields, w.RelatedModel().ScalarField(field.FieldID))
	}
	return fields
}

func (w RelationFieldWalker) ReferencedFieldNames() []string {
	var fields []string
	for _, field := range w.ReferencedFields() {
		fields = append(fields, field.Name())
	}
	return fields
}

func (w RelationFieldWalker) ReferencingFields() []ScalarFieldWalker {
	return w.Fields()
}

func (w RelationFieldWalker) ReferentialArity() nodes.FieldArity {
	for _, field := range w.Fields() {
		if field.IsRequired() {
			return nodes.Required
		}
	}
	return w.AstField().Type.Arity
}

func (w RelationFieldWalker) Fields() []ScalarFieldWalker {
	var fields []ScalarFieldWalker
	for _, field := range w.Get().Fields {
		fields = append(fields, w.db.WalkModel(field.ModelID).ScalarField(field.FieldID))
	}
	return fields
}

func (w RelationFieldWalker) Get() *RelationField {
	return w.relationField
}

func (w RelationFieldWalker) FieldNames() []string {
	var fields []string
	for _, field := range w.Fields() {
		fields = append(fields, field.Name())
	}
	return fields
}

func (w RelationFieldWalker) OnDelete() mo.Option[ReferentialAction] {
	if w.relationField.OnDelete != nil {
		return mo.Some(w.relationField.OnDelete.Action)
	}
	return mo.None[ReferentialAction]()
}

func (w RelationFieldWalker) OnDeleteSpan() mo.Option[ksl.Range] {
	if w.relationField.OnDelete != nil {
		return mo.Some(w.relationField.OnDelete.Span)
	}
	return mo.None[ksl.Range]()
}

func (w RelationFieldWalker) OnUpdate() mo.Option[ReferentialAction] {
	if w.relationField.OnUpdate != nil {
		return mo.Some(w.relationField.OnUpdate.Action)
	}
	return mo.None[ReferentialAction]()
}

func (w RelationFieldWalker) OnUpdateSpan() mo.Option[ksl.Range] {
	if w.relationField.OnUpdate != nil {
		return mo.Some(w.relationField.OnUpdate.Span)
	}
	return mo.None[ksl.Range]()
}
