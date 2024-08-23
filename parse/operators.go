package parse

import (
	"fmt"

	"github.com/kwilteam/kwil-db/parse/common"
)

var arithmeticOps = map[ArithmeticOperator]common.ArithmeticOp{
	ArithmeticOperatorAdd:      common.Add,
	ArithmeticOperatorSubtract: common.Sub,
	ArithmeticOperatorMultiply: common.Mul,
	ArithmeticOperatorDivide:   common.Div,
	ArithmeticOperatorModulo:   common.Mod,
	ArithmeticOperatorConcat:   common.Concat,
}

var unaryOps = map[UnaryOperator]common.UnaryOp{
	UnaryOperatorNot: common.Not,
	UnaryOperatorNeg: common.Neg,
	UnaryOperatorPos: common.Pos,
}

// ConvertArithmeticOp converts an arithmetic operator from the parser to the interpreter.
func ConvertArithmeticOp(op ArithmeticOperator) common.ArithmeticOp {
	ar, ok := arithmeticOps[op]
	if !ok {
		panic(fmt.Sprintf("unknown ast arithmetic operator: %v", op))
	}
	return ar
}

// ConvertUnaryOp converts a unary operator from the parser to the interpreter.
func ConvertUnaryOp(op UnaryOperator) common.UnaryOp {
	ar, ok := unaryOps[op]
	if !ok {
		panic(fmt.Sprintf("unknown ast unary operator: %v", op))
	}

	return ar
}

// GetComparisonOps gets the comparison operators for the given operator.
// Since the interpreter has a restricted subset of comparison operators compared to the parser,
// it is possible that one parser operator maps to multiple interpreter operators (which should be
// combined using OR). It also returns a boolean indicating if the operator should be negated.
func GetComparisonOps(op ComparisonOperator) (ops []common.ComparisonOp, negate bool) {
	switch op {
	case ComparisonOperatorEqual:
		return []common.ComparisonOp{common.Equal}, false
	case ComparisonOperatorNotEqual:
		return []common.ComparisonOp{common.Equal}, true
	case ComparisonOperatorLessThan:
		return []common.ComparisonOp{common.LessThan}, false
	case ComparisonOperatorLessThanOrEqual:
		return []common.ComparisonOp{common.LessThan, common.Equal}, false
	case ComparisonOperatorGreaterThan:
		return []common.ComparisonOp{common.GreaterThan}, false
	case ComparisonOperatorGreaterThanOrEqual:
		return []common.ComparisonOp{common.GreaterThan, common.Equal}, false
	}

	panic(fmt.Sprintf("unknown ast comparison operator: %v", op))
}
