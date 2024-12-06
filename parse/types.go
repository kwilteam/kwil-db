package parse

import (
	"github.com/antlr4-go/antlr/v4"
)

// Position is a Position in the parse tree. It represents a range of line and column
// values in Kuneiform source code.
type Position struct {
	// Set is true if the position of the Position has been set.
	// This is useful for testing parsers.
	IsSet     bool `json:"-"`
	StartLine int  `json:"start_line"`
	StartCol  int  `json:"start_col"`
	EndLine   int  `json:"end_line"`
	EndCol    int  `json:"end_col"`
}

// Set sets the position of the Position based on the given parser rule context.
func (n *Position) Set(r antlr.ParserRuleContext) {
	n.IsSet = true
	n.StartLine = r.GetStart().GetLine()
	n.StartCol = r.GetStart().GetColumn()
	n.EndLine = r.GetStop().GetLine()
	n.EndCol = r.GetStop().GetColumn()
}

// SetToken sets the position of the Position based on the given token.
func (n *Position) SetToken(t antlr.Token) {
	n.IsSet = true
	n.StartLine = t.GetLine()
	n.StartCol = t.GetColumn()
	n.EndLine = t.GetLine()
	n.EndCol = t.GetColumn()
}

// GetPosition returns the Position.
// It is useful if the Position is embedded in another struct.
func (n *Position) GetPosition() *Position {
	return n
}

// Clear clears the position of the Position.
func (n *Position) Clear() {
	n.IsSet = false
	n.StartLine = 0
	n.StartCol = 0
	n.EndLine = 0
	n.EndCol = 0
}

// unaryNode creates a Position with the same start and end position.
func unaryNode(start, end int) *Position {
	return &Position{
		StartLine: start,
		StartCol:  end,
		EndLine:   start,
		EndCol:    end,
	}
}
