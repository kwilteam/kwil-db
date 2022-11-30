package pdb

import (
	"fmt"
	"ksl"
	"ksl/syntax/nodes"
	"strings"
	"unicode"
	"unicode/utf8"
)

type NamesContext struct {
	ModelEnums  map[string]NodeID
	Blocks      map[string]map[string]NodeID
	ModelFields map[string]map[string]NodeID
	EnumFields  map[string]map[string]NodeID
}

// ResolveNames is responsible for validating that there are no name collisions in the following namespaces:
//   - Model and enum names
//   - Block names
//   - Model fields for each model
//   - Enum variants for each enum
func (ctx *context) ResolveNames() {
	insertName := func(entry nodes.NamedNode, eid NodeID, names map[string]NodeID, schemaType string) {
		name := entry.GetName()
		if oid, ok := names[name]; ok {
			other := nodes.GetTopLevelType(ctx.Ast.GetNode(oid).(nodes.TopLevel))
			ctx.diag(&ksl.Diagnostic{
				Severity: ksl.DiagError,
				Summary:  "Duplicate name",
				Detail:   fmt.Sprintf("The name %q is already defined as a %s.", name, other),
				Subject:  entry.Range().Ptr(),
			})
		} else {
			names[name] = eid
		}
	}

	for eid, entry := range ctx.Ast.Tops {
		ctx.validateIsNotReservedScalarType(entry.GetNameNode())

		switch entry := entry.(type) {
		case *nodes.Enum:
			ctx.validateName(entry.GetNameNode(), "enum")
			ctx.validateEnumName(entry)
			ctx.validateAnnotationNames(entry)
			insertName(entry, EnumID(eid), ctx.Names.ModelEnums, "enum")

			names := make(map[string]NodeID)
			for evid, ev := range entry.Values {
				ctx.validateName(ev.GetNameNode(), "enum value")
				ctx.validateAnnotationNames(ev)
				insertName(ev, EnumID(evid), names, "enum value")
			}
			ctx.Names.EnumFields[entry.GetName()] = names

		case *nodes.Model:
			ctx.validateName(entry.GetNameNode(), "model")
			ctx.validateModelName(entry)
			ctx.validateAnnotationNames(entry)
			insertName(entry, ModelID(eid), ctx.Names.ModelEnums, "model")

			names := make(map[string]NodeID)
			for fid, fld := range entry.Fields {
				mfid := MakeModelFieldID(ModelID(eid), FieldID(fid))
				ctx.validateName(fld.GetNameNode(), "field")
				ctx.validateAnnotationNames(fld)
				insertName(fld, mfid, names, "field")
			}
			ctx.Names.ModelFields[entry.GetName()] = names

		case *nodes.Block:
			typ := entry.GetType()
			if _, ok := ctx.Names.Blocks[typ]; !ok {
				ctx.Names.Blocks[typ] = make(map[string]NodeID)
			}

			ctx.validateName(entry.GetNameNode(), entry.GetType())
			ctx.validateBlockName(entry)
			ctx.checkForDuplicateProperties(entry, entry.Properties, entry.GetType())
			insertName(entry, BlockID(eid), ctx.Names.Blocks[typ], entry.GetType())

		case *nodes.Annotation:
			ctx.validateName(entry.GetNameNode(), "annotation")
		}
	}
}

func (c *context) validateName(ident *nodes.Name, schemaType string) {
	name := ident.String()
	if name == "" {
		c.diag(&ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  "Missing name",
			Detail:   "Every " + schemaType + " must have a name.",
			Subject:  ident.Span.Ptr(),
		})
	} else if rn, _ := utf8.DecodeRuneInString(name); unicode.IsDigit(rn) {
		c.diag(&ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  "Invalid name",
			Detail:   "The name of a " + schemaType + " must not start with a number.",
			Subject:  ident.Span.Ptr(),
		})
	} else if strings.ContainsAny(name, "-") {
		c.diag(&ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  "Invalid name",
			Detail:   "The name of a " + schemaType + " must not contain a dash.",
			Subject:  ident.Span.Ptr(),
		})
	}
}

func (c *context) checkForDuplicateProperties(entry nodes.NamedNode, props nodes.Properties, schemaType string) {
	seen := map[string]struct{}{}
	for _, prop := range props {
		if _, ok := seen[prop.GetName()]; ok {
			c.diag(&ksl.Diagnostic{
				Severity: ksl.DiagError,
				Summary:  "Duplicate property",
				Detail:   fmt.Sprintf("The property %q is already defined on the %s %q.", prop.GetName(), schemaType, entry.GetName()),
				Subject:  prop.Range().Ptr(),
			})
		} else {
			seen[prop.GetName()] = struct{}{}
		}
	}
}

func (c *context) validateAnnotationNames(entry nodes.Annotated) {
	for _, annotation := range entry.GetAnnotations() {
		c.validateName(annotation.GetNameNode(), "annotation")
	}
}

func (c *context) validateEnumName(enum *nodes.Enum) {
	c.validateNameIsNotReserved(enum.GetNameNode(), "enum")
}

func (c *context) validateModelName(model *nodes.Model) {
	c.validateNameIsNotReserved(model.GetNameNode(), "model")
}

func (c *context) validateBlockName(blk *nodes.Block) {
	c.validateNameIsNotReserved(blk.GetNameNode(), blk.GetType())
}

func (c *context) validateNameIsNotReserved(name *nodes.Name, schemaType string) {
	if _, ok := reservedNames[name.String()]; ok {
		c.diag(&ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  "Reserved name",
			Detail:   fmt.Sprintf("The %s name %q is a reserved name.", schemaType, name.String()),
			Subject:  name.Span.Ptr(),
		})
	}
}

func (c *context) validateIsNotReservedScalarType(ident *nodes.Name) {
	if _, ok := ksl.BuiltIns.From(ident.String()); ok {
		c.diag(&ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  "Reserved name",
			Detail:   "The name " + ident.String() + " is reserved for a scalar type.",
			Subject:  ident.Span.Ptr(),
		})
	}
}

var reservedNames = map[string]struct{}{
	"true":  {},
	"false": {},
	"null":  {},
}
