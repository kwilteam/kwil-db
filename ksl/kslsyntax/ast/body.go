package ast

import (
	"ksl"
	"ksl/kslsyntax/lex"
)

var _ Node = (*Body)(nil)

type Body struct {
	LBrace      lex.Token
	Annotations Annotations
	Attributes  Attributes
	EnumValues  []*Str
	Definitions Definitions
	Blocks      Blocks
	RBrace      lex.Token

	SrcRange ksl.Range
}

func (b *Body) GetAnnotations() Annotations {
	if b == nil {
		return nil
	}
	return b.Annotations
}

func (b *Body) GetAttributes() Attributes {
	if b == nil {
		return nil
	}
	return b.Attributes
}

func (b *Body) GetEnumValues() []string {
	if b == nil {
		return nil
	}
	ret := make([]string, len(b.EnumValues))
	for i := range ret {
		ret[i] = b.EnumValues[i].GetString()
	}
	return ret
}

func (b *Body) GetDefinitions() Definitions {
	if b == nil {
		return nil
	}
	return b.Definitions
}

func (b *Body) GetBlocks() Blocks {
	if b == nil {
		return nil
	}
	return b.Blocks
}
