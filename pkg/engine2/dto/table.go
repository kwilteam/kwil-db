package dto

import (
	"fmt"
)

type Table struct {
	Name    string    `json:"name" clean:"lower"`
	Columns []*Column `json:"columns"`
	Indexes []*Index  `json:"indexes"`
}

func (t *Table) GetColumn(c string) *Column {
	for _, col := range t.Columns {
		if col.Name == c {
			return col
		}
	}
	return nil
}

func (t *Table) ListColumns() []string {
	var columns []string
	for _, col := range t.Columns {
		columns = append(columns, col.Name)
	}
	return columns
}

type Column struct {
	Name       string       `json:"name" clean:"lower"`
	Type       DataType     `json:"type" clean:"is_enum,data_type"`
	Attributes []*Attribute `json:"attributes,omitempty" traverse:"shallow"`
}

func (c *Column) GetAttribute(attrType AttributeType) *Attribute {
	for _, attr := range c.Attributes {
		if attr.Type == attrType {
			return attr
		}
	}
	return nil
}

type Attribute struct {
	Type  AttributeType `json:"type" clean:"is_enum,attribute_type"`
	Value any           `json:"value"`
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
