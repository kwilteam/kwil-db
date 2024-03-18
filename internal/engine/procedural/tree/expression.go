package tree

import (
	"fmt"
	"strings"

	"github.com/kwilteam/kwil-db/internal/engine/procedural/types"
)

type Expression interface {
	PGMarshaler
	expression()
	assigns(info *SystemInfo) (ReturnType, types.DataType, error)
}

type ExpressionArithmetic struct {
	Left     Expression
	Operator ArithmeticOperator
	Right    Expression
}

type ArithmeticOperator string

const (
	ArithmeticOperatorAdd ArithmeticOperator = "+"
	ArithmeticOperatorSub ArithmeticOperator = "-"
	ArithmeticOperatorMul ArithmeticOperator = "*"
	ArithmeticOperatorDiv ArithmeticOperator = "/"
	ArithmeticOperatorMod ArithmeticOperator = "%"
)

func (e *ExpressionArithmetic) expression() {}
func (e *ExpressionArithmetic) assigns(info *SystemInfo) (ReturnType, types.DataType, error) {
	// we read the types from both sides since we will support
	// several types of numbers (int, uint256, etc.)

	l, tp, err := e.Left.assigns(info)
	if err != nil {
		return 0, nil, err
	}

	if l != ReturnTypeValue {
		return 0, nil, fmt.Errorf("left side of arithmetic expression is not a value")
	}

	r, tp2, err := e.Right.assigns(info)
	if err != nil {
		return 0, nil, err
	}

	if r != ReturnTypeValue {
		return 0, nil, fmt.Errorf("right side of arithmetic expression is not a value")
	}

	if tp != tp2 {
		return 0, nil, fmt.Errorf("left and right side of arithmetic expression have different types")
	}

	return ReturnTypeValue, tp, nil
}

func (e *ExpressionArithmetic) MarshalPG(info *SystemInfo) (string, error) {
	str := strings.Builder{}
	l, err := e.Left.MarshalPG(info)
	if err != nil {
		return "", err
	}

	r, err := e.Right.MarshalPG(info)
	if err != nil {
		return "", err
	}

	str.WriteString(l)
	str.WriteString(" ")
	str.WriteString(string(e.Operator))
	str.WriteString(" ")
	str.WriteString(r)

	return str.String(), nil
}

type ExpressionBoolean struct {
	Left     Expression
	Operator ComparisonOperator
	Right    Expression
}

func (e *ExpressionBoolean) expression() {}

func (e *ExpressionBoolean) assigns(info *SystemInfo) (ReturnType, types.DataType, error) {
	l, _, err := e.Left.assigns(info)
	if err != nil {
		return 0, nil, err
	}

	r, _, err := e.Right.assigns(info)
	if err != nil {
		return 0, nil, err
	}

	if l != ReturnTypeValue || r != ReturnTypeValue {
		return 0, nil, fmt.Errorf("left and right side of boolean expression are not values")
	}

	return ReturnTypeValue, types.TypeBoolean, nil
}

func (e *ExpressionBoolean) MarshalPG(info *SystemInfo) (string, error) {
	str := strings.Builder{}
	l, err := e.Left.MarshalPG(info)
	if err != nil {
		return "", err
	}

	r, err := e.Right.MarshalPG(info)
	if err != nil {
		return "", err
	}

	str.WriteString(l)
	str.WriteString(" ")
	str.WriteString(string(e.Operator))
	str.WriteString(" ")
	str.WriteString(r)

	return str.String(), nil
}

type ComparisonOperator string

const (
	ComparisonOperatorEqual              ComparisonOperator = "="
	ComparisonOperatorNotEqual           ComparisonOperator = "!="
	ComparisonOperatorGreaterThan        ComparisonOperator = ">"
	ComparisonOperatorLessThan           ComparisonOperator = "<"
	ComparisonOperatorGreaterThanOrEqual ComparisonOperator = ">="
	ComparisonOperatorLessThanOrEqual    ComparisonOperator = "<="
)
