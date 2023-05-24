package tree

type BinaryOperator interface {
	Binary()
	String() string
}

type ArithmeticOperator string

const (
	ArithmeticOperatorAdd      ArithmeticOperator = "+"
	ArithmeticOperatorSubtract ArithmeticOperator = "-"
	ArithmeticOperatorMultiply ArithmeticOperator = "*"
	ArithmeticOperatorDivide   ArithmeticOperator = "/"
	ArithmeticOperatorModulus  ArithmeticOperator = "%"
)

func (a ArithmeticOperator) Binary() {}
func (a ArithmeticOperator) String() string {
	return string(a)
}

type ComparisonOperator string

const (
	ComparisonOperatorEqual              ComparisonOperator = "="
	ComparisonOperatorNotEqual           ComparisonOperator = "!="
	ComparisonOperatorGreaterThan        ComparisonOperator = ">"
	ComparisonOperatorLessThan           ComparisonOperator = "<"
	ComparisonOperatorGreaterThanOrEqual ComparisonOperator = ">="
	ComparisonOperatorLessThanOrEqual    ComparisonOperator = "<="
	ComparisonOperatorIs                 ComparisonOperator = "IS"
	ComparisonOperatorIsNot              ComparisonOperator = "IS NOT"
	ComparisonOperatorIn                 ComparisonOperator = "IN"
	ComparisonOperatorNotIn              ComparisonOperator = "NOT IN"
	ComparisonOperatorBetween            ComparisonOperator = "BETWEEN"
	ComparisonOperatorNotBetween         ComparisonOperator = "NOT BETWEEN"
)

func (c ComparisonOperator) Binary() {}
func (c ComparisonOperator) String() string {
	return string(c)
}

type BitwiseOperator string

const (
	BitwiseOperatorAnd        BitwiseOperator = "&"
	BitwiseOperatorOr         BitwiseOperator = "|"
	BitwiseOperatorXor        BitwiseOperator = "^"
	BitwiseOperatorNot        BitwiseOperator = "~"
	BitwiseOperatorLeftShift  BitwiseOperator = "<<"
	BitwiseOperatorRightShift BitwiseOperator = ">>"
)

func (b BitwiseOperator) Binary() {}
func (b BitwiseOperator) String() string {
	return string(b)
}

type LogicalOperator string

const (
	LogicalOperatorAnd LogicalOperator = "AND"
	LogicalOperatorOr  LogicalOperator = "OR"
)

func (l LogicalOperator) Binary() {}
func (l LogicalOperator) String() string {
	return string(l)
}

type StringOperator string

const (
	StringOperatorLike      StringOperator = "LIKE"
	StringOperatorNotLike   StringOperator = "NOT LIKE"
	StringOperatorGlob      StringOperator = "GLOB"
	StringOperatorNotGlob   StringOperator = "NOT GLOB"
	StringOperatorRegexp    StringOperator = "REGEXP"
	StringOperatorNotRegexp StringOperator = "NOT REGEXP"
	StringOperatorMatch     StringOperator = "MATCH"
	StringOperatorNotMatch  StringOperator = "NOT MATCH"
)

func (s StringOperator) Binary() {}
func (s StringOperator) String() string {
	return string(s)
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
	UnaryOperatorPlus   UnaryOperator = "+"
	UnaryOperatorMinus  UnaryOperator = "-"
	UnaryOperatorNot    UnaryOperator = "NOT"
	UnaryOperatorBitNot UnaryOperator = "~"
)

func (u UnaryOperator) String() string {
	return string(u)
}
