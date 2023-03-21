package types

import "fmt"

/*
###########################################################################
Attributes
###########################################################################
*/
type AttributeType int

const (
	INVALID_ATTRIBUTE_TYPE AttributeType = iota + 100
	PRIMARY_KEY
	UNIQUE
	NOT_NULL
	DEFAULT
	MIN        // Min allowed value
	MAX        // Max allowed value
	MIN_LENGTH // Min allowed length
	MAX_LENGTH // Max allowed length
	END_ATTRIBUTE_TYPE
)

func (a *AttributeType) String() string {
	switch *a {
	case PRIMARY_KEY:
		return `primary_key`
	case UNIQUE:
		return `unique`
	case NOT_NULL:
		return `not_null`
	case DEFAULT:
		return `default`
	case MIN:
		return `min`
	case MAX:
		return `max`
	case MIN_LENGTH:
		return `min_length`
	case MAX_LENGTH:
		return `max_length`
	}
	return `unknown`
}

func (a *AttributeType) Int() int {
	return int(*a)
}

func (a *AttributeType) IsValid() bool {
	return *a > INVALID_ATTRIBUTE_TYPE && *a < END_ATTRIBUTE_TYPE
}

/*
	###########################################################################
	Comparison Operators
	###########################################################################
*/

type ComparisonOperatorType int

const (
	INVALID_COMPARISON_OPERATOR_TYPE ComparisonOperatorType = iota + 100
	EQUAL
	NOT_EQUAL
	GREATER_THAN
	GREATER_THAN_OR_EQUAL
	LESS_THAN
	LESS_THAN_OR_EQUAL
	IN
	NOT_IN
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

/*
	###########################################################################
	Indexes
	###########################################################################
*/

type IndexType int

const (
	INVALID_INDEX_TYPE IndexType = iota + 100
	BTREE
	UNIQUE_BTREE
	END_INDEX_TYPE
)

func (i *IndexType) String() string {
	switch *i {
	case BTREE:
		return "btree"
	case UNIQUE_BTREE:
		return "unique_btree"
	}
	return "unknown"
}

func (i *IndexType) Int() int {
	return int(*i)
}

func (i *IndexType) IsValid() bool {
	return *i > INVALID_INDEX_TYPE && *i < END_INDEX_TYPE
}

/*
	###########################################################################
	Modifiers
	###########################################################################
*/

type ModifierType int

// Modifiers
const (
	NO_MODIFIER           ModifierType = 0
	INVALID_MODIFIER_TYPE ModifierType = iota + 99
	CALLER
	END_MODIFIER_TYPE
)

func (m ModifierType) String() string {
	switch m {
	case CALLER:
		return "caller"
	}
	return "unknown"
}

func (m ModifierType) Int() int {
	return int(m)
}

func (m ModifierType) IsValid() bool {
	if m == NO_MODIFIER {
		return true
	}
	return m > INVALID_MODIFIER_TYPE && m < END_MODIFIER_TYPE
}

/*
	###########################################################################
	Query Types
	###########################################################################
*/

type QueryType int

// Queries
const (
	INVALID_QUERY_TYPE QueryType = iota + 100
	INSERT
	UPDATE
	DELETE
	SELECT
	END_QUERY_TYPE
)

func (q *QueryType) Int() int {
	return int(*q)
}

func (q *QueryType) String() (string, error) {
	switch *q {
	case INSERT:
		return "insert", nil
	case UPDATE:
		return "update", nil
	case DELETE:
		return "delete", nil
	}
	return "", fmt.Errorf("unknown query type")
}

func (q *QueryType) IsValid() bool {
	return *q > INVALID_QUERY_TYPE && *q < END_QUERY_TYPE
}
