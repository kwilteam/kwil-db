package ast

import "ksl"

type Fields []*Field

func (a Fields) Get(name string) (*Field, bool) {
	if a == nil {
		return nil, false
	}

	for _, fld := range a {
		if fld.GetName() == name {
			return fld, true
		}
	}
	return nil, false
}

type FieldArity int

const (
	Required FieldArity = iota
	Optional
	Repeated
)

func (f FieldArity) IsRequired() bool { return f == Required }
func (f FieldArity) IsOptional() bool { return f == Optional }
func (f FieldArity) IsRepeated() bool { return f == Repeated }
func (f FieldArity) IsAny(o ...FieldArity) bool {
	for _, a := range o {
		if f == a {
			return true
		}
	}
	return false
}

type Field struct {
	Name        *Name
	Type        *FieldType
	Annotations Annotations
	Comment     *CommentGroup
	Span        ksl.Range
}

func (f Field) Range() ksl.Range               { return f.Span }
func (e *Field) GetName() string               { return e.Name.String() }
func (d *Field) GetNameNode() *Name            { return d.Name }
func (e *Field) Documentation() string         { return e.Comment.String() }
func (m *Field) GetAnnotations() []*Annotation { return m.Annotations }

func (f *Field) IsRequired() bool { return f.Type.Arity.IsRequired() }
func (f *Field) IsOptional() bool { return f.Type.Arity.IsOptional() }
func (f *Field) IsRepeated() bool { return f.Type.Arity.IsRepeated() }

type FieldType struct {
	Name  *Name
	Arity FieldArity
	Span  ksl.Range
}

func (f FieldType) Range() ksl.Range    { return f.Span }
func (f *FieldType) GetName() string    { return f.Name.String() }
func (d *FieldType) GetNameNode() *Name { return d.Name }
