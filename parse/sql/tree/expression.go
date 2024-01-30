package tree

import (
	"fmt"
	"strings"

	sqlwriter "github.com/kwilteam/kwil-db/parse/sql/tree/sql-writer"

	"github.com/cstockton/go-conv"
)

type Expression interface {
	isExpression() // private function to prevent external packages from implementing this interface
	ToSQL() string
	Walk(w AstWalker) error
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

func (e *expressionBase) Walk(w AstWalker) error {
	return fmt.Errorf("expressionBase: Walk() must be implemented by child")
}

type Wrapped bool

type ExpressionLiteral struct {
	node

	expressionBase
	Wrapped
	Value string
	// TODO: add type
}

func (e *ExpressionLiteral) Accept(v AstVisitor) any {
	return v.VisitExpressionLiteral(e)
}

func (e *ExpressionLiteral) Walk(w AstWalker) error {
	return run(
		w.EnterExpressionLiteral(e),
		w.ExitExpressionLiteral(e),
	)
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
	node

	expressionBase
	Wrapped
	Parameter string
}

func (e *ExpressionBindParameter) Accept(v AstVisitor) any {
	return v.VisitExpressionBindParameter(e)
}

func (e *ExpressionBindParameter) Walk(w AstWalker) error {
	return run(
		w.EnterExpressionBindParameter(e),
		w.ExitExpressionBindParameter(e),
	)
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
	node

	expressionBase
	Wrapped
	Table  string
	Column string
}

func (e *ExpressionColumn) Accept(v AstVisitor) any {
	return v.VisitExpressionColumn(e)
}

func (e *ExpressionColumn) Walk(w AstWalker) error {
	return run(
		w.EnterExpressionColumn(e),
		w.ExitExpressionColumn(e),
	)
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
	node

	expressionBase
	Wrapped
	Operator UnaryOperator
	Operand  Expression
}

func (e *ExpressionUnary) Accept(v AstVisitor) any {
	return v.VisitExpressionUnary(e)
}

func (e *ExpressionUnary) Walk(w AstWalker) error {
	return run(
		w.EnterExpressionUnary(e),
		w.ExitExpressionUnary(e),
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
	node

	expressionBase
	Wrapped
	Left     Expression
	Operator BinaryOperator
	Right    Expression
}

func (e *ExpressionBinaryComparison) Accept(v AstVisitor) any {
	return v.VisitExpressionBinaryComparison(e)
}

func (e *ExpressionBinaryComparison) Walk(w AstWalker) error {
	return run(
		w.EnterExpressionBinaryComparison(e),
		accept(w, e.Left),
		accept(w, e.Right),
		w.ExitExpressionBinaryComparison(e),
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
	node

	expressionBase
	Wrapped
	Function SQLFunction
	Inputs   []Expression
	Distinct bool
}

func (e *ExpressionFunction) Accept(v AstVisitor) any {
	return v.VisitExpressionFunction(e)
}

func (e *ExpressionFunction) Walk(w AstWalker) error {
	return run(
		w.EnterExpressionFunction(e),
		acceptMany(w, e.Inputs),
		w.ExitExpressionFunction(e),
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
		stringToWrite = e.Function.ToString(e.Inputs...)
	}

	stmt.WriteString(stringToWrite)
	return stmt.String()
}

type ExpressionList struct {
	node

	expressionBase
	Wrapped
	Expressions []Expression
}

func (e *ExpressionList) Accept(v AstVisitor) any {
	return v.VisitExpressionList(e)
}

func (e *ExpressionList) Walk(w AstWalker) error {
	return run(
		w.EnterExpressionList(e),
		acceptMany(w, e.Expressions),
		w.ExitExpressionList(e),
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
	node

	expressionBase
	Wrapped
	Expression Expression
	Collation  CollationType
}

func (e *ExpressionCollate) Accept(v AstVisitor) any {
	return v.VisitExpressionCollate(e)
}

func (e *ExpressionCollate) Walk(w AstWalker) error {
	return run(
		w.EnterExpressionCollate(e),
		accept(w, e.Expression),
		w.ExitExpressionCollate(e),
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
	node

	expressionBase
	Wrapped
	Left     Expression
	Operator StringOperator
	Right    Expression
	Escape   Expression // can only be used with LIKE or NOT LIKE
}

func (e *ExpressionStringCompare) Accept(v AstVisitor) any {
	return v.VisitExpressionStringCompare(e)
}

func (e *ExpressionStringCompare) Walk(w AstWalker) error {
	return run(
		w.EnterExpressionStringCompare(e),
		accept(w, e.Left),
		accept(w, e.Right),
		accept(w, e.Escape),
		w.ExitExpressionStringCompare(e),
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
	node

	expressionBase
	Wrapped
	Expression Expression
	IsNull     bool
}

func (e *ExpressionIsNull) Walk(w AstWalker) error {
	return run(
		w.EnterExpressionIsNull(e),
		accept(w, e.Expression),
		w.ExitExpressionIsNull(e),
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
	node

	expressionBase
	Wrapped
	Left  Expression
	Right Expression
	IsNot bool
}

func (e *ExpressionDistinct) Accept(v AstVisitor) any {
	return v.VisitExpressionDistinct(e)
}

func (e *ExpressionDistinct) Walk(w AstWalker) error {
	return run(
		w.EnterExpressionDistinct(e),
		accept(w, e.Left),
		accept(w, e.Right),
		w.ExitExpressionDistinct(e),
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
	node

	expressionBase
	Wrapped
	Expression Expression
	NotBetween bool
	Left       Expression
	Right      Expression
}

func (e *ExpressionBetween) Accept(v AstVisitor) any {
	return v.VisitExpressionBetween(e)
}

func (e *ExpressionBetween) Walk(w AstWalker) error {
	return run(
		w.EnterExpressionBetween(e),
		accept(w, e.Expression),
		accept(w, e.Left),
		accept(w, e.Right),
		w.ExitExpressionBetween(e),
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
	node

	expressionBase
	Wrapped
	IsNot    bool
	IsExists bool
	Select   *SelectStmt
}

func (e *ExpressionSelect) Accept(v AstVisitor) any {
	return v.VisitExpressionSelect(e)
}

func (e *ExpressionSelect) Walk(w AstWalker) error {
	return run(
		w.EnterExpressionSelect(e),
		accept(w, e.Select),
		w.ExitExpressionSelect(e),
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
	node

	expressionBase
	Wrapped
	CaseExpression Expression
	WhenThenPairs  [][2]Expression
	ElseExpression Expression
}

func (e *ExpressionCase) Accept(v AstVisitor) any {
	return v.VisitExpressionCase(e)
}

func (e *ExpressionCase) Walk(w AstWalker) error {
	return run(
		w.EnterExpressionCase(e),
		accept(w, e.CaseExpression),
		func() error {
			for _, whenThen := range e.WhenThenPairs {
				err := accept(w, whenThen[0])
				if err != nil {
					return err
				}
				err = accept(w, whenThen[1])
				if err != nil {
					return err
				}
			}
			return nil
		}(),
		accept(w, e.ElseExpression),
		w.ExitExpressionCase(e),
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
	node

	expressionBase
	Wrapped
	Left     Expression
	Operator ArithmeticOperator
	Right    Expression
}

func (e *ExpressionArithmetic) Accept(v AstVisitor) any {
	return v.VisitExpressionArithmetic(e)
}

func (e *ExpressionArithmetic) Walk(w AstWalker) error {
	return run(
		w.EnterExpressionArithmetic(e),
		accept(w, e.Left),
		accept(w, e.Right),
		w.ExitExpressionArithmetic(e),
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
