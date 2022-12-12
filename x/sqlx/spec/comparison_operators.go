package types

import "fmt"

type ComparisonOperatorType int

const (
	INVALID_COMPARISON_OPERATOR ComparisonOperatorType = iota
	EQUAL
	NOT_EQUAL
	GREATER_THAN
	GREATER_THAN_OR_EQUAL
	LESS_THAN
	LESS_THAN_OR_EQUAL
)

func (c *ComparisonOperatorType) String() string {
	switch *c {
	case EQUAL:
		return "="
	case NOT_EQUAL:
		return "!="
	case GREATER_THAN:
		return ">"
	case GREATER_THAN_OR_EQUAL:
		return ">="
	case LESS_THAN:
		return "<"
	case LESS_THAN_OR_EQUAL:
		return "<="
	}
	return "unknown"
}

// ConvertComparisonOperator converts a string to a ComparisonOperatorType
func (c *conversion) ConvertComparisonOperator(s string) (ComparisonOperatorType, error) {
	switch s {
	case "=":
		return EQUAL, nil
	case "!=":
		return NOT_EQUAL, nil
	case ">":
		return GREATER_THAN, nil
	case ">=":
		return GREATER_THAN_OR_EQUAL, nil
	case "<":
		return LESS_THAN, nil
	case "<=":
		return LESS_THAN_OR_EQUAL, nil
	}
	return INVALID_COMPARISON_OPERATOR, fmt.Errorf("unknown comparison operator type")
}
