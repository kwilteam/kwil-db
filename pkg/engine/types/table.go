package types

import (
	"fmt"
)

type Table struct {
	Name        string        `json:"name" clean:"lower"`
	Columns     []*Column     `json:"columns"`
	Indexes     []*Index      `json:"indexes,omitempty"`
	ForeignKeys []*ForeignKey `json:"foreign_keys"`
}

func (t *Table) Identifier() string {
	return t.Name
}

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
			primaryKey = idx.Columns
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

type Column struct {
	Name       string       `json:"name" clean:"lower"`
	Type       DataType     `json:"type" clean:"is_enum,data_type"`
	Attributes []*Attribute `json:"attributes,omitempty" traverse:"shallow"`
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

func (c *Column) hasPrimary() bool {
	for _, attr := range c.Attributes {
		if attr.Type == PRIMARY_KEY {
			return true
		}
	}
	return false
}

type Attribute struct {
	Type  AttributeType `json:"type" clean:"is_enum,attribute_type"`
	Value any           `json:"value"`
}

func (a *Attribute) Clean() error {
	if a.Value == nil {
		return a.Type.Clean()
	}

	return runCleans(
		a.Type.Clean(),
		cleanScalar(&a.Value),
	)
}
