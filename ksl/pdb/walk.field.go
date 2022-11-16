package pdb

import (
	"fmt"
	"ksl/syntax/ast"
)

type FieldWalker interface {
	ID() FieldID
	AstField() *ast.Field
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

func (w ScalarFieldWalker) Db() *Db                { return w.db }
func (w ScalarFieldWalker) ID() FieldID            { return w.field }
func (w ScalarFieldWalker) Model() ModelWalker     { return ModelWalker{db: w.db, id: w.model} }
func (w ScalarFieldWalker) ModelID() ModelID       { return w.model }
func (w ScalarFieldWalker) Attribute() ScalarField { return *w.scalarField }
func (w ScalarFieldWalker) Documentation() string  { return w.AstField().Documentation() }

func (w ScalarFieldWalker) Arity() ast.FieldArity { return w.AstField().Type.Arity }
func (w ScalarFieldWalker) AstField() *ast.Field {
	return w.db.Ast.GetModelField(MakeModelFieldID(w.model, w.field))
}

func (w ScalarFieldWalker) Name() string { return w.AstField().GetName() }
func (w ScalarFieldWalker) DatabaseName() string {
	if name := w.Attribute().MappedName; name != "" {
		return name
	}
	return w.Name()
}
func (w ScalarFieldWalker) IsOptional() bool { return w.AstField().IsOptional() }
func (w ScalarFieldWalker) IsRequired() bool { return w.AstField().IsRequired() }
func (w ScalarFieldWalker) IsRepeated() bool { return w.AstField().IsRepeated() }
func (w ScalarFieldWalker) NativeType() (*NativeTypeAnnotation, bool) {
	if nt := w.Attribute().NativeType; nt != nil {
		return nt, true
	}
	return nil, false
}

func (w ScalarFieldWalker) NativeTypeAnnotation() *ast.Annotation {
	if nativeType, ok := w.NativeType(); ok {
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
	return w.Attribute().FieldType
}

type RelationFieldWalker struct {
	model         ModelID
	field         FieldID
	db            *Db
	relationField *RelationField
}

func (w RelationFieldWalker) ID() FieldID        { return w.field }
func (w RelationFieldWalker) Model() ModelWalker { return ModelWalker{db: w.db, id: w.model} }
func (w RelationFieldWalker) ModelID() ModelID   { return w.model }
func (w RelationFieldWalker) IsIgnored() bool    { return w.relationField.Ignore }
func (w RelationFieldWalker) IsRequired() bool   { return w.AstField().IsRequired() }
func (w RelationFieldWalker) IsOptional() bool   { return w.AstField().IsOptional() }
func (w RelationFieldWalker) IsRepeated() bool   { return w.AstField().IsRepeated() }

func (w RelationFieldWalker) ReferencesSingularIDField() bool {
	refs := w.Get().References
	if len(refs) != 1 {
		return len(refs) == 0
	}

	fieldID := refs[0].FieldID

	if pk, ok := w.RelatedModel().PrimaryKey(); ok {
		return pk.ContainsExactlyFieldsByID(fieldID)
	}
	return false
}

func (w RelationFieldWalker) AstField() *ast.Field {
	return w.db.Ast.GetModelField(MakeModelFieldID(w.model, w.field))
}
func (w RelationFieldWalker) AstAnnotation() *ast.Annotation {
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

func (w RelationFieldWalker) ReferentialArity() ast.FieldArity {
	for _, field := range w.Fields() {
		if field.IsRequired() {
			return ast.Required
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

func (w RelationFieldWalker) OnDelete() (RefAction, bool) {
	if w.relationField.OnDelete != nil {
		return *w.relationField.OnDelete, true
	}
	return RefAction{}, false
}

func (w RelationFieldWalker) OnUpdate() (RefAction, bool) {
	if w.relationField.OnUpdate != nil {
		return *w.relationField.OnUpdate, true
	}
	return RefAction{}, false
}
