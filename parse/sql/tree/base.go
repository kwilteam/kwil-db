package tree

type Position struct {
	StartLine   int
	EndLine     int
	StartColumn int
	EndColumn   int
}

// node is the base struct implementing the Node interface.
// Node implementations should have this embedded.
type node struct {
	text string
	pos  *Position
}

func (n *node) SetText(text string) {
	n.text = text
}

func (n *node) Text() string {
	return n.text
}

func (n *node) Position() *Position {
	return n.pos
}

func (n *node) SetPosition(pos *Position) {
	n.pos = pos
}
