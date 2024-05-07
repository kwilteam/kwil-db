package types

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/antlr4-go/antlr/v4"
)

// ParseErrors is a collection of parse errors.
type ParseErrors []*ParseError

var _ interface{ Unwrap() []error } = (*ParseErrors)(nil)

// Unwrap allows errors.Is and error.As to identify wrapped errors.
func (p ParseErrors) Unwrap() []error {
	errs := make([]error, len(p))
	for i := range p {
		errs[i] = p[i]
	}
	return errs
}

// Err returns all the errors as a single error.
func (p ParseErrors) Err() error {
	if len(p) == 0 {
		return nil
	}
	return &p
}

// The zero value of a ParseErrors instance intentionally does not implement the
// error interface. The Err method will return nil if the length is zero.
var _ error = (*ParseErrors)(nil)

// Error implements the error interface.
func (p *ParseErrors) Error() string {
	errs := *p
	switch len(errs) {
	case 0: // use Err and this won't happen
		return "<nil>"
	case 1:
		return errs[0].Error()
	default:
		var str strings.Builder
		str.WriteString("detected multiple parse errors:")
		for i, err := range errs {
			str.WriteString(fmt.Sprintf("\n%d: %s", i, err.Error()))
		}
		return str.String()
	}
}

// Add adds errors to the collection.
func (p *ParseErrors) Add(errs ...*ParseError) {
	*p = append(*p, errs...)
}

// ParseError is an error that occurred during parsing.
type ParseError struct {
	ParserName string         `json:"parser_name,omitempty"`
	Type       ParseErrorType `json:"type"`
	Err        ErrMsg         `json:"error"`
	Node       *Node          `json:"node,omitempty"`
}

// Unwrap() allows errors.Is and errors.As to find wrapped errors.
func (p ParseError) Unwrap() error {
	return p.Err
}

// Error satisfies the standard library error interface.
func (p *ParseError) Error() string {
	// Add 1 to the line and column numbers to make them 1-indexed.
	return fmt.Sprintf("(%s) %s error: start %d:%d end %d:%d: %s", p.ParserName, p.Type,
		p.Node.StartLine+1, p.Node.StartCol+1,
		p.Node.EndLine+1, p.Node.EndCol+1, p.Err)
}

// ErrMsg is a type that can be used to create error messages.
// It marshals and unmarshals to and from a string.
type ErrMsg struct {
	error
}

func (e ErrMsg) Error() string {
	return e.error.Error()
}

func (e *ErrMsg) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.Error())
}

func (e *ErrMsg) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	e.error = fmt.Errorf(s)
	return nil
}

// ParseErrorTypes are used to group errors into categories.
type ParseErrorType string

const (
	ParseErrorTypeSyntax   ParseErrorType = "syntax"
	ParseErrorTypeType     ParseErrorType = "type"
	ParseErrorTypeSemantic ParseErrorType = "semantic"
	ParseErrorTypeUnknown  ParseErrorType = "unknown"
)

// Node is a node in the parse tree. It represents a range of line and column
// values in Kuneiform source code.
type Node struct {
	// Set is true if the position of the node has been set.
	// This is useful for testing parsers.
	IsSet     bool `json:"-"`
	StartLine int  `json:"start_line"`
	StartCol  int  `json:"start_col"`
	EndLine   int  `json:"end_line"`
	EndCol    int  `json:"end_col"`
}

// Set sets the position of the node based on the given parser rule context.
func (n *Node) Set(r antlr.ParserRuleContext) {
	n.IsSet = true
	n.StartLine = r.GetStart().GetLine() - 1
	n.StartCol = r.GetStart().GetColumn()
	n.EndLine = r.GetStop().GetLine() - 1
	n.EndCol = r.GetStop().GetColumn()
}

// SetToken sets the position of the node based on the given token.
func (n *Node) SetToken(t antlr.Token) {
	n.IsSet = true
	n.StartLine = t.GetLine() - 1
	n.StartCol = t.GetColumn()
	n.EndLine = t.GetLine() - 1
	n.EndCol = t.GetColumn()
}

// GetNode returns the node.
// It is useful if the node is embedded in another struct.
func (n Node) GetNode() *Node {
	return &n
}

// unaryNode creates a node with the same start and end position.
func unaryNode(start, end int) *Node {
	return &Node{
		StartLine: start,
		StartCol:  end,
		EndLine:   start,
		EndCol:    end,
	}
}

// MergeNodes merges two nodes into a single node.
// It starts at the left node and ends at the right node.
func MergeNodes(left, right *Node) *Node {
	return &Node{
		StartLine: left.StartLine,
		StartCol:  left.StartCol,
		EndLine:   right.EndLine,
		EndCol:    right.EndCol,
	}
}

// SchemaInfo contains information about a parsed schema
type SchemaInfo struct {
	// Blocks maps declared block names to their nodes.
	// Block names include:
	// - tables
	// - extensions
	// - actions
	// - procedures
	// - foreign procedures
	Blocks map[string]*Block `json:"blocks"`
}

type Block struct {
	Node
	// AbsStart is the absolute start position of the block in the source code.
	AbsStart int `json:"abs_start"`
	// AbsEnd is the absolute end position of the block in the source code.
	AbsEnd int `json:"abs_end"`
}
