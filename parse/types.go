package parse

import (
	"github.com/antlr4-go/antlr/v4"
	"github.com/kwilteam/kwil-db/core/types"
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

// GetNode returns the Position.
// It is useful if the Position is embedded in another struct.
func (n Position) GetNode() *Position {
	return &n
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

// MergeNodes merges two Positions into a single Position.
// It starts at the left Position and ends at the right Position.
func MergeNodes(left, right *Position) *Position {
	return &Position{
		StartLine: left.StartLine,
		StartCol:  left.StartCol,
		EndLine:   right.EndLine,
		EndCol:    right.EndCol,
	}
}

// SchemaInfo contains information about a parsed schema
type SchemaInfo struct {
	// Blocks maps declared block names to their Positions.
	// Block names include:
	// - tables
	// - extensions
	// - actions
	// - procedures
	// - foreign procedures
	Blocks map[string]*Block `json:"blocks"`
}

type Block struct {
	Position
	// AbsStart is the absolute start position of the block in the source code.
	AbsStart int `json:"abs_start"`
	// AbsEnd is the absolute end position of the block in the source code.
	AbsEnd int `json:"abs_end"`
}

// Relation represents a relation in a sql statement.
// It is meant to represent the shape of a relation, not the data.
type Relation struct {
	Name string
	// Attributes holds the attributes of the relation.
	Attributes []*Attribute
}

// ShapesMatch checks if the shapes of two relations match.
// It only checks types and order, and ignores the names of the attributes,
// as well as uniques and primary keys.
func (r *Relation) ShapesMatch(r2 *Relation) bool {
	return ShapesMatch(r.Attributes, r2.Attributes)
}

// FindAttribute finds an attribute by name.
// If it is not found, it returns nil and false
func (r *Relation) FindAttribute(name string) (*Attribute, bool) {
	for _, a := range r.Attributes {
		if a.Name == name {
			return a, true
		}
	}
	return nil, false
}

// Copy returns a copy of the relation.
func (r *Relation) Copy() *Relation {
	attrs := make([]*Attribute, len(r.Attributes))
	for i, a := range r.Attributes {
		attrs[i] = &Attribute{
			Name: a.Name,
			Type: a.Type.Copy(),
		}
	}

	return &Relation{
		Name:       r.Name,
		Attributes: attrs,
	}
}

// Attribute represents an attribute in a relation.
type Attribute struct {
	Name string
	Type *types.DataType
}

// ShapesMatch checks if the shapes of two relations match.
// It only checks types and order, and ignores the names of the attributes,
// as well as uniques and primary keys.
func ShapesMatch(a1, a2 []*Attribute) bool {
	if len(a1) != len(a2) {
		return false
	}
	for i := range a1 {
		if !a1[i].Type.Equals(a2[i].Type) {
			return false
		}
	}
	return true
}

// Flatten flattens many relations into one unnamed relation.
// If there are column conflicts / ambiguities, it will return an error,
// and the name that caused the conflict. It will also discard any
// primary keys, and will unmark any unique columns.
func Flatten(rels ...*Relation) (res []*Attribute, col string, err error) {
	for _, r := range rels {
		res, col, err = Coalesce(append(res, r.Attributes...)...)
		if err != nil {
			return nil, col, err
		}
	}

	return res, "", nil
}

// Coalesce coalesces sets of attributes. If there are ambiguities, it will
// return an error, and the name that caused the conflict. It will also discard
// any primary keys, and will unmark any unique columns.
func Coalesce(attrs ...*Attribute) (res []*Attribute, ambigousCol string, err error) {
	colNames := make(map[string]struct{})

	for _, a := range attrs {
		if _, ok := colNames[a.Name]; ok {
			return nil, a.Name, ErrDuplicateResultColumnName
		}

		colNames[a.Name] = struct{}{}
		res = append(res, &Attribute{
			Name: a.Name,
			Type: a.Type.Copy(),
		})
	}

	return res, "", nil
}

// findAttribute finds an attribute by name.
func findAttribute(attrs []*Attribute, name string) *Attribute {
	for _, a := range attrs {
		if a.Name == name {
			return a
		}
	}
	return nil
}
