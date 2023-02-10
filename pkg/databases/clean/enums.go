package clean

import (
	execution2 "kwil/pkg/databases/spec"
	"reflect"
)

// supported enums
const (
	// Kwil supported data types
	enumDataType = "data_type"

	// Attribute types
	enumAttributeType = "attribute_type"

	// Query Types
	enumQueryType = "query_type"

	// Modifiers
	enumModifier = "modifier_type"

	// Comparison Operators
	enumComparisonOperator = "comparison_operator_type"

	// Index Types
	enumIndexType = "index_type"
)

func cleanEnum(val reflect.Value, tags []string) {
	switch tags[1] {
	case enumDataType:
		inenumRange(val, int64(execution2.INVALID_DATA_TYPE), int64(execution2.END_DATA_TYPE))
	case enumAttributeType:
		inenumRange(val, int64(execution2.INVALID_ATTRIBUTE_TYPE), int64(execution2.END_ATTRIBUTE_TYPE))
	case enumQueryType:
		inenumRange(val, int64(execution2.INVALID_QUERY_TYPE), int64(execution2.END_QUERY_TYPE))
	case enumModifier:
		inenumRange(val, int64(execution2.NO_MODIFIER), int64(execution2.END_MODIFIER_TYPE))
	case enumComparisonOperator:
		inenumRange(val, int64(execution2.INVALID_COMPARISON_OPERATOR_TYPE), int64(execution2.END_COMPARISON_OPERATOR_TYPE))
	case enumIndexType:
		inenumRange(val, int64(execution2.INVALID_INDEX_TYPE), int64(execution2.END_INDEX_TYPE))
	default:
		panic("Unknown enum type: " + tags[1]) // since it is scanning our own struct tags, we should never get here
	}
}

// inenumRange checks if the value is in the range of the enum
// the first value is the received value
// the second is the minimum enum,
// the third is the maximum enum
// If the value is <= the minimum or >= the maximum, it is not in the range,
// and will be set to the minimum
func inenumRange(val reflect.Value, min, max int64) {
	num := val.Int()
	if num < min || num >= max {
		val.SetInt(min)
	}
}
