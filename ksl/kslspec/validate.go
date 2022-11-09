package kslspec

import (
	"fmt"
	"strings"

	"ksl"
	"ksl/kslsyntax/ast"
)

type Validator interface {
	Validate(*ast.Document) ksl.Diagnostics
}

type validator struct {
	schema   *DocumentSpec
	refTypes map[string]struct{}
}

func NewValidator(schema *DocumentSpec) Validator {
	return &validator{schema: schema, refTypes: map[string]struct{}{}}
}

func Validate(doc *ast.Document, schema *DocumentSpec) ksl.Diagnostics {
	return NewValidator(schema).Validate(doc)
}

func (v *validator) Validate(doc *ast.Document) ksl.Diagnostics {
	schemaDirectives := map[string]*DirectiveSpec{}
	schemaBlocks := map[string]*BlockSpec{}

	for _, dir := range v.schema.Directives {
		schemaDirectives[dir.Name] = dir
	}

	for _, block := range v.schema.Blocks {
		schemaBlocks[block.Type] = block
	}

	var diags ksl.Diagnostics

	for _, d := range doc.Directives {
		dirSpec := schemaDirectives[d.Name.GetString()]
		if dirSpec == nil {
			diags = append(diags, &ksl.Diagnostic{
				Severity: ksl.DiagError,
				Summary:  DiagUnknownDirective,
				Detail:   fmt.Sprintf("%q is not a valid directive type.", d.Name.GetString()),
				Subject:  d.Name.Range().Ptr(),
			})
		}
		diags = append(diags, v.validateDirective(d, dirSpec)...)
	}

	groupedBlocks := map[string][]*ast.Block{}
	for _, b := range doc.Blocks {
		typ := b.GetType()
		if s, ok := schemaBlocks[typ]; ok {
			if s.CanReferenceAsType {
				v.refTypes[b.GetName()] = struct{}{}
			}
		}
		groupedBlocks[typ] = append(groupedBlocks[typ], b)
	}

	for _, grp := range groupedBlocks {
		seenBlocks := map[string]struct{}{}
		for _, b := range grp {
			blockSpec := schemaBlocks[b.Type.Value]
			if blockSpec == nil {
				diags = append(diags, &ksl.Diagnostic{
					Severity: ksl.DiagError,
					Summary:  DiagUnknownBlockType,
					Detail:   fmt.Sprintf("%q is not a valid block type.", b.Type.Value),
					Subject:  b.Type.Range().Ptr(),
				})
				continue
			}
			if _, ok := seenBlocks[b.GetName()]; ok && !blockSpec.AllowMultiple {
				diags = append(diags, &ksl.Diagnostic{
					Severity: ksl.DiagError,
					Summary:  DiagDuplicateBlock,
					Detail:   fmt.Sprintf("A block of type %q with the name %q was already defined.", b.GetType(), b.GetName()),
					Subject:  b.Range().Ptr(),
				})
				continue
			}

			diags = append(diags, v.validateBlock(b, blockSpec)...)
		}
	}

	return diags
}

func (v *validator) validateDirective(dir *ast.Directive, schema *DirectiveSpec) ksl.Diagnostics {
	if schema == nil {
		return nil
	}

	var diags ksl.Diagnostics

	switch {
	case schema.Key.IsRequired() && dir.GetKey() == "":
		diags = append(diags, &ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  DiagMissingDirectiveKey,
			Detail:   fmt.Sprintf("A key is required for %q directives.", dir.GetName()),
			Subject:  dir.Range().Ptr(),
		})
	case schema.Key.IsNotAllowed() && dir.GetKey() != "":
		diags = append(diags, &ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  DiagUnexpectedDirectiveKey,
			Detail:   fmt.Sprintf("Keys are not valid for %q directives.", dir.GetName()),
			Subject:  dir.Key.Range().Ptr(),
		})
	}

	switch {
	case schema.ValueOption.IsNotAllowed() && dir.Value != nil:
		diags = append(diags, &ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  DiagUnexpectedDirectiveValue,
			Detail:   fmt.Sprintf("%q directives do not have values.", dir.GetName()),
			Subject:  dir.Value.Range().Ptr(),
		})
	case schema.ValueOption.IsRequired() && dir.Value == nil:
		diags = append(diags, &ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  DiagMissingDirectiveValue,
			Detail:   fmt.Sprintf("%q directives require a value.", dir.GetName()),
			Subject:  dir.Range().Ptr(),
		})
	case dir.Value != nil && !schema.ValueOption.IsNotAllowed():
		diags = append(diags, schema.Value.Validate(dir.Value)...)
	}

	return diags
}

func (v *validator) validateBlock(block *ast.Block, schema *BlockSpec) ksl.Diagnostics {
	if schema == nil {
		return nil
	}

	types := map[string]*TypeSpec{}

	for _, t := range schema.Types {
		types[t.Name] = t
	}

	labelSpecs := map[string]*LabelSpec{}
	modifiers := map[string]struct{}{}

	var diags ksl.Diagnostics

	for _, a := range schema.Labels {
		labelSpecs[a.Name] = a
	}
	for _, a := range schema.Modifiers {
		modifiers[a] = struct{}{}
	}

	switch {
	case schema.Name.IsRequired() && block.GetName() == "":
		diags = append(diags, &ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  DiagMissingBlockName,
			Detail:   fmt.Sprintf("A name is required for %q blocks.", block.GetType()),
			Subject:  ksl.Range{Start: block.Type.SrcRange.End, End: block.Type.SrcRange.End}.Ptr(),
		})
	case schema.Name.IsNotAllowed() && block.GetName() != "":
		diags = append(diags, &ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  DiagUnexpectedBlockName,
			Detail:   fmt.Sprintf("Names are not valid for %q blocks.", block.GetType()),
			Subject:  block.Name.Range().Ptr(),
		})
	}

	if block.Modifier != nil {
		if _, ok := modifiers[block.Modifier.Value]; !ok {
			var detail string
			if len(modifiers) == 0 {
				detail = fmt.Sprintf("%q is not a valid modifier. %q blocks do not allow modifier keywords.", block.Modifier.Value, block.GetType())
			} else {
				detail = fmt.Sprintf("%q is not a valid modifier. Valid modifiers for %q blocks are: %s.", block.Modifier.Value, block.GetType(), strings.Join(sortedKeys(modifiers), ", "))
			}

			diags = append(diags, &ksl.Diagnostic{
				Severity: ksl.DiagError,
				Summary:  DiagUnknownBlockModifier,
				Detail:   detail,
				Subject:  block.Modifier.Range().Ptr(),
			})
		}
	}

	if block.Labels != nil {
		for _, label := range block.Labels.Values {
			labelSpec := labelSpecs[label.Name.Value]
			if labelSpec == nil {
				diags = append(diags, &ksl.Diagnostic{
					Severity: ksl.DiagError,
					Summary:  DiagUnknownBlockLabel,
					Detail:   fmt.Sprintf("%q is not a valid label. Valid labels for %q blocks are: %s.", label.Name.Value, block.GetType(), strings.Join(sortedKeys(labelSpecs), ", ")),
					Subject:  label.Name.Range().Ptr(),
				})
			}
			diags = append(diags, v.validateLabel(label, labelSpec)...)
		}
	}

	diags = append(diags, v.validateBlockBody(block.Body, schema, types)...)

	return diags
}

func (v *validator) validateBlockBody(body *ast.Body, schema *BlockSpec, types map[string]*TypeSpec) ksl.Diagnostics {
	var diags ksl.Diagnostics

	enumValues := map[string]*ast.Str{}
	annotationSpecs := map[string]*FunctionSpec{}
	blockAnnotationSpecs := map[string]*FunctionSpec{}
	blockSpecs := map[string]*BlockSpec{}
	attrSpecs := map[string]*AttributeSpec{}

	for _, a := range schema.Annotations {
		annotationSpecs[a.Name] = a
	}
	for _, a := range schema.BlockAnnotations {
		blockAnnotationSpecs[a.Name] = a
	}
	for _, a := range schema.Blocks {
		blockSpecs[a.Type] = a
	}
	for _, a := range schema.Attributes {
		attrSpecs[a.Key] = a
	}

	if schema.Mode.IsConfiguration() {
		diags = append(diags, v.validateAttributes(body.Attributes, attrSpecs, schema)...)
	} else {
		for _, a := range body.Attributes {
			diags = append(diags, &ksl.Diagnostic{
				Severity: ksl.DiagError,
				Summary:  DiagUnexpectedAttribute,
				Detail:   fmt.Sprintf("%q blocks do not allow attributes.", schema.Type),
				Subject:  a.Range().Ptr(),
			})
		}
	}

	if schema.Mode.IsConfiguration() || schema.Mode.IsDefinition() {
		for _, b := range body.Blocks {
			blockSpec := blockSpecs[b.Type.Value]
			if blockSpec == nil {
				diags = append(diags, &ksl.Diagnostic{
					Severity: ksl.DiagError,
					Summary:  DiagUnknownBlockType,
					Detail:   fmt.Sprintf("%q is not a valid child block type. Valid child block types for %q blocks are: %s.", b.GetType(), schema.Type, strings.Join(sortedKeys(blockSpecs), ", ")),
					Subject:  b.Type.Range().Ptr(),
				})
			}
			diags = append(diags, v.validateBlock(b, blockSpec)...)
		}
	} else {
		for _, b := range body.Blocks {
			diags = append(diags, &ksl.Diagnostic{
				Severity: ksl.DiagError,
				Summary:  DiagUnexpectedBlock,
				Detail:   fmt.Sprintf("%q blocks do not allow nested blocks.", schema.Type),
				Subject:  b.Range().Ptr(),
			})
		}
	}

	if schema.Mode.IsDefinition() {
		for _, a := range body.Annotations {
			annotationSpec := blockAnnotationSpecs[a.Name.Value]
			if annotationSpec == nil {
				diags = append(diags, &ksl.Diagnostic{
					Severity: ksl.DiagError,
					Summary:  DiagUnknownAnnotation,
					Detail:   fmt.Sprintf("%q is not a valid annotation. Valid annotations for %q blocks are: %s.", a.GetName(), schema.Type, strings.Join(sortedKeys(blockAnnotationSpecs), ", ")),
					Subject:  a.Name.Range().Ptr(),
				})
			}
			diags = append(diags, v.validateAnnotation(a, annotationSpec)...)
		}

		seendDefinitions := map[string]struct{}{}
		for _, decl := range body.Definitions {
			if _, ok := seendDefinitions[decl.Name.Value]; ok && schema.UniqueKeys {
				diags = append(diags, &ksl.Diagnostic{
					Severity: ksl.DiagError,
					Summary:  DiagDuplicateDefinition,
					Detail:   fmt.Sprintf("Duplicate definition for %q.", decl.Name.Value),
					Subject:  decl.Name.Range().Ptr(),
				})
			}
			diags = append(diags, v.validateDefinition(decl, annotationSpecs, types)...)

		}
	} else {
		for _, a := range body.Annotations {
			diags = append(diags, &ksl.Diagnostic{
				Severity: ksl.DiagError,
				Summary:  DiagUnexpectedAnnotation,
				Detail:   fmt.Sprintf("Annotations are not allowed in %q blocks.", schema.Type),
				Subject:  a.Range().Ptr(),
			})
		}

		for _, decl := range body.Definitions {
			diags = append(diags, &ksl.Diagnostic{
				Severity: ksl.DiagError,
				Summary:  DiagUnexpectedDeclaration,
				Detail:   fmt.Sprintf("Declarations are not allowed in %q blocks.", schema.Type),
				Subject:  decl.Range().Ptr(),
			})
		}
	}

	if schema.Mode.IsEnum() && schema.UniqueKeys {
		for _, v := range body.EnumValues {
			if _, ok := enumValues[v.Value]; ok {
				diags = append(diags, &ksl.Diagnostic{
					Severity: ksl.DiagError,
					Summary:  DiagDuplicateEnumValue,
					Detail:   fmt.Sprintf("%q has already been declared in this enum block. Enum values must be unique.", v.Value),
					Subject:  v.Range().Ptr(),
				})
				continue
			}
			enumValues[v.Value] = v
		}
	} else {
		for _, v := range body.EnumValues {
			diags = append(diags, &ksl.Diagnostic{
				Severity: ksl.DiagError,
				Summary:  DiagUnexpectedEnumValue,
				Detail:   fmt.Sprintf("Enum values are not allowed in %q blocks.", schema.Type),
				Subject:  v.Range().Ptr(),
			})
		}
	}

	return diags
}

func (v *validator) validateLabel(label *ast.Attribute, schema *LabelSpec) ksl.Diagnostics {
	if schema == nil {
		return nil
	}
	var diags ksl.Diagnostics

	switch {
	case schema.ValueOption.IsNotAllowed() && label.Value != nil:
		diags = append(diags, &ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  DiagUnexpectedLabelValue,
			Detail:   fmt.Sprintf("The %q label does not expect a value.", label.Name.Value),
			Subject:  label.Value.Range().Ptr(),
		})
	case schema.ValueOption.IsRequired() && label.Value == nil:
		diags = append(diags, &ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  DiagMissingLabelValue,
			Detail:   fmt.Sprintf("The %q label expects a value.", label.Name.Value),
			Subject:  ksl.Range{Start: label.Name.Range().End, End: label.Name.Range().End}.Ptr(),
		})
	case label.Value != nil && !schema.ValueOption.IsNotAllowed():
		diags = append(diags, schema.Value.Validate(label.Value)...)
	}

	return diags
}

func (v *validator) validateAnnotation(ann *ast.Annotation, schema *FunctionSpec) ksl.Diagnostics {
	if schema == nil {
		return nil
	}

	var diags ksl.Diagnostics
	if ann.ArgList != nil {
		diags = append(diags, v.validateKwargs(ann, schema.Kwargs)...)
		diags = append(diags, v.validateArgs(ann, schema.Args)...)
	}

	return diags
}

func (v *validator) validateArgs(fn *ast.Annotation, argSpecs []*ArgSpec) ksl.Diagnostics {
	var diags ksl.Diagnostics

	last := len(fn.ArgList.Args)
	if last > len(argSpecs) {
		last = len(argSpecs)
	}

	for i := 0; i < last; i++ {
		arg := fn.ArgList.Args[i]
		sch := argSpecs[i]
		diags = append(diags, sch.Value.Validate(arg)...)
	}

	if len(fn.ArgList.Args) > len(argSpecs) {
		for i := last; i < len(fn.ArgList.Args); i++ {
			arg := fn.ArgList.Args[i]
			diags = append(diags, &ksl.Diagnostic{
				Severity: ksl.DiagError,
				Summary:  DiagUnexpectedArg,
				Detail:   fmt.Sprintf("Invalid argument at position %d. %q expects at most %d positional arguments.", i, fn.GetName(), len(argSpecs)),
				Subject:  arg.Range().Ptr(),
			})
		}
	} else if len(fn.ArgList.Args) < len(argSpecs) {
		var start, end ksl.Pos
		if len(fn.ArgList.Args) == 0 {
			start, end = fn.ArgList.LParen.Range.End, fn.ArgList.LParen.Range.End
		} else if len(fn.ArgList.Kwargs) > 0 {
			rng := fn.ArgList.Kwargs[0].Range()
			start, end = rng.Start, rng.Start
		} else {
			rng := fn.ArgList.Args[len(fn.ArgList.Args)-1].Range()
			start, end = rng.End, rng.End
		}
		rng := ksl.Range{Start: start, End: end}

		for i := last; i < len(argSpecs); i++ {
			s := argSpecs[i]
			if !s.Optional {
				diags = append(diags, &ksl.Diagnostic{
					Severity: ksl.DiagError,
					Summary:  DiagMissingRequiredArg,
					Detail:   fmt.Sprintf("Missing required argument %q at position %d.", s.Name, i),
					Subject:  rng.Ptr(),
				})
			}
		}
	}

	return diags
}

func (v *validator) validateKwargs(fn *ast.Annotation, specs []*ArgSpec) ksl.Diagnostics {
	var diags ksl.Diagnostics

	kwargs := make(map[string]*ast.Attribute)
	seenKwargs := make(map[string]struct{})

	for _, kwarg := range fn.ArgList.Kwargs {
		kwargs[kwarg.Name.Value] = kwarg
	}

	for _, s := range specs {
		if kwarg, ok := kwargs[s.Name]; ok {
			diags = append(diags, s.Value.Validate(kwarg.Value)...)
			seenKwargs[s.Name] = struct{}{}
		} else if !s.Optional {
			diags = append(diags, &ksl.Diagnostic{
				Severity: ksl.DiagError,
				Summary:  DiagMissingRequiredKeywordArg,
				Detail:   fmt.Sprintf("Missing required keyword argument %q for function %q.", s.Name, fn.GetName()),
				Subject:  fn.Name.Range().Ptr(),
			})
		}
	}

	for _, kw := range fn.ArgList.Kwargs {
		if _, ok := seenKwargs[kw.Name.Value]; !ok {
			diags = append(diags, &ksl.Diagnostic{
				Severity: ksl.DiagError,
				Summary:  DiagUnexpectedKeywordArg,
				Detail:   fmt.Sprintf("Unexpected keyword argument %q for function %q.", kw.Name.Value, fn.GetName()),
				Subject:  kw.Name.Range().Ptr(),
			})
		}
	}

	return diags
}

func (v *validator) validateAttributes(props ast.Attributes, specs map[string]*AttributeSpec, blkSpec *BlockSpec) ksl.Diagnostics {
	var diags ksl.Diagnostics
	propsMap := make(map[string]*ast.Attribute)

	for _, prop := range props {
		if !blkSpec.Mode.IsConfiguration() {
			diags = append(diags, &ksl.Diagnostic{
				Severity: ksl.DiagError,
				Summary:  DiagUnexpectedAttribute,
				Detail:   fmt.Sprintf("%q blocks do not allow attributes.", blkSpec.Type),
				Subject:  prop.Range().Ptr(),
			})
			continue
		}

		if blkSpec.UniqueKeys {
			if kv, ok := propsMap[prop.GetName()]; ok {
				diags = append(diags, &ksl.Diagnostic{
					Severity: ksl.DiagError,
					Summary:  DiagDuplicateAttribute,
					Detail:   fmt.Sprintf("%q has already been set in this block. Attribute keys must be unique.", prop.GetName()),
					Subject:  prop.Range().Ptr(),
					Context:  kv.Range().Ptr(),
				})
				continue
			}
		}

		propsMap[prop.GetName()] = prop
		if s, ok := specs[prop.GetName()]; ok {
			valDiags := s.Value.Validate(prop.Value)
			diags = append(diags, valDiags...)
		} else {
			if !blkSpec.ExtraProps {
				diags = append(diags, &ksl.Diagnostic{
					Severity: ksl.DiagError,
					Summary:  DiagUnknownAttribute,
					Detail:   fmt.Sprintf("%q is not a known attribute of %q.", prop.GetName(), blkSpec.Type),
					Subject:  prop.Name.Range().Ptr(),
				})
			}
		}
	}

	if blkSpec.Mode.IsConfiguration() {
		for _, s := range specs {
			prop, ok := propsMap[s.Key]
			if !ok && !s.Optional {
				diags = append(diags, &ksl.Diagnostic{
					Severity: ksl.DiagError,
					Summary:  DiagMissingRequiredAttribute,
					Detail:   fmt.Sprintf("%q blocks require an attribute named %q to be set.", blkSpec.Type, s.Key),
					Subject:  prop.Range().Ptr(),
				})
			}
		}
	}

	return diags
}

func (v *validator) validateDefinition(decl *ast.Definition, annotations map[string]*FunctionSpec, types map[string]*TypeSpec) ksl.Diagnostics {
	if len(annotations) == 0 {
		return nil
	}

	var diags ksl.Diagnostics
	seenAnnots := make(map[string]*ast.Annotation)
	typeAnnots := make(map[string]*TypeAnnotation)

	if typeSpec, ok := types[decl.GetTypeName()]; ok {
		for _, annot := range typeSpec.Annotations {
			typeAnnots[annot.Name] = annot
		}
	} else if _, ok := v.refTypes[decl.GetTypeName()]; !ok {
		diags = append(diags, &ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  DiagUnknownDeclarationType,
			Detail:   fmt.Sprintf("%q is not a valid declaration type.", decl.GetTypeName()),
			Subject:  decl.Type.Range().Ptr(),
		})
	}

	for _, a := range decl.Annotations {
		if len(annotations) == 0 {
			diags = append(diags, &ksl.Diagnostic{
				Severity: ksl.DiagError,
				Summary:  DiagUnexpectedAnnotation,
				Detail:   "Declaration annotations are not allowed in this block.",
				Subject:  a.Name.Range().Ptr(),
			})
		}
		if _, ok := seenAnnots[a.GetName()]; ok {
			diags = append(diags, &ksl.Diagnostic{
				Severity: ksl.DiagError,
				Summary:  DiagDuplicateAnnotation,
				Detail:   fmt.Sprintf("An annotation named %q has already been declared for this declaration.", a.GetName()),
				Subject:  a.Name.Range().Ptr(),
			})
			continue
		}
		seenAnnots[a.GetName()] = a
		if _, ok := annotations[a.GetName()]; !ok {
			if _, ok := typeAnnots[a.GetName()]; !ok {
				diags = append(diags, &ksl.Diagnostic{
					Severity: ksl.DiagError,
					Summary:  DiagUnknownAnnotation,
					Detail:   fmt.Sprintf("%q is not a valid annotation for type %q.", a.GetName(), decl.GetTypeName()),
					Subject:  a.Name.Range().Ptr(),
				})
			}
		}
		diags = append(diags, v.validateAnnotation(a, annotations[a.GetName()])...)
	}

	for name, typeAnnot := range typeAnnots {
		if annot, ok := seenAnnots[name]; ok {
			diags = append(diags, v.validateAnnotation(annot, typeAnnot.FuncSpec())...)
		} else {
			if typeAnnot.IsRequired {
				diags = append(diags, &ksl.Diagnostic{
					Severity: ksl.DiagError,
					Summary:  DiagMissingRequiredAnnotation,
					Detail:   fmt.Sprintf("The %q annotation is required for type %q.", name, decl.GetTypeName()),
					Subject:  decl.Type.Range().Ptr(),
				})
			}
		}
	}

	return diags
}
