package typing

import (
	"fmt"

	"github.com/kwilteam/kwil-db/core/types"
)

// Relation is a set of attributes.
type Relation struct {
	// attributes are the attributes in the relation.
	attributes map[string]*Attribute

	// OrdinalPositions tracks the ordinal positions of the attributes.
	// It is used to loop through the attributes in the order they were added.
	oridinalPositions []string
}

// newRelation creates a new relation.
func newRelation() *Relation {
	return &Relation{
		attributes:        make(map[string]*Attribute),
		oridinalPositions: []string{},
	}
}

// Attribute returns an attribute from the relation.
// It also returns whether the attribute was found.
func (q *Relation) Attribute(name string) (*Attribute, bool) {
	attr, ok := q.attributes[name]
	return attr, ok
}

// AddAttribute adds an attribute to the relation.
// If there is a conflict, it will return an error.
func (q *Relation) AddAttribute(a *QualifiedAttribute) error {
	if a.Name == "" {
		return fmt.Errorf("returned columns cannot be anonymous, and should be aliased")
	}

	if _, ok := q.attributes[a.Name]; ok {
		return fmt.Errorf(`ambiguous return column "%s"`, a.Name)
	}

	q.attributes[a.Name] = a.Copy()
	q.oridinalPositions = append(q.oridinalPositions, a.Name)

	return nil
}

// Merge merges two relations. If there is a conflict,
// an error is returned.
func (q *Relation) Merge(other *Relation) error {
	for _, attrName := range other.oridinalPositions {
		if _, ok := q.attributes[attrName]; ok {
			return fmt.Errorf("ambiguous attribute: %s", attrName)
		}

		q.attributes[attrName] = other.attributes[attrName].Copy()
	}

	q.oridinalPositions = append(q.oridinalPositions, other.oridinalPositions...)

	return nil
}

// Copy returns a copy of the relation.
func (q *Relation) Copy() *Relation {
	copied := make(map[string]*Attribute)
	for k, v := range q.attributes {
		copied[k] = v.Copy()
	}

	return &Relation{
		attributes:        copied,
		oridinalPositions: copyArr(q.oridinalPositions),
	}
}

func copyArr(arr []string) []string {
	copied := make([]string, len(arr))
	copy(copied, arr)
	return copied
}

// Loop loops through the attributes of the relation,
// in the order of their ordinal position.
// Returning an error will stop the loop.
func (q *Relation) Loop(f func(string, *Attribute) error) error {
	for _, attrName := range q.oridinalPositions {
		attr, ok := q.attributes[attrName]
		if !ok {
			panic("attribute not found during ordered loop")
		}

		err := f(attrName, attr)
		if err != nil {
			return err
		}
	}

	return nil
}

// Shape returns the shape of the relation.
func (q *Relation) Shape() []*types.DataType {
	var res []*types.DataType

	err := q.Loop(func(s string, a *Attribute) error {
		res = append(res, a.Type)
		return nil
	})
	if err != nil {
		panic(err) // this will never happen since the loop function does not return an error
	}

	return res
}

// Attribute is an anonymous attribute in a relation.
type Attribute struct {
	// Type is the type of the attribute.
	Type *types.DataType
}

// Copy returns a copy of the attribute.
func (a *Attribute) Copy() *Attribute {
	copied := *a.Type
	return &Attribute{
		Type: &copied,
	}
}

// QualifiedRelation is a relation that has a name.
// It is used for subqueries and joins.
type QualifiedRelation struct {
	*Relation
	Name string
}

// QualifiedAttribute represents an attribute in a relation
type QualifiedAttribute struct {
	*Attribute
	Name string
}

// tableToRelation converts a table to a relation.
// It is only used in the constructor of the visitor,
// and therefore could probably be deleted.
func tableToRelation(t *types.Table) *QualifiedRelation {
	attrs := make(map[string]*Attribute)
	columnNames := make([]string, len(t.Columns))

	for i, col := range t.Columns {
		attrs[col.Name] = &Attribute{
			Type: col.Type.Copy(),
		}

		columnNames[i] = col.Name
	}

	return &QualifiedRelation{
		Name: t.Name,
		Relation: &Relation{
			attributes:        attrs,
			oridinalPositions: columnNames,
		},
	}
}
