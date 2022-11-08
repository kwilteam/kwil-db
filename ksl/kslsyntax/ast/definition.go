package ast

import "ksl"

var _ Node = (*Definition)(nil)

type Definition struct {
	Name        *Str
	Type        *Type
	Annotations Annotations

	SrcRange ksl.Range
}

func (d *Definition) GetName() string {
	if d == nil {
		return ""
	}
	return d.Name.GetString()
}

func (d *Definition) IsNullable() bool {
	if d == nil || d.Type == nil {
		return false
	}
	return d.Type.Nullable
}

func (d *Definition) IsArray() bool {
	if d == nil || d.Type == nil {
		return false
	}
	return d.Type.IsArray
}

func (d *Definition) GetTypeName() string {
	if d == nil || d.Type == nil {
		return ""
	}
	return d.Type.Name.GetString()
}

func (d *Definition) GetAnnotations() Annotations {
	if d == nil {
		return nil
	}
	return d.Annotations
}

func (d *Definition) Annotation(key string) (*Annotation, bool) { return d.Annotations.Get(key) }

type Definitions []*Definition

func (d Definitions) ByName() map[string]*Definition {
	ret := make(map[string]*Definition)
	for _, def := range d {
		ret[def.GetName()] = def
	}
	return ret
}
