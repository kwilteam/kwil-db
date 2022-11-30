package spec

import (
	"fmt"
	"strings"
)

type SchemaSpec struct {
	Directives map[string]AnnotSpec
	Blocks     map[string]BlockSpec

	ModelBlockAnnotations         map[string]AnnotSpec
	ModelFieldScalarAnnotations   map[string]AnnotSpec
	ModelFieldRelationAnnotations map[string]AnnotSpec
	EnumBlockAnnotations          map[string]AnnotSpec
	EnumFieldAnnotations          map[string]AnnotSpec
}

type BlockSpec struct {
	Type         string
	NamePresence Presence
	Body         ObjectSpec
}

type AnnotSpec struct {
	Name       string
	Arguments  map[string]ArgSpec
	Required   bool
	Singular   bool
	DefaultArg string
}

type FuncSpec struct {
	Name       string
	Arguments  map[string]ArgSpec
	DefaultArg string
}

type ArgSpec struct {
	Name     string
	Value    ValueSpec
	Required bool
	Default  bool
}

type PropertySpec struct {
	Name     string
	Value    ValueSpec
	Required bool
}

type OneOfSpec struct {
	Options []ValueSpec
}

type ValueSpec interface{ val() }

type ScalarValueSpec struct {
	Kind ScalarKind
}

type ConstantValueSpec struct {
	Value string
}

type EnumSpec struct {
	Values []string
}

type ListSpec struct {
	ElementType ValueSpec
}

type ObjectSpec struct {
	AllowExtraProps bool
	Properties      map[string]PropertySpec
}

type Presence int

func (o Presence) IsOptional() bool   { return o == OptionalPresence }
func (o Presence) IsRequired() bool   { return o == RequiredPresence }
func (o Presence) IsNotAllowed() bool { return o == NoPresence }

const (
	OptionalPresence Presence = iota
	RequiredPresence
	NoPresence
)

type FieldType uint

const (
	FieldTypeUnknown FieldType = 0
	FieldTypeScalar  FieldType = 1 << (iota - 1)
	FieldTypeEnum
	FieldTypeRelation
)

type ScalarKind uint

func (k ScalarKind) String() string {
	switch k {
	case NoKind:
		return "Null"
	case AnyKind:
		return "Any"
	case IntKind:
		return "Int"
	case FloatKind:
		return "Float"
	case StringLitKind:
		return "Literal"
	case QuotedStringKind:
		return "String"
	case BoolKind:
		return "Bool"
	default:
		return "Unknown"
	}
}

const (
	NoKind  ScalarKind = 0
	IntKind ScalarKind = 1 << (iota - 1)
	FloatKind
	QuotedStringKind
	BoolKind
	StringLitKind
	NumberKind            = IntKind | FloatKind
	AnyKind    ScalarKind = (1 << 64) - 1
)

func (ScalarValueSpec) val()   {}
func (ConstantValueSpec) val() {}
func (ListSpec) val()          {}
func (ObjectSpec) val()        {}
func (OneOfSpec) val()         {}
func (FuncSpec) val()          {}
func (EnumSpec) val()          {}

func GetTypeDescription(v ValueSpec) string {
	switch v := v.(type) {
	case ScalarValueSpec:
		return v.Kind.String()
	case ConstantValueSpec:
		return fmt.Sprintf("%q", v.Value)
	case ListSpec:
		return fmt.Sprintf("%s[]", GetTypeDescription(v.ElementType))
	case ObjectSpec:
		return "Object"
	case OneOfSpec:
		possibilities := make([]string, len(v.Options))
		for i, opt := range v.Options {
			possibilities[i] = GetTypeDescription(opt)
		}
		return strings.Join(possibilities, " | ")
	case FuncSpec:
		fn := v.Name
		if fn == "" {
			fn = "func"
		}
		args := make([]string, 0, len(v.Arguments))
		for _, arg := range v.Arguments {
			argStr := fmt.Sprintf("%s: %s", arg.Name, GetTypeDescription(arg.Value))
			if !arg.Required {
				argStr += "?"
			}
			args = append(args, argStr)
		}
		return fmt.Sprintf("%s(%s)", fn, strings.Join(args, ", "))
	case EnumSpec:
		return strings.Join(v.Values, " | ")
	default:
		return "unknown"
	}
}

func Schema(opts ...SchemaOption) SchemaSpec {
	s := SchemaSpec{}
	for _, opt := range opts {
		opt(&s)
	}
	return s
}

func Block(typ string, opts ...BlockOption) BlockSpec {
	b := BlockSpec{Type: typ, NamePresence: RequiredPresence}
	for _, opt := range opts {
		opt(&b)
	}
	return b
}

func Func(opts ...FuncOption) FuncSpec {
	f := FuncSpec{}
	for _, opt := range opts {
		opt(&f)
	}
	for _, arg := range f.Arguments {
		if arg.Default {
			f.DefaultArg = arg.Name
			break
		}
	}
	return f
}

func Annot(name string, opts ...AnnotOption) AnnotSpec {
	a := AnnotSpec{Name: name}
	for _, opt := range opts {
		opt(&a)
	}
	for _, arg := range a.Arguments {
		if arg.Default {
			a.DefaultArg = arg.Name
			break
		}
	}
	return a
}

func RequiredAnnot(name string, opts ...AnnotOption) AnnotSpec {
	a := AnnotSpec{Name: name, Required: true}
	for _, opt := range opts {
		opt(&a)
	}
	return a
}

func Prop(name string, value ValueSpec, opts ...PropOption) PropertySpec {
	p := PropertySpec{Name: name, Value: value}
	for _, opt := range opts {
		opt(&p)
	}
	return p
}

func RequiredProp(name string, value ValueSpec, opts ...PropOption) PropertySpec {
	p := PropertySpec{Name: name, Value: value, Required: true}
	for _, opt := range opts {
		opt(&p)
	}
	return p
}

func Arg(name string, value ValueSpec, opts ...ArgOption) ArgSpec {
	a := ArgSpec{Name: name, Required: false, Value: value}
	for _, opt := range opts {
		opt(&a)
	}
	return a
}

func RequiredArg(name string, value ValueSpec, opts ...ArgOption) ArgSpec {
	a := ArgSpec{Name: name, Required: true, Value: value}
	for _, opt := range opts {
		opt(&a)
	}
	return a
}

func Enum(values ...string) EnumSpec { return EnumSpec{Values: values} }

func FieldRefs() ValueSpec {
	allowed := OneOf(
		StringLit(),
		Func(
			OptFunc.Args(
				RequiredArg("sort", Enum("Asc", "Desc")),
			),
		),
	)
	return OneOf(List(allowed), allowed)
}

func OneOf(specs ...ValueSpec) ValueSpec { return OneOfSpec{Options: specs} }
func Constant(value string) ValueSpec    { return ConstantValueSpec{Value: value} }
func DbGenerated() FuncSpec {
	return Func(
		OptFunc.Named("dbgenerated"),
		OptFunc.Args(
			RequiredArg("value", String(), OptArg.Default()),
		),
	)
}

func AnyScalar() ValueSpec       { return ScalarValueSpec{Kind: AnyKind} }
func Int() ValueSpec             { return ScalarValueSpec{Kind: IntKind} }
func Float() ValueSpec           { return ScalarValueSpec{Kind: FloatKind} }
func StringLit() ValueSpec       { return ScalarValueSpec{Kind: StringLitKind} }
func String() ValueSpec          { return ScalarValueSpec{Kind: QuotedStringKind} }
func Bool() ValueSpec            { return ScalarValueSpec{Kind: BoolKind} }
func List(v ValueSpec) ValueSpec { return ListSpec{ElementType: v} }
func Object(opts ...ObjectOption) ValueSpec {
	o := ObjectSpec{}
	for _, opt := range opts {
		opt(&o)
	}
	return o
}

var OptSchema schemaopts
var OptAnnot annotation
var OptFunc function
var OptArg argopts
var OptProp propopts
var OptObj objopts
var OptBlock blockopts

type BlockOption func(*BlockSpec)
type ObjectOption func(*ObjectSpec)
type PropOption func(*PropertySpec)
type ArgOption func(*ArgSpec)
type FuncOption func(*FuncSpec)
type AnnotOption func(*AnnotSpec)
type SchemaOption func(*SchemaSpec)

type schemaopts struct{}

func (schemaopts) WithDirectives(directives ...AnnotSpec) SchemaOption {
	return func(s *SchemaSpec) {
		s.Directives = make(map[string]AnnotSpec)
		for _, d := range directives {
			s.Directives[d.Name] = d
		}
	}
}

func (schemaopts) WithBlocks(blocks ...BlockSpec) SchemaOption {
	return func(s *SchemaSpec) {
		s.Blocks = make(map[string]BlockSpec)
		for _, b := range blocks {
			s.Blocks[b.Type] = b
		}
	}
}

func (schemaopts) WithModelBlockAnnotations(annots ...AnnotSpec) SchemaOption {
	return func(s *SchemaSpec) {
		annotMap := make(map[string]AnnotSpec, len(annots))
		for _, a := range annots {
			annotMap[a.Name] = a
		}
		s.ModelBlockAnnotations = annotMap
	}
}

func (schemaopts) WithModelFieldScalarAnnotations(annots ...AnnotSpec) SchemaOption {
	return func(s *SchemaSpec) {
		annotMap := make(map[string]AnnotSpec, len(annots))
		for _, a := range annots {
			annotMap[a.Name] = a
		}
		s.ModelFieldScalarAnnotations = annotMap
	}
}

func (schemaopts) WithModelFieldRelationAnnotations(annots ...AnnotSpec) SchemaOption {
	return func(s *SchemaSpec) {
		annotMap := make(map[string]AnnotSpec, len(annots))
		for _, a := range annots {
			annotMap[a.Name] = a
		}
		s.ModelFieldRelationAnnotations = annotMap
	}
}

func (schemaopts) WithEnumBlockAnnotations(annots ...AnnotSpec) SchemaOption {
	return func(s *SchemaSpec) {
		annotMap := make(map[string]AnnotSpec, len(annots))
		for _, a := range annots {
			annotMap[a.Name] = a
		}
		s.EnumBlockAnnotations = annotMap
	}
}

func (schemaopts) WithEnumFieldAnnotations(annots ...AnnotSpec) SchemaOption {
	return func(s *SchemaSpec) {
		annotMap := make(map[string]AnnotSpec, len(annots))
		for _, a := range annots {
			annotMap[a.Name] = a
		}
		s.EnumFieldAnnotations = annotMap
	}
}

type annotation struct{}

func (annotation) Single() AnnotOption {
	return func(a *AnnotSpec) {
		a.Singular = true
	}
}

func (annotation) Required() AnnotOption {
	return func(a *AnnotSpec) {
		a.Required = true
	}
}

func (annotation) Args(args ...ArgSpec) AnnotOption {
	return func(f *AnnotSpec) {
		f.Arguments = map[string]ArgSpec{}
		for _, a := range args {
			f.Arguments[a.Name] = a
		}
	}
}

type function struct{}

func (function) Args(args ...ArgSpec) FuncOption {
	return func(f *FuncSpec) {
		f.Arguments = make(map[string]ArgSpec)
		for _, a := range args {
			f.Arguments[a.Name] = a
		}
	}
}

func (function) Named(n string) FuncOption {
	return func(f *FuncSpec) {
		f.Name = n
	}
}

type argopts struct{}

func (argopts) Default() ArgOption {
	return func(a *ArgSpec) {
		a.Default = true
	}
}

func (argopts) Required() ArgOption {
	return func(a *ArgSpec) {
		a.Required = true
	}
}

type propopts struct{}

func (propopts) Required() PropOption {
	return func(a *PropertySpec) {
		a.Required = true
	}
}

type objopts struct{}

func (propopts) AllowExtraProperties() ObjectOption {
	return func(a *ObjectSpec) {
		a.AllowExtraProps = true
	}
}

func (propopts) Props(props ...PropertySpec) ObjectOption {
	return func(a *ObjectSpec) {
		a.Properties = map[string]PropertySpec{}
		for _, p := range props {
			a.Properties[p.Name] = p
		}
	}
}

type blockopts struct{}

func (blockopts) Nameless() BlockOption {
	return func(b *BlockSpec) {
		b.NamePresence = NoPresence
	}
}

func (blockopts) Props(props ...PropertySpec) BlockOption {
	return func(b *BlockSpec) {
		b.Body.Properties = map[string]PropertySpec{}
		for _, p := range props {
			b.Body.Properties[p.Name] = p
		}
	}
}

func (blockopts) AllowExtraProperties() BlockOption {
	return func(b *BlockSpec) {
		b.Body.AllowExtraProps = true
	}
}
