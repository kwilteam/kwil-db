package dto

import "kwil/x/execution"

type Input interface {
	GetName() string
	GetColumn() string
	GetStatic() bool
	GetValue() any
	GetModifier() execution.ModifierType
}

type Parameter struct {
	Name     string                 `json:"name"`
	Column   string                 `json:"column"`
	Static   bool                   `json:"static"`
	Value    any                    `json:"value,omitempty"`
	Modifier execution.ModifierType `json:"modifier,omitempty"`
}

func (p *Parameter) GetName() string {
	return p.Name
}

func (p *Parameter) GetColumn() string {
	return p.Column
}

func (p *Parameter) GetStatic() bool {
	return p.Static
}

func (p *Parameter) GetValue() any {
	return p.Value
}

func (p *Parameter) GetModifier() execution.ModifierType {
	return p.Modifier
}

type WhereClause struct {
	Name     string                           `json:"name"`
	Column   string                           `json:"column"`
	Static   bool                             `json:"static"`
	Operator execution.ComparisonOperatorType `json:"operator,omitempty"`
	Value    any                              `json:"value,omitempty"`
	Modifier execution.ModifierType           `json:"modifier,omitempty"`
}

func (w *WhereClause) GetName() string {
	return w.Name
}

func (w *WhereClause) GetColumn() string {
	return w.Column
}

func (w *WhereClause) GetStatic() bool {
	return w.Static
}

func (w *WhereClause) GetValue() any {
	return w.Value
}

func (w *WhereClause) GetModifier() execution.ModifierType {
	return w.Modifier
}
