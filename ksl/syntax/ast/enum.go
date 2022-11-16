package ast

import "ksl"

type Enum struct {
	Name        *Name
	Values      EnumValues
	Annotations Annotations
	Comment     *CommentGroup
	Span        ksl.Range
}

func (e Enum) Range() ksl.Range               { return e.Span }
func (e *Enum) GetName() string               { return e.Name.String() }
func (d *Enum) GetNameNode() *Name            { return d.Name }
func (e *Enum) GetAnnotations() []*Annotation { return e.Annotations }
func (e *Enum) Documentation() string         { return e.Comment.String() }
func (e *Enum) GetValues() []string {
	values := make([]string, len(e.Values))
	for i, v := range e.Values {
		values[i] = v.Name.String()
	}
	return values
}

type EnumValues []*EnumValue

func (e EnumValues) Get(name string) (*EnumValue, bool) {
	if e == nil {
		return nil, false
	}

	for _, ev := range e {
		if ev.GetName() == name {
			return ev, true
		}
	}
	return nil, false
}

type EnumValue struct {
	Name        *Name
	Annotations Annotations
	Comment     *CommentGroup
	Span        ksl.Range
}

func (e EnumValue) Range() ksl.Range               { return e.Span }
func (e *EnumValue) GetName() string               { return e.Name.String() }
func (d *EnumValue) GetNameNode() *Name            { return d.Name }
func (e *EnumValue) GetAnnotations() []*Annotation { return e.Annotations }
func (e *EnumValue) Documentation() string         { return e.Comment.String() }
