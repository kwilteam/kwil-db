package pdb

import (
	"fmt"
	"ksl"
	"strings"
)

func (v *ValidationContext) runModelValidators(model ModelWalker) {
	for _, validator := range v.models {
		v.diag(validator(model)...)
	}
}

func (v *ValidationContext) validateModelDatabaseNameClashes(db *Db) {
	dbnames := map[string]ModelID{}

	for _, model := range db.WalkModels() {
		key := model.DatabaseName()
		if mid, ok := dbnames[key]; ok {
			if mappedName := model.MappedName(); mappedName != "" {
				existingModelName := db.Ast.GetModel(mid).GetName()
				for _, annot := range model.AstModel().GetAnnotations() {
					if annot.GetName() == "map" {
						v.diag(&ksl.Diagnostic{
							Severity: ksl.DiagError,
							Summary:  "Duplicate database name",
							Detail:   fmt.Sprintf("Model %q has a database name clash with model %q", model.Name(), existingModelName),
							Subject:  annot.Range().Ptr(),
						})
						break
					}
				}
			} else {
				existingModel := db.Ast.GetModel(mid)
				for _, annot := range existingModel.GetAnnotations() {
					if annot.GetName() == "map" {
						v.diag(&ksl.Diagnostic{
							Severity: ksl.DiagError,
							Summary:  "Duplicate database name",
							Detail:   fmt.Sprintf("Model %q has a database name clash with model %q", model.Name(), model.DatabaseName()),
							Subject:  annot.Range().Ptr(),
						})
						break
					}
				}
			}
		}
		dbnames[key] = model.ID()
	}
}

func (v *ValidationContext) validateModelHasStrictUniqueCriteria(model ModelWalker) {
	if model.IsIgnored() {
		return
	}

	for _, criteria := range model.UniqueCriterias() {
		if criteria.IsStrictCriteria() {
			return
		}
	}

	var looseCriteria []string
	for _, criteria := range model.UniqueCriterias() {
		fieldNames := make([]string, len(criteria.Fields()))
		for i, field := range criteria.Fields() {
			fieldNames[i] = field.Name()
		}
		looseCriteria = append(looseCriteria, "- "+strings.Join(fieldNames, ", "))
	}
	message := "Each model must have at least one unique criteria that has only required fields. Either mark a single field with `@id`, `@unique` or add a multi field criterion with `@@id([])` or `@@unique([])` to the model."
	if len(looseCriteria) > 0 {
		suffix := fmt.Sprintf("The following unique criterias were not considered as they contain fields that are not required:\n%s", strings.Join(looseCriteria, "\n"))
		message = fmt.Sprintf("%s %s", message, suffix)
	}
	v.diag(&ksl.Diagnostic{
		Severity: ksl.DiagError,
		Summary:  "Model has no unique criteria",
		Detail:   message,
		Subject:  model.AstModel().Range().Ptr(),
	})
}

func (v *ValidationContext) validateModelHasUniquePrimaryKeyName(model ModelWalker) {
	pk, ok := model.PrimaryKey().Get()
	if !ok {
		return
	}

	for _, violation := range v.constraintNameScopeViolations(model.ID(), PrimaryKeyConstraintName(pk.ConstraintName())) {
		v.diag(&ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  "Duplicate primary key name",
			Detail:   fmt.Sprintf("The primary key name %q is not unique for %s.", pk.ConstraintName(), violation),
			Subject:  pk.AstAnnotation().Range().Ptr(),
		})
	}
}

func (v *ValidationContext) validateModelIdHasFields(model ModelWalker) {
	if pk, ok := model.PrimaryKey().Get(); ok {
		if len(pk.Fields()) == 0 {
			v.diag(&ksl.Diagnostic{
				Severity: ksl.DiagError,
				Summary:  "Invalid primary key",
				Detail:   "The primary key must have at least one field.",
				Subject:  pk.AstAnnotation().Range().Ptr(),
			})
		}
	}
}

func (v *ValidationContext) validateModelIdNameDoesNotClashWithField(model ModelWalker) {
	if pk, ok := model.PrimaryKey().Get(); ok {
		fields := pk.Fields()
		if len(fields) > 1 {
			fieldNames := make([]string, len(fields))
			for i, field := range fields {
				fieldNames[i] = field.Name()
			}
			idName := strings.Join(fieldNames, "_")
			for _, field := range model.ScalarFields() {
				if field.Name() == idName {
					v.diag(&ksl.Diagnostic{
						Severity: ksl.DiagError,
						Summary:  "Invalid primary key",
						Detail:   fmt.Sprintf("The primary key name %q clashes with a field name.", idName),
						Subject:  pk.AstAnnotation().Range().Ptr(),
					})
				}
			}
		}
	}
}

func (v *ValidationContext) constraintNameScopeViolations(modelID ModelID, name ConstraintName) []ConstraintScope {
	var violations []ConstraintScope

	for _, scope := range possibleScopes(name) {
		constraint := LocalConstraint{Model: modelID, Name: name.String(), Scope: scope}
		if v.names.ConstraintNamespace.Local[constraint] > 1 {
			violations = append(violations, scope)
		}
	}
	return violations
}

func possibleScopes(name ConstraintName) []ConstraintScope {
	var scopes []ConstraintScope

	switch name.(type) {
	case IndexConstraintName:
		scopes = append(scopes, GlobalKeyIndex, GlobalPrimaryKeyKeyIndex, ModelKeyIndex, ModelPrimaryKeyKeyIndex)
	case RelationConstraintName:
		scopes = append(scopes, GlobalForeignKey, GlobalPrimaryKeyForeignKeyDefault, ModelPrimaryKeyKeyIndexForeignKey)
	case PrimaryKeyConstraintName:
		scopes = append(scopes, GlobalPrimaryKeyKeyIndex, ModelPrimaryKeyKeyIndex, GlobalPrimaryKeyForeignKeyDefault, ModelPrimaryKeyKeyIndexForeignKey)
	case DefaultConstraintName:
		scopes = append(scopes, GlobalPrimaryKeyForeignKeyDefault)
	}
	return scopes
}
