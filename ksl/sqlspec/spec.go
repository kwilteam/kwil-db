package sqlspec

import (
	"ksl"
	"ksl/kslparse"
	"ksl/kslspec"
)

var sp = kslspec.SpecBuilder{}

func NewSpec() *kslspec.DocumentSpec {
	referenceOptions := sp.Enum(string(NoAction), string(Restrict), string(Cascade), string(SetNull), string(SetDefault))

	return sp.Document().
		WithBlocks(
			sp.DefinitionBlock("table").
				RequireName().
				WithTypes(TypeRegistry.Specs()...).
				WithBlockAnnotations(
					sp.Func("id").WithArgs(sp.Arg("columns", sp.List(sp.Ref()))),
					sp.Func("index").WithArgs(
						sp.OptionalArg("name", sp.String()),
						sp.Kwarg("columns", sp.List(sp.Ref())),
						sp.OptionalKwarg("type", sp.Enum(IndexTypeBTree, IndexTypeHash, IndexTypeGIN, IndexTypeGiST, IndexTypeBRIN)),
						sp.OptionalKwarg("unique", sp.Bool()),
					),
					sp.Func("foreign_key").WithArgs(
						sp.OptionalArg("name", sp.String()),
						sp.Kwarg("columns", sp.List(sp.Ref())),
						sp.Kwarg("references", sp.List(sp.Ref())),
						sp.OptionalKwarg("on_delete", referenceOptions),
						sp.OptionalKwarg("on_update", referenceOptions),
					),
				).
				WithAnnotations(
					sp.Func("id"),
					sp.Func("default").WithArgs(sp.Arg("value", sp.Any())),
					sp.Func("unique"),
					sp.Func("index").WithArgs(
						sp.OptionalArg("name", sp.String()),
						sp.OptionalKwarg("type", sp.Enum(IndexTypeBTree, IndexTypeHash, IndexTypeGIN, IndexTypeGiST, IndexTypeBRIN)),
						sp.OptionalKwarg("unique", sp.Bool()),
					),
					sp.Func("foreign_key").WithArgs(
						sp.Arg("column", sp.Ref()),
						sp.OptionalKwarg("name", sp.String()),
						sp.OptionalKwarg("on_delete", referenceOptions),
						sp.OptionalKwarg("on_update", referenceOptions),
					),
				),
			sp.EnumBlock("enum").RequireName(),
			sp.ConfigBlock("query").
				RequireName().
				WithAttributes(sp.Attr("statement", sp.String())),
			sp.ConfigBlock("role").
				RequireName().
				WithModifiers("extends").
				WithLabels(sp.Label("default").WithOptionalValue(sp.Bool())).
				WithAttributes(
					sp.Attr("allow", sp.List(sp.Ref())),
					sp.OptionalAttr("default", sp.Bool()),
				),
		)
}

func Decode(file *kslspec.FileSet) (*Realm, ksl.Diagnostics) {
	return newDecoder(Spec, TypeRegistry).decodeFileSet(file)
}

func Unmarshal(data []byte, filename string) (*Realm, ksl.Diagnostics) {
	parser := kslparse.NewParser()
	file, diags := parser.Parse(data, filename)
	if diags.HasErrors() {
		return nil, diags
	}
	return newDecoder(Spec, TypeRegistry).decode(file)
}

func UnmarshalFile(path string) (*Realm, ksl.Diagnostics) {
	parser := kslparse.NewParser()
	file, diags := parser.ParseFile(path)
	if diags.HasErrors() {
		return nil, diags
	}
	return newDecoder(Spec, TypeRegistry).decode(file)
}

var Spec = NewSpec()
var TypeRegistry = kslspec.NewRegistry(
	kslspec.WithParser(ParseType),
	kslspec.WithSpecFunc(TypeSpec),
	kslspec.WithFormatter(PrintType),
	kslspec.WithSpecs(
		sp.Type("date", TypeDate),
		sp.Type("time", TypeTime).Mappings(TypeTimeTZ, TypeTimeWOTZ, TypeTimeWTZ).WithAnnots(sp.TypeAnnot("precision", sp.Int())),
		sp.Type("datetime", TypeTimestamp).Mappings(TypeTimestampTZ, TypeTimestampWOTZ, TypeTimestampWTZ),
		sp.Type("string", TypeVarChar).Mappings(TypeCharVar, TypeChar, TypeCharacter, TypeText, TypeUUID, TypeXML, TypeJSON).WithAnnots(sp.TypeAnnot("size", sp.Int())),
		sp.Type("int32", TypeInt4).Mappings(TypeInteger, TypeSerial, TypeSerial2, TypeSerial4, TypeInt, TypeNumeric, TypeSmallInt, TypeSmallSerial, TypeInt2),
		sp.Type("int64", TypeInt8).Mappings(TypeBigInt, TypeBigSerial),
		sp.Type("bool", TypeBool).Mappings(TypeBoolean),
	),
)
