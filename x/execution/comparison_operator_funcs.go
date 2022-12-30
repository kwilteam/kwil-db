package execution

import "fmt"

type comparsionOperators struct{}

var ComparisonOperators = &comparsionOperators{}

// ConvertComparisonOperator converts a string to a ComparisonOperatorType
func (c *comparsionOperators) ConvertComparisonOperator(s string) (ComparisonOperatorType, error) {
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
