package pdb

import (
	"fmt"
	"ksl"
	"strings"
)

func (v *ValidationContext) runIndexValidators(index IndexWalker) {
	for _, validator := range v.indexes {
		v.diag(validator(index)...)
	}
}

func (v *ValidationContext) validateIndexHasFields(index IndexWalker) {
	if len(index.Fields()) > 0 {
		return
	}
	v.diag(&ksl.Diagnostic{
		Severity: ksl.DiagError,
		Summary:  "Invalid index",
		Detail:   "The list of fields in an index cannot be empty. Please specify at least one field.",
		Subject:  index.AstAnnotation().Range().Ptr(),
	})
}

func (v *ValidationContext) validateIndexHasUniqueConstraintName(index IndexWalker) {
	model := index.Model()
	name := index.ConstraintName()

	for _, violation := range v.constraintNameScopeViolations(model.ID(), IndexConstraintName(name)) {
		v.diag(&ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  "Duplicate primary key name",
			Detail:   fmt.Sprintf("The index name %q is not unique for %s.", name, violation),
			Subject:  index.AstAnnotation().Range().Ptr(),
		})
	}
}

func (v *ValidationContext) validateIndexUniqueClientNameDoesNotClashWithField(index IndexWalker) {
	if !index.IsUnique() {
		return
	}

	fields := index.Fields()
	// Only compound indexes can clash with fields.
	if len(fields) <= 1 {
		return
	}

	fieldNames := make([]string, len(fields))
	for i, field := range fields {
		fieldNames[i] = field.Name()
	}
	idxName := strings.Join(fieldNames, "_")
	for _, field := range index.Model().ScalarFields() {
		if field.Name() == idxName {
			v.diag(&ksl.Diagnostic{
				Severity: ksl.DiagError,
				Summary:  "Invalid index",
				Detail:   fmt.Sprintf("The index name %q clashes with a field name.", idxName),
				Subject:  index.AstAnnotation().Range().Ptr(),
			})
		}
	}
}

func (v *ValidationContext) validateIndexHasUniqueCustomNamePerModel(index IndexWalker) {
	model := index.Model()
	name := index.Name()
	if name != "" {
		if v.names.ConstraintNamespace.LocalCustom[LocalCustomConstraint{Model: model.ID(), Name: name}] > 1 {
			v.diag(&ksl.Diagnostic{
				Severity: ksl.DiagError,
				Summary:  "Duplicate constraint name",
				Detail:   fmt.Sprintf("The index name %q is not unique for %s.", name, model.Name()),
				Subject:  index.AstAnnotation().Range().Ptr(),
			})
		}
	}
}

func (v *ValidationContext) validateHashIndexMustNotUseSortParam(index IndexWalker) {
	if index.Algorithm().IsHash() {
		for _, field := range index.Attribute().Fields {
			if field.Sort != "" {
				v.diag(&ksl.Diagnostic{
					Severity: ksl.DiagError,
					Summary:  "Invalid index",
					Detail:   "The sort parameter is not allowed for hash indices.",
					Subject:  index.AstAnnotation().Range().Ptr(),
				})
				return
			}
		}
	}
}

// func (v *ValidationContext) validateCompatibleNativeTypes(index IndexWalker) ksl.Diagnostics {
// 	var diags ksl.Diagnostics

// 	for _, field := range index.Fields() {
// 		if nativeType, ok := field.NativeType(); ok {
// 			typ, err := Types.ScalarFrom(nativeType.Name, nativeType.Args...)
// 			if err != nil {
// 				continue
// 			}

// 			if typ.(scalartype).name == TypeXML {
// 				diags = append(diags, &ksl.Diagnostic{
// 					Severity: ksl.DiagError,
// 					Summary:  "Invalid native type",
// 					Detail:   "The native type xml is not supported for indexing.",
// 					Subject:  field.NativeTypeAnnotation().Range().Ptr(),
// 				})
// 			}
// 		}
// 	}
// 	return diags
// }

// func (v *ValidationContext) validateSPGistIndexedColumnCount(index IndexWalker) ksl.Diagnostics {
// 	var diags ksl.Diagnostics

// 	if index.Algorithm() == pdb.SpGist && len(index.Fields()) > 1 {
// 		diags = append(diags, &ksl.Diagnostic{
// 			Severity: ksl.DiagError,
// 			Summary:  "Invalid index",
// 			Detail:   "The spgist index algorithm does not support indexing multiple columns.",
// 			Subject:  index.AstAnnotation().Range().Ptr(),
// 		})
// 	}

// 	return diags
// }
