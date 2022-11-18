package pdb

type UniqueCriteriaWalker struct {
	db     *Db
	model  ModelID
	fields []FieldRef
}

func (w UniqueCriteriaWalker) Db() *Db { return w.db }
func (w UniqueCriteriaWalker) Model() ModelWalker {
	return ModelWalker{db: w.db, id: w.model}
}

func (w UniqueCriteriaWalker) Fields() []ScalarFieldWalker {
	var fields []ScalarFieldWalker
	for _, field := range w.fields {
		fields = append(fields, w.Model().ScalarField(field.FieldID))
	}
	return fields
}

func (w UniqueCriteriaWalker) FieldNames() []string {
	var names []string
	for _, field := range w.fields {
		names = append(names, w.Model().ScalarField(field.FieldID).Name())
	}
	return names
}

func (w UniqueCriteriaWalker) IsStrictCriteria() bool {
	for _, field := range w.Fields() {
		if field.IsOptional() {
			return false
		}
	}
	return true
}

func (w UniqueCriteriaWalker) ContainsExactlyFields(fields []ScalarFieldWalker) bool {
	if len(fields) != len(w.fields) {
		return false
	}
	for i, field := range w.Fields() {
		if field.ID() != fields[i].ID() {
			return false
		}
	}
	return true
}
