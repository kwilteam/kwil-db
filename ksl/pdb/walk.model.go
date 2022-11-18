package pdb

import (
	"ksl/syntax/ast"
)

type ModelWalker struct {
	db *Db
	id ModelID
}

func (w ModelWalker) Db() *Db                      { return w.db }
func (w ModelWalker) ID() ModelID                  { return w.id }
func (w ModelWalker) Name() string                 { return w.AstModel().GetName() }
func (w ModelWalker) AstModel() *ast.Model         { return w.db.Ast.GetModel(w.id) }
func (w ModelWalker) Documentation() string        { return w.AstModel().Documentation() }
func (w ModelWalker) Fields() []*ast.Field         { return w.AstModel().Fields }
func (w ModelWalker) Attributes() ModelAnnotations { return w.db.Types.ModelAnnotations[w.id] }
func (w ModelWalker) IsIgnored() bool              { return w.Attributes().Ignored }

func (w ModelWalker) HasSingleIDField() bool {
	if pk := w.Attributes().PrimaryKey; pk != nil {
		return len(pk.Fields) == 1
	}
	return false
}

func (w ModelWalker) PrimaryKey() (PrimaryKeyWalker, bool) {
	if pk := w.Attributes().PrimaryKey; pk != nil {
		return PrimaryKeyWalker{db: w.db, model: w.id, pk: *pk}, true
	}
	return PrimaryKeyWalker{}, false
}

func (w ModelWalker) MustPrimaryKey() PrimaryKeyWalker {
	return PrimaryKeyWalker{db: w.db, model: w.id, pk: *w.Attributes().PrimaryKey}
}

func (w ModelWalker) DatabaseName() string {
	if name := w.MappedName(); name != "" {
		return name
	}
	return w.AstModel().GetName()
}

func (w ModelWalker) FieldDatabaseName(field FieldID) string {
	if fld, ok := w.db.Types.ScalarFields[MakeModelFieldID(w.id, field)]; ok {
		if fld.MappedName != "" {
			return fld.MappedName
		}
	}
	return w.db.Ast.GetModelField(MakeModelFieldID(w.id, field)).GetName()
}

func (w ModelWalker) ScalarField(field FieldID) ScalarFieldWalker {
	return ScalarFieldWalker{
		db:          w.db,
		model:       w.id,
		field:       field,
		scalarField: w.db.Types.ScalarFields[MakeModelFieldID(w.id, field)],
	}
}

func (w ModelWalker) ScalarFields() []ScalarFieldWalker {
	var fields []ScalarFieldWalker
	for i, field := range w.db.Types.ScalarFields {
		if i.Model() == w.id {
			fields = append(fields, ScalarFieldWalker{db: w.db, model: i.Model(), field: i.Field(), scalarField: field})
		}
	}
	return fields
}

func (w ModelWalker) UniqueCriterias() []UniqueCriteriaWalker {
	var criterias []UniqueCriteriaWalker
	if pk := w.Attributes().PrimaryKey; pk != nil {
		criterias = append(criterias, UniqueCriteriaWalker{db: w.db, model: w.id, fields: pk.Fields})
	}
	for _, idx := range w.Indexes() {
		if idx.IsUnique() {
			criterias = append(criterias, UniqueCriteriaWalker{db: w.db, model: w.id, fields: idx.Attribute().Fields})
		}
	}
	return criterias
}

func (w ModelWalker) Indexes() []IndexWalker {
	info := w.Attributes()
	indexes := make([]IndexWalker, len(info.Indexes))
	for i, idx := range info.Indexes {
		indexes[i] = IndexWalker{db: w.db, model: w.id, index: idx}
	}
	return indexes
}

func (w ModelWalker) RelationFields() []RelationFieldWalker {
	var relations []RelationFieldWalker
	for i, relation := range w.db.Types.RelationFields {
		if i.Model() == w.id {
			relations = append(relations, RelationFieldWalker{db: w.db, model: w.id, field: i.Field(), relationField: relation})
		}
	}
	return relations
}

func (w ModelWalker) RelationField(field FieldID) RelationFieldWalker {
	return RelationFieldWalker{
		db:            w.db,
		model:         w.id,
		field:         field,
		relationField: w.db.Types.RelationFields[MakeModelFieldID(w.id, field)],
	}
}

func (w ModelWalker) RelationsFrom() []RelationWalker {
	var relations []RelationWalker
	for _, relation := range w.db.Relations.FromModel(w.id) {
		relations = append(relations, RelationWalker{db: w.db, id: relation})
	}
	return relations
}

func (w ModelWalker) RelationsTo() []RelationWalker {
	var relations []RelationWalker
	for _, relation := range w.db.Relations.ToModel(w.id) {
		relations = append(relations, RelationWalker{db: w.db, id: relation})
	}
	return relations
}

func (w ModelWalker) MappedName() string { return w.Attributes().Name }
