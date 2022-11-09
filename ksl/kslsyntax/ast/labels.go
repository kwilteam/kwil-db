package ast

import (
	"ksl"
	"ksl/kslsyntax/lex"
)

var _ Node = (*BlockLabels)(nil)

type BlockLabels struct {
	LBrack lex.Token
	Values []*Attribute
	RBrack lex.Token

	SrcRange ksl.Range
}

func (l *BlockLabels) Label(key string) (*Attribute, bool) {
	for _, kv := range l.GetValues() {
		if kv.Name.GetString() == key {
			return kv, true
		}
	}
	return nil, false
}

func (l *BlockLabels) MustLabel(key string) *Attribute {
	r, _ := l.Label(key)
	return r
}

func (l *BlockLabels) GetValues() []*Attribute {
	if l == nil {
		return nil
	}
	return l.Values
}
