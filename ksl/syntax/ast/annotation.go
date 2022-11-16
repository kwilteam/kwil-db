package ast

import "ksl"

type Annotations []*Annotation

func (a Annotations) Get(name string) (*Annotation, bool) {
	if a == nil {
		return nil, false
	}

	for _, annot := range a {
		if annot.GetName() == name {
			return annot, true
		}
	}
	return nil, false
}

type Annotation struct {
	Name *Name
	Args *ArgumentList
	Span ksl.Range
}

func (a *Annotation) GetName() string    { return a.Name.String() }
func (a *Annotation) GetNameNode() *Name { return a.Name }
func (a Annotation) Range() ksl.Range    { return a.Span }

func (f *Annotation) GetArgs() Arguments {
	if f == nil || f.Args == nil {
		return nil
	}
	return f.Args.Args()
}

func (a *Annotation) HasArg(name string) bool {
	if a == nil {
		return false
	}
	_, ok := a.Arg(name)
	return ok
}

func (a *Annotation) Arg(name string) (*Argument, bool) {
	if a == nil {
		return nil, false
	}
	return a.Args.Arg(name)
}

func (a *Annotation) DefaultArg(name string) (*Argument, bool) {
	if a == nil {
		return nil, false
	}
	if arg, ok := a.Args.Arg(name); ok {
		return arg, true
	}
	for _, arg := range a.GetArgs() {
		if arg.GetName() == "" {
			return arg, true
		}
	}
	return nil, false
}
