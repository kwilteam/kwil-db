package ast

import "ksl"

type Model struct {
	Name        *Name
	Fields      Fields
	Annotations Annotations
	Comment     *CommentGroup
	Span        ksl.Range
}

func (m Model) Range() ksl.Range               { return m.Span }
func (m *Model) GetName() string               { return m.Name.String() }
func (d *Model) GetNameNode() *Name            { return d.Name }
func (m *Model) Documentation() string         { return m.Comment.String() }
func (m *Model) GetAnnotations() []*Annotation { return m.Annotations }
