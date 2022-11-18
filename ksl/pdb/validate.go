package pdb

import (
	"ksl"
)

type Validator interface{ validator() }

type GlobalValidator func(*Db) ksl.Diagnostics
type ModelValidator func(ModelWalker) ksl.Diagnostics
type EnumValidator func(EnumWalker) ksl.Diagnostics
type IndexValidator func(IndexWalker) ksl.Diagnostics
type RelationValidator func(RelationWalker) ksl.Diagnostics
type ScalarFieldValidator func(ScalarFieldWalker) ksl.Diagnostics
type RelationFieldValidator func(RelationFieldWalker) ksl.Diagnostics
type FieldValidator func(FieldWalker) ksl.Diagnostics

func (fn GlobalValidator) validator()        {}
func (fn ModelValidator) validator()         {}
func (fn EnumValidator) validator()          {}
func (fn IndexValidator) validator()         {}
func (fn RelationValidator) validator()      {}
func (fn ScalarFieldValidator) validator()   {}
func (fn RelationFieldValidator) validator() {}
func (fn FieldValidator) validator()         {}

type validationoptions struct {
	validators     []GlobalValidator
	models         []ModelValidator
	enums          []EnumValidator
	indexes        []IndexValidator
	relations      []RelationValidator
	scalarFields   []ScalarFieldValidator
	relationFields []RelationFieldValidator
	fields         []FieldValidator
}
type ValidateOption func(*validationoptions)

func Validations(validators ...Validator) ValidateOption {
	return func(v *validationoptions) {
		for _, validator := range validators {
			switch validator := validator.(type) {
			case GlobalValidator:
				v.validators = append(v.validators, validator)
			case ModelValidator:
				v.models = append(v.models, validator)
			case EnumValidator:
				v.enums = append(v.enums, validator)
			case IndexValidator:
				v.indexes = append(v.indexes, validator)
			case RelationValidator:
				v.relations = append(v.relations, validator)
			case ScalarFieldValidator:
				v.scalarFields = append(v.scalarFields, validator)
			case RelationFieldValidator:
				v.relationFields = append(v.relationFields, validator)
			case FieldValidator:
				v.fields = append(v.fields, validator)
			}
		}
	}
}

type ValidationContext struct {
	validationoptions
	db    *Db
	names *nameValidationContext
	diags ksl.Diagnostics
}

func NewValidationContext(db *Db, diags ksl.Diagnostics, opts ...ValidateOption) *ValidationContext {
	var validateOpts validationoptions
	for _, opt := range opts {
		opt(&validateOpts)
	}

	names := newNameValidationContext(db)
	ctx := &ValidationContext{validationoptions: validateOpts, db: db, diags: diags, names: names}
	return ctx
}

func (v *ValidationContext) Validate(db *Db) ksl.Diagnostics {
	if v.diags.HasErrors() {
		return v.diags
	}

	v.runValidators(db)
	v.validateModelDatabaseNameClashes(db)

	for _, model := range db.WalkModels() {
		v.validateModelHasStrictUniqueCriteria(model)
		v.validateModelHasUniquePrimaryKeyName(model)
		v.validateModelIdHasFields(model)
		v.validateModelIdNameDoesNotClashWithField(model)
		v.runModelValidators(model)

		for _, field := range model.ScalarFields() {
			v.runScalarFieldValidators(field)
			v.runFieldValidators(field)
			v.validateFieldClientName(field)
			v.validateScalarFieldDefaultValue(field)
		}

		for _, field := range model.RelationFields() {
			v.runRelationFieldValidators(field)
			v.runFieldValidators(field)

			v.validateFieldClientName(field)
			v.validateRelationFieldIgnoredRelatedModels(field)

			if !v.validateRelationFieldAmbiguity(field) {
				return v.diags
			}
		}

		for _, index := range model.Indexes() {
			v.runIndexValidators(index)

			v.validateIndexHasFields(index)
			v.validateIndexHasUniqueConstraintName(index)
			v.validateIndexUniqueClientNameDoesNotClashWithField(index)
			v.validateIndexHasUniqueCustomNamePerModel(index)
			v.validateHashIndexMustNotUseSortParam(index)
		}
	}

	v.validateEnumDatabaseNameClashes(db)
	for _, enum := range db.WalkEnums() {
		v.runEnumValidators(enum)
		v.validateEnumHasValues(enum)
	}

	for _, relation := range db.WalkRelations() {
		v.runRelationValidators(relation)
		switch refined := relation.Refine().(type) {
		case InlineRelationWalker:
			v.validateRelationReferencesUniqueFields(refined)
			v.validateRelationSameLengthInReferencingAndReferenced(refined)
			v.validateRelationFieldArity(refined)
			v.validateRelationReferencingScalarFieldTypes(refined)
			v.validateRelationHasUniqueConstraintName(refined)
			v.validateRequiredRelationCannotUseSetNull(refined)

			if refined.IsOneToOne() {
				v.validateRelationOneToOneBothSidesAreDefined(refined)
				v.validateRelationOneToOneFieldsAndReferencesAreDefined(refined)
				v.validateRelationOneToOneFieldsAndReferencesDefinedOnOneSideOnly(refined)
				v.validateRelationOneToOneReferentialActions(refined)
				v.validateRelationOneToOneFieldsMustBeUniqueConstraint(refined)
				v.validateRelationOneToOneFieldsReferencesMixups(refined)
				v.validateRelationOneToOneBackRelationArityIsOptional(refined)
				v.validateRelationOneToOneFieldsAndReferencesOnWrongSide(refined)
			} else {
				v.validateRelationOneToManyBothSidesAreDefined(refined)
				v.validateRelationOneToManyFieldsAndReferencesAreDefined(refined)
				v.validateRelationOneToManyReferentialActions(refined)
			}

		case ImplicitManyToManyRelationWalker:
			v.validateRelationImplicitManyToManySingularId(refined)
			v.validateRelationImplicitManyToManyNoReferentialActions(refined)
			v.validateRelationImplicitManyToManyCannotDefineReferencesArgument(refined)
		}
	}

	return v.diags
}

func (v *ValidationContext) runValidators(db *Db) {
	for _, validator := range v.validators {
		v.diag(validator(db)...)
	}
}

func (v *ValidationContext) diag(diags ...*ksl.Diagnostic) bool {
	v.diags = append(v.diags, diags...)
	return ksl.Diagnostics(diags).HasErrors()
}
