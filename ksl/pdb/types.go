package pdb

import (
	"fmt"
	"ksl"
	"ksl/syntax/ast"
)

type TypesContext struct {
	ScalarFields            map[ModelFieldID]*ScalarField
	RelationFields          map[ModelFieldID]*RelationField
	EnumFields              map[ModelFieldID]*ScalarField
	EnumAnnotations         map[EnumID]EnumAnnotations
	ModelAnnotations        map[ModelID]ModelAnnotations
	UnknownFunctionDefaults []ModelFieldID
}

type ScalarField struct {
	FieldType  ScalarFieldType
	Ignored    bool
	Default    *DefaultAnnotation
	MappedName string
	NativeType *NativeTypeAnnotation
}

type NativeTypeAnnotation struct {
	Name string
	Args []string

	SourceAnnotation AnnotID
}

func (ctx *context) ResolveTypes() {
	for eid, entry := range ctx.Ast.Tops {
		switch entry := entry.(type) {
		case *ast.Model:
			ctx.visitModel(ModelID(eid), entry)
		case *ast.Enum:
			ctx.visitEnum(EnumID(eid), entry)
		case *ast.Block:
			ctx.visitBlock(BlockID(eid), entry)
		}
	}
}

func (ctx *context) visitModel(modelID ModelID, model *ast.Model) {
	for fid, field := range model.Fields {
		mfid := MakeModelFieldID(modelID, FieldID(fid))
		ft := ctx.getFieldType(field.Type.GetNameNode())

		switch ft := ft.(type) {
		case ModelFieldType:
			ctx.Types.RelationFields[mfid] = &RelationField{ModelID: mfid.Model(), FieldID: mfid.Field(), RefModelID: ft.Model}
		case ScalarFieldType:
			ctx.Types.ScalarFields[mfid] = &ScalarField{FieldType: ft}
		default:
			ctx.diag(&ksl.Diagnostic{
				Severity: ksl.DiagError,
				Summary:  "Unknown type",
				Detail:   fmt.Sprintf("Type %q is neither a built-in type, nor refers to another model or enum.", field.Type.GetName()),
				Subject:  field.Type.Span.Ptr(),
			})
		}
	}
}

func (ctx *context) visitEnum(enumID EnumID, enum *ast.Enum) {
	if len(enum.Values) == 0 {
		ctx.diag(&ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  "Empty enum",
			Detail:   "Enums must have at least one value.",
			Subject:  enum.Span.Ptr(),
		})
	}
}

func (ctx *context) visitBlock(blockID BlockID, block *ast.Block) {}

func (ctx *context) getFieldType(typ *ast.Name) FieldType {
	typeStr := typ.String()

	if typ, ok := ksl.BuiltIns.From(typeStr); ok {
		return ScalarFieldType{Type: BuiltInScalarType(typ)}
	}
	if nodeID, ok := ctx.Names.ModelEnums[typeStr]; ok {
		switch nodeID := nodeID.(type) {
		case EnumID:
			return ScalarFieldType{Type: EnumFieldType{Enum: nodeID}}
		case ModelID:
			return ModelFieldType{Model: nodeID}
		}
	}
	return nil
}

type BuiltInScalarType ksl.BuiltInScalar

func (t BuiltInScalarType) String() string {
	return ksl.BuiltInScalar(t).Name()
}

func (t BuiltInScalarType) Deref() ksl.BuiltInScalar { return ksl.BuiltInScalar(t) }

type FieldType interface{ fieldtyp() }
type ModelFieldType struct{ Model ModelID }
type ScalarFieldType struct{ Type ScalarType }

type ScalarType interface{ scalartyp() }
type EnumFieldType struct{ Enum EnumID }

func (BuiltInScalarType) scalartyp() {}
func (EnumFieldType) scalartyp()     {}

func (ModelFieldType) fieldtyp()  {}
func (ScalarFieldType) fieldtyp() {}
