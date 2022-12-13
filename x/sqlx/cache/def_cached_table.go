package cache

import (
	"fmt"
	spec "kwil/x/sqlx"
	"kwil/x/sqlx/models"
)

type Table struct {
	Name    string
	Columns []*Column
}

type Column struct {
	Name       string
	Type       spec.DataType
	Attributes []*Attribute
}

type Attribute struct {
	Type  spec.AttributeType
	Value any
}

func (t *Table) From(m *models.Table) error {
	t.Name = m.Name
	t.Columns = make([]*Column, len(m.Columns))
	for i, col := range m.Columns {
		t.Columns[i] = &Column{}
		if err := t.Columns[i].From(col); err != nil {
			return err
		}
	}
	return nil
}

func (c *Column) From(m *models.Column) error {
	typ, err := spec.Conversion.StringToKwilType(m.Type)
	if err != nil {
		return fmt.Errorf("failed to convert column type: %s", err.Error())
	}

	c.Name = m.Name
	c.Type = typ
	c.Attributes = make([]*Attribute, len(m.Attributes))
	for i, attr := range m.Attributes {
		c.Attributes[i] = &Attribute{}
		if err := c.Attributes[i].From(attr); err != nil {
			return err
		}
	}
	return nil
}

func (a *Attribute) From(m *models.Attribute) error {
	attrbt, err := spec.Conversion.ConvertAttribute(m.Type)
	if err != nil {
		return fmt.Errorf("failed to convert attribute type: %s", err.Error())
	}

	a.Type = attrbt
	a.Value = m.Value
	return nil
}
