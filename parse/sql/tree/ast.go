package tree

// Node defines interface for all nodes.
type Node interface {
	// Text returns the original text of the node.
	Text() string
	// SetText sets original text to the node.
	SetText(string)
	// Position returns the position of the node.
	Position() *Position
	// SetPosition sets position to the node.
	SetPosition(*Position)
}

// AstNode represents an AST node.
type AstNode interface {
	Node
	Walker

	// ToSQL converts the node to a SQL string.
	ToSQL() string
	// Accept accepts an AstVisitor to visit itself.
	Accept(AstVisitor) any
}
