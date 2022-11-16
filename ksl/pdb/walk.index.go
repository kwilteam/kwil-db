package pdb

import (
	"ksl/constraints"
	"ksl/syntax/ast"
)

type PrimaryKeyWalker struct {
	db    *Db
	model ModelID
	pk    IDAnnotation
}

func (w PrimaryKeyWalker) Db() *Db      { return w.db }
func (w PrimaryKeyWalker) Name() string { return w.pk.Name }
func (w PrimaryKeyWalker) AstAnnotation() *ast.Annotation {
	return w.db.Ast.GetAnnotation(w.pk.SourceAnnot)
}
func (w PrimaryKeyWalker) Model() ModelWalker     { return ModelWalker{db: w.db, id: w.model} }
func (w PrimaryKeyWalker) IsDefinedOnField() bool { return w.pk.SourceField != nil }
func (w PrimaryKeyWalker) AnnotationName() string {
	if w.IsDefinedOnField() {
		return "@id"
	}
	return "@@id"
}
func (w PrimaryKeyWalker) Fields() []ScalarFieldWalker {
	model := w.Model()
	fields := make([]ScalarFieldWalker, len(w.pk.Fields))
	for i, field := range w.pk.Fields {
		fields[i] = model.ScalarField(field.FieldID)
	}
	return fields
}

func (w PrimaryKeyWalker) FirstField() ScalarFieldWalker {
	return w.Model().ScalarField(w.pk.Fields[0].FieldID)
}

func (w PrimaryKeyWalker) ScalarFieldAttributes() []ScalarFieldAttributeWalker {
	var attrs []ScalarFieldAttributeWalker
	for _, field := range w.pk.Fields {
		attrs = append(attrs, ScalarFieldAttributeWalker{db: w.db, model: w.model, field: field.FieldID, ref: field})
	}
	return attrs
}

func (w PrimaryKeyWalker) ConstraintName() string {
	if name := w.Name(); name != "" {
		return name
	}
	return constraints.PrimaryKeyName(w.Model().DatabaseName(), MaxIdentifierLength)
}

func (w PrimaryKeyWalker) ContainsExactlyFieldsByID(fields ...FieldID) bool {
	if len(fields) != len(w.pk.Fields) {
		return false
	}
	for i, field := range w.pk.Fields {
		if field.FieldID != fields[i] {
			return false
		}
	}
	return true
}

type ScalarFieldAttributeWalker struct {
	db    *Db
	model ModelID
	field FieldID
	ref   FieldRef
}

func (w ScalarFieldAttributeWalker) Db() *Db                  { return w.db }
func (w ScalarFieldAttributeWalker) Model() ModelWalker       { return w.db.WalkModel(w.model) }
func (w ScalarFieldAttributeWalker) Field() ScalarFieldWalker { return w.Model().ScalarField(w.field) }
func (w ScalarFieldAttributeWalker) SortOrder() string        { return w.ref.Sort }

type IndexWalker struct {
	model ModelID
	annot AnnotID
	db    *Db
	index *IndexAnnotation
}

func (w IndexWalker) Db() *Db                        { return w.db }
func (w IndexWalker) ModelID() ModelID               { return w.model }
func (w IndexWalker) Model() ModelWalker             { return ModelWalker{db: w.db, id: w.model} }
func (w IndexWalker) AnnotID() AnnotID               { return w.annot }
func (w IndexWalker) AstAnnotation() *ast.Annotation { return w.db.Ast.GetAnnotation(w.annot) }
func (w IndexWalker) Attribute() *IndexAnnotation    { return w.index }
func (w IndexWalker) Name() string                   { return w.index.Name }
func (w IndexWalker) IsUnique() bool                 { return w.index.Type == IndexTypeUnique }
func (w IndexWalker) Algorithm() IndexAlgorithm      { return w.index.Algorithm }

func (w IndexWalker) ConstraintName() string {
	if name := w.Name(); name != "" {
		return name
	}
	model := w.Model()
	modelDbName := model.DatabaseName()

	var fields []string
	for _, field := range w.index.Fields {
		scalar := model.ScalarField(field.FieldID)
		name := scalar.Attribute().MappedName
		if name == "" {
			name = scalar.Name()
		}
		fields = append(fields, name)
	}
	if w.IsUnique() {
		return constraints.UniqueIndexName(modelDbName, fields, MaxIdentifierLength)
	} else {
		return constraints.NonUniqueIndexName(modelDbName, fields, MaxIdentifierLength)
	}
}

func (w IndexWalker) Fields() []ScalarFieldWalker {
	var fields []ScalarFieldWalker
	model := w.Model()
	for _, field := range w.index.Fields {
		fields = append(fields, model.ScalarField(field.FieldID))
	}
	return fields
}

func (w IndexWalker) ScalarFieldAttributes() []ScalarFieldAttributeWalker {
	var attrs []ScalarFieldAttributeWalker
	for _, field := range w.index.Fields {
		attrs = append(attrs, ScalarFieldAttributeWalker{db: w.db, model: w.model, field: field.FieldID, ref: field})
	}
	return attrs
}
