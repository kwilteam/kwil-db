package ddlbuilder

import (
	"fmt"
	"kwil/pkg/databases/spec"
	"strings"

	conv "github.com/cstockton/go-conv"
)

type Column struct {
	name       string
	typ        string
	attributes map[spec.AttributeType]any
}

type namePicker interface {
	Name(string) typePicker
}

type typePicker interface {
	Type(string) ColumnBuilder
}

type ColumnBuilder interface {
	builder
	WithAttribute(spec.AttributeType, any) ColumnBuilder
	BuildAttributes(schema string, table string) ([]string, error)
	GetName() string
}

func NewColumnBuilder() namePicker {
	return &Column{
		attributes: make(map[spec.AttributeType]any),
	}
}

func (b *Column) Name(name string) typePicker {
	b.name = name
	return b
}

func (b *Column) Type(typ string) ColumnBuilder {
	b.typ = typ
	return b
}

func (b *Column) GetName() string {
	return b.name
}

// Attributes
func (c *Column) WithAttribute(attr spec.AttributeType, value any) ColumnBuilder {
	c.attributes[attr] = value
	return c
}

func (b *Column) Build() string {

	sb := &strings.Builder{}
	sb.WriteString(b.name)
	sb.WriteString(" ")
	sb.WriteString(b.typ)
	return sb.String()
}

func (c *Column) BuildAttributes(schema, table string) ([]string, error) {
	var attributes []string

	for attr, value := range c.attributes {
		ab := NewAttributeBuilder()
		if schema != "" {
			ab.Schema(schema)
		}
		if table == "" {
			return attributes, fmt.Errorf("table name is required for attribute %s", attr.String())
		}

		// attribute picker
		ap := ab.Table(table)

		var res string
		switch attr {
		case spec.PRIMARY_KEY:
			res = ap.PrimaryKey(c.name).Build()
		case spec.UNIQUE:
			res = ap.Unique(c.name).Build()
		case spec.NOT_NULL:
			res = ap.NotNull(c.name).Build()
		case spec.DEFAULT:
			res = ap.Default(c.name, value).Build()
		case spec.MIN:
			val, err := conv.Int(value)
			if err != nil {
				return attributes, fmt.Errorf("min attribute value must be an integer")
			}

			res = ap.Min(c.name, val).Build()
		case spec.MAX:
			val, err := conv.Int(value)
			if err != nil {
				return attributes, fmt.Errorf("max attribute value must be an integer")
			}

			res = ap.Max(c.name, val).Build()
		case spec.MIN_LENGTH:
			val, err := conv.Int(value)
			if err != nil {
				return attributes, fmt.Errorf("min_length attribute value must be an integer")
			}

			res = ap.MinLength(c.name, val).Build()
		case spec.MAX_LENGTH:
			val, err := conv.Int(value)
			if err != nil {
				return attributes, fmt.Errorf("max_length attribute value must be an integer")
			}
			res = ap.MaxLength(c.name, val).Build()
		default:
			return attributes, fmt.Errorf("unknown attribute %d", attr.Int())
		}

		attributes = append(attributes, res)
	}

	return attributes, nil
}
