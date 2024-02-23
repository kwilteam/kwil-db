package tree

import "fmt"

type BinaryOperator interface {
	binary()
	String() string
	Valid() error
}

type ArithmeticOperator string

const (
	ArithmeticOperatorAdd      ArithmeticOperator = "+"
	ArithmeticOperatorSubtract ArithmeticOperator = "-"
	ArithmeticOperatorMultiply ArithmeticOperator = "*"
	ArithmeticOperatorDivide   ArithmeticOperator = "/"
	ArithmeticOperatorModulus  ArithmeticOperator = "%"
)

func (a ArithmeticOperator) String() string {
	return string(a)
}

func (a ArithmeticOperator) Valid() error {
	switch a {
	case ArithmeticOperatorAdd, ArithmeticOperatorSubtract, ArithmeticOperatorMultiply, ArithmeticOperatorDivide, ArithmeticOperatorModulus:
		return nil
	default:
		return fmt.Errorf("invalid arithmetic operator: %s", a)
	}
}

type ComparisonOperator string

const (
	ComparisonOperatorEqual              ComparisonOperator = "="
	ComparisonOperatorNotEqualDiamond    ComparisonOperator = "<>"
	ComparisonOperatorNotEqual           ComparisonOperator = "!="
	ComparisonOperatorGreaterThan        ComparisonOperator = ">"
	ComparisonOperatorLessThan           ComparisonOperator = "<"
	ComparisonOperatorGreaterThanOrEqual ComparisonOperator = ">="
	ComparisonOperatorLessThanOrEqual    ComparisonOperator = "<="
	ComparisonOperatorIn                 ComparisonOperator = "IN"
	ComparisonOperatorNotIn              ComparisonOperator = "NOT IN"
)

func (c ComparisonOperator) binary() {}
func (c ComparisonOperator) String() string {
	return string(c)
}

func (c ComparisonOperator) Valid() error {
	switch c {
	case ComparisonOperatorEqual, ComparisonOperatorNotEqualDiamond, ComparisonOperatorNotEqual, ComparisonOperatorGreaterThan, ComparisonOperatorLessThan, ComparisonOperatorGreaterThanOrEqual, ComparisonOperatorLessThanOrEqual, ComparisonOperatorIn, ComparisonOperatorNotIn:
		return nil
	default:
		return fmt.Errorf("invalid comparison operator: %s", c)
	}
}

type LogicalOperator string

const (
	LogicalOperatorAnd LogicalOperator = "AND"
	LogicalOperatorOr  LogicalOperator = "OR"
)

func (l LogicalOperator) binary() {}
func (l LogicalOperator) String() string {
	return string(l)
}

func (l LogicalOperator) Valid() error {
	switch l {
	case LogicalOperatorAnd, LogicalOperatorOr:
		return nil
	default:
		return fmt.Errorf("invalid logical operator: %s", l)
	}
}

type StringOperator string

const (
	StringOperatorLike    StringOperator = "LIKE"
	StringOperatorNotLike StringOperator = "NOT LIKE"
)

func (s StringOperator) binary() {}
func (s StringOperator) String() string {
	return string(s)
}
func (s StringOperator) Valid() error {
	switch s {
	case StringOperatorLike, StringOperatorNotLike:
		return nil
	default:
		return fmt.Errorf("invalid string operator: %s", s)
	}
}
func (s StringOperator) Escapable() bool {
	switch s {
	case StringOperatorLike, StringOperatorNotLike:
		return true
	default:
		return false
	}
}

type UnaryOperator string

const (
	UnaryOperatorPlus  UnaryOperator = "+"
	UnaryOperatorMinus UnaryOperator = "-"
	UnaryOperatorNot   UnaryOperator = "NOT"
	//UnaryOperatorBitNot UnaryOperator = "~"
)

func (u UnaryOperator) String() string {
	return string(u)
}

func (u UnaryOperator) Valid() error {
	switch u {
	case UnaryOperatorPlus, UnaryOperatorMinus, UnaryOperatorNot:
		return nil
	default:
		return fmt.Errorf("invalid unary operator: %s", u)
	}
}
