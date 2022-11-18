package ast

import (
	"ksl"
	"strings"
)

type Expression interface {
	Node
	expr()
}

type Number struct {
	Value string
	Span  ksl.Range
}

type String struct {
	Value string
	Span  ksl.Range
}

type Heredoc struct {
	Marker      string
	Values      []string
	StripIndent bool
	Span        ksl.Range
}

func (n *Heredoc) String() string {
	if n == nil {
		return ""
	}

	lines := make([]string, 0, len(n.Values))
	for _, line := range n.Values {
		if n.StripIndent {
			lines = append(lines, strings.TrimLeft(line, " \t"))
		} else {
			lines = append(lines, line)
		}
	}
	return strings.Join(lines, "")
}

type Literal struct {
	Value string
	Span  ksl.Range
}

type Object struct {
	Properties Properties
	Span       ksl.Range
}

type Function struct {
	Name      *Name
	Arguments *ArgumentList
	Span      ksl.Range
}

func (f *Function) GetName() string    { return f.Name.String() }
func (a *Function) GetNameNode() *Name { return a.Name }
func (f *Function) GetArgs() Arguments {
	if f == nil || f.Arguments == nil {
		return nil
	}
	return f.Arguments.Args()
}

type List struct {
	Elements []Expression
	Span     ksl.Range
}

type Variable struct {
	Name *Name
	Span ksl.Range
}

func (v *Variable) GetName() string { return v.Name.String() }

func (Number) expr()   {}
func (String) expr()   {}
func (Literal) expr()  {}
func (Function) expr() {}
func (Heredoc) expr()  {}
func (List) expr()     {}
func (Object) expr()   {}
func (Variable) expr() {}

func (v Variable) Range() ksl.Range { return v.Span }
func (o Object) Range() ksl.Range   { return o.Span }
func (f Heredoc) Range() ksl.Range  { return f.Span }
func (f Function) Range() ksl.Range { return f.Span }
func (c Literal) Range() ksl.Range  { return c.Span }
func (l List) Range() ksl.Range     { return l.Span }
func (s String) Range() ksl.Range   { return s.Span }
func (n Number) Range() ksl.Range   { return n.Span }
