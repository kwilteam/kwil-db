package interpreter

import (
	"fmt"

	"github.com/kwilteam/kwil-db/internal/engine2/parse"
)

type ComparisonOp uint8

const (
	equal ComparisonOp = iota
	lessThan
	greaterThan
	is
	isDistinctFrom
)

type UnaryOp uint8

const (
	not UnaryOp = iota
	neg
	pos
)

func (op UnaryOp) String() string {
	switch op {
	case not:
		return "NOT"
	case neg:
		return "-"
	case pos:
		return "+"
	}

	panic(fmt.Sprintf("unknown unary operator: %d", op))
}

type ArithmeticOp uint8

const (
	add ArithmeticOp = iota
	sub
	mul
	div
	mod
	concat
)

func (op ArithmeticOp) String() string {
	switch op {
	case add:
		return "+"
	case sub:
		return "-"
	case mul:
		return "*"
	case div:
		return "/"
	case mod:
		return "%"
	case concat:
		return "||"
	}

	panic(fmt.Sprintf("unknown arithmetic operator: %d", op))
}

func (op ComparisonOp) String() string {
	switch op {
	case equal:
		return "="
	case lessThan:
		return "<"
	case greaterThan:
		return ">"
	case is:
		return "IS"
	case isDistinctFrom:
		return "IS DISTINCT FROM"
	}

	panic(fmt.Sprintf("unknown comparison operator: %d", op))
}

// GetComparisonOps gets the comparison operators for the given operator.
// Since the interpreter has a restricted subset of comparison operators compared to the parser,
// it is possible that one parser operator maps to multiple interpreter operators (which should be
// combined using OR). It also returns a boolean indicating if the operator should be negated.
func getComparisonOps(op parse.ComparisonOperator) (ops []ComparisonOp, negate bool) {
	switch op {
	case parse.ComparisonOperatorEqual:
		return []ComparisonOp{equal}, false
	case parse.ComparisonOperatorNotEqual:
		return []ComparisonOp{equal}, true
	case parse.ComparisonOperatorLessThan:
		return []ComparisonOp{lessThan}, false
	case parse.ComparisonOperatorLessThanOrEqual:
		return []ComparisonOp{lessThan, equal}, false
	case parse.ComparisonOperatorGreaterThan:
		return []ComparisonOp{greaterThan}, false
	case parse.ComparisonOperatorGreaterThanOrEqual:
		return []ComparisonOp{greaterThan, equal}, false
	}

	panic(fmt.Sprintf("unknown ast comparison operator: %v", op))
}

// ConvertArithmeticOp converts an arithmetic operator from the parser to the interpreter.
func convertArithmeticOp(op parse.ArithmeticOperator) ArithmeticOp {
	ar, ok := arithmeticOps[op]
	if !ok {
		panic(fmt.Sprintf("unknown ast arithmetic operator: %v", op))
	}
	return ar
}

// ConvertUnaryOp converts a unary operator from the parser to the interpreter.
func convertUnaryOp(op parse.UnaryOperator) UnaryOp {
	ar, ok := unaryOps[op]
	if !ok {
		panic(fmt.Sprintf("unknown ast unary operator: %v", op))
	}

	return ar
}

var arithmeticOps = map[parse.ArithmeticOperator]ArithmeticOp{
	parse.ArithmeticOperatorAdd:      add,
	parse.ArithmeticOperatorSubtract: sub,
	parse.ArithmeticOperatorMultiply: mul,
	parse.ArithmeticOperatorDivide:   div,
	parse.ArithmeticOperatorModulo:   mod,
	parse.ArithmeticOperatorConcat:   concat,
}

var unaryOps = map[parse.UnaryOperator]UnaryOp{
	parse.UnaryOperatorNot: not,
	parse.UnaryOperatorNeg: neg,
	parse.UnaryOperatorPos: pos,
}
