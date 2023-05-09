package databases

import "github.com/kwilteam/kwil-db/pkg/databases/spec"

// Input is a schema input; either a parameter or a where clause
type Input[T spec.AnyValue] interface {
	GetName() string
	GetColumn() string
	GetStatic() bool
	GetValue() T
	GetModifier() spec.ModifierType
}

type Parameter[T spec.AnyValue] struct {
	Name     string            `json:"name" clean:"lower"`
	Column   string            `json:"column" clean:"lower"`
	Static   bool              `json:"static"`
	Value    T                 `json:"value,omitempty" traverse:"shallow"`
	Modifier spec.ModifierType `json:"modifier,omitempty" clean:"is_enum,modifier_type"`
}

func (p *Parameter[T]) GetName() string {
	return p.Name
}

func (p *Parameter[T]) GetColumn() string {
	return p.Column
}

func (p *Parameter[T]) GetStatic() bool {
	return p.Static
}

func (p *Parameter[T]) GetValue() T {
	return p.Value
}

func (p *Parameter[T]) GetModifier() spec.ModifierType {
	return p.Modifier
}

type WhereClause[T spec.AnyValue] struct {
	Name     string                      `json:"name" clean:"lower"`
	Column   string                      `json:"column" clean:"lower"`
	Static   bool                        `json:"static"`
	Operator spec.ComparisonOperatorType `json:"operator,omitempty" clean:"is_enum,comparison_operator_type"`
	Value    T                           `json:"value,omitempty" traverse:"shallow"`
	Modifier spec.ModifierType           `json:"modifier,omitempty" clean:"is_enum,modifier_type"`
}

func (w *WhereClause[T]) GetName() string {
	return w.Name
}

func (w *WhereClause[T]) GetColumn() string {
	return w.Column
}

func (w *WhereClause[T]) GetStatic() bool {
	return w.Static
}

func (w *WhereClause[T]) GetValue() T {
	return w.Value
}

func (w *WhereClause[T]) GetModifier() spec.ModifierType {
	return w.Modifier
}
