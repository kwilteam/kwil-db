package ast

import "ksl"

type Block struct {
	Type       *TypeName
	Name       *Name
	Properties Properties
	Comment    *CommentGroup
	Span       ksl.Range
}

func (b Block) Range() ksl.Range       { return b.Span }
func (b *Block) Documentation() string { return b.Comment.String() }
func (a *Block) GetNameNode() *Name    { return a.Name }
func (b *Block) GetName() string       { return b.Name.String() }
func (b *Block) GetType() string       { return b.Type.String() }

func (b *Block) Property(name string) (*Property, bool) {
	if b == nil {
		return nil, false
	}
	return b.Properties.Get(name)
}

type Properties []*Property

func (p Properties) Get(name string) (*Property, bool) {
	if p == nil {
		return nil, false
	}

	for _, prop := range p {
		if prop.GetName() == name {
			return prop, true
		}
	}
	return nil, false
}

type Property struct {
	Name  *Name
	Value Expression
	Span  ksl.Range
}

func (p Property) Range() ksl.Range    { return p.Span }
func (p *Property) GetName() string    { return p.Name.String() }
func (a *Property) GetNameNode() *Name { return a.Name }
func (a *Property) Identifier() *Name  { return a.Name }
