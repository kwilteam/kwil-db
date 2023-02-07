package databases

import (
	execution2 "kwil/pkg/execution"
	"kwil/pkg/types/data_types/any_type"
)

type Input[T anytype.AnyValue] interface {
	GetName() string
	GetColumn() string
	GetStatic() bool
	GetValue() *T
	GetModifier() execution2.ModifierType
}

type Parameter[T anytype.AnyValue] struct {
	Name     string                  `json:"name" clean:"lower"`
	Column   string                  `json:"column" clean:"lower"`
	Static   bool                    `json:"static"`
	Value    T                       `json:"value,omitempty" traverse:"shallow"`
	Modifier execution2.ModifierType `json:"modifier,omitempty" clean:"is_enum,modifier_type"`
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

func (p *Parameter[T]) GetValue() *T {
	return &p.Value
}

func (p *Parameter[T]) GetModifier() execution2.ModifierType {
	return p.Modifier
}

type WhereClause[T anytype.AnyValue] struct {
	Name     string                            `json:"name" clean:"lower"`
	Column   string                            `json:"column" clean:"lower"`
	Static   bool                              `json:"static"`
	Operator execution2.ComparisonOperatorType `json:"operator,omitempty" clean:"is_enum,comparison_operator_type"`
	Value    T                                 `json:"value,omitempty" traverse:"shallow"`
	Modifier execution2.ModifierType           `json:"modifier,omitempty" clean:"is_enum,modifier_type"`
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

func (w *WhereClause[T]) GetValue() *T {
	return &w.Value
}

func (w *WhereClause[T]) GetModifier() execution2.ModifierType {
	return w.Modifier
}
