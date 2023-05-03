package tree

import (
	"kwil/pkg/engine/tree/types"

	"github.com/doug-martin/goqu/v9"
)

type Expression interface {
	IsExpression()
	// ToSqlStruct converts the expression to a struct that can be used to generate SQL
	ToSqlStruct() any
}

// Literal, BindParameter, ExpressionSelect
type InsertExpression interface {
	Expression
	// Insert doesn't do anything, but it makes it clear to package consumers what can be used in an INSERT statement
	Insert()
}

type ExpressionLiteral struct {
	Value interface{}
}

func (e *ExpressionLiteral) Insert()       {}
func (e *ExpressionLiteral) IsExpression() {}
func (e *ExpressionLiteral) ToSqlStruct() any {
	return e.Value
}

type ExpressionBindParameter struct {
	Parameter string
}

func (e *ExpressionBindParameter) Insert()       {}
func (e *ExpressionBindParameter) IsExpression() {}
func (e *ExpressionBindParameter) ToSqlStruct() any {
	return goqu.L(e.Parameter)
}

type ExpressionBinaryComparison struct {
	Left     Expression
	Operator types.BinaryOperator
	Right    Expression
}

func (e *ExpressionBinaryComparison) IsExpression() {}
func (e *ExpressionBinaryComparison) ToSqlStruct() any {
	return nil
}

type ExpressionFunction struct {
	Function Function
}

func (e *ExpressionFunction) IsExpression() {}

type ExpressionExpressionList struct {
	Expressions []Expression
}

func (e *ExpressionExpressionList) IsExpression() {}

type ExpressionCollate struct {
	Expression Expression
	Collation  types.CollationType
}

func (e *ExpressionCollate) IsExpression() {}

type ExpressionStringCompare struct {
	Left     Expression
	Operator types.StringOperator
	Right    Expression
	Escape   Expression // can only be used with LIKE or NOT LIKE
}

func (e *ExpressionStringCompare) IsExpression() {}

type ExpressionIsNull struct {
	Expression Expression
	IsNull     bool
}

func (e *ExpressionIsNull) IsExpression() {}

type ExpressionDistinct struct {
	Left     Expression
	Right    Expression
	IsNot    bool
	Distinct bool
}

func (e *ExpressionDistinct) IsExpression() {}

type ExpressionBetween struct {
	Expression Expression
	IsBetween  bool
	Left       Expression
	Right      Expression
}

func (e *ExpressionBetween) IsExpression() {}

type ExpressionIn struct {
	Expression  Expression
	IsIn        bool
	Expressions []Expression
}

func (e *ExpressionIn) IsExpression() {}

type ExpressionSelect struct {
	IsNot    bool
	IsExists bool
	Select   *SelectStatement
}

func (e *ExpressionSelect) Insert()       {}
func (e *ExpressionSelect) IsExpression() {}

type ExpressionCase struct {
	CaseExpression Expression
	WhenThenPairs  [][2]Expression
	ElseExpression Expression
}

func (e *ExpressionCase) IsExpression() {}
