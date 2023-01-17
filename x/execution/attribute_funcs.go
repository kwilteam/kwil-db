package execution

import (
	"fmt"

	"github.com/cstockton/go-conv"
)

type attributes struct{}

var Attributes = &attributes{}

// ConvertAttribute converts a string to an AttributeType
func (c *attributes) ConvertAttribute(s string) (AttributeType, error) {
	switch s {
	case `primary_key`:
		return PRIMARY_KEY, nil
	case `unique`:
		return UNIQUE, nil
	case `not_null`:
		return NOT_NULL, nil
	case `default`:
		return DEFAULT, nil
	case `min`:
		return MIN, nil
	case `max`:
		return MAX, nil
	case `min_length`:
		return MIN_LENGTH, nil
	case `max_length`:
		return MAX_LENGTH, nil
	}
	return INVALID_ATTRIBUTE_TYPE, fmt.Errorf(`unknown attribute: "%s"`, s)
}

// converts the passed type to the required type for the specified attribute
func (c *attributes) ConvertAnyToValidAttributeType(v any, attribute AttributeType) (any, error) {
	switch attribute {
	case PRIMARY_KEY:
		return nil, nil
	case UNIQUE:
		return nil, nil
	case NOT_NULL:
		return nil, nil
	case DEFAULT:
		return v, nil
	case MIN:
		return conv.Int64(v)
	case MAX:
		return conv.Int64(v)
	case MIN_LENGTH:
		return conv.Int64(v)
	case MAX_LENGTH:
		return conv.Int64(v)
	}

	return nil, fmt.Errorf(`unknown attribute: "%s"`, attribute.String())
}
