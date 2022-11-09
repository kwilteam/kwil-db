package ast

import (
	"ksl"
	"ksl/kslsyntax/lex"
)

var _ Node = (*Attribute)(nil)

type Attribute struct {
	Name  *Str
	Eq    lex.Token
	Value Expr

	SrcRange ksl.Range
}

func (kv *Attribute) GetName() string {
	if kv == nil {
		return ""
	}
	return kv.Name.GetString()
}

func (kv *Attribute) HasValue() bool {
	if kv == nil {
		return false
	}
	return kv.Value != nil
}

func (kv *Attribute) GetValue() Expr {
	if kv == nil {
		return nil
	}
	return kv.Value
}

type Attributes []*Attribute

func (a Attributes) ByName() map[string]*Attribute {
	ret := make(map[string]*Attribute)
	for _, kv := range a {
		ret[kv.Name.Value] = kv
	}
	return ret
}
