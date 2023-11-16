package types

import (
	"fmt"
)

type Table struct {
	Name        string        `json:"name"`
	Columns     []*Column     `json:"columns"`
	Indexes     []*Index      `json:"indexes,omitempty"`
	ForeignKeys []*ForeignKey `json:"foreign_keys"`
}

// Clean validates rules about the data in the struct (naming conventions, syntax, etc.).
func (t *Table) Clean() error {
	hasPrimaryAttribute := false
	for _, col := range t.Columns {
		if err := col.Clean(); err != nil {
			return err
		}
		if col.hasPrimary() {
			if hasPrimaryAttribute {
				return fmt.Errorf("table %s has multiple primary attributes", t.Name)
			}
			hasPrimaryAttribute = true
		}
	}

	hasPrimaryIndex := false
	for _, idx := range t.Indexes {
		if err := idx.Clean(); err != nil {
			return err
		}

		if idx.Type == PRIMARY {
			if hasPrimaryIndex {
				return fmt.Errorf("table %s has multiple primary indexes", t.Name)
			}
			hasPrimaryIndex = true
		}
	}

	if !hasPrimaryAttribute && !hasPrimaryIndex {
		return fmt.Errorf("table %s has no primary key", t.Name)
	}

	if hasPrimaryAttribute && hasPrimaryIndex {
		return fmt.Errorf("table %s has both primary attribute and primary index", t.Name)
	}

	_, err := t.GetPrimaryKey()
	if err != nil {
		return err
	}

	return runCleans(
		cleanIdent(&t.Name),
	)
}

// GetPrimaryKey returns the names of the column(s) that make up the primary key.
// If there is more than one, or no primary key, an error is returned.
func (t *Table) GetPrimaryKey() ([]string, error) {
	var primaryKey []string

	hasAttributePrimaryKey := false
	for _, col := range t.Columns {
		for _, attr := range col.Attributes {
			if attr.Type == PRIMARY_KEY {
				if hasAttributePrimaryKey {
					return nil, fmt.Errorf("table %s has multiple primary attributes", t.Name)
				}
				hasAttributePrimaryKey = true
				primaryKey = []string{col.Name}
			}
		}
	}

	hasIndexPrimaryKey := false
	for _, idx := range t.Indexes {
		if idx.Type == PRIMARY {
			if hasIndexPrimaryKey {
				return nil, fmt.Errorf("table %s has multiple primary indexes", t.Name)
			}
			hasIndexPrimaryKey = true

			// copy
			// if we do not copy, then the returned slice will allow modification of the index
			primaryKey = make([]string, len(idx.Columns))
			copy(primaryKey, idx.Columns)
		}
	}

	if !hasAttributePrimaryKey && !hasIndexPrimaryKey {
		return nil, fmt.Errorf("table %s has no primary key", t.Name)
	}

	if hasAttributePrimaryKey && hasIndexPrimaryKey {
		return nil, fmt.Errorf("table %s has both primary attribute and primary index", t.Name)
	}

	return primaryKey, nil
}

// Copy returns a copy of the table
func (t *Table) Copy() *Table {
	res := &Table{
		Name: t.Name,
	}

	for _, col := range t.Columns {
		res.Columns = append(res.Columns, col.Copy())
	}

	for _, idx := range t.Indexes {
		res.Indexes = append(res.Indexes, idx.Copy())
	}

	for _, fk := range t.ForeignKeys {
		res.ForeignKeys = append(res.ForeignKeys, fk.Copy())
	}

	return res
}

type Column struct {
	Name       string       `json:"name"`
	Type       DataType     `json:"type"`
	Attributes []*Attribute `json:"attributes,omitempty"`
}

func (c *Column) Clean() error {
	for _, attr := range c.Attributes {
		if err := attr.Clean(); err != nil {
			return err
		}
	}

	return runCleans(
		cleanIdent(&c.Name),
		c.Type.Clean(),
	)
}

// Copy returns a copy of the column
func (c *Column) Copy() *Column {
	res := &Column{
		Name: c.Name,
		Type: c.Type,
	}

	for _, attr := range c.Attributes {
		res.Attributes = append(res.Attributes, attr.Copy())
	}

	return res
}

func (c *Column) hasPrimary() bool {
	for _, attr := range c.Attributes {
		if attr.Type == PRIMARY_KEY {
			return true
		}
	}
	return false
}

type Attribute struct {
	Type  AttributeType `json:"type"`
	Value string        `json:"value,omitempty"`
}

func (a *Attribute) Clean() error {
	return runCleans(
		a.Type.Clean(),
	)
}

// Copy returns a copy of the attribute
func (a *Attribute) Copy() *Attribute {
	return &Attribute{
		Type:  a.Type,
		Value: a.Value,
	}
}
