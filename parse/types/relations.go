package types

import "github.com/kwilteam/kwil-db/core/types"

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
			return nil, a.Name, ErrAmbiguousColumn
		}

		colNames[a.Name] = struct{}{}
		res = append(res, &Attribute{
			Name: a.Name,
			Type: a.Type.Copy(),
		})
	}

	return res, "", nil
}
