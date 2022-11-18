package pdb

import (
	"fmt"
	"ksl"
	"sort"
	"strings"

	"golang.org/x/exp/slices"
)

func (v *ValidationContext) runRelationValidators(relation RelationWalker) {
	for _, validator := range v.relations {
		v.diag(validator(relation)...)
	}
}

func (v *ValidationContext) validateRelationReferencesUniqueFields(relation InlineRelationWalker) {
	relationField, ok := relation.ForwardRelationField().Get()
	if !ok {
		return
	}

	if len(relationField.ReferencedFields()) == 0 {
		return
	}

	refModel := relation.ReferencedModel()

	for _, criteria := range refModel.UniqueCriterias() {
		fieldNames := criteria.FieldNames()
		sort.StringSlice(fieldNames).Sort()

		referencedFieldNames := relation.ReferencedFieldNames()
		sort.StringSlice(referencedFieldNames).Sort()

		if slices.Equal(fieldNames, referencedFieldNames) {
			return
		}
	}

	fields := relation.ReferencedFieldNames()
	model := relation.ReferencedModel().Name()

	var message string
	if len(fields) == 1 {
		message = fmt.Sprintf("The argument `references` must refer to a unique criteria in the related model. Consider adding a `@unique` annotation to the field %q in the model %q.", strings.Join(fields, ", "), model)
	} else {
		message = fmt.Sprintf("The argument `references` must refer to a unique criteria in the related model. Consider adding a `@@unique([%s])` annotation to the model %q.", strings.Join(fields, ", "), model)
	}

	v.diag(&ksl.Diagnostic{
		Severity: ksl.DiagError,
		Summary:  "Invalid relation",
		Detail:   message,
		Subject:  relationField.AstField().Range().Ptr(),
	})
}

func (v *ValidationContext) validateRelationSameLengthInReferencingAndReferenced(relation InlineRelationWalker) {
	relationField, ok := relation.ForwardRelationField().Get()
	if !ok {
		return
	}

	if len(relationField.ReferencedFields()) != len(relationField.Fields()) {
		v.diag(&ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  "Invalid relation",
			Detail:   "The number of fields in the `fields` argument must match the number of fields in the `references` argument.",
			Subject:  relationField.AstAnnotation().Range().Ptr(),
		})
	}
}

func (v *ValidationContext) validateRelationFieldArity(relation InlineRelationWalker) {
	field, ok := relation.ForwardRelationField().Get()
	if !ok || !field.AstField().IsRequired() {
		return
	}

	refFields := relation.ReferencingFields()
	if len(refFields) == 0 {
		return
	}

	allRequired := true
	for _, f := range refFields {
		if f.IsOptional() || f.IsRepeated() {
			allRequired = false
			break
		}
	}

	if allRequired {
		return
	}

	fieldNames := relation.ReferencingFieldNames()
	message := fmt.Sprintf(
		"The relation field %q uses the scalar fields %s. At least one of those fields is optional. Hence the relation field must be optional as well.",
		field.Name(),
		strings.Join(fieldNames, ", "),
	)
	v.diag(&ksl.Diagnostic{
		Severity: ksl.DiagError,
		Summary:  "Invalid relation",
		Detail:   message,
		Subject:  field.AstField().Range().Ptr(),
	})
}

func (v *ValidationContext) validateRelationReferencingScalarFieldTypes(relation InlineRelationWalker) {
	referencingFields := relation.ReferencingFields()
	referencedFields := relation.ReferencedFields()

	if len(referencedFields) != len(referencingFields) {
		return
	}

	for i := range referencedFields {
		referenced := referencedFields[i]
		referencing := referencingFields[i]

		if referenced.ScalarFieldType() != referencing.ScalarFieldType() {
			message := fmt.Sprintf(
				"The type of the field %q in the model %q is not matching the type of the referenced field %q in model %q.",
				referencing.Name(),
				referencing.Model().Name(),
				referenced.Name(),
				referenced.Model().Name(),
			)
			v.diag(&ksl.Diagnostic{
				Severity: ksl.DiagError,
				Summary:  "Invalid relation",
				Detail:   message,
				Subject:  referencing.AstField().Range().Ptr(),
			})
		}
	}
}
func (v *ValidationContext) validateRelationHasUniqueConstraintName(relation InlineRelationWalker) {
	name := relation.ConstraintName()
	model := relation.ReferencingModel()

	field, fok := relation.ForwardRelationField().Get()

	for _, violation := range v.constraintNameScopeViolations(model.ID(), RelationConstraintName(name)) {
		var span ksl.Range
		if fok {
			span = field.AstAnnotation().Range()
		} else {
			span = relation.ReferencedModel().AstModel().Range()
		}
		v.diag(&ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  "Duplicate relation name",
			Detail:   fmt.Sprintf("The relation name %q is not unique for %s.", name, violation),
			Subject:  &span,
		})
	}
}

func (v *ValidationContext) validateRequiredRelationCannotUseSetNull(relation InlineRelationWalker) {
	forward, ok := relation.ForwardRelationField().Get()
	if !ok {
		return
	}

	allOptional := true
	// return early if no referencing field is required
	for _, field := range forward.ReferencingFields() {
		if field.IsRequired() {
			allOptional = false
			break
		}
	}

	if allOptional {
		return
	}

	if onDelete, ok := forward.OnDelete().Get(); ok && onDelete == SetNull {
		v.diag(&ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  "Invalid relation",
			Detail:   "The `onDelete` referential action of a relation must not be set to `SetNull` when a referenced field is required. Either choose another referential action, or make the referenced fields optional.",
			Subject:  forward.AstAnnotation().Range().Ptr(),
		})
	}
	if onUpdate, ok := forward.OnUpdate().Get(); ok && onUpdate == SetNull {
		v.diag(&ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  "Invalid relation",
			Detail:   "The `onUpdate` referential action of a relation must not be set to `SetNull` when a referenced field is required. Either choose another referential action, or make the referenced fields optional.",
			Subject:  forward.AstAnnotation().Range().Ptr(),
		})
	}
}

func (v *ValidationContext) validateRelationOneToOneBothSidesAreDefined(relation InlineRelationWalker) {
	_, ok := relation.BackRelationField().Get()
	if ok {
		return
	}
	forward, _ := relation.ForwardRelationField().Get()

	v.diag(&ksl.Diagnostic{
		Severity: ksl.DiagError,
		Summary:  "Invalid relation",
		Detail: fmt.Sprintf(
			"The relation field %q is missing an opposite relation field on the model %q.",
			fmt.Sprintf("%s.%s", forward.Model().Name(), forward.Name()),
			relation.ReferencedModel().Name(),
		),
		Subject: forward.AstField().Range().Ptr(),
	})
}

func (v *ValidationContext) validateRelationOneToOneFieldsAndReferencesAreDefined(relation InlineRelationWalker) {
	forward, fok := relation.ForwardRelationField().Get()
	back, bok := relation.BackRelationField().Get()
	if !fok || !bok {
		return
	}

	if len(forward.ReferencingFields()) == 0 && len(back.ReferencingFields()) == 0 {
		v.diag(&ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  "Invalid relation",
			Detail: fmt.Sprintf(
				"The relation fields %q and %q do not provide the `fields` argument in the @ref annotation. You have to provide it on one of the two fields.",
				fmt.Sprintf("%s.%s", forward.Model().Name(), forward.Name()),
				fmt.Sprintf("%s.%s", back.Model().Name(), back.Name()),
			),
			Subject: forward.AstField().Range().Ptr(),
		})

		v.diag(&ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  "Invalid relation",
			Detail: fmt.Sprintf(
				"The relation fields %q and %q do not provide the `fields` argument in the @ref annotation. You have to provide it on one of the two fields.",
				fmt.Sprintf("%s.%s", back.Model().Name(), back.Name()),
				fmt.Sprintf("%s.%s", forward.Model().Name(), forward.Name()),
			),
			Subject: back.AstField().Range().Ptr(),
		})
	}

	if len(forward.ReferencedFields()) == 0 && len(back.ReferencedFields()) == 0 {
		v.diag(&ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  "Invalid relation",
			Detail: fmt.Sprintf(
				"The relation fields %q and %q do not provide the `references` argument in the @ref annotation. You have to provide it on one of the two fields.",
				fmt.Sprintf("%s.%s", forward.Model().Name(), forward.Name()),
				fmt.Sprintf("%s.%s", back.Model().Name(), back.Name()),
			),
			Subject: forward.AstField().Range().Ptr(),
		})

		v.diag(&ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  "Invalid relation",
			Detail: fmt.Sprintf(
				"The relation fields %q and %q do not provide the `references` argument in the @ref annotation. You have to provide it on one of the two fields.",
				fmt.Sprintf("%s.%s", back.Model().Name(), back.Name()),
				fmt.Sprintf("%s.%s", forward.Model().Name(), forward.Name()),
			),
			Subject: back.AstField().Range().Ptr(),
		})
	}
}
func (v *ValidationContext) validateRelationOneToOneFieldsAndReferencesDefinedOnOneSideOnly(relation InlineRelationWalker) {
	forward, fok := relation.ForwardRelationField().Get()
	back, bok := relation.BackRelationField().Get()
	if !fok || !bok {
		return
	}

	if len(forward.ReferencedFields()) > 0 && len(back.ReferencedFields()) > 0 {
		message := fmt.Sprintf(
			"The relation fields %q and %q both provide the `references` argument in the @ref attribute. You have to provide it only on one of the two fields.",
			fmt.Sprintf("%s.%s", forward.Model().Name(), forward.Name()),
			fmt.Sprintf("%s.%s", back.Model().Name(), back.Name()),
		)

		v.diag(&ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  "Invalid relation",
			Detail:   message,
			Subject:  forward.AstField().Range().Ptr(),
		})
		v.diag(&ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  "Invalid relation",
			Detail:   message,
			Subject:  back.AstField().Range().Ptr(),
		})
	}

	if len(forward.ReferencingFields()) > 0 && len(back.ReferencingFields()) > 0 {
		message := fmt.Sprintf(
			"The relation fields %q and %q both provide the `fields` argument in the @ref attribute. You have to provide it only on one of the two fields.",
			fmt.Sprintf("%s.%s", forward.Model().Name(), forward.Name()),
			fmt.Sprintf("%s.%s", back.Model().Name(), back.Name()),
		)

		v.diag(&ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  "Invalid relation",
			Detail:   message,
			Subject:  forward.AstField().Range().Ptr(),
		})
		v.diag(&ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  "Invalid relation",
			Detail:   message,
			Subject:  back.AstField().Range().Ptr(),
		})
	}
}

func (v *ValidationContext) validateRelationOneToOneReferentialActions(relation InlineRelationWalker) {
	forward, fok := relation.ForwardRelationField().Get()
	back, bok := relation.BackRelationField().Get()
	if !fok || !bok {
		return
	}

	_, fod := forward.OnDelete().Get()
	_, fou := forward.OnUpdate().Get()
	_, bod := back.OnDelete().Get()
	_, bou := back.OnUpdate().Get()

	if (fod || fou) && (bod || bou) {
		v.diag(&ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  "Invalid relation",
			Detail: fmt.Sprintf(
				"The relation fields %q and %q both provide the `onDelete` or `onUpdate` argument in the @ref attribute. You have to provide it only on one of the two fields.",
				fmt.Sprintf("%s.%s", back.Model().Name(), back.Name()),
				fmt.Sprintf("%s.%s", forward.Model().Name(), forward.Name()),
			),
			Subject: back.AstField().Range().Ptr(),
		})
		v.diag(&ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  "Invalid relation",
			Detail: fmt.Sprintf(
				"The relation fields %q and %q both provide the `onDelete` or `onUpdate` argument in the @ref attribute. You have to provide it only on one of the two fields.",
				fmt.Sprintf("%s.%s", forward.Model().Name(), forward.Name()),
				fmt.Sprintf("%s.%s", back.Model().Name(), back.Name()),
			),
			Subject: forward.AstField().Range().Ptr(),
		})
	} else if bod || bou {
		v.diag(&ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  "Invalid relation",
			Detail: fmt.Sprintf(
				"The relation field %q must not specify the `onDelete` or `onUpdate` argument in the @ref attribute. You must only specify it on the opposite field %q.",
				fmt.Sprintf("%s.%s", back.Model().Name(), back.Name()),
				fmt.Sprintf("%s.%s", forward.Model().Name(), forward.Name()),
			),
			Subject: back.AstField().Range().Ptr(),
		})
	}
}

// A 1:1 relation is enforced with a unique constraint. The referencing side must use a unique constraint to enforce the relation.
func (v *ValidationContext) validateRelationOneToOneFieldsMustBeUniqueConstraint(relation InlineRelationWalker) {
	forward, fok := relation.ForwardRelationField().Get()
	if !fok {
		return
	}

	model := relation.ReferencingModel()
	referencing := relation.ReferencingFields()

	isUnique := false
	for _, criteria := range model.UniqueCriterias() {
		if len(referencing) == 0 || criteria.ContainsExactlyFields(referencing) {
			isUnique = true
			break
		}
	}

	if isUnique {
		return
	}

	fieldNames := relation.ReferencingFieldNames()
	var message string
	if len(fieldNames) == 1 {
		message = fmt.Sprintf("A one-to-one relation must use unique fields on the defining side. Either add a `@unique` attribute to the field %q, or change the relation to one-to-many.", strings.Join(fieldNames, ", "))
	} else {
		message = fmt.Sprintf("A one-to-one relation must use unique fields on the defining side. Either add a `@@unique[%s]` attribute to the model, or change the relation to one-to-many.", strings.Join(fieldNames, ", "))
	}

	v.diag(&ksl.Diagnostic{
		Severity: ksl.DiagError,
		Summary:  "Invalid relation",
		Detail:   message,
		Subject:  forward.AstField().Range().Ptr(),
	})
}

// Validation of some crazy things, such as definining `fields` and `references` on different sides in the relation.
func (v *ValidationContext) validateRelationOneToOneFieldsReferencesMixups(relation InlineRelationWalker) {
	if v.diags.HasErrors() {
		return
	}

	forward, fok := relation.ForwardRelationField().Get()
	back, bok := relation.BackRelationField().Get()
	if !fok || !bok {
		return
	}

	if len(forward.ReferencingFields()) > 0 && len(back.ReferencedFields()) > 0 {
		message := fmt.Sprintf(
			"The relation field %q provides the `fields` argument in the @ref attribute. And the related field %q provides the `references` argument. You must provide both arguments on the same side.",
			fmt.Sprintf("%s.%s", forward.Model().Name(), forward.Name()),
			fmt.Sprintf("%s.%s", back.Model().Name(), back.Name()),
		)

		v.diag(&ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  "Invalid relation",
			Detail:   message,
			Subject:  forward.AstField().Range().Ptr(),
		})

		v.diag(&ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  "Invalid relation",
			Detail:   message,
			Subject:  back.AstField().Range().Ptr(),
		})
	}

	if len(forward.ReferencedFields()) > 0 && len(back.ReferencingFields()) > 0 {
		message := fmt.Sprintf(
			"The relation field %q provides the `references` argument in the @ref attribute. And the related field %q provides the `fields` argument. You must provide both arguments on the same side.",
			fmt.Sprintf("%s.%s", forward.Model().Name(), forward.Name()),
			fmt.Sprintf("%s.%s", back.Model().Name(), back.Name()),
		)

		v.diag(&ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  "Invalid relation",
			Detail:   message,
			Subject:  forward.AstField().Range().Ptr(),
		})
	}
}

func (v *ValidationContext) validateRelationOneToOneBackRelationArityIsOptional(relation InlineRelationWalker) {
	if v.diags.HasErrors() {
		return
	}

	forward, fok := relation.ForwardRelationField().Get()
	back, bok := relation.BackRelationField().Get()
	if !fok || !bok {
		return
	}

	if back.AstField().IsRequired() {
		v.diag(&ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  "Invalid relation",
			Detail: fmt.Sprintf(
				"The relation field %q is required. This is not possible to enforce on the database level. Please change the field type from %q to %q to fix this.",
				fmt.Sprintf("%s.%s", back.Model().Name(), back.Name()),
				forward.Model().Name(),
				forward.Model().Name()+"?",
			),
			Subject: back.AstField().Range().Ptr(),
		})
	}
}

func (v *ValidationContext) validateRelationOneToOneFieldsAndReferencesOnWrongSide(relation InlineRelationWalker) {
	if v.diags.HasErrors() {
		return
	}

	forward, fok := relation.ForwardRelationField().Get()
	back, bok := relation.BackRelationField().Get()
	if !fok || !bok {
		return
	}

	if forward.IsRequired() && (len(back.ReferencingFields()) > 0 || len(back.ReferencedFields()) > 0) {
		v.diag(&ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  "Invalid relation",
			Detail: fmt.Sprintf(
				"The relation field %q defines the `fields` and/or `references` argument. You must set them on the required side of the relation (%q) in order for the constraints to be enforced. Alternatively, you can change this field to be required and the opposite optional, or make both sides of the relation optional.",
				fmt.Sprintf("%s.%s", back.Model().Name(), back.Name()),
				fmt.Sprintf("%s.%s", forward.Model().Name(), forward.Name()),
			),
			Subject: back.AstField().Range().Ptr(),
		})
	}
}

func (v *ValidationContext) validateRelationOneToManyBothSidesAreDefined(relation InlineRelationWalker) {
	forward, fok := relation.ForwardRelationField().Get()
	back, bok := relation.BackRelationField().Get()

	switch {
	case fok && !bok:
		v.diag(&ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  "Invalid relation",
			Detail: fmt.Sprintf(
				"The relation field %q is missing an opposite relation field on the model %q.",
				fmt.Sprintf("%s.%s", forward.Model().Name(), forward.Name()),
				forward.RelatedModel().Name(),
			),
			Subject: forward.AstField().Range().Ptr(),
		})
	case bok && !fok:
		v.diag(&ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  "Invalid relation",
			Detail: fmt.Sprintf(
				"The relation field %q is missing an opposite relation field on the model %q.",
				fmt.Sprintf("%s.%s", back.Model().Name(), back.Name()),
				back.RelatedModel().Name(),
			),
			Subject: back.AstField().Range().Ptr(),
		})
	}
}

func (v *ValidationContext) validateRelationOneToManyFieldsAndReferencesAreDefined(relation InlineRelationWalker) {
	forward, fok := relation.ForwardRelationField().Get()
	back, bok := relation.BackRelationField().Get()
	if !fok || !bok {
		return
	}

	// fields argument should not be empty
	if len(forward.ReferencingFields()) == 0 {
		v.diag(&ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  "Invalid relation",
			Detail: fmt.Sprintf(
				"The relation field %q is missing the `fields` argument in the @ref attribute.",
				fmt.Sprintf("%s.%s", forward.Model().Name(), forward.Name()),
			),
			Subject: forward.AstField().Range().Ptr(),
		})
	}

	// references argument should not be empty
	if len(forward.ReferencedFields()) == 0 {
		v.diag(&ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  "Invalid relation",
			Detail: fmt.Sprintf(
				"The relation field %q is missing the `references` argument in the @ref attribute.",
				fmt.Sprintf("%s.%s", forward.Model().Name(), forward.Name()),
			),
			Subject: forward.AstField().Range().Ptr(),
		})
	}

	if len(back.ReferencingFields()) > 0 || len(back.ReferencedFields()) > 0 {
		v.diag(&ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  "Invalid relation",
			Detail: fmt.Sprintf(
				"The relation field %q must not specify the `fields` or `references` argument in the @ref attribute. You must only specify it on the opposite field %q.",
				fmt.Sprintf("%s.%s", back.Model().Name(), back.Name()),
				fmt.Sprintf("%s.%s", forward.Model().Name(), forward.Name()),
			),
			Subject: forward.AstField().Range().Ptr(),
		})
	}
}

func (v *ValidationContext) validateRelationOneToManyReferentialActions(relation InlineRelationWalker) {
	forward, fok := relation.ForwardRelationField().Get()
	back, bok := relation.BackRelationField().Get()
	if !fok || !bok {
		return
	}

	_, bod := back.OnDelete().Get()
	_, bou := back.OnUpdate().Get()

	if bod || bou {
		v.diag(&ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  "Invalid relation",
			Detail: fmt.Sprintf(
				"The relation field %q must not specify the `onDelete` or `onUpdate` argument in the @ref attribute. You must only specify it on the opposite field %q, or in case of a many to many relation, in an explicit join table.",
				fmt.Sprintf("%s.%s", back.Model().Name(), back.Name()),
				fmt.Sprintf("%s.%s", forward.Model().Name(), forward.Name()),
			),
			Subject: back.AstField().Range().Ptr(),
		})
	}
}

func (v *ValidationContext) validateRelationImplicitManyToManySingularId(relation ImplicitManyToManyRelationWalker) {
	for _, field := range []RelationFieldWalker{relation.FieldA(), relation.FieldB()} {
		if !field.RelatedModel().HasSingleIDField() {
			v.diag(&ksl.Diagnostic{
				Severity: ksl.DiagError,
				Summary:  "Invalid relation",
				Detail: fmt.Sprintf(
					"The relation field %q references %q which does not have an @id field. Models without @id cannot be part of a many to many relation. Use an explicit intermediate model to represent this relationship.",
					fmt.Sprintf("%s.%s", field.Model().Name(), field.Name()),
					field.RelatedModel().Name(),
				),
				Subject: field.AstField().Range().Ptr(),
			})
			continue
		}

		if !field.ReferencesSingularIDField() {
			v.diag(&ksl.Diagnostic{
				Severity: ksl.DiagError,
				Summary:  "Invalid relation",
				Detail: fmt.Sprintf(
					"Implicit many-to-many relations must always reference the id field of the related model. Change the argument `references` to use the id field of the related model %q. But it is referencing the following fields that are not the id: %s",
					field.RelatedModel().Name(),
					strings.Join(field.ReferencedFieldNames(), ", "),
				),
				Subject: field.AstField().Range().Ptr(),
			})
		}
	}
}

func (v *ValidationContext) validateRelationImplicitManyToManyNoReferentialActions(relation ImplicitManyToManyRelationWalker) {
	var refactions []RefAction
	for _, field := range []RelationFieldWalker{relation.FieldA(), relation.FieldB()} {
		if ondelete, ok := field.OnDelete().Get(); ok {
			refactions = append(refactions, RefAction{Action: ondelete, Span: field.OnDeleteSpan().MustGet()})
		}
		if onupdate, ok := field.OnUpdate().Get(); ok {
			refactions = append(refactions, RefAction{Action: onupdate, Span: field.OnUpdateSpan().MustGet()})
		}
	}

	for _, action := range refactions {
		v.diag(&ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  "Invalid relation",
			Detail:   "Referential actions on implicit many-to-many relations are not supported.",
			Subject:  &action.Span,
		})
	}
}

func (v *ValidationContext) validateRelationImplicitManyToManyCannotDefineReferencesArgument(relation ImplicitManyToManyRelationWalker) {
	for _, field := range []RelationFieldWalker{relation.FieldA(), relation.FieldB()} {
		if len(field.ReferencedFields()) > 0 {
			v.diag(&ksl.Diagnostic{
				Severity: ksl.DiagError,
				Summary:  "Invalid relation",
				Detail: fmt.Sprintf(
					"The relation field %q must not specify the `references` argument in the @ref attribute. The referenced fields are automatically inferred from the related model.",
					fmt.Sprintf("%s.%s", field.Model().Name(), field.Name()),
				),
				Subject: field.AstField().Range().Ptr(),
			})
		}
	}
}
