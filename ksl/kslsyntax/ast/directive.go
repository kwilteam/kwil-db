package ast

import (
	"ksl"
	"ksl/kslsyntax/lex"
)

var _ Node = (*Directive)(nil)

type Directive struct {
	At    lex.Token
	Name  *Str
	Key   *Str
	Value Expr

	SrcRange ksl.Range
}

func (d *Directive) GetName() string {
	if d == nil {
		return ""
	}
	return d.Name.GetString()
}

func (d *Directive) GetKey() string {
	if d == nil {
		return ""
	}
	return d.Key.GetString()
}

func (d *Directive) HasValue() bool {
	if d == nil {
		return false
	}
	return d.Value != nil
}

type Directives []*Directive

func (els Directives) ByType() map[string]Directives {
	ret := make(map[string]Directives)
	for _, el := range els {
		ty := el.Name.Value
		if ret[ty] == nil {
			ret[ty] = make(Directives, 0, 1)
		}
		ret[ty] = append(ret[ty], el)
	}
	return ret
}
