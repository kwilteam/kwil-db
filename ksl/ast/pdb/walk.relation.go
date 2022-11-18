package pdb

import (
	"fmt"
	"ksl/constraints"

	"github.com/samber/mo"
)

type RelationWalker struct {
	db *Db
	id RelationID
}

func (w RelationWalker) Db() *Db { return w.db }
func (w RelationWalker) HasField(mfid ModelFieldID) bool {
	return w.db.Relations.Get(w.id).HasField(mfid)
}

func (w RelationWalker) Refine() RefinedRelationWalker {
	rel := w.db.Relations.Get(w.id)
	if rel.IsImplicitManyToMany() {
		return ImplicitManyToManyRelationWalker(w)
	} else {
		return InlineRelationWalker(w)
	}
}

type CompleteInlineRelationWalker struct {
	modelA ModelID
	fieldA FieldID
	modelB ModelID
	fieldB FieldID
	db     *Db
}

func (w CompleteInlineRelationWalker) Db() *Db { return w.db }

type RefinedRelationWalker interface{ rrl() }

type InlineRelationWalker struct {
	db *Db
	id RelationID
}

func (w InlineRelationWalker) Db() *Db { return w.db }
func (w InlineRelationWalker) IsOneToOne() bool {
	rel := w.db.Relations.Get(w.id)
	return rel.IsOneToOneBoth() || rel.IsOneToOneForward()
}
func (w InlineRelationWalker) Get() Relation { return w.db.Relations.Get(w.id) }
func (w InlineRelationWalker) ReferencingModel() ModelWalker {
	return ModelWalker{w.db, w.Get().ModelA}
}
func (w InlineRelationWalker) ReferencingFields() []ScalarFieldWalker {
	var fields []ScalarFieldWalker
	if field, ok := w.ForwardRelationField().Get(); ok {
		return field.Fields()
	}
	return fields
}
func (w InlineRelationWalker) ReferencingFieldNames() []string {
	var fields []string
	for _, field := range w.ReferencingFields() {
		fields = append(fields, field.Name())
	}
	return fields
}
func (w InlineRelationWalker) ReferencingFieldDatabaseNames() []string {
	var fields []string
	for _, field := range w.ReferencingFields() {
		fields = append(fields, field.DatabaseName())
	}
	return fields
}
func (w InlineRelationWalker) ReferencedModel() ModelWalker {
	return ModelWalker{w.db, w.Get().ModelB}
}
func (w InlineRelationWalker) ReferencedFields() []ScalarFieldWalker {
	var fields []ScalarFieldWalker
	if field, ok := w.ForwardRelationField().Get(); ok {
		return field.ReferencedFields()
	} else {
		for _, criteria := range w.ReferencedModel().UniqueCriterias() {
			if criteria.IsStrictCriteria() {
				fields = append(fields, criteria.Fields()...)
			}
		}
	}
	return fields
}
func (w InlineRelationWalker) ReferencedFieldNames() []string {
	var fields []string
	for _, field := range w.ReferencedFields() {
		fields = append(fields, field.Name())
	}
	return fields
}

func (w InlineRelationWalker) ForwardRelationField() mo.Option[RelationFieldWalker] {
	model := w.ReferencingModel()
	switch typ := w.Get().Type.(type) {
	case OneToOneForward:
		return mo.Some(model.RelationField(typ.Field))
	case OneToOneBoth:
		return mo.Some(model.RelationField(typ.FieldA))
	case OneToManyBoth:
		return mo.Some(model.RelationField(typ.FieldA))
	case OneToManyForward:
		return mo.Some(model.RelationField(typ.Field))
	}

	return mo.None[RelationFieldWalker]()
}

func (w InlineRelationWalker) BackRelationField() mo.Option[RelationFieldWalker] {
	model := w.ReferencedModel()
	switch typ := w.Get().Type.(type) {
	case OneToOneBoth:
		return mo.Some(model.RelationField(typ.FieldB))
	case OneToManyBoth:
		return mo.Some(model.RelationField(typ.FieldB))
	case OneToManyBack:
		return mo.Some(model.RelationField(typ.Field))
	}

	return mo.None[RelationFieldWalker]()

}

func (w InlineRelationWalker) ForeignKeyName() mo.Option[string] {
	fkName := w.ConstraintName()
	defaultName := constraints.ForeignKeyConstraintName(
		w.ReferencingModel().DatabaseName(),
		w.ReferencingFieldDatabaseNames(),
		MaxIdentifierLength,
	)

	if fkName != defaultName {
		return mo.Some(fkName)
	}
	return mo.None[string]()
}

func (w InlineRelationWalker) RelationName() RelationName {
	name := w.Get().Name
	if name != "" {
		return ExplicitRelationName{Name: name}
	}
	modelName := w.ReferencingModel().Name()
	relatedModelName := w.ReferencedModel().Name()
	if modelName < relatedModelName {
		return GeneratedRelationName{Name: fmt.Sprintf("%sTo%s", modelName, relatedModelName)}
	} else {
		return GeneratedRelationName{Name: fmt.Sprintf("%sTo%s", relatedModelName, modelName)}
	}
}

func (w InlineRelationWalker) AsComplete() mo.Option[CompleteInlineRelationWalker] {
	forward := w.ForwardRelationField()
	back := w.BackRelationField()
	if forward.IsPresent() && back.IsPresent() {
		return mo.Some(CompleteInlineRelationWalker{
			modelA: w.Get().ModelA,
			fieldA: forward.MustGet().ID(),
			modelB: w.Get().ModelB,
			fieldB: back.MustGet().ID(),
			db:     w.db,
		})
	}
	return mo.None[CompleteInlineRelationWalker]()
}

func (w InlineRelationWalker) MappedName() string {
	if f, ok := w.ForwardRelationField().Get(); ok {
		return f.MappedName()
	}
	return ""
}

func (w InlineRelationWalker) ConstraintName() string {
	if name := w.MappedName(); name != "" {
		return name
	}
	modelDbName := w.ReferencingModel().DatabaseName()
	referencedFields := w.ReferencingFields()
	fieldNames := make([]string, 0, len(referencedFields))
	for _, field := range referencedFields {
		fieldNames = append(fieldNames, field.DatabaseName())
	}
	return constraints.ForeignKeyConstraintName(modelDbName, fieldNames, MaxIdentifierLength)
}

type ImplicitManyToManyRelationWalker struct {
	db *Db
	id RelationID
}

func (w ImplicitManyToManyRelationWalker) Db() *Db       { return w.db }
func (w ImplicitManyToManyRelationWalker) Get() Relation { return w.db.Relations.Get(w.id) }

func (w ImplicitManyToManyRelationWalker) ModelA() ModelWalker {
	return ModelWalker{w.db, w.Get().ModelA}
}

func (w ImplicitManyToManyRelationWalker) ModelB() ModelWalker {
	return ModelWalker{w.db, w.Get().ModelB}
}

func (w ImplicitManyToManyRelationWalker) FieldA() RelationFieldWalker {
	return w.ModelA().RelationField(w.Get().Type.(ImplicitManyToMany).FieldA)
}

func (w ImplicitManyToManyRelationWalker) FieldB() RelationFieldWalker {
	return w.ModelB().RelationField(w.Get().Type.(ImplicitManyToMany).FieldB)
}

func (w ImplicitManyToManyRelationWalker) RelationName() RelationName {
	return w.FieldA().RelationName()
}

func (InlineRelationWalker) rrl()             {}
func (ImplicitManyToManyRelationWalker) rrl() {}
