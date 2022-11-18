package pdb

import (
	"encoding/base64"
	"fmt"
	"ksl"
	"ksl/syntax/nodes"
	"strings"
	"time"

	"golang.org/x/exp/slices"
)

func (v *ValidationContext) runFieldValidators(field FieldWalker) {
	for _, validator := range v.fields {
		v.diag(validator(field)...)
	}
}

func (v *ValidationContext) runScalarFieldValidators(field ScalarFieldWalker) {
	for _, validator := range v.scalarFields {
		v.diag(validator(field)...)
	}
}

func (v *ValidationContext) runRelationFieldValidators(field RelationFieldWalker) {
	for _, validator := range v.relationFields {
		v.diag(validator(field)...)
	}
}

func (v *ValidationContext) validateFieldClientName(field FieldWalker) {
	model := field.Model()

	if _, ok := v.names.IndexNames[IndexName{Model: model.ID(), Name: field.Name()}]; ok {
		v.diag(&ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  "Invalid index name",
			Detail:   fmt.Sprintf("The custom name %q specified for the @@index annotation is already used as a name for a field. Please choose a different name.", field.Name()),
			Subject:  model.AstModel().Range().Ptr(),
		})
	}

	if _, ok := v.names.UniqueNames[IndexName{Model: model.ID(), Name: field.Name()}]; ok {
		v.diag(&ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  "Invalid index name",
			Detail:   fmt.Sprintf("The custom name %q specified for the @@unique annotation is already used as a name for a field. Please choose a different name.", field.Name()),
			Subject:  model.AstModel().Range().Ptr(),
		})
	}

	if pk, ok := v.names.PrimaryKeyNames[model.ID()]; ok && pk == field.Name() {
		v.diag(&ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  "Invalid index name",
			Detail:   fmt.Sprintf("The custom name %q specified for the @@id annotation is already used as a name for a field. Please choose a different name.", field.Name()),
			Subject:  model.AstModel().Range().Ptr(),
		})
	}

}

func (v *ValidationContext) validateScalarFieldDefaultValue(fld ScalarFieldWalker) {
	scalar := fld.Get()
	field := fld.AstField()
	if scalar.Default == nil {
		return
	}

	value := scalar.Default.Value

	switch fldType := scalar.FieldType.Type.(type) {
	case EnumFieldType:
		enum := v.db.Ast.GetEnum(fldType.Enum)
		enumValues := enum.GetValues()

		validateEnumValue := func(expr ...nodes.Expression) {
			for _, elem := range expr {
				var enumValue string
				v.diag(v.db.Eval(elem, &enumValue)...)
				if !slices.Contains(enumValues, enumValue) {
					v.diag(&ksl.Diagnostic{
						Severity: ksl.DiagError,
						Summary:  "Invalid default value",
						Detail:   fmt.Sprintf("The default value %q is not a valid enum value.", enumValue),
						Subject:  elem.Range().Ptr(),
					})
				}
			}
		}

		switch {
		case field.IsRepeated():
			if val, ok := value.(*nodes.List); ok {
				validateEnumValue(val.Elements...)
			} else {
				v.diag(&ksl.Diagnostic{
					Severity: ksl.DiagError,
					Summary:  "Invalid default value",
					Detail:   "Default value for repeated enum field must be a list.",
					Subject:  value.Range().Ptr(),
				})
			}
		default:
			validateEnumValue(value)
		}
	case BuiltInScalarType:
		invalidDefault := func(expectedType string, fld ksl.BuiltInScalar, rng ksl.Range) {
			v.diag(&ksl.Diagnostic{
				Severity: ksl.DiagError,
				Summary:  "Invalid default value",
				Detail:   fmt.Sprintf("Default value for %s fields must be a %s.", fld.Name(), expectedType),
				Subject:  &rng,
			})
		}
		validateScalarValue := func(expr ...nodes.Expression) {
			for _, elem := range expr {
				switch fldType := ksl.BuiltInScalar(fldType); fldType {
				case ksl.BuiltIns.Int, ksl.BuiltIns.BigInt, ksl.BuiltIns.Float, ksl.BuiltIns.Decimal:
					if !nodes.IsNumericValue(elem) {
						invalidDefault("number", fldType, elem.Range())
					}
				case ksl.BuiltIns.String:
					if !nodes.IsStringValue(elem) {
						invalidDefault("string", fldType, elem.Range())
					}
				case ksl.BuiltIns.Bytes:
					var data string
					v.diag(v.db.Eval(elem, &data)...)
					if _, err := base64.StdEncoding.DecodeString(data); err != nil {
						invalidDefault("base64-encoded string", fldType, elem.Range())
					}

				case ksl.BuiltIns.DateTime:
					var dt string
					v.diag(v.db.Eval(elem, &dt)...)
					if _, err := time.Parse(time.RFC3339, dt); err != nil {
						invalidDefault("string in RFC3339 layout", fldType, elem.Range())
					}

				case ksl.BuiltIns.Date:
					var dt string
					v.diag(v.db.Eval(elem, &dt)...)
					if _, err := time.Parse("2006-01-02", dt); err != nil {
						invalidDefault("string in YYYY-MM-DD layout", fldType, elem.Range())
					}
				case ksl.BuiltIns.Time:
					var dt string
					v.diag(v.db.Eval(elem, &dt)...)
					if _, err := time.Parse("15:04:05.000", dt); err != nil {
						invalidDefault("string in HH:MM:SS.sss format", fldType, elem.Range())
					}

				case ksl.BuiltIns.Bool:
					allowed := []string{"true", "false"}
					if lit, ok := elem.(*nodes.Literal); !ok || !slices.Contains(allowed, lit.Value) {
						invalidDefault("\"true\" or \"false\"", fldType, elem.Range())
					}
				}
			}
		}

		switch {
		case field.IsRepeated():
			if val, ok := value.(*nodes.List); ok {
				validateScalarValue(val.Elements...)
			} else {
				v.diag(&ksl.Diagnostic{
					Severity: ksl.DiagError,
					Summary:  "Invalid default value",
					Detail:   "Default value for repeated field must be a list.",
					Subject:  value.Range().Ptr(),
				})
			}
		default:
			validateScalarValue(value)
		}
	}
}

func (v *ValidationContext) validateRelationFieldAmbiguity(field RelationFieldWalker) bool {
	model := field.Model()
	relatedModel := field.RelatedModel()
	relationName := field.RelationName()
	identifier := RelationIdentifier{ModelA: model.ID(), ModelB: relatedModel.ID(), Name: relationName}
	selfRelation := model.ID() == relatedModel.ID()

	if fields, ok := v.names.RelationNames[identifier]; ok && len(fields) > 1 {
		var message string
		switch {
		case relationName.IsGenerated() && selfRelation && len(fields) == 2:
			message = fmt.Sprintf(
				"Ambiguous self relation detected. The fields %s in model %q both refer to %q.\nIf they are part of the same relation, add the same relation name for them with `@ref(<name>)`.",
				formatFieldsAmbiguous(model, fields),
				model.Name(),
				relatedModel.Name(),
			)
		case relationName.IsGenerated() && selfRelation && len(fields) > 2:
			message = fmt.Sprintf(
				"Unnamed self relation detected. The fields %s in model %q have no relation name. Please provide a relation name for one of them by adding @ref(<name>).",
				formatFieldsAmbiguous(model, fields),
				model.Name(),
			)
		case relationName.IsExplicit() && selfRelation && len(fields) > 2:
			message = fmt.Sprintf(
				"Wrongly named self relation detected. The fields %s in model %q have the same relation name. At most two relation fields can belong to the same relation and therefore have the same name. Please assign a different relation name to one of them.",
				formatFieldsAmbiguous(model, fields),
				model.Name(),
			)

		case relationName.IsExplicit() && selfRelation && len(fields) == 2:
			return true

		case relationName.IsGenerated():
			message = fmt.Sprintf(
				"Ambiguous relation detected. The fields %s in model %q both refer to %q. Please provide different relation names for them by adding @ref(<name>).",
				formatFieldsAmbiguous(model, fields),
				model.Name(),
				relatedModel.Name(),
			)
		case relationName.IsExplicit():
			message = fmt.Sprintf(
				"Wrongly named relation detected. The fields %s in model %q both use the same relation name. Please provide different relation names for them through @ref(<name>).",
				formatFieldsAmbiguous(model, fields),
				model.Name(),
			)
		}
		v.diag(&ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  "Ambiguous relation detected",
			Detail:   message,
			Subject:  field.AstField().Range().Ptr(),
		})
		return false
	}
	return true
}

func (v *ValidationContext) validateRelationFieldIgnoredRelatedModels(field RelationFieldWalker) {
	model := field.Model()
	relatedModel := field.RelatedModel()

	if !relatedModel.IsIgnored() || field.IsIgnored() || model.IsIgnored() {
		return
	}
	message := fmt.Sprintf(
		"The relation field %q on model %q must specify the @ignore annotation, because the model %q it is pointing to is marked ignored.",
		field.Name(),
		model.Name(),
		relatedModel.Name(),
	)

	v.diag(&ksl.Diagnostic{
		Severity: ksl.DiagError,
		Summary:  "Invalid relation field",
		Detail:   message,
		Subject:  field.AstField().Range().Ptr(),
	})
}

// func (v *ValidationContext) validateNativeTypes(field ScalarFieldWalker) ksl.Diagnostics {
// 	var diags ksl.Diagnostics

// 	if nativeType, ok := field.NativeType(); ok {
// 		annot := field.NativeTypeAnnotation()
// 		if !slices.Contains(nativeTypeAnnots[:], nativeType.Name) {
// 			diags = append(diags, &ksl.Diagnostic{
// 				Severity: ksl.DiagError,
// 				Summary:  "Unknown native type",
// 				Detail:   fmt.Sprintf("The native type %q is not known by the postgres connector.", nativeType.Name),
// 				Subject:  annot.Range().Ptr(),
// 			})
// 			return diags
// 		}

// 		args := annot.GetArgs()
// 		if len(args) == 0 {
// 			return diags
// 		}

// 		switch aliasType(nativeType.Name) {
// 		case TypeNumeric:
// 			if len(args) > 2 {
// 				diags = append(diags, &ksl.Diagnostic{
// 					Severity: ksl.DiagError,
// 					Summary:  "Invalid native type",
// 					Detail:   "The decimal native type takes at most two numeric arguments.",
// 					Subject:  annot.Range().Ptr(),
// 				})
// 				return diags
// 			}
// 			for _, arg := range args {
// 				if _, ok := arg.Value.(*ast.Number); !ok {
// 					diags = append(diags, &ksl.Diagnostic{
// 						Severity: ksl.DiagError,
// 						Summary:  "Invalid native type",
// 						Detail:   fmt.Sprintf("The %s native type takes only numeric arguments.", nativeType.Name),
// 						Subject:  arg.Range().Ptr(),
// 					})
// 					return diags
// 				}
// 			}

// 		case TypeTimestamp, TypeTimestampTZ, TypeTime, TypeTimeTZ:
// 			if len(args) > 1 {
// 				diags = append(diags, &ksl.Diagnostic{
// 					Severity: ksl.DiagError,
// 					Summary:  "Invalid native type",
// 					Detail:   fmt.Sprintf("The %s native type takes at most one numeric argument.", nativeType.Name),
// 					Subject:  annot.Range().Ptr(),
// 				})
// 				return diags
// 			}

// 			if _, ok := args[0].Value.(*ast.Number); !ok {
// 				diags = append(diags, &ksl.Diagnostic{
// 					Severity: ksl.DiagError,
// 					Summary:  "Invalid native type",
// 					Detail:   fmt.Sprintf("The %s native type takes only numeric arguments.", nativeType.Name),
// 					Subject:  args[0].Range().Ptr(),
// 				})
// 				return diags
// 			}

// 		case TypeBit, TypeVarBit, TypeVarChar, TypeChar:
// 			if len(args) > 1 {
// 				diags = append(diags, &ksl.Diagnostic{
// 					Severity: ksl.DiagError,
// 					Summary:  "Invalid native type",
// 					Detail:   fmt.Sprintf("The %s native type takes at most one numeric argument.", nativeType.Name),
// 					Subject:  annot.Range().Ptr(),
// 				})
// 				return diags
// 			}

// 			if _, ok := args[0].Value.(*ast.Number); !ok {
// 				diags = append(diags, &ksl.Diagnostic{
// 					Severity: ksl.DiagError,
// 					Summary:  "Invalid native type",
// 					Detail:   fmt.Sprintf("The %s native type takes only numeric arguments.", nativeType.Name),
// 					Subject:  args[0].Range().Ptr(),
// 				})
// 				return diags
// 			}
// 		default:
// 			diags = append(diags, &ksl.Diagnostic{
// 				Severity: ksl.DiagError,
// 				Summary:  "Invalid native type",
// 				Detail:   fmt.Sprintf("The %s native type does not take any arguments.", nativeType.Name),
// 				Subject:  annot.Range().Ptr(),
// 			})
// 			return diags
// 		}
// 		fieldType := field.ScalarFieldType().Type
// 		builtin, ok := fieldType.(pdb.BuiltInScalarType)
// 		if !ok {
// 			diags = append(diags, &ksl.Diagnostic{
// 				Severity: ksl.DiagError,
// 				Summary:  "Invalid native type",
// 				Detail:   fmt.Sprintf("The %s native type can only be used with built-in scalar types.", nativeType.Name),
// 				Subject:  annot.Range().Ptr(),
// 			})
// 			return diags
// 		}
// 		if CompatibleScalar(nativeType.Name) != ksl.BuiltInScalar(builtin) {
// 			diags = append(diags, &ksl.Diagnostic{
// 				Severity: ksl.DiagError,
// 				Summary:  "Invalid native type",
// 				Detail:   fmt.Sprintf("The %s native type is not compatible with the %q scalar type.", nativeType.Name, builtin),
// 				Subject:  annot.Range().Ptr(),
// 			})
// 			return diags
// 		}
// 	}

// 	return diags
// }

func formatFieldsAmbiguous(model ModelWalker, fields []FieldID) string {
	var names []string
	for _, field := range fields {
		names = append(names, fmt.Sprintf("%q", model.RelationField(field).Name()))
	}
	switch {
	case len(fields) < 2:
		return strings.Join(names, ", ")
	case len(fields) == 2:
		return strings.Join(names, " and ")
	default:
		return strings.Join(names[:len(names)-1], ", ") + " and " + names[len(names)-1]
	}
}
