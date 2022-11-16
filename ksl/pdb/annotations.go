package pdb

import (
	"fmt"
	"ksl"
	"ksl/spec"
	"ksl/syntax/ast"
	"strings"
)

func (ctx *context) ResolveAnnotations() {
	visitedDirectives := map[string]struct{}{}
	for eid, entry := range ctx.Ast.Tops {
		switch entry := entry.(type) {
		case *ast.Model:
			ctx.resolveModelAnnotations(ModelID(eid), entry)
		case *ast.Enum:
			ctx.resolveEnumAnnotations(EnumID(eid), entry)
		case *ast.Annotation:
			ctx.resolveDirective(DirectiveID(eid), entry, visitedDirectives)
		}
	}
}

func (ctx *context) resolveDirective(id DirectiveID, annot *ast.Annotation, visited map[string]struct{}) {
	sp, ok := ctx.Spec.Directives[annot.GetName()]
	if !ok {
		ctx.diag(&ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  "Invalid directive",
			Detail:   fmt.Sprintf("Directive %q is not a known directive.", annot.GetName()),
			Subject:  annot.Range().Ptr(),
		})
		return
	}
	if sp.Singular {
		if _, ok := visited[annot.GetName()]; ok {
			ctx.diag(&ksl.Diagnostic{
				Severity: ksl.DiagError,
				Summary:  "Duplicate directive",
				Detail:   fmt.Sprintf("Directive %q is already set.", annot.GetName()),
				Subject:  annot.Range().Ptr(),
			})
			return
		}
		visited[annot.GetName()] = struct{}{}
	}

	if ctx.diag(spec.ValidateArgs(sp.Arguments, sp.DefaultArg, annot)...) {
		return
	}

	switch annot.GetName() {
	case "backend":
		backend := Backend{}
		if arg, ok := annot.DefaultArg("name"); ok {
			ctx.diag(Eval(arg.Value, ctx.Context, &backend.Name)...)
		}
		backend.Source = id
		ctx.Config.Backend = &backend
	}
}

func (ctx *context) resolveModelAnnotations(modelID ModelID, model *ast.Model) {
	attr := ModelAnnotations{}

	// First resolve all the annotations defined on fields **in isolation**.
	for fid, field := range model.Fields {
		fid := FieldID(fid)
		if scalar, ok := ctx.Types.ScalarFields[MakeModelFieldID(modelID, fid)]; ok {
			ctx.visitModelScalarFieldAnnots(modelID, fid, model, field, &attr, scalar)
		} else if relation, ok := ctx.Types.RelationFields[MakeModelFieldID(modelID, fid)]; ok {
			ctx.visitModelRelationFieldAnnots(modelID, fid, model, field, &attr, relation)
		} else {
			ctx.diag(&ksl.Diagnostic{
				Severity: ksl.DiagError,
				Summary:  "Unknown field",
				Detail:   fmt.Sprintf("Field %q is neither a scalar, nor a relation, nor an enum.", field.GetName()),
				Subject:  field.Span.Ptr(),
			})
		}
	}

	// Resolve all the attributes defined on the model itself **in isolation**.
	visited := map[string]struct{}{}
	visit := func(annot *ast.Annotation, single bool) bool { return visitannot(ctx, annot, single, visited) }

	for aid, annot := range model.GetAnnotations() {
		sp, hasSpec := ctx.Spec.ModelBlockAnnotations[annot.GetName()]
		if !visit(annot, hasSpec && sp.Singular) {
			continue
		}

		switch annot.GetName() {
		case "ignore", "unique", "id", "map", "index":
		default:
			ctx.diag(&ksl.Diagnostic{
				Severity: ksl.DiagError,
				Summary:  "Invalid annotation",
				Detail:   fmt.Sprintf("The model %q cannot be annotated with %q.", model.GetName(), annot.GetName()),
				Subject:  annot.Span.Ptr(),
			})
			continue
		}

		if hasSpec && ctx.diag(spec.ValidateArgs(sp.Arguments, sp.DefaultArg, annot)...) {
			continue
		}

		switch annot.GetName() {
		case "ignore":
			attr.Ignored = true
		case "unique":
			var index = IndexAnnotation{
				Type:             IndexTypeUnique,
				SourceAnnotation: MakeAnnotID(modelID, IndexID(aid)),
			}
			if arg, ok := annot.DefaultArg("fields"); ok {
				index.Fields = ctx.resolveFields(arg, modelID)
			}

			if arg, ok := annot.Arg("name"); ok {
				ctx.diag(Eval(arg.Value, ctx.Context, &index.Name)...)
			}

			if arg, ok := annot.Arg("type"); ok {
				ctx.diag(Eval(arg.Value, ctx.Context, &index.Algorithm)...)
			}

			attr.Indexes = append(attr.Indexes, &index)

		case "id":
			id := IDAnnotation{
				SourceAnnot: MakeAnnotID(modelID, IndexID(aid)),
			}

			if arg, ok := annot.Arg("name"); ok {
				ctx.diag(Eval(arg.Value, ctx.Context, &id.Name)...)
			}

			if arg, ok := annot.Arg("map"); ok {
				ctx.diag(Eval(arg.Value, ctx.Context, &id.MappedName)...)
			}

			if arg, ok := annot.DefaultArg("fields"); ok {
				id.Fields = ctx.resolveFields(arg, modelID)
			}

			attr.PrimaryKey = &id
		case "map":
			if arg, ok := annot.Arg("name"); ok {
				ctx.diag(Eval(arg.Value, ctx.Context, &attr.Name)...)
			}
		case "index":
			var index = IndexAnnotation{
				Type:             IndexTypeNormal,
				SourceAnnotation: MakeAnnotID(modelID, IndexID(aid)),
			}

			if arg, ok := annot.Arg("name"); ok {
				ctx.diag(Eval(arg.Value, ctx.Context, &index.Name)...)
			}

			if arg, ok := annot.Arg("type"); ok {
				ctx.diag(Eval(arg.Value, ctx.Context, &index.Algorithm)...)
			}

			if arg, ok := annot.DefaultArg("fields"); ok {
				index.Fields = ctx.resolveFields(arg, modelID)
			}

			attr.Indexes = append(attr.Indexes, &index)
		}
	}

	ctx.validateIdFieldArities(modelID, attr)
	ctx.Types.ModelAnnotations[modelID] = attr
}

func (ctx *context) validateIdFieldArities(modelID ModelID, attr ModelAnnotations) {
	if attr.Ignored || attr.PrimaryKey == nil || attr.PrimaryKey.SourceField == nil {
		return
	}

	sourceField := ctx.Ast.GetModelField(MakeModelFieldID(modelID, *attr.PrimaryKey.SourceField))
	if sourceField.Type.Arity.IsAny(ast.Repeated, ast.Optional) {
		ctx.diag(&ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  "Invalid ID field",
			Detail:   fmt.Sprintf("The ID field %q cannot be repeated or optional.", sourceField.GetName()),
			Subject:  sourceField.Span.Ptr(),
		})
	}
}

func (ctx *context) resolveEnumAnnotations(enumID EnumID, enum *ast.Enum) {
	attr := EnumAnnotations{
		MappedValues: map[EnumValueID]string{},
	}

	for eid, field := range enum.Values {
		valueID := MakeEnumValueID(enumID, IndexID(eid))
		ctx.visitEnumFieldAnnots(valueID, enum, field, &attr)
	}

	visited := map[string]struct{}{}
	visit := func(annot *ast.Annotation, single bool) bool { return visitannot(ctx, annot, single, visited) }

	for _, annot := range enum.GetAnnotations() {
		sp, hasSpec := ctx.Spec.EnumBlockAnnotations[annot.GetName()]
		if !visit(annot, hasSpec && sp.Singular) {
			continue
		}

		switch annot.GetName() {
		case "map":
		default:
			ctx.diag(&ksl.Diagnostic{
				Severity: ksl.DiagError,
				Summary:  "Invalid annotation",
				Detail:   fmt.Sprintf("The enum value %q cannot be annotated with %q.", enum.GetName(), annot.GetName()),
				Subject:  annot.Span.Ptr(),
			})
			continue
		}

		if hasSpec && ctx.diag(spec.ValidateArgs(sp.Arguments, sp.DefaultArg, annot)...) {
			continue
		}

		switch annot.GetName() {
		case "map":
			if arg, ok := annot.DefaultArg("name"); ok {
				ctx.diag(Eval(arg.Value, ctx.Context, &attr.MappedName)...)
			}
		}
	}

	ctx.Types.EnumAnnotations[enumID] = attr
}

func (ctx *context) visitEnumFieldAnnots(enumValueID EnumValueID, enum *ast.Enum, field *ast.EnumValue, attrs *EnumAnnotations) {
	visited := map[string]struct{}{}
	visit := func(annot *ast.Annotation, single bool) bool { return visitannot(ctx, annot, single, visited) }

	for _, annot := range field.GetAnnotations() {
		sp, hasSpec := ctx.Spec.EnumFieldAnnotations[annot.GetName()]
		if !visit(annot, hasSpec && sp.Singular) {
			continue
		}

		switch annot.GetName() {
		case "map":
		default:
			ctx.diag(&ksl.Diagnostic{
				Severity: ksl.DiagError,
				Summary:  "Invalid annotation",
				Detail:   fmt.Sprintf("The enum value %q cannot be annotated with %q.", field.GetName(), annot.GetName()),
				Subject:  annot.Span.Ptr(),
			})
			continue
		}

		if hasSpec && ctx.diag(spec.ValidateArgs(sp.Arguments, sp.DefaultArg, annot)...) {
			continue
		}

		switch annot.GetName() {
		case "map":
			if arg, ok := annot.DefaultArg("name"); ok {
				var name string
				ctx.diag(Eval(arg.Value, ctx.Context, &name)...)
				for _, v := range attrs.MappedValues {
					if v == name {
						ctx.diag(&ksl.Diagnostic{
							Severity: ksl.DiagError,
							Summary:  "Duplicate enum value",
							Detail:   fmt.Sprintf("The enum value %q is already mapped to %q.", field.GetName(), name),
							Subject:  annot.Span.Ptr(),
						})
						continue
					}
				}
				if _, ok := enum.Values.Get(name); ok {
					ctx.diag(&ksl.Diagnostic{
						Severity: ksl.DiagError,
						Summary:  "Duplicate mapped name",
						Detail:   fmt.Sprintf("The mapped name %q is already in use.", name),
						Subject:  annot.Span.Ptr(),
					})
					continue
				}
				attrs.MappedValues[enumValueID] = name
			}
		}
	}
}

func (ctx *context) visitModelRelationFieldAnnots(modelID ModelID, fieldID FieldID, model *ast.Model, field *ast.Field, attrs *ModelAnnotations, relation *RelationField) {
	visited := map[string]struct{}{}
	visit := func(annot *ast.Annotation, single bool) bool { return visitannot(ctx, annot, single, visited) }

	for aid, annot := range field.GetAnnotations() {
		sp, hasSpec := ctx.Spec.ModelFieldRelationAnnotations[annot.GetName()]
		if !visit(annot, hasSpec && sp.Singular) {
			continue
		}

		switch annot.GetName() {
		case "ref", "ignore":
		case "unique", "default", "id", "map":
			ctx.diag(&ksl.Diagnostic{
				Severity: ksl.DiagError,
				Summary:  "Invalid annotation",
				Detail:   fmt.Sprintf("The field %q is a relation field and cannot be annotated with %q.", field.GetName(), annot.GetName()),
				Subject:  annot.Span.Ptr(),
			})
			continue
		default:
			ctx.diag(&ksl.Diagnostic{
				Severity: ksl.DiagError,
				Summary:  "Unknown annotation",
				Detail:   fmt.Sprintf("Annotation %q is not a known annotation for this field.", annot.GetName()),
				Subject:  annot.Span.Ptr(),
			})
			continue
		}

		if hasSpec && ctx.diag(spec.ValidateArgs(sp.Arguments, sp.DefaultArg, annot)...) {
			continue
		}

		switch annot.GetName() {
		case "ref":
			relation.AnnotationID = MakeAnnotID(MakeModelFieldID(modelID, fieldID), IndexID(aid))

			if arg, ok := annot.DefaultArg("name"); ok {
				ctx.diag(Eval(arg.Value, ctx.Context, &relation.Name)...)
			}

			if arg, ok := annot.Arg("fields"); ok {
				relation.Fields = ctx.resolveFields(arg, modelID)
			}

			if arg, ok := annot.Arg("references"); ok {
				relation.References = ctx.resolveFields(arg, relation.RefModelID)
			}

			if arg, ok := annot.Arg("onDelete"); ok {
				relation.OnDelete = &RefAction{Span: arg.Value.Range()}
				ctx.diag(Eval(arg.Value, ctx.Context, &relation.OnDelete.Action)...)
			}

			if arg, ok := annot.Arg("onUpdate"); ok {
				relation.OnUpdate = &RefAction{Span: arg.Value.Range()}
				ctx.diag(Eval(arg.Value, ctx.Context, &relation.OnUpdate.Action)...)
			}

		case "ignore":
			relation.Ignore = true
		}
	}
}

func (ctx *context) resolveFields(arg *ast.Argument, modelID ModelID) []FieldRef {
	model := ctx.Ast.GetModel(modelID)
	var refs []ast.Expression
	switch arg := arg.Value.(type) {
	case *ast.List:
		refs = arg.Elements[:]
	case *ast.Literal:
		refs = []ast.Expression{arg}
	}

	fields := map[ModelFieldID]FieldRef{}

	for _, ref := range refs {
		var name, sort string
		if ctx.decodeRefField(ref, &name, &sort) {
			// Does the field exist?
			mfid, ok := ctx.Ast.FindModelField(modelID, name)
			if !ok {
				ctx.diag(&ksl.Diagnostic{
					Severity: ksl.DiagError,
					Summary:  "Unknown field",
					Detail:   fmt.Sprintf("Field %q does not exist in model %q.", name, model.GetName()),
					Subject:  ref.Range().Ptr(),
				})
				continue
			}

			// Is the field a scalar field?
			if _, ok := ctx.Types.ScalarFields[mfid]; !ok {
				ctx.diag(&ksl.Diagnostic{
					Severity: ksl.DiagError,
					Summary:  "Invalid field",
					Detail:   "The argument fields must refer only to scalar fields.",
					Subject:  ref.Range().Ptr(),
				})
				continue
			}

			// Is the field used multiple times?
			if _, ok := fields[mfid]; ok {
				ctx.diag(&ksl.Diagnostic{
					Severity: ksl.DiagError,
					Summary:  "Duplicate field",
					Detail:   fmt.Sprintf("Field %q is referenced multiple times.", name),
					Subject:  ref.Range().Ptr(),
				})
				continue
			}
			fields[mfid] = FieldRef{ModelID: mfid.Model(), FieldID: mfid.Field(), Sort: sort}
		}
	}

	return valuesOf(fields)
}

func (ctx *context) visitModelScalarFieldAnnots(modelID ModelID, fieldID FieldID, model *ast.Model, field *ast.Field, attrs *ModelAnnotations, scalar *ScalarField) {
	visited := map[string]struct{}{}
	visit := func(annot *ast.Annotation, single bool) bool { return visitannot(ctx, annot, single, visited) }

	modelFieldNames := ctx.Names.ModelFields[model.GetName()]
	mappedFieldNames := map[string]struct{}{}

	for aid, annot := range field.GetAnnotations() {
		annotName := annot.GetName()
		if strings.HasPrefix(annotName, "db.") {
			ctx.visitNativeTypeAnnotation(field, annot, MakeAnnotID(MakeModelFieldID(modelID, fieldID), IndexID(aid)), scalar)
			continue
		}

		sp, hasSpec := ctx.Spec.ModelFieldScalarAnnotations[annotName]
		if !visit(annot, hasSpec && sp.Singular) {
			continue
		}

		switch annotName {
		case "id", "map", "default", "unique", "ignore":
		case "ref":
			ctx.diag(&ksl.Diagnostic{
				Severity: ksl.DiagError,
				Summary:  "Invalid annotation",
				Detail:   "Invalid field type, not a relation.",
				Subject:  annot.Span.Ptr(),
			})
			continue
		default:
			ctx.diag(&ksl.Diagnostic{
				Severity: ksl.DiagError,
				Summary:  "Unknown annotation",
				Detail:   fmt.Sprintf("Annotation %q is not a known annotation for this field.", annotName),
				Subject:  annot.Span.Ptr(),
			})
			continue
		}

		if hasSpec && ctx.diag(spec.ValidateArgs(sp.Arguments, sp.DefaultArg, annot)...) {
			continue
		}

		switch annotName {
		case "id":
			if attrs.PrimaryKey != nil {
				ctx.diag(&ksl.Diagnostic{
					Severity: ksl.DiagError,
					Summary:  "Invalid annotation",
					Detail:   "At most one field must be marked as the id field with the `@id` attribute.",
					Subject:  annot.Span.Ptr(),
				})
				continue
			}

			var fieldRef = FieldRef{ModelID: modelID, FieldID: fieldID}
			id := IDAnnotation{
				SourceAnnot: MakeAnnotID(MakeModelFieldID(modelID, fieldID), IndexID(aid)),
				SourceField: &fieldID,
			}
			if arg, ok := annot.DefaultArg("name"); ok {
				ctx.diag(Eval(arg.Value, ctx.Context, &id.Name)...)
			}
			if arg, ok := annot.Arg("map"); ok {
				ctx.diag(Eval(arg.Value, ctx.Context, &id.MappedName)...)
			}
			if arg, ok := annot.Arg("sort"); ok {
				ctx.diag(Eval(arg.Value, ctx.Context, &fieldRef.Sort)...)
			}
			id.Fields = append(id.Fields, fieldRef)
			attrs.PrimaryKey = &id

		case "map":
			if arg, ok := annot.DefaultArg("name"); ok {
				var name string
				ctx.diag(Eval(arg.Value, ctx.Context, &name)...)
				if _, ok := mappedFieldNames[name]; ok {
					ctx.diag(&ksl.Diagnostic{
						Severity: ksl.DiagError,
						Summary:  "Duplicate mapped name",
						Detail:   fmt.Sprintf("The mapped name %q is already mapped by another field.", name),
						Subject:  annot.Span.Ptr(),
					})
				} else if nodeID, ok := modelFieldNames[name]; ok {
					if _, ok := ctx.Types.ScalarFields[nodeID.(ModelFieldID)]; ok {
						ctx.diag(&ksl.Diagnostic{
							Severity: ksl.DiagError,
							Summary:  "Unknown mapped name",
							Detail:   fmt.Sprintf("The mapped name %q is already used by another field.", name),
							Subject:  annot.Span.Ptr(),
						})
					}
				} else {
					scalar.MappedName = name
				}
				mappedFieldNames[name] = struct{}{}
			}

		case "unique":
			var fieldRef = FieldRef{ModelID: modelID, FieldID: fieldID}
			index := IndexAnnotation{
				Type:             IndexTypeUnique,
				SourceAnnotation: MakeAnnotID(fieldID, IndexID(aid)),
				SourceField:      &fieldID,
			}
			if arg, ok := annot.DefaultArg("name"); ok {
				ctx.diag(Eval(arg.Value, ctx.Context, &index.Name)...)
			}
			if arg, ok := annot.Arg("sort"); ok {
				ctx.diag(Eval(arg.Value, ctx.Context, &fieldRef.Sort)...)
			}
			index.Fields = append(index.Fields, fieldRef)
			attrs.Indexes = append(attrs.Indexes, &index)

		case "ignore":
			scalar.Ignored = true

		case "default":
			annotID := MakeAnnotID(MakeModelFieldID(modelID, fieldID), IndexID(aid))
			scalar.Default = &DefaultAnnotation{SourceAnnot: annotID}
			if arg, ok := annot.DefaultArg("value"); ok {
				scalar.Default.Value = arg.Value
			}
		}
	}
}
func (ctx *context) visitNativeTypeAnnotation(field *ast.Field, annot *ast.Annotation, annotID AnnotID, scalar *ScalarField) {
	if _, ok := scalar.FieldType.Type.(BuiltInScalarType); !ok {
		ctx.diag(&ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  "Invalid annotation",
			Detail:   fmt.Sprintf("The field %q is not a scalar field and cannot be annotated with \"@%s\".", field.GetName(), annot.GetName()),
			Subject:  annot.Span.Ptr(),
		})
		return
	}

	annotName := annot.GetName()
	if scalar.NativeType != nil {
		ctx.diag(&ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  "Invalid annotation",
			Detail:   fmt.Sprintf("The field %q is already annotated with a native type.", field.GetName()),
			Subject:  annot.Span.Ptr(),
		})
		return
	}

	name := strings.TrimPrefix(annotName, "db.")
	if strings.Contains(name, ".") {
		ctx.diag(&ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  "Invalid annotation",
			Detail:   fmt.Sprintf("The annotation %q is not a valid native type.", annotName),
			Subject:  annot.Span.Ptr(),
		})
		return
	}

	args := make([]string, len(annot.GetArgs()))
	for i, arg := range annot.GetArgs() {
		ctx.diag(Eval(arg.Value, ctx.Context, &args[i])...)
	}
	scalar.NativeType = &NativeTypeAnnotation{
		Name:             name,
		Args:             args,
		SourceAnnotation: annotID,
	}
}

func valuesOf[K comparable, V any](m map[K]V) []V {
	vals := make([]V, 0, len(m))
	for _, v := range m {
		vals = append(vals, v)
	}
	return vals
}

// func keysOf[K comparable, V any](m map[K]V) []K {
// 	keys := make([]K, 0, len(m))
// 	for k := range m {
// 		keys = append(keys, k)
// 	}
// 	return keys
// }

func visitannot(ctx *context, annot *ast.Annotation, single bool, visited map[string]struct{}) bool {
	valid := true
	if _, ok := visited[annot.GetName()]; ok && single {
		ctx.diag(&ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  "Duplicate annotation",
			Detail:   fmt.Sprintf("Annotation %q is specified more than once.", annot.GetName()),
			Subject:  annot.Span.Ptr(),
		})
		valid = false
	}
	visited[annot.GetName()] = struct{}{}
	return valid
}

func (ctx *context) decodeRefField(expr ast.Expression, name, sort *string) bool {
	switch e := expr.(type) {
	case *ast.List:
		switch len(e.Elements) {
		case 0:
		case 1:
			return ctx.decodeRefField(e.Elements[0], name, sort)
		default:
			ctx.diag(&ksl.Diagnostic{
				Severity: ksl.DiagError,
				Summary:  "Invalid argument",
				Detail:   "Expected a single reference field.",
				Subject:  e.Range().Ptr(),
			})
			return false
		}
	case *ast.Literal:
		*name = e.Value
	case *ast.String:
		*name = e.Value
	case *ast.Function:
		*name = e.GetName()
		var seenSort bool
		for _, arg := range e.GetArgs() {
			switch arg.GetName() {
			case "sort":
				if seenSort {
					ctx.diag(&ksl.Diagnostic{
						Severity: ksl.DiagError,
						Summary:  "Duplicate sort argument",
						Detail:   "The sort argument is specified more than once.",
						Subject:  arg.Span.Ptr(),
					})
					continue
				}
				seenSort = true
				ctx.diag(Eval(arg.Value, ctx.Context, sort)...)
			default:
				ctx.diag(&ksl.Diagnostic{
					Severity: ksl.DiagError,
					Summary:  "Invalid field reference argument",
					Detail:   fmt.Sprintf("Invalid field reference argument %q", arg.GetName()),
					Subject:  arg.Range().Ptr(),
				})
			}
		}
	default:
		ctx.diag(&ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  "Invalid field reference",
			Subject:  expr.Range().Ptr(),
		})
		return false
	}
	return true
}
