package tree

import (
	"errors"
	"fmt"
	"strings"

	sqlwriter "github.com/kwilteam/kwil-db/pkg/engine/sqlparser/tree/sql-writer"

	"github.com/cstockton/go-conv"
)

type Expression interface {
	isExpression() // private function to prevent external packages from implementing this interface
	ToSQL() string
	Accept(visitor Visitor) error
	joinable
}

type expressionBase struct{}

func (e *expressionBase) isExpression() {}

func (e *expressionBase) joinable() joinableStatus {
	return joinableStatusInvalid
}

func (e *expressionBase) ToSQL() string {
	panic("expressionBase: ToSQL() must be implemented by child")
}

func (e *expressionBase) Accept(visitor Visitor) error {
	return fmt.Errorf("expressionBase: Accept() must be implemented by child")
}

type Wrapped bool

type ExpressionLiteral struct {
	expressionBase
	Wrapped
	Value string
}

func (e *ExpressionLiteral) Accept(visitor Visitor) error {
	return visitor.VisitExpressionLiteral(e)
}

func (e *ExpressionLiteral) ToSQL() string {
	validateIsNonStringLiteral(e.Value)
	stmt := sqlwriter.NewWriter()

	if e.Wrapped {
		stmt.WrapParen()
	}

	stmt.WriteString(e.Value)

	return stmt.String()
}

func isStringLiteral(str string) bool {
	return str[0] == '\'' && str[len(str)-1] == '\''
}

func validateIsNonStringLiteral(str string) {
	if isStringLiteral(str) {
		return
	}

	if strings.EqualFold(str, "null") {
		return
	}
	if strings.EqualFold(str, "true") || strings.EqualFold(str, "false") {
		return
	}

	if len(strings.Split(str, ".")) > 1 {
		panic(fmt.Errorf("literal cannot be float.  received: %s", str))
	}

	_, err := conv.Int64(str)
	if err == nil {
		return
	}

	panic(fmt.Errorf("invalid literal: %s", str))
}

type ExpressionBindParameter struct {
	expressionBase
	Wrapped
	Parameter string
}

func (e *ExpressionBindParameter) Accept(visitor Visitor) error {
	return visitor.VisitExpressionBindParameter(e)
}

func (e *ExpressionBindParameter) ToSQL() string {
	if e.Parameter == "" {
		panic("ExpressionBindParameter: bind parameter cannot be empty")
	}
	if e.Parameter[0] != '$' && e.Parameter[0] != '@' {
		panic("ExpressionBindParameter: bind parameter must start with $")
	}

	stmt := sqlwriter.NewWriter()

	if e.Wrapped {
		stmt.WrapParen()
	}

	stmt.WriteString(e.Parameter)

	return stmt.String()
}

type ExpressionColumn struct {
	expressionBase
	Wrapped
	Table  string
	Column string
}

func (e *ExpressionColumn) Accept(visitor Visitor) error {
	return visitor.VisitExpressionColumn(e)
}

func (e *ExpressionColumn) ToSQL() string {
	stmt := sqlwriter.NewWriter()

	if e.Wrapped {
		stmt.WrapParen()
	}

	if e.Table != "" {
		stmt.WriteIdent(e.Table)
		stmt.Token.Period()
	}

	if e.Column == "" {
		panic("ExpressionColumn: column cannot be empty")
	}

	stmt.WriteIdent(e.Column)
	return stmt.String()
}

type ExpressionUnary struct {
	expressionBase
	Wrapped
	Operator UnaryOperator
	Operand  Expression
}

func (e *ExpressionUnary) Accept(visitor Visitor) error {
	return errors.Join(
		visitor.VisitExpressionUnary(e),
		accept(visitor, e.Operand),
	)
}

func (e *ExpressionUnary) ToSQL() string {
	stmt := sqlwriter.NewWriter()

	if e.Wrapped {
		stmt.WrapParen()
	}

	stmt.WriteString(e.Operator.String())
	stmt.WriteString(e.Operand.ToSQL())
	return stmt.String()
}

type ExpressionBinaryComparison struct {
	expressionBase
	Wrapped
	Left     Expression
	Operator BinaryOperator
	Right    Expression
}

func (e *ExpressionBinaryComparison) Accept(visitor Visitor) error {
	return errors.Join(
		visitor.VisitExpressionBinaryComparison(e),
		accept(visitor, e.Left),
		accept(visitor, e.Right),
	)
}

func (e *ExpressionBinaryComparison) ToSQL() string {
	stmt := sqlwriter.NewWriter()

	if e.Wrapped {
		stmt.WrapParen()
	}

	stmt.WriteString(e.Left.ToSQL())
	stmt.WriteString(e.Operator.String())
	stmt.WriteString(e.Right.ToSQL())
	return stmt.String()
}

type ExpressionFunction struct {
	expressionBase
	Wrapped
	Function SQLFunction
	Inputs   []Expression
	Distinct bool
}

func (e *ExpressionFunction) Accept(visitor Visitor) error {
	return errors.Join(
		visitor.VisitExpressionFunction(e),
		acceptMany(visitor, e.Inputs),
	)
}

func (e *ExpressionFunction) ToSQL() string {
	stmt := sqlwriter.NewWriter()

	if e.Wrapped {
		stmt.WrapParen()
	}

	var stringToWrite string

	if e.Distinct {
		exprFunc, ok := e.Function.(distinctable)
		if !ok {
			panic("ExpressionFunction: function '" + e.Function.Name() + "' does not support DISTINCT")
		}
		stringToWrite = exprFunc.stringDistinct(e.Inputs...)
	} else {
		stringToWrite = e.Function.String(e.Inputs...)
	}

	stmt.WriteString(stringToWrite)
	return stmt.String()
}

type ExpressionList struct {
	expressionBase
	Wrapped
	Expressions []Expression
}

func (e *ExpressionList) Accept(visitor Visitor) error {
	return errors.Join(
		visitor.VisitExpressionList(e),
		acceptMany(visitor, e.Expressions),
	)
}

func (e *ExpressionList) ToSQL() string {
	stmt := sqlwriter.NewWriter()

	if e.Wrapped {
		stmt.WrapParen()
	}

	if len(e.Expressions) == 0 {
		panic("ExpressionExpressionList: expressions cannot be empty")
	}

	stmt.WriteParenList(len(e.Expressions), func(i int) {
		stmt.WriteString(e.Expressions[i].ToSQL())
	})

	return stmt.String()
}

type ExpressionCollate struct {
	expressionBase
	Wrapped
	Expression Expression
	Collation  CollationType
}

func (e *ExpressionCollate) Accept(visitor Visitor) error {
	return errors.Join(
		visitor.VisitExpressionCollate(e),
		accept(visitor, e.Expression),
	)
}

func (e *ExpressionCollate) ToSQL() string {
	if e.Expression == nil {
		panic("ExpressionCollate: expression cannot be nil")
	}
	if e.Collation == "" {
		panic("ExpressionCollate: collation name cannot be empty")
	}

	stmt := sqlwriter.NewWriter()

	if e.Wrapped {
		stmt.WrapParen()
	}

	stmt.WriteString(e.Expression.ToSQL())
	stmt.Token.Collate()
	stmt.WriteString(e.Collation.String())
	return stmt.String()
}

type ExpressionStringCompare struct {
	expressionBase
	Wrapped
	Left     Expression
	Operator StringOperator
	Right    Expression
	Escape   Expression // can only be used with LIKE or NOT LIKE
}

func (e *ExpressionStringCompare) Accept(visitor Visitor) error {
	return errors.Join(
		visitor.VisitExpressionStringCompare(e),
		accept(visitor, e.Left),
		accept(visitor, e.Right),
		accept(visitor, e.Escape),
	)
}

func (e *ExpressionStringCompare) ToSQL() string {
	stmt := sqlwriter.NewWriter()

	if e.Wrapped {
		stmt.WrapParen()
	}

	stmt.WriteString(e.Left.ToSQL())
	stmt.WriteString(e.Operator.String())
	stmt.WriteString(e.Right.ToSQL())
	if e.Escape != nil {
		if !e.Operator.Escapable() {
			panic("ExpressionStringCompare: escape can only be used with LIKE or NOT LIKE")
		}

		stmt.Token.Escape()
		stmt.WriteString(e.Escape.ToSQL())
	}
	return stmt.String()
}

type ExpressionIsNull struct {
	expressionBase
	Wrapped
	Expression Expression
	IsNull     bool
}

func (e *ExpressionIsNull) Accept(visitor Visitor) error {
	return errors.Join(
		visitor.VisitExpressionIsNull(e),
		accept(visitor, e.Expression),
	)
}

func (e *ExpressionIsNull) ToSQL() string {
	if e.Expression == nil {
		panic("ExpressionIsNull: expression cannot be nil")
	}

	stmt := sqlwriter.NewWriter()

	if e.Wrapped {
		stmt.WrapParen()
	}

	stmt.WriteString(e.Expression.ToSQL())
	if e.IsNull {
		stmt.Token.Is().Null()
	} else {
		stmt.Token.Not().Null()
	}
	return stmt.String()
}

type ExpressionDistinct struct {
	expressionBase
	Wrapped
	Left  Expression
	Right Expression
	IsNot bool
}

func (e *ExpressionDistinct) Accept(visitor Visitor) error {
	return errors.Join(
		visitor.VisitExpressionDistinct(e),
		accept(visitor, e.Left),
		accept(visitor, e.Right),
	)
}

func (e *ExpressionDistinct) ToSQL() string {
	if e.Left == nil {
		panic("ExpressionDistinct: left expression cannot be nil")
	}
	if e.Right == nil {
		panic("ExpressionDistinct: right expression cannot be nil")
	}

	stmt := sqlwriter.NewWriter()

	if e.Wrapped {
		stmt.WrapParen()
	}

	stmt.WriteString(e.Left.ToSQL())
	stmt.Token.Is()
	if e.IsNot {
		stmt.Token.Not()
	}
	stmt.Token.Distinct().From()
	stmt.WriteString(e.Right.ToSQL())
	return stmt.String()
}

type ExpressionBetween struct {
	expressionBase
	Wrapped
	Expression Expression
	NotBetween bool
	Left       Expression
	Right      Expression
}

func (e *ExpressionBetween) Accept(visitor Visitor) error {
	return errors.Join(
		visitor.VisitExpressionBetween(e),
		accept(visitor, e.Expression),
		accept(visitor, e.Left),
		accept(visitor, e.Right),
	)
}

func (e *ExpressionBetween) ToSQL() string {
	if e.Expression == nil {
		panic("ExpressionBetween: expression cannot be nil")
	}
	if e.Left == nil {
		panic("ExpressionBetween: left expression cannot be nil")
	}
	if e.Right == nil {
		panic("ExpressionBetween: right expression cannot be nil")
	}

	stmt := sqlwriter.NewWriter()

	if e.Wrapped {
		stmt.WrapParen()
	}

	stmt.WriteString(e.Expression.ToSQL())
	if e.NotBetween {
		stmt.Token.Not()
	}
	stmt.Token.Between()
	stmt.WriteString(e.Left.ToSQL())
	stmt.Token.And()
	stmt.WriteString(e.Right.ToSQL())
	return stmt.String()
}

type ExpressionSelect struct {
	expressionBase
	Wrapped
	IsNot    bool
	IsExists bool
	Select   *SelectStmt
}

func (e *ExpressionSelect) Accept(visitor Visitor) error {
	return errors.Join(
		visitor.VisitExpressionSelect(e),
		accept(visitor, e.Select),
	)
}

func (e *ExpressionSelect) ToSQL() string {
	e.check()

	stmt := sqlwriter.NewWriter()

	if e.Wrapped {
		stmt.WrapParen()
	}

	if e.IsNot {
		stmt.Token.Not()
	}
	if e.IsExists {
		stmt.Token.Exists()
	}
	stmt.Token.Lparen()

	selectSql := e.Select.ToSQL()

	stmt.WriteString(selectSql)
	stmt.Token.Rparen()
	return stmt.String()
}

func (e *ExpressionSelect) check() {
	if e.Select == nil {
		panic("ExpressionSelect: select cannot be nil")
	}

	if e.IsNot {
		if !e.IsExists {
			panic("ExpressionSelect: NOT can only be used with EXISTS")
		}
	}
}

type ExpressionCase struct {
	expressionBase
	Wrapped
	CaseExpression Expression
	WhenThenPairs  [][2]Expression
	ElseExpression Expression
}

func (e *ExpressionCase) Accept(visitor Visitor) error {
	return errors.Join(
		visitor.VisitExpressionCase(e),
		accept(visitor, e.CaseExpression),
		func() error {
			for _, whenThen := range e.WhenThenPairs {
				err := accept(visitor, whenThen[0])
				if err != nil {
					return err
				}
				err = accept(visitor, whenThen[1])
				if err != nil {
					return err
				}
			}
			return nil
		}(),
		accept(visitor, e.ElseExpression),
	)
}

func (e *ExpressionCase) ToSQL() string {
	if len(e.WhenThenPairs) == 0 {
		panic("ExpressionCase: must contain at least 1 when-then pair")
	}

	stmt := sqlwriter.NewWriter()

	if e.Wrapped {
		stmt.WrapParen()
	}

	stmt.Token.Case()
	if e.CaseExpression != nil {
		stmt.WriteString(e.CaseExpression.ToSQL())
	}

	for _, whenThen := range e.WhenThenPairs {
		stmt.Token.When()
		stmt.WriteString(whenThen[0].ToSQL())
		stmt.Token.Then()
		stmt.WriteString(whenThen[1].ToSQL())
	}

	if e.ElseExpression != nil {
		stmt.Token.Else()
		stmt.WriteString(e.ElseExpression.ToSQL())
	}

	stmt.Token.End()
	return stmt.String()
}

type ExpressionArithmetic struct {
	expressionBase
	Wrapped
	Left     Expression
	Operator ArithmeticOperator
	Right    Expression
}

func (e *ExpressionArithmetic) Accept(visitor Visitor) error {
	return errors.Join(
		visitor.VisitExpressionArithmetic(e),
		accept(visitor, e.Left),
		accept(visitor, e.Right),
	)
}

func (e *ExpressionArithmetic) ToSQL() string {
	stmt := sqlwriter.NewWriter()

	if e.Wrapped {
		stmt.WrapParen()
	}

	// TODO: this should be removed once we handle * properly
	// right now, * is treated as an ArithmeticOperator, but it should be its own type
	// if * is passed, left and right are not defined

	stmt.WriteString(e.Left.ToSQL())
	stmt.WriteString(e.Operator.String())
	stmt.WriteString(e.Right.ToSQL())

	return stmt.String()
}
