package tree

import (
	"fmt"
	"strconv"
	"strings"

	sqlwriter "github.com/kwilteam/kwil-db/parse/sql/tree/sql-writer"
)

/*
	'type hint' works as below:
	- it can directly be applied to literal, bind parameter, column, and function expression
    - case expression does not have type hint
	- other expressions can have type hint only when wrapped

    This logic is also enforced in the sqlparser.
*/

type Expression interface {
	isExpression() // private function to prevent external packages from implementing this interface
	ToSQL() string
	Accept(w Walker) error
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

func (e *expressionBase) Accept(w Walker) error {
	return fmt.Errorf("expressionBase: Accept() must be implemented by child")
}

// suffixTypeHint adds a type hint to `expression` if it is not empty
// NOTE: `::` is used to indicate type hint in SQL
func suffixTypeHint(expr string, typeHint string) string {
	if typeHint != "" {
		return fmt.Sprintf("%s ::%s", expr, typeHint)
	}
	return expr
}

type Wrapped bool

type ExpressionLiteral struct {
	expressionBase
	Wrapped
	Value    string
	TypeHint string // TODO: use predefined types?
}

func (e *ExpressionLiteral) Accept(w Walker) error {
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
	return suffixTypeHint(stmt.String(), e.TypeHint)
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

	_, err := strconv.ParseInt(str, 10, 64) // go-conv: conv.Int64(str)
	if err == nil {
		return
	}

	panic(fmt.Errorf("invalid literal: %s", str))
}

type ExpressionBindParameter struct {
	expressionBase
	Wrapped
	Parameter string
	TypeHint  string
}

func (e *ExpressionBindParameter) Accept(w Walker) error {
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

	return suffixTypeHint(stmt.String(), e.TypeHint)
}

type ExpressionColumn struct {
	expressionBase
	Wrapped
	Table    string
	Column   string
	TypeHint string
}

func (e *ExpressionColumn) Accept(w Walker) error {
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

	return suffixTypeHint(stmt.String(), e.TypeHint)
}

type ExpressionUnary struct {
	expressionBase
	Wrapped
	Operator UnaryOperator
	Operand  Expression
	TypeHint string
}

func (e *ExpressionUnary) Accept(w Walker) error {
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

	return suffixTypeHint(stmt.String(), e.TypeHint)
}

type ExpressionBinaryComparison struct {
	expressionBase
	Wrapped
	Left     Expression
	Operator BinaryOperator
	Right    Expression
	// NOTE: type hint only makes sense when wrapped,
	TypeHint string
}

func (e *ExpressionBinaryComparison) Accept(w Walker) error {
	return run(
		w.EnterExpressionBinaryComparison(e),
		accept(w, e.Left),
		accept(w, e.Right),
		w.ExitExpressionBinaryComparison(e),
	)
}

func (e *ExpressionBinaryComparison) ToSQL() string {
	if e.TypeHint != "" && !e.Wrapped {
		panic("ExpressionCollate: type hint need wrapped")
	}

	stmt := sqlwriter.NewWriter()

	if e.Wrapped {
		stmt.WrapParen()
	}

	stmt.WriteString(e.Left.ToSQL())
	stmt.WriteString(e.Operator.String())
	stmt.WriteString(e.Right.ToSQL())

	return suffixTypeHint(stmt.String(), e.TypeHint)
}

type ExpressionFunction struct {
	expressionBase
	Wrapped
	Function SQLFunction
	Inputs   []Expression
	Distinct bool
	TypeHint string
}

func (e *ExpressionFunction) Accept(w Walker) error {
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
		stringToWrite = e.Function.String(e.Inputs...)
	}

	stmt.WriteString(stringToWrite)

	return suffixTypeHint(stmt.String(), e.TypeHint)
}

type ExpressionList struct {
	expressionBase
	Wrapped
	Expressions []Expression
	TypeHint    string
}

func (e *ExpressionList) Accept(w Walker) error {
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

	return suffixTypeHint(stmt.String(), e.TypeHint)
}

type ExpressionCollate struct {
	expressionBase
	Wrapped
	Expression Expression
	Collation  CollationType
	// NOTE: type hint only makes sense when wrapped,
	TypeHint string
}

func (e *ExpressionCollate) Accept(w Walker) error {
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
	if e.TypeHint != "" && !e.Wrapped {
		panic("ExpressionCollate: type hint need wrapped")
	}

	stmt := sqlwriter.NewWriter()

	if e.Wrapped {
		stmt.WrapParen()
	}

	stmt.WriteString(e.Expression.ToSQL())
	stmt.Token.Collate()
	stmt.WriteString(e.Collation.String())

	return suffixTypeHint(stmt.String(), e.TypeHint)
}

type ExpressionStringCompare struct {
	expressionBase
	Wrapped
	Left     Expression
	Operator StringOperator
	Right    Expression
	Escape   Expression // can only be used with LIKE or NOT LIKE
	// NOTE: type hint only makes sense when wrapped,
	TypeHint string
}

func (e *ExpressionStringCompare) Accept(w Walker) error {
	return run(
		w.EnterExpressionStringCompare(e),
		accept(w, e.Left),
		accept(w, e.Right),
		accept(w, e.Escape),
		w.ExitExpressionStringCompare(e),
	)
}

func (e *ExpressionStringCompare) ToSQL() string {
	if e.TypeHint != "" && !e.Wrapped {
		panic("ExpressionStringCompare: type hint need wrapped")
	}

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

	return suffixTypeHint(stmt.String(), e.TypeHint)
}

type ExpressionIsNull struct {
	expressionBase
	Wrapped
	Expression Expression
	IsNull     bool
	// NOTE: type hint only makes sense when wrapped,
	TypeHint string
}

func (e *ExpressionIsNull) Accept(w Walker) error {
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
	if e.TypeHint != "" && !e.Wrapped {
		panic("ExpressionIsNull: type hint need wrapped")
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

	return suffixTypeHint(stmt.String(), e.TypeHint)
}

type ExpressionDistinct struct {
	expressionBase
	Wrapped
	Left  Expression
	Right Expression
	IsNot bool
	// NOTE: type hint only makes sense when wrapped,
	TypeHint string
}

func (e *ExpressionDistinct) Accept(w Walker) error {
	return run(
		w.EnterExpressionDistinct(e),
		accept(w, e.Left),
		accept(w, e.Right),
		w.ExitExpressionDistinct(e),
	)
}

func (e *ExpressionDistinct) ToSQL() string {
	// TODO: those validation should be done in analyzer
	if e.Left == nil {
		panic("ExpressionDistinct: left expression cannot be nil")
	}
	if e.Right == nil {
		panic("ExpressionDistinct: right expression cannot be nil")
	}
	if e.TypeHint != "" && !e.Wrapped {
		panic("ExpressionDistinct: type hint need wrapped")
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

	return suffixTypeHint(stmt.String(), e.TypeHint)
}

type ExpressionBetween struct {
	expressionBase
	Wrapped
	Expression Expression
	NotBetween bool
	Left       Expression
	Right      Expression
	// NOTE: type hint only makes sense when wrapped,
	TypeHint string
}

func (e *ExpressionBetween) Accept(w Walker) error {
	return run(
		w.EnterExpressionBetween(e),
		accept(w, e.Expression),
		accept(w, e.Left),
		accept(w, e.Right),
		w.ExitExpressionBetween(e),
	)
}

func (e *ExpressionBetween) ToSQL() string {
	// TODO: those validation should be done in analyzer
	if e.Expression == nil {
		panic("ExpressionBetween: expression cannot be nil")
	}
	if e.Left == nil {
		panic("ExpressionBetween: left expression cannot be nil")
	}
	if e.Right == nil {
		panic("ExpressionBetween: right expression cannot be nil")
	}
	if e.TypeHint != "" && !e.Wrapped {
		panic("ExpressionBetween: type hint need wrapped")
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

	return suffixTypeHint(stmt.String(), e.TypeHint)
}

type ExpressionSelect struct {
	expressionBase
	Wrapped
	IsNot    bool
	IsExists bool
	Select   *SelectStmt
	// NOTE: type hint only makes sense when wrapped,
	TypeHint string
}

func (e *ExpressionSelect) Accept(w Walker) error {
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

	return suffixTypeHint(stmt.String(), e.TypeHint)
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

	if e.TypeHint != "" && !e.Wrapped {
		panic("ExpressionSelect: type hint need wrapped")
	}
}

type ExpressionCase struct {
	expressionBase
	Wrapped
	CaseExpression Expression
	WhenThenPairs  [][2]Expression
	ElseExpression Expression
	// NOTE: type hint does not apply to the whole case expression
}

func (e *ExpressionCase) Accept(w Walker) error {
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
	expressionBase
	Wrapped
	Left     Expression
	Operator ArithmeticOperator
	Right    Expression
	// NOTE: type hint only makes sense when wrapped,
	TypeHint string
}

func (e *ExpressionArithmetic) Accept(w Walker) error {
	return run(
		w.EnterExpressionArithmetic(e),
		accept(w, e.Left),
		accept(w, e.Right),
		w.ExitExpressionArithmetic(e),
	)
}

func (e *ExpressionArithmetic) ToSQL() string {
	if e.TypeHint != "" && !e.Wrapped {
		panic("ExpressionArithmetic: type hint need wrapped")
	}

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

	return suffixTypeHint(stmt.String(), e.TypeHint)
}
