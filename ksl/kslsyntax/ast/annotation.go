package ast

import (
	"ksl"
	"ksl/kslsyntax/lex"
)

var _ Node = (*Annotation)(nil)
var _ Node = (*FunctionCall)(nil)
var _ Node = (*ArgList)(nil)

type Annotation struct {
	Marker  *Str
	Name    *Str
	ArgList *ArgList

	SrcRange ksl.Range
}

func (f *Annotation) Arg(i int) (Expr, bool) {
	if f == nil {
		return nil, false
	}
	return f.ArgList.Arg(i)
}

func (f *Annotation) MustArg(i int) Expr {
	a, _ := f.Arg(i)
	return a
}

func (f *Annotation) Kwarg(name string) (Expr, bool) {
	return f.ArgList.Kwarg(name)
}

func (f *Annotation) MustKwarg(name string) Expr {
	k, _ := f.Kwarg(name)
	return k
}

func (a *Annotation) GetName() string {
	if a == nil {
		return ""
	}
	return a.Name.GetString()
}

func (a *Annotation) GetArgs() []Expr {
	if a == nil {
		return nil
	}
	return a.ArgList.GetArgs()
}

func (a *Annotation) GetKwargs() map[string]*Attribute {
	if a == nil {
		return nil
	}
	return a.ArgList.GetKwargs()
}

type FunctionCall struct {
	Name    *Str
	ArgList *ArgList

	SrcRange ksl.Range
}

func (f *FunctionCall) GetName() string {
	if f == nil {
		return ""
	}
	return f.Name.GetString()
}

func (f *FunctionCall) Arg(i int) (Expr, bool) {
	if f == nil {
		return nil, false
	}
	return f.ArgList.Arg(i)
}

func (f *FunctionCall) MustArg(i int) Expr {
	a, _ := f.Arg(i)
	return a
}

func (f *FunctionCall) Kwarg(name string) (Expr, bool) {
	if f == nil {
		return nil, false
	}
	return f.ArgList.Kwarg(name)
}

func (f *FunctionCall) MustKwarg(name string) Expr {
	k, _ := f.Kwarg(name)
	return k
}

func (a *FunctionCall) GetArgs() []Expr {
	if a == nil {
		return nil
	}
	return a.ArgList.GetArgs()
}

func (a *FunctionCall) GetKwargs() map[string]*Attribute {
	if a == nil {
		return nil
	}
	return a.ArgList.GetKwargs()
}

type ArgList struct {
	LParen lex.Token
	Args   []Expr
	Kwargs []*Attribute
	RParen lex.Token

	SrcRange ksl.Range
}

func (f *ArgList) MustArg(i int) Expr {
	a, _ := f.Arg(i)
	return a
}

func (f *ArgList) Arg(i int) (Expr, bool) {
	if f == nil {
		return nil, false
	}
	args := f.GetArgs()
	if i >= len(args) {
		return nil, false
	}
	return args[i], true
}

func (f *ArgList) MustKwarg(name string) Expr {
	a, _ := f.Kwarg(name)
	return a
}

func (f *ArgList) Kwarg(name string) (Expr, bool) {
	if f == nil {
		return nil, false
	}

	for _, kv := range f.Kwargs {
		if kv.GetName() == name {
			return kv.GetValue(), true
		}
	}
	return nil, false
}

func (a *ArgList) GetArgs() []Expr {
	if a == nil {
		return nil
	}
	return a.Args
}

func (a *ArgList) GetKwargs() map[string]*Attribute {
	if a == nil {
		return nil
	}
	ret := make(map[string]*Attribute, len(a.Kwargs))
	for _, kv := range a.Kwargs {
		ret[kv.GetName()] = kv
	}
	return ret
}

type Annotations []*Annotation

func (a Annotations) Must(name string) *Annotation {
	annot, _ := a.Get(name)
	return annot
}

func (a Annotations) Get(name string) (*Annotation, bool) {
	for _, ann := range a {
		if ann.Name.GetString() == name {
			return ann, true
		}
	}
	return nil, false
}
