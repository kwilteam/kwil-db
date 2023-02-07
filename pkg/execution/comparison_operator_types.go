package execution

type ComparisonOperatorType int

const (
	INVALID_COMPARISON_OPERATOR_TYPE ComparisonOperatorType = iota + 100
	EQUAL
	NOT_EQUAL
	GREATER_THAN
	GREATER_THAN_OR_EQUAL
	LESS_THAN
	LESS_THAN_OR_EQUAL
	END_COMPARISON_OPERATOR_TYPE
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

func (c *ComparisonOperatorType) Int() int {
	return int(*c)
}

func (c *ComparisonOperatorType) IsValid() bool {
	return *c > INVALID_COMPARISON_OPERATOR_TYPE && *c < END_COMPARISON_OPERATOR_TYPE
}
