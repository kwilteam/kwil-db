package dto

import (
	"fmt"
)

type Table struct {
	Name    string    `json:"name" clean:"lower"`
	Columns []*Column `json:"columns"`
	Indexes []*Index  `json:"indexes"`
}

func (t *Table) Clean() error {
	for _, col := range t.Columns {
		if err := col.Clean(); err != nil {
			return err
		}
	}

	for _, idx := range t.Indexes {
		if err := idx.Clean(); err != nil {
			return err
		}
	}

	return runCleans(
		cleanIdent(&t.Name),
	)
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

// IsType will coerce the attribute value to the correct data type, depending on the attribute type.
// It takes an input of a column type, which is used in the case that the attribute type is DEFAULT
func (a *Attribute) IsType(columnType DataType) error {
	switch a.Type {
	case PRIMARY_KEY, UNIQUE, NOT_NULL:
		a.Value = nil
	case DEFAULT:
		// default must be the same type as the column
		return a.assertType(columnType)
	case MIN, MAX, MIN_LENGTH, MAX_LENGTH:
		// min, max, min_length, max_length must be int, regardless of column type
		return a.assertType(INT)
	default:
		return fmt.Errorf("invalid attribute type: %s", a.Type)
	}
	return nil
}

// assertType will convert the attribute value to the correct serialized type if it is not already
func (a *Attribute) assertType(typ DataType) error {
	newVal, err := typ.Coerce(a.Value)
	if err != nil {
		return err
	}

	a.Value = newVal
	return nil
}
