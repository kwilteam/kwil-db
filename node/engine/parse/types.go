package parse

import (
	antlr "github.com/antlr4-go/antlr/v4"
)

// Position is a Position in the parse tree. It represents a range of line and column
// values in Kuneiform source code.
type Position struct {
	// Set is true if the position of the Position has been set.
	// This is useful for testing parsers.
	isSet     bool
	StartLine *int `json:"start_line,omitempty"`
	StartCol  *int `json:"start_col,omitempty"`
	EndLine   *int `json:"end_line,omitempty"`
	EndCol    *int `json:"end_col,omitempty"`
}

func (p *Position) nilStart() bool {
	return p.StartLine == nil || p.StartCol == nil
}

func (p *Position) nilEnd() bool {
	return p.EndLine == nil || p.EndCol == nil
}

// Set sets the position of the Position based on the given parser rule context.
func (n *Position) Set(r antlr.ParserRuleContext) {
	n.isSet = true
	n.StartLine = intPtr(r.GetStart().GetLine())
	n.StartCol = intPtr(r.GetStart().GetColumn())
	n.EndLine = intPtr(r.GetStop().GetLine())
	n.EndCol = intPtr(r.GetStop().GetColumn())
}

func intPtr(i int) *int {
	return &i
}

// SetToken sets the position of the Position based on the given token.
func (n *Position) SetToken(t antlr.Token) {
	n.isSet = true
	n.StartLine = intPtr(t.GetLine())
	n.StartCol = intPtr(t.GetColumn())
	n.EndLine = intPtr(t.GetLine())
	n.EndCol = intPtr(t.GetColumn())
}

// GetPosition returns the Position.
// It is useful if the Position is embedded in another struct.
func (n *Position) GetPosition() *Position {
	return n
}

// Clear clears the position of the Position.
func (n *Position) Clear() {
	n.isSet = false
	n.StartLine = nil
	n.StartCol = nil
	n.EndLine = nil
	n.EndCol = nil
}

// unaryNode creates a Position with the same start and end position.
func unaryNode(start, end int) *Position {
	return &Position{
		StartLine: intPtr(start),
		StartCol:  intPtr(end),
		// EndLine:   intPtr(start),
		// EndCol:    intPtr(end),
	}
}
