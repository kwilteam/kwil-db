package kslspec

import (
	"fmt"
	"strings"

	"golang.org/x/exp/slices"
	"ksl"
	"ksl/kslsyntax/ast"
)

type File struct {
	Body  *ast.Document
	Bytes []byte
}

type FileSet struct {
	Files map[string]*File
}

func (f *FileSet) Data() []byte {
	var data []byte
	for _, file := range f.Files {
		data = append(data, file.Bytes...)
		data = append(data, '\n')
	}
	return data
}

func (f *FileSet) Directives() ast.Directives {
	var directives ast.Directives
	for _, file := range f.Files {
		directives = append(directives, file.Body.Directives...)
	}
	return directives
}

func (f *FileSet) Blocks() ast.Blocks {
	var blocks ast.Blocks
	for _, file := range f.Files {
		blocks = append(blocks, file.Body.Blocks...)
	}
	return blocks
}

type ExpressionSpec interface {
	Validate(ast.Expr) ksl.Diagnostics
}

func NoSpec() *DocumentSpec { return nil }

type DocumentSpec struct {
	Directives []*DirectiveSpec
	Blocks     []*BlockSpec
}

func (d *DocumentSpec) WithBlocks(blocks ...*BlockSpec) *DocumentSpec {
	d.Blocks = blocks
	return d
}

func (d *DocumentSpec) WithDirectives(directives ...*DirectiveSpec) *DocumentSpec {
	d.Directives = directives
	return d
}

type DirectiveSpec struct {
	Name        string
	Key         Option
	Value       ExpressionSpec
	ValueOption Option
}

func (d *DirectiveSpec) RequireKey() *DirectiveSpec {
	d.Key = RequiredOption
	return d
}

func (d *DirectiveSpec) NoKey() *DirectiveSpec {
	d.Key = NotAllowedOption
	return d
}

func (d *DirectiveSpec) WithValue(value ExpressionSpec) *DirectiveSpec {
	d.Value = value
	return d
}

func (d *DirectiveSpec) RequireValue() *DirectiveSpec {
	d.ValueOption = RequiredOption
	return d
}

func (d *DirectiveSpec) OptionalValue() *DirectiveSpec {
	d.ValueOption = OptionalOption
	return d
}

func (d *DirectiveSpec) NoValue() *DirectiveSpec {
	d.ValueOption = NotAllowedOption
	return d
}

type BlockMode int

func (m BlockMode) IsConfiguration() bool { return m&BlockModeConfig == BlockModeConfig }
func (m BlockMode) IsDefinition() bool    { return m&BlockModeDefine == BlockModeDefine }
func (m BlockMode) IsEnum() bool          { return m&BlockModeEnum == BlockModeEnum }

const (
	BlockModeConfig = 1 << iota
	BlockModeDefine
	BlockModeEnum

	BlockModeAny = BlockModeConfig | BlockModeDefine | BlockModeEnum
)

type BlockSpec struct {
	Type string
	Name Option
	Mode BlockMode

	ExtraProps         bool
	UniqueKeys         bool
	AllowMultiple      bool
	CanReferenceAsType bool

	Labels           []*LabelSpec
	Modifiers        []string
	Annotations      []*FunctionSpec
	BlockAnnotations []*FunctionSpec
	Blocks           []*BlockSpec
	Attributes       []*AttributeSpec
	Types            []*TypeSpec
}

func (b *BlockSpec) Reference() *BlockSpec {
	b.CanReferenceAsType = true
	return b
}

func (b *BlockSpec) NonReference() *BlockSpec {
	b.CanReferenceAsType = false
	return b
}

func (d *BlockSpec) NonUniqueKeys() *BlockSpec {
	d.UniqueKeys = false
	return d
}

func (d *BlockSpec) AllowMultipleSameName() *BlockSpec {
	d.AllowMultiple = true
	return d
}

func (d *BlockSpec) WithTypes(types ...*TypeSpec) *BlockSpec {
	d.Types = types
	return d
}

func (b *BlockSpec) WithExtraAttributes() *BlockSpec {
	b.ExtraProps = true
	return b
}

func (b *BlockSpec) WithBlocks(blocks ...*BlockSpec) *BlockSpec {
	b.Blocks = blocks
	return b
}

func (b *BlockSpec) WithAnnotations(annotations ...*FunctionSpec) *BlockSpec {
	b.Annotations = annotations
	return b
}

func (b *BlockSpec) WithBlockAnnotations(annotations ...*FunctionSpec) *BlockSpec {
	b.BlockAnnotations = annotations
	return b
}

func (b *BlockSpec) WithModifiers(modifiers ...string) *BlockSpec {
	b.Modifiers = modifiers
	return b
}

func (b *BlockSpec) WithLabels(labels ...*LabelSpec) *BlockSpec {
	b.Labels = labels
	return b
}

func (b *BlockSpec) RequireName() *BlockSpec {
	b.Name = RequiredOption
	return b
}

func (b *BlockSpec) NoName() *BlockSpec {
	b.Name = NotAllowedOption
	return b
}

func (b *BlockSpec) WithAttributes(attrs ...*AttributeSpec) *BlockSpec {
	b.Attributes = attrs
	return b
}

type AttributeSpec struct {
	Key      string
	Value    ExpressionSpec
	Optional bool
}

type LabelSpec struct {
	Name        string
	Value       ExpressionSpec
	Optional    bool
	ValueOption Option
}

func (l *LabelSpec) WithOptionalValue(value ExpressionSpec) *LabelSpec {
	l.Value = value
	l.ValueOption = OptionalOption
	return l
}

func (l *LabelSpec) WithRequiredValue(value ExpressionSpec) *LabelSpec {
	l.Value = value
	l.ValueOption = RequiredOption
	return l
}

type FunctionSpec struct {
	Name   string
	Args   []*ArgSpec
	Kwargs []*ArgSpec
}

func (f *FunctionSpec) WithArgs(specs ...*ArgSpec) *FunctionSpec {
	var args []*ArgSpec
	var kwargs []*ArgSpec

	for _, a := range specs {
		if a.Positional {
			args = append(args, a)
		} else {
			kwargs = append(kwargs, a)
		}
	}

	f.Args = args
	f.Kwargs = kwargs
	return f
}

type ArgSpec struct {
	Name       string
	Value      ExpressionSpec
	Optional   bool
	Positional bool
}

type TypeSpec struct {
	Name        string
	Type        string
	Aliases     []string
	Annotations map[string]*TypeAnnotation
}

func (t *TypeSpec) WithAnnots(annots ...*TypeAnnotation) *TypeSpec {
	t.Annotations = make(map[string]*TypeAnnotation)
	for _, a := range annots {
		t.Annotations[a.Name] = a
	}
	return t
}

func (t *TypeSpec) Mappings(mappings ...string) *TypeSpec {
	t.Aliases = mappings
	return t
}

func (t *TypeSpec) HasAnnotation(name string) bool {
	_, ok := t.Annotations[name]
	return ok
}

type TypeSpecOption func(*TypeSpec)

type TypeAnnotation struct {
	Name       string
	Kind       TypeKind
	IsRequired bool
}

func (t *TypeAnnotation) Required() *TypeAnnotation {
	t.IsRequired = true
	return t
}

func (t *TypeAnnotation) FuncSpec() *FunctionSpec {
	return &FunctionSpec{Name: t.Name, Args: []*ArgSpec{{Name: t.Name, Value: &LiteralExprSpec{Kind: t.Kind}}}}
}

type Option int

func (o Option) IsOptional() bool   { return o == OptionalOption }
func (o Option) IsRequired() bool   { return o == RequiredOption }
func (o Option) IsNotAllowed() bool { return o == NotAllowedOption }

const (
	OptionalOption Option = iota
	RequiredOption
	NotAllowedOption
)

type SpecBuilder struct{}

func (SpecBuilder) Document() *DocumentSpec {
	return &DocumentSpec{}
}

func (SpecBuilder) Directive(name string) *DirectiveSpec {
	return &DirectiveSpec{Name: name, Key: OptionalOption, Value: &LiteralExprSpec{Kind: AnyKind}, ValueOption: RequiredOption}
}

func (SpecBuilder) ConfigBlock(typeName string) *BlockSpec {
	return &BlockSpec{
		Type:               typeName,
		Name:               OptionalOption,
		Mode:               BlockModeConfig,
		UniqueKeys:         true,
		CanReferenceAsType: true,
	}
}

func (SpecBuilder) DefinitionBlock(typeName string) *BlockSpec {
	return &BlockSpec{
		Type:               typeName,
		Name:               OptionalOption,
		Mode:               BlockModeDefine,
		UniqueKeys:         true,
		CanReferenceAsType: true,
	}
}

func (SpecBuilder) EnumBlock(typeName string) *BlockSpec {
	return &BlockSpec{
		Type:               typeName,
		Name:               RequiredOption,
		Mode:               BlockModeEnum,
		UniqueKeys:         true,
		CanReferenceAsType: true,
	}
}

func (SpecBuilder) Attr(key string, v ExpressionSpec) *AttributeSpec {
	return &AttributeSpec{Key: key, Value: v}
}

func (SpecBuilder) OptionalAttr(name string, v ExpressionSpec) *AttributeSpec {
	return &AttributeSpec{Key: name, Value: v, Optional: true}
}

func (SpecBuilder) Label(name string) *LabelSpec {
	return &LabelSpec{Name: name, ValueOption: NotAllowedOption, Value: &LiteralExprSpec{Kind: AnyKind}}
}

func (SpecBuilder) Arg(name string, v ExpressionSpec) *ArgSpec {
	return &ArgSpec{Name: name, Optional: false, Value: v, Positional: true}
}

func (SpecBuilder) OptionalArg(name string, v ExpressionSpec) *ArgSpec {
	return &ArgSpec{Name: name, Optional: true, Value: v, Positional: true}
}

func (SpecBuilder) Kwarg(name string, v ExpressionSpec) *ArgSpec {
	return &ArgSpec{Name: name, Optional: false, Value: v, Positional: false}
}

func (SpecBuilder) OptionalKwarg(name string, v ExpressionSpec) *ArgSpec {
	return &ArgSpec{Name: name, Optional: true, Value: v, Positional: false}
}

func (s SpecBuilder) Type(name string, dbtype string, opts ...TypeSpecOption) *TypeSpec {
	t := &TypeSpec{Name: name, Type: dbtype}
	for _, opt := range opts {
		opt(t)
	}
	return t
}

func (s SpecBuilder) TypeAnnot(name string, val *LiteralExprSpec) *TypeAnnotation {
	return &TypeAnnotation{Name: name, Kind: val.Kind}
}

func (s SpecBuilder) RequiredTypeAnnot(name string, val *LiteralExprSpec) *TypeAnnotation {
	return &TypeAnnotation{Name: name, Kind: val.Kind, IsRequired: true}
}

func (SpecBuilder) Enum(values ...string) *EnumSpec {
	return &EnumSpec{Values: values}
}

func (SpecBuilder) Object(attrs ...*AttributeSpec) *ObjectSpec {
	attrMap := make(map[string]*AttributeSpec, len(attrs))
	for _, attr := range attrs {
		attrMap[attr.Key] = attr
	}
	return &ObjectSpec{Attributes: attrMap}
}

func (SpecBuilder) OneOf(e ...ExpressionSpec) *OneOfExpression     { return &OneOfExpression{Specs: e} }
func (s SpecBuilder) ExprOrList(e ExpressionSpec) *OneOfExpression { return s.OneOf(s.List(e), e) }
func (SpecBuilder) Func(name string) *FunctionSpec                 { return &FunctionSpec{Name: name} }
func (SpecBuilder) List(v ExpressionSpec) *ListSpec                { return &ListSpec{Elem: v} }
func (SpecBuilder) Any() *LiteralExprSpec                          { return &LiteralExprSpec{Kind: AnyKind} }
func (SpecBuilder) Int() *LiteralExprSpec                          { return &LiteralExprSpec{Kind: IntKind} }
func (SpecBuilder) Float() *LiteralExprSpec                        { return &LiteralExprSpec{Kind: FloatKind} }
func (SpecBuilder) String() *LiteralExprSpec                       { return &LiteralExprSpec{Kind: StringKind} }
func (SpecBuilder) Bool() *LiteralExprSpec                         { return &LiteralExprSpec{Kind: BoolKind} }
func (SpecBuilder) Ref() *LiteralExprSpec                          { return &LiteralExprSpec{Kind: TypeRefKind} }

type TypeKind uint

func (k TypeKind) Has(n TypeKind) bool { return k&n != 0 }
func (k TypeKind) String() string {
	if k == 0 {
		return "unknown"
	}

	var kinds []string
	if k.Has(IntKind) {
		kinds = append(kinds, "int")
	}
	if k.Has(FloatKind) {
		kinds = append(kinds, "float")
	}
	if k.Has(StringKind) {
		kinds = append(kinds, "string")
	}
	if k.Has(BoolKind) {
		kinds = append(kinds, "bool")
	}
	if k.Has(ObjectKind) {
		kinds = append(kinds, "object")
	}
	if k.Has(TypeRefKind) {
		kinds = append(kinds, "type")
	}
	if k.Has(NullKind) {
		kinds = append(kinds, "null")
	}
	return strings.Join(kinds, "|")
}

const NoneKind TypeKind = 0

const (
	IntKind TypeKind = 1 << iota
	FloatKind
	StringKind
	BoolKind
	TypeRefKind
	NullKind
	VariableKind

	FuncKind
	ObjectKind
	ListKind

	PrimitiveKind = IntKind | FloatKind | StringKind | BoolKind | NullKind | TypeRefKind | VariableKind
	NumberKind    = IntKind | FloatKind
	AnyKind       = 1<<64 - 1
)

type OneOfExpression struct {
	Specs   []ExpressionSpec
	Message string
}

func (v *OneOfExpression) WithMessage(msg string) *OneOfExpression {
	v.Message = msg
	return v
}

func (v *OneOfExpression) Validate(node ast.Expr) ksl.Diagnostics {
	for _, v := range v.Specs {
		if !v.Validate(node).HasErrors() {
			return nil
		}
	}

	return ksl.Diagnostics{
		&ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  v.Message,
			Subject:  node.Range().Ptr(),
		},
	}
}

type ListSpec struct {
	Elem ExpressionSpec
}

func (s *ListSpec) Validate(e ast.Expr) ksl.Diagnostics {
	if array, ok := e.(*ast.List); ok {
		var diags ksl.Diagnostics
		for _, elem := range array.Values {
			elemDiags := s.Elem.Validate(elem)
			diags = append(diags, elemDiags...)
		}
		return diags
	} else {
		return ksl.Diagnostics{{
			Severity: ksl.DiagError,
			Summary:  "Invalid list",
			Detail:   "Expected a list",
			Subject:  e.Range().Ptr(),
		}}
	}
}

type ObjectSpec struct {
	Attributes map[string]*AttributeSpec
}

func (v *ObjectSpec) Validate(e ast.Expr) ksl.Diagnostics {
	var diags ksl.Diagnostics

	if obj, ok := e.(*ast.Object); ok {
		items := make(map[string]ksl.Expression, len(obj.Attributes))
		for _, kv := range obj.Attributes {
			attrSpec, ok := v.Attributes[kv.Name.Value]
			if !ok && !attrSpec.Optional {
				diags = append(diags, &ksl.Diagnostic{
					Severity: ksl.DiagError,
					Summary:  "Invalid attribute",
					Detail:   fmt.Sprintf("Attribute %q is not expected here.", kv.Name.Value),
					Subject:  kv.Range().Ptr(),
				})
				continue
			}
			if _, ok := items[kv.Name.Value]; ok {
				diags = append(diags, &ksl.Diagnostic{
					Severity: ksl.DiagError,
					Summary:  "Duplicate attribute",
					Detail:   fmt.Sprintf("Attribute %q is already set.", kv.Name.Value),
					Subject:  kv.Range().Ptr(),
				})
				continue
			}

			diags = append(diags, attrSpec.Value.Validate(kv.Value)...)
		}
		for _, attrSpec := range v.Attributes {
			if _, ok := items[attrSpec.Key]; !ok && !attrSpec.Optional {
				diags = append(diags, &ksl.Diagnostic{
					Severity: ksl.DiagError,
					Summary:  "Missing attribute",
					Detail:   fmt.Sprintf("Attribute %q is required.", attrSpec.Key),
					Subject:  e.Range().Ptr(),
				})
			}
		}
	} else {
		diags = append(diags, &ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  "Invalid object",
			Detail:   "Expected an object",
			Subject:  e.Range().Ptr(),
		})
	}

	return diags
}

type EnumSpec struct {
	Values []string
}

func (e *EnumSpec) WithValues(values ...string) *EnumSpec {
	e.Values = values
	return e
}

func (v *EnumSpec) Validate(node ast.Expr) ksl.Diagnostics {
	var diags ksl.Diagnostics
	var value string
	switch node := node.(type) {
	case *ast.Str:
		value = node.Value
	case *ast.QuotedStr:
		value = node.Value
	default:
		diags = append(diags, &ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  "Invalid enum value",
			Detail:   fmt.Sprintf("Value must be one of %s", strings.Join(v.Values, ", ")),
			Subject:  node.Range().Ptr(),
		})
	}

	if value != "" && !slices.Contains(v.Values, value) {
		diags = append(diags, &ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  "Invalid enum value",
			Detail:   fmt.Sprintf("%q is not valid for this annotation. Expected one of %s", value, strings.Join(v.Values, ", ")),
			Subject:  node.Range().Ptr(),
		})
	}

	return diags
}

type LiteralExprSpec struct {
	Kind      TypeKind
	CanBeNull bool
}

func (v *LiteralExprSpec) Nullable(val bool) *LiteralExprSpec {
	v.CanBeNull = val
	return v
}

func (v *LiteralExprSpec) Validate(e ast.Expr) ksl.Diagnostics {
	var kind TypeKind
	var diags ksl.Diagnostics

	switch e := e.(type) {
	case *ast.Float:
		kind = FloatKind
	case *ast.Int:
		kind = IntKind
	case *ast.Number:
		kind = NumberKind
	case *ast.Str:
		kind = TypeRefKind
	case *ast.QuotedStr:
		kind = StringKind
	case *ast.Heredoc:
		kind = StringKind
	case *ast.Bool:
		kind = BoolKind
	case *ast.Null:
		kind = NullKind
	case *ast.Object:
		kind = ObjectKind
	case *ast.List:
		kind = ListKind
	case *ast.FunctionCall:
		kind = FuncKind
	case *ast.Var:
		kind = VariableKind
	case nil:
		if v.CanBeNull {
			return nil
		}
		kind = NullKind
	default:
		diags = append(diags, &ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  "Invalid value type",
			Detail:   fmt.Sprintf("Expected %s, but got unknown", v.Kind),
			Subject:  e.Range().Ptr(),
		})
		return diags
	}

	if kind&v.Kind != kind {
		diags = append(diags, &ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  "Invalid value type",
			Detail:   fmt.Sprintf("Expected %s, but got %s", v.Kind, kind),
			Subject:  e.Range().Ptr(),
		})
	}

	return diags
}
