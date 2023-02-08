package clean

import (
	execution2 "kwil/pkg/databases"
	"kwil/pkg/types/data_types"
	"reflect"
)

// supported enums
const (
	// Kwil supported data types
	EnumDataType = "data_type"

	// Attribute types
	EnumAttributeType = "attribute_type"

	// Query Types
	EnumQueryType = "query_type"

	// Modifiers
	EnumModifier = "modifier_type"

	// Comparison Operators
	EnumComparisonOperator = "comparison_operator_type"

	// Index Types
	EnumIndexType = "index_type"
)

func cleanEnum(val reflect.Value, tags []string) {
	switch tags[1] {
	case EnumDataType:
		inEnumRange(val, int64(datatypes.INVALID_DATA_TYPE), int64(datatypes.END_DATA_TYPE))
	case EnumAttributeType:
		inEnumRange(val, int64(execution2.INVALID_ATTRIBUTE_TYPE), int64(execution2.END_ATTRIBUTE_TYPE))
	case EnumQueryType:
		inEnumRange(val, int64(execution2.INVALID_QUERY_TYPE), int64(execution2.END_QUERY_TYPE))
	case EnumModifier:
		inEnumRange(val, int64(execution2.NO_MODIFIER), int64(execution2.END_MODIFIER_TYPE))
	case EnumComparisonOperator:
		inEnumRange(val, int64(execution2.INVALID_COMPARISON_OPERATOR_TYPE), int64(execution2.END_COMPARISON_OPERATOR_TYPE))
	case EnumIndexType:
		inEnumRange(val, int64(execution2.INVALID_INDEX_TYPE), int64(execution2.END_INDEX_TYPE))
	default:
		panic("Unknown enum type: " + tags[1])
	}
}

// inEnumRange checks if the value is in the range of the enum
// the first value is the received value
// the second is the minimum enum,
// the third is the maximum enum
// If the value is <= the minimum or >= the maximum, it is not in the range,
// and will be set to the minimum
func inEnumRange(val reflect.Value, min, max int64) {
	num := val.Int()
	if num < min || num >= max {
		val.SetInt(min)
	}
}
