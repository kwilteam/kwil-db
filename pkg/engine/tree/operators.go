package tree

import "github.com/doug-martin/goqu/v9"

type BinaryOperator interface {
	Compare(left, right Expression) map[string]interface{}
}

type ArithmeticOperator uint8

const (
	ArithmeticOperatorAdd ArithmeticOperator = iota
	ArithmeticOperatorSubtract
	ArithmeticOperatorMultiply
	ArithmeticOperatorDivide
	ArithmeticOperatorModulus
)

func (a ArithmeticOperator) Compare(left, right Expression) map[string]interface{} {
	return goqu.Ex{}
}

type ComparisonOperator uint8

const (
	ComparisonOperatorEqual ComparisonOperator = iota
	ComparisonOperatorNotEqual
	ComparisonOperatorGreaterThan
	ComparisonOperatorLessThan
	ComparisonOperatorGreaterThanOrEqual
	ComparisonOperatorLessThanOrEqual
	ComparisonOperatorIs
	ComparisonOperatorIsNot
	ComparisonOperatorIn
	ComparisonOperatorNotIn
	ComparisonOperatorBetween
	ComparisonOperatorNotBetween
)

type BitwiseOperator uint8

const (
	BitwiseOperatorAnd BitwiseOperator = iota
	BitwiseOperatorOr
	BitwiseOperatorXor
	BitwiseOperatorNot
	BitwiseOperatorLeftShift
	BitwiseOperatorRightShift
)

type LogicalOperator uint8

const (
	LogicalOperatorAnd LogicalOperator = iota
	LogicalOperatorOr
)

type StringOperator uint8

const (
	ComparisonOperatorLike StringOperator = iota
	ComparisonOperatorNotLike
	ComparisonOperatorGlob
	ComparisonOperatorNotGlob
	ComparisonOperatorRegexp
	ComparisonOperatorNotRegexp
	StringOperatorMatch
	StringOperatorNotMatch
)
