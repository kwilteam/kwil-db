package spec

import (
	"fmt"

	"github.com/cstockton/go-conv"
)

// attributeFuncs are functions that apply the business logic of attributes
type attributeFuncs map[AttributeType]func(v *KwilAny, attribute *KwilAny) (*KwilAny, error)

var AttributeFuncs = attributeFuncs{
	PRIMARY_KEY: primaryKey,
	UNIQUE:      unique,
	NOT_NULL:    notNull,
	DEFAULT:     defaultValue,
	MIN:         min,
	MAX:         max,
	MIN_LENGTH:  minLength,
	MAX_LENGTH:  maxLength,
}

// Below I have included functions for all attributes, even if the attribute does not change the value.

// PrimaryKey does not modify the value
func primaryKey(v *KwilAny, attribute *KwilAny) (*KwilAny, error) {
	return v, nil
}

// Unique does not modify the value
func unique(v *KwilAny, attribute *KwilAny) (*KwilAny, error) {
	return v, nil
}

// NotNull checks that the value is not NULL or empty
func notNull(v *KwilAny, attribute *KwilAny) (*KwilAny, error) {
	// must check is not NULL type
	if v.Type() == NULL {
		return v, fmt.Errorf(`received NULL value on column containing attribute "not_null"`)
	}

	// must check bytewise that it is not empty after the first byte (since the first byte is the type)
	if v.Bytes()[:1] == nil {
		return v, fmt.Errorf(`received empty value on column containing attribute "not_null"`)
	}

	return v, nil
}

// Default sets the value to the default value
func defaultValue(v *KwilAny, attribute *KwilAny) (*KwilAny, error) {
	// copy the value
	newValue := *attribute
	return &newValue, nil
}

// Min checks that the value is greater than or equal to the min value
func min(v *KwilAny, attribute *KwilAny) (*KwilAny, error) {
	value, attr, err := asInts(v, attribute)
	if err != nil {
		return v, fmt.Errorf(`failed to convert values to ints: %w`, err)
	}

	if value < attr {
		return v, fmt.Errorf(`received value %v on column containing attribute "min"`, value)
	}

	return v, nil
}

// Max checks that the value is greater than or equal to the min value
func max(v *KwilAny, attribute *KwilAny) (*KwilAny, error) {
	value, attr, err := asInts(v, attribute)
	if err != nil {
		return v, fmt.Errorf(`failed to convert values to ints: %w`, err)
	}

	if value > attr {
		return v, fmt.Errorf(`received value %v on column containing attribute "max"`, value)
	}

	return v, nil
}

// MinLength checks that the value is greater than or equal to the min value
func minLength(v *KwilAny, attribute *KwilAny) (*KwilAny, error) {
	strVal, attrVal, err := asStringAndInt(v, attribute)
	if err != nil {
		return v, fmt.Errorf(`failed to convert values to strings and ints: %w`, err)
	}

	if len(strVal) < attrVal {
		return v, fmt.Errorf(`column has min_length of %v but received value %v`, attrVal, strVal)
	}

	return v, nil
}

// MaxLength checks that the value is greater than or equal to the min value
func maxLength(v *KwilAny, attribute *KwilAny) (*KwilAny, error) {
	strVal, attrVal, err := asStringAndInt(v, attribute)
	if err != nil {
		return v, fmt.Errorf(`failed to convert values to strings and ints: %w`, err)
	}

	if len(strVal) > attrVal {
		return v, fmt.Errorf(`column has min_length of %v but received value %v`, attrVal, strVal)
	}

	return v, nil
}

// utility functions
func asInts(v *KwilAny, attribute *KwilAny) (int, int, error) {
	dataType := v.Type()
	if !dataType.IsNumeric() {
		return 0, 0, fmt.Errorf(`received non-numeric value on column containing attribute "min"`)
	}
	// convert the value to the same type as the attribute
	intVal, err := conv.Int(v.Value())
	if err != nil {
		return 0, 0, fmt.Errorf(`failed to convert value to int: %w`, err)
	}

	// convert the attribute to the same type as the value
	intAttr, err := conv.Int(attribute.Value())
	if err != nil {
		return 0, 0, fmt.Errorf(`failed to convert attribute to int: %w`, err)
	}

	return intVal, intAttr, nil
}

func asStringAndInt(v *KwilAny, attribute *KwilAny) (string, int, error) {
	dataType := v.Type()
	if !dataType.IsText() {
		return "", 0, fmt.Errorf(`received non-string value on column containing attribute "min_length"`)
	}
	// convert the value to the same type as the attribute
	strVal, err := v.AsString()
	if err != nil {
		return "", 0, fmt.Errorf(`failed to convert value to string: %w`, err)
	}

	// convert the attribute to the same type as the value
	intAttr, err := attribute.AsInt()
	if err != nil {
		return "", 0, fmt.Errorf(`failed to convert attribute to int: %w`, err)
	}

	return strVal, intAttr, nil
}
