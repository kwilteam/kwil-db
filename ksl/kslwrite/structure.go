package kslwrite

import (
	"io"
	"strings"
)

var bld = Builder{}

type Node interface {
	Format(io.Writer) error
}

type Attributes []*Attribute
type Kwargs []*Kwarg
type Annotations []*Annotation
type Definitions []*Definition
type Blocks []*Block
type Directives []*Directive

type File struct {
	Directives Directives
	Blocks     Blocks
}

func (f *File) AddDirectives(d ...*Directive) *File {
	f.Directives = append(f.Directives, d...)
	return f
}

func (f *File) NewDirective(key string) *Directive {
	d := bld.Directive(key)
	f.AddDirectives(d)
	return d
}

func (f *File) AddBlocks(b ...*Block) *File {
	f.Blocks = append(f.Blocks, b...)
	return f
}

func (f *File) NewBlock(typ, name string) *Block {
	blk := bld.Block(typ, name)
	f.AddBlocks(blk)
	return blk
}

func (f *File) GetBlock(typ, name string) (*Block, bool) {
	for _, b := range f.Blocks {
		if b.Type == typ && b.Name == name {
			return b, true
		}
	}
	return nil, false
}

func (f *File) GetOrCreateBlock(typ, name string) *Block {
	if blk, ok := f.GetBlock(typ, name); ok {
		return blk
	}
	blk := bld.Block(typ, name)
	f.AddBlocks(blk)
	return blk
}

type KeyValue struct {
	Key   string
	Value string
}

type Attribute KeyValue
type Kwarg Attribute

func (a *Attribute) SetValue(value string) *Attribute {
	a.Value = value
	return a
}

func (a *Kwarg) SetValue(value string) *Kwarg {
	a.Value = value
	return a
}

type Directive struct {
	Name  string
	Key   string
	Value string
}

func (d *Directive) SetKeyValue(key string, value string) {
	d.Key = key
	d.Value = value
}

func (d *Directive) SetValue(value string) *Directive {
	d.Value = value
	return d
}

type Block struct {
	Type     string
	Name     string
	Modifier string
	Target   string
	Labels   Kwargs
	Body     *BlockBody
}

func (b *Block) Mods(mod, target string) *Block {
	b.Modifier = mod
	b.Target = target
	return b
}

func (b *Block) AddLabels(labels ...*Kwarg) *Block {
	b.Labels = labels
	return b
}

func (b *Block) AddAnnotations(anns ...*Annotation) *Block {
	if b.Body == nil {
		b.Body = new(BlockBody)
	}
	b.Body.Annotations = append(b.Body.Annotations, anns...)
	return b
}

func (b *Block) AddAttributes(attrs ...*Attribute) *Block {
	if b.Body == nil {
		b.Body = new(BlockBody)
	}
	b.Body.Attributes = append(b.Body.Attributes, attrs...)
	return b
}

func (b *Block) AddBlocks(blocks ...*Block) *Block {
	if b.Body == nil {
		b.Body = new(BlockBody)
	}
	b.Body.Blocks = append(b.Body.Blocks, blocks...)
	return b
}

func (b *Block) AddDefinitions(defs ...*Definition) *Block {
	if b.Body == nil {
		b.Body = new(BlockBody)
	}
	b.Body.Definitions = append(b.Body.Definitions, defs...)
	return b
}

func (b *Block) AddEnumValues(values ...string) *Block {
	if b.Body == nil {
		b.Body = new(BlockBody)
	}
	b.Body.EnumValues = append(b.Body.EnumValues, values...)
	return b
}

func (b *Block) GetDefinitions() Definitions {
	if b.Body == nil {
		b.Body = &BlockBody{}
	}
	return b.Body.Definitions
}
func (b *Block) GetAnnotations() Annotations {
	if b.Body == nil {
		b.Body = &BlockBody{}
	}
	return b.Body.Annotations
}
func (b *Block) GetAttributes() Attributes {
	if b.Body == nil {
		b.Body = &BlockBody{}
	}
	return b.Body.Attributes
}
func (b *Block) GetBlocks() Blocks {
	if b.Body == nil {
		b.Body = &BlockBody{}
	}
	return b.Body.Blocks
}
func (b *Block) GetEnumValues() []string {
	if b.Body == nil {
		b.Body = &BlockBody{}
	}
	return b.Body.EnumValues
}

func (b *Block) GetOrCreateDefinition(name string) *Definition {
	for _, d := range b.GetDefinitions() {
		if d.Name == name {
			return d
		}
	}
	d := bld.Def(name)
	b.AddDefinitions(d)
	return d
}

func (b *Block) GetOrCreateAnnotation(name string) *Annotation {
	for _, a := range b.GetAnnotations() {
		if a.Name == name {
			return a
		}
	}
	a := bld.Annot(name)
	b.AddAnnotations(a)
	return a
}

type BlockBody struct {
	Attributes  Attributes
	Annotations Annotations
	Definitions Definitions
	Blocks      Blocks
	EnumValues  []string
}

type Annotation struct {
	Name string
	Args *ArgList
}

func (a *Annotation) AddArgs(args ...string) *Annotation {
	if a.Args == nil {
		a.Args = new(ArgList)
	}
	a.Args.Args = append(a.Args.Args, args...)
	return a
}

func (a *Annotation) AddKwarg(key string, value string) *Annotation {
	return a.AddKwargs(&Kwarg{Key: key, Value: value})
}

func (a *Annotation) AddKwargs(kwargs ...*Kwarg) *Annotation {
	if a.Args == nil {
		a.Args = new(ArgList)
	}
	a.Args.Kwargs = append(a.Args.Kwargs, kwargs...)
	return a
}

type ArgList struct {
	Args   []string
	Kwargs []*Kwarg
}

type Definition struct {
	Name       string
	Type       string
	IsOptional bool
	IsArray    bool

	Annotations []*Annotation
}

func (d *Definition) SetType(typ string) *Definition {
	d.Type = typ
	return d
}

func (d *Definition) AddAnnotations(anns ...*Annotation) *Definition {
	d.Annotations = append(d.Annotations, anns...)
	return d
}

func (d *Definition) GetOrCreateAnnotation(name string) *Annotation {
	for _, a := range d.Annotations {
		if a.Name == name {
			return a
		}
	}
	a := bld.Annot(name)
	d.AddAnnotations(a)
	return a
}

func (d *Definition) SetOptional(flag bool) *Definition {
	d.IsOptional = flag
	return d
}

func (d *Definition) SetArray(flag bool) *Definition {
	d.IsArray = flag
	return d
}

type Builder struct{}

func (Builder) File() *File                              { return &File{} }
func (Builder) Directive(name string) *Directive         { return &Directive{Name: name} }
func (Builder) Block(typ, name string) *Block            { return &Block{Type: typ, Name: name} }
func (Builder) Annot(name string) *Annotation            { return &Annotation{Name: name} }
func (Builder) Def(name string) *Definition              { return &Definition{Name: name} }
func (Builder) Attr(key string, value string) *Attribute { return &Attribute{Key: key, Value: value} }

func Quoted(v string) string {
	return `"` + strings.ReplaceAll(v, `"`, `\"`) + `"`
}

func List(values ...string) string {
	return "[" + strings.Join(values, ", ") + "]"
}
