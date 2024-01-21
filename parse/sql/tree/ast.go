package tree

import (
	"fmt"
)

// Node defines interface for all tree nodes in the AST.
type Node interface {
	fmt.Stringer
	//GetChildren() []Node
	//SetParent(Node)
	//GetParent() Node
}

type Position struct {
	StartLine   int
	EndLine     int
	StartColumn int
	EndColumn   int
}

// ParseNode defines interface for a parsed node.
type ParseNode interface {
	Node

	// for validation logic
	SetPosition(*Position)
	GetPosition() *Position
	//AddChild(ParseNode)
}

// AstNode defines interface for all AST nodes.
type AstNode interface {
	ParseNode

	Accept(AstVisitor) any
	Walker
	ToSQL() string
}

type BaseAstNode struct {
	//children []Node
	//parent Node
	pos *Position
}

//func NewBaseAstNode(pos *Position) *BaseAstNode {
//	return &BaseAstNode{
//		pos: pos,
//	}
//}

//func (b *BaseAstNode) GetChildren() []Node {
//	return b.children
//}

//func (b *BaseAstNode) SetParent(parent Node) {
//	b.parent = parent
//}
//
//func (b *BaseAstNode) GetParent() Node {
//	return b.parent
//}

func (b *BaseAstNode) SetPosition(pos *Position) {
	b.pos = pos
}

func (b *BaseAstNode) GetPosition() *Position {
	if b.pos == nil {
		return &Position{}
	}
	return b.pos
}

//func (b *BaseAstNode) AddChild(node ParseNode) {
//	if b.children == nil {
//		b.children = make([]Node, 0)
//	}
//	b.children = append(b.children, node)
//}

func (b *BaseAstNode) Walk(w AstWalker) error {
	return nil
}

func (b *BaseAstNode) Accept(v AstVisitor) any {
	return v.Visit(b)
}

func (b *BaseAstNode) ToSQL() string {
	return ""
}

func (b *BaseAstNode) String() string {
	return ""
}
