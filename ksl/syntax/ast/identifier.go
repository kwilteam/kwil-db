package ast

import "ksl"

type Name struct {
	Name string
	Span ksl.Range
}

func (i Name) Range() ksl.Range { return i.Span }
func (i *Name) String() string {
	if i == nil {
		return ""
	}
	return i.Name
}

type TypeName struct {
	Name string
	Span ksl.Range
}

func (i TypeName) Range() ksl.Range { return i.Span }
func (i *TypeName) String() string {
	if i == nil {
		return ""
	}
	return i.Name
}
