package execution

import (
	"fmt"
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
	return INVALID_ATTRIBUTE, fmt.Errorf(`unknown attribute: "%s"`, s)
}
