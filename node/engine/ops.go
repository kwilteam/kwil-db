package engine

import (
	"fmt"

	"github.com/kwilteam/kwil-db/node/engine/parse"
)

type ComparisonOp uint8

const (
	EQUAL ComparisonOp = iota
	LESS_THAN
	GREATER_THAN
	IS
	IS_DISTINCT_FROM
)

type UnaryOp uint8

const (
	NOT UnaryOp = iota
	NEG
	POS
)

func (op UnaryOp) String() string {
	switch op {
	case NOT:
		return "NOT"
	case NEG:
		return "-"
	case POS:
		return "+"
	}

	panic(fmt.Sprintf("unknown unary operator: %d", op))
}

type ArithmeticOp uint8

const (
	ADD ArithmeticOp = iota
	SUB
	MUL
	DIV
	MOD
	EXP
	CONCAT
)

func (op ArithmeticOp) String() string {
	switch op {
	case ADD:
		return "+"
	case SUB:
		return "-"
	case MUL:
		return "*"
	case DIV:
		return "/"
	case MOD:
		return "%"
	case EXP:
		return "^"
	case CONCAT:
		return "||"
	}

	panic(fmt.Sprintf("unknown arithmetic operator: %d", op))
}

func (op ComparisonOp) String() string {
	switch op {
	case EQUAL:
		return "="
	case LESS_THAN:
		return "<"
	case GREATER_THAN:
		return ">"
	case IS:
		return "IS"
	case IS_DISTINCT_FROM:
		return "IS DISTINCT FROM"
	}

	panic(fmt.Sprintf("unknown comparison operator: %d", op))
}

// ConvertComparisonOps gets the engine comparison operators for the given parse operator.
// Since the interpreter has a restricted subset of comparison operators compared to the parser,
// it is possible that one parser operator maps to multiple interpreter operators (which should be
// combined using OR). It also returns a boolean indicating if the operator should be negated.
func ConvertComparisonOps(op parse.ComparisonOperator) (ops []ComparisonOp, negate bool) {
	switch op {
	case parse.ComparisonOperatorEqual:
		return []ComparisonOp{EQUAL}, false
	case parse.ComparisonOperatorNotEqual:
		return []ComparisonOp{EQUAL}, true
	case parse.ComparisonOperatorLessThan:
		return []ComparisonOp{LESS_THAN}, false
	case parse.ComparisonOperatorLessThanOrEqual:
		return []ComparisonOp{LESS_THAN, EQUAL}, false
	case parse.ComparisonOperatorGreaterThan:
		return []ComparisonOp{GREATER_THAN}, false
	case parse.ComparisonOperatorGreaterThanOrEqual:
		return []ComparisonOp{GREATER_THAN, EQUAL}, false
	}

	panic(fmt.Sprintf("unknown ast comparison operator: %v", op))
}

// ConvertArithmeticOp converts an arithmetic operator from the parser to the interpreter.
func ConvertArithmeticOp(op parse.ArithmeticOperator) ArithmeticOp {
	ar, ok := arithmeticOps[op]
	if !ok {
		panic(fmt.Sprintf("unknown ast arithmetic operator: %v", op))
	}
	return ar
}

// ConvertUnaryOp converts a unary operator from the parser to the interpreter.
func ConvertUnaryOp(op parse.UnaryOperator) UnaryOp {
	ar, ok := unaryOps[op]
	if !ok {
		panic(fmt.Sprintf("unknown ast unary operator: %v", op))
	}

	return ar
}

var arithmeticOps = map[parse.ArithmeticOperator]ArithmeticOp{
	parse.ArithmeticOperatorAdd:      ADD,
	parse.ArithmeticOperatorSubtract: SUB,
	parse.ArithmeticOperatorMultiply: MUL,
	parse.ArithmeticOperatorDivide:   DIV,
	parse.ArithmeticOperatorModulo:   MOD,
	parse.ArithmeticOperatorConcat:   CONCAT,
	parse.ArithmeticOperatorExponent: EXP,
}

var unaryOps = map[parse.UnaryOperator]UnaryOp{
	parse.UnaryOperatorNot: NOT,
	parse.UnaryOperatorNeg: NEG,
	parse.UnaryOperatorPos: POS,
}
