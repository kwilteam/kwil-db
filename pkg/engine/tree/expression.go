package tree

import (
	"fmt"
	"math"
	"reflect"

	"github.com/cstockton/go-conv"
)

type Expression interface {
	IsExpression()
	ToSQL() string
	Joinable() bool
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
func (e *ExpressionLiteral) ToSQL() string {
	dataType := reflect.TypeOf(e.Value)
	switch dataType.Kind() {
	case reflect.String:
		return fmt.Sprintf("'%s'", toString(e.Value))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Bool, reflect.Array, reflect.Slice:
		return toString(e.Value)
	case reflect.Float32, reflect.Float64:
		fl64, err := conv.Float64(e.Value)
		if err != nil {
			panic(fmt.Errorf("ExpressionLiteral: failed to convert float literal to float64: %w", err))
		}

		return toString(math.Floor(fl64))
	default:
		panic(fmt.Errorf("ExpressionLiteral: unsupported literal type: %s", dataType.Kind()))
	}
}

func (e *ExpressionLiteral) Joinable() bool {
	return false
}

func toString(val any) string {
	strVal, err := conv.String(val)
	if err != nil {
		panic(fmt.Errorf("failed to convert literal to string: %w", err))
	}
	return strVal
}

type ExpressionBindParameter struct {
	Parameter string
}

func (e *ExpressionBindParameter) Insert()       {}
func (e *ExpressionBindParameter) IsExpression() {}
func (e *ExpressionBindParameter) ToSQL() string {
	if e.Parameter == "" {
		panic("ExpressionBindParameter: bind parameter cannot be empty")
	}
	if e.Parameter[0] != '$' && e.Parameter[0] != '@' {
		panic("ExpressionBindParameter: bind parameter must start with $")
	}

	return e.Parameter
}

func (e *ExpressionBindParameter) Joinable() bool {
	return false
}

type ExpressionColumn struct {
	Table  string
	Column string
}

func (e *ExpressionColumn) Insert()       {}
func (e *ExpressionColumn) IsExpression() {}
func (e *ExpressionColumn) ToSQL() string {
	stmt := newSQLBuilder()
	stmt.Write(SPACE)

	if e.Table != "" {
		stmt.WriteIdent(e.Table)
		stmt.Write(PERIOD)
	}

	if e.Column == "" {
		panic("ExpressionColumn: column cannot be empty")
	}

	stmt.WriteIdent(e.Column)

	stmt.Write(SPACE)
	return stmt.String()
}

func (e *ExpressionColumn) Joinable() bool {
	return true
}

type ExpressionBinaryComparison struct {
	Left     Expression
	Operator BinaryOperator
	Right    Expression
}

func (e *ExpressionBinaryComparison) IsExpression() {}
func (e *ExpressionBinaryComparison) ToSQL() string {
	stmt := newSQLBuilder()
	stmt.Write(SPACE)
	stmt.WriteString(e.Left.ToSQL())
	stmt.Write(SPACE)
	stmt.WriteString(e.Operator.String())
	stmt.Write(SPACE)
	stmt.WriteString(e.Right.ToSQL())
	stmt.Write(SPACE)
	return stmt.String()
}

func (e *ExpressionBinaryComparison) Joinable() bool {
	if e.Operator != ComparisonOperatorEqual {
		return false
	}
	if e.Left.Joinable() && e.Right.Joinable() {
		return true
	}
	return false
}

type ExpressionFunction struct {
	Function SQLFunction
	Inputs   []Expression
}

func (e *ExpressionFunction) IsExpression() {}
func (e *ExpressionFunction) Insert()       {}
func (e *ExpressionFunction) ToSQL() string {
	stmt := newSQLBuilder()
	stmt.Write(SPACE)
	stmt.WriteString(e.Function.String(e.Inputs))
	stmt.Write(SPACE)
	return stmt.String()
}

func (e *ExpressionFunction) Joinable() bool {
	for _, input := range e.Inputs { // I don't know if this is actually safe, there is likely a way to hack this.
		if input.Joinable() {
			return true
		}
	}
	return false
}

type ExpressionExpressionList struct {
	Expressions []Expression
}

func (e *ExpressionExpressionList) IsExpression() {}
func (e *ExpressionExpressionList) Insert()       {}
func (e *ExpressionExpressionList) ToSQL() string {
	stmt := newSQLBuilder()
	stmt.Write(SPACE)
	stmt.Write(LPAREN)
	for i, expr := range e.Expressions {
		if i > 0 && i < len(e.Expressions) {
			stmt.Write(COMMA, SPACE)
		}
		stmt.WriteString(expr.ToSQL())
	}
	stmt.Write(RPAREN, SPACE)
	return stmt.String()
}

func (e *ExpressionExpressionList) Joinable() bool {
	return false
}

type ExpressionCollate struct {
	Expression Expression
	Collation  CollationType
}

func (e *ExpressionCollate) IsExpression() {}
func (e *ExpressionCollate) ToSQL() string {
	if e.Expression == nil {
		panic("ExpressionCollate: expression cannot be nil")
	}
	if e.Collation == "" {
		panic("ExpressionCollate: collation name cannot be empty")
	}

	stmt := newSQLBuilder()
	stmt.Write(SPACE)
	stmt.WriteString(e.Expression.ToSQL())
	stmt.Write(SPACE, COLLATE, SPACE)
	stmt.WriteString(e.Collation.String())
	stmt.Write(SPACE)
	return stmt.String()
}

func (e *ExpressionCollate) Joinable() bool {
	return false
}

type ExpressionStringCompare struct {
	Left     Expression
	Operator StringOperator
	Right    Expression
	Escape   Expression // can only be used with LIKE or NOT LIKE
}

func (e *ExpressionStringCompare) IsExpression() {}
func (e *ExpressionStringCompare) ToSQL() string {
	stmt := newSQLBuilder()
	stmt.Write(SPACE)
	stmt.WriteString(e.Left.ToSQL())
	stmt.Write(SPACE)
	stmt.WriteString(e.Operator.String())
	stmt.Write(SPACE)
	stmt.WriteString(e.Right.ToSQL())
	if e.Escape != nil {
		if !e.Operator.Escapable() {
			panic("ExpressionStringCompare: escape can only be used with LIKE or NOT LIKE")
		}

		stmt.Write(SPACE, ESCAPE, SPACE)
		stmt.WriteString(e.Escape.ToSQL())
	}
	stmt.Write(SPACE)
	return stmt.String()
}

func (e *ExpressionStringCompare) Joinable() bool {
	return false
}

type ExpressionIsNull struct {
	Expression Expression
	IsNull     bool
}

func (e *ExpressionIsNull) IsExpression() {}
func (e *ExpressionIsNull) ToSQL() string {
	if e.Expression == nil {
		panic("ExpressionIsNull: expression cannot be nil")
	}

	stmt := newSQLBuilder()
	stmt.Write(SPACE)
	stmt.WriteString(e.Expression.ToSQL())
	stmt.Write(SPACE)
	if e.IsNull {
		stmt.Write(IS, SPACE, NULL)
	} else {
		stmt.Write(IS, SPACE, NOT, SPACE, NULL)
	}
	stmt.Write(SPACE)
	return stmt.String()
}

func (e *ExpressionIsNull) Joinable() bool {
	return false
}

type ExpressionDistinct struct {
	Left     Expression
	Right    Expression
	IsNot    bool
	Distinct bool
}

func (e *ExpressionDistinct) IsExpression() {}
func (e *ExpressionDistinct) ToSQL() string {
	if e.Left == nil {
		panic("ExpressionDistinct: left expression cannot be nil")
	}
	if e.Right == nil {
		panic("ExpressionDistinct: right expression cannot be nil")
	}

	stmt := newSQLBuilder()
	stmt.Write(SPACE)
	stmt.WriteString(e.Left.ToSQL())
	stmt.Write(SPACE, IS, SPACE)
	if e.IsNot {
		stmt.Write(NOT, SPACE)
	}
	if e.Distinct {
		stmt.Write(DISTINCT, SPACE, FROM, SPACE)
	}
	stmt.WriteString(e.Right.ToSQL())
	stmt.Write(SPACE)
	return stmt.String()
}

func (e *ExpressionDistinct) Joinable() bool {
	return false
}

type ExpressionBetween struct {
	Expression Expression
	NotBetween bool
	Left       Expression
	Right      Expression
}

func (e *ExpressionBetween) IsExpression() {}
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

	stmt := newSQLBuilder()
	stmt.Write(SPACE)
	stmt.WriteString(e.Expression.ToSQL())
	stmt.Write(SPACE)
	if e.NotBetween {
		stmt.Write(NOT, SPACE)
	}
	stmt.Write(BETWEEN, SPACE)
	stmt.WriteString(e.Left.ToSQL())
	stmt.Write(SPACE, AND, SPACE)
	stmt.WriteString(e.Right.ToSQL())
	stmt.Write(SPACE)
	return stmt.String()
}

func (e *ExpressionBetween) Joinable() bool {
	if e.Expression.Joinable() {
		return true
	}
	return false
}

type ExpressionIn struct {
	Expression    Expression
	NotIn         bool
	InExpressions []Expression
}

func (e *ExpressionIn) IsExpression() {}
func (e *ExpressionIn) ToSQL() string {
	if e.Expression == nil {
		panic("ExpressionIn: expression cannot be nil")
	}
	if len(e.InExpressions) == 0 {
		panic("ExpressionIn: expressions cannot be empty")
	}

	stmt := newSQLBuilder()
	stmt.Write(SPACE)
	stmt.WriteString(e.Expression.ToSQL())
	stmt.Write(SPACE)
	if e.NotIn {
		stmt.Write(NOT, SPACE)
	}
	stmt.Write(IN, SPACE)
	stmt.Write(LPAREN)
	for i, expr := range e.InExpressions {
		if i > 0 && i < len(e.InExpressions) {
			stmt.Write(COMMA, SPACE)
		}
		stmt.WriteString(expr.ToSQL())
	}
	stmt.Write(RPAREN, SPACE)
	return stmt.String()
}

func (e *ExpressionIn) Joinable() bool {
	return false
}

type ExpressionSelect struct {
	IsNot    bool
	IsExists bool
	Select   *Select
}

func (e *ExpressionSelect) Insert()       {}
func (e *ExpressionSelect) IsExpression() {}
func (e *ExpressionSelect) ToSQL() string {
	if e.Select == nil {
		panic("ExpressionSelect: select cannot be nil")
	}

	stmt := newSQLBuilder()
	stmt.Write(SPACE)
	if e.IsNot {
		stmt.Write(NOT, SPACE)
		if !e.IsExists {
			panic("ExpressionSelect: NOT can only be used with EXISTS")
		}
	}
	if e.IsExists {
		stmt.Write(EXISTS, SPACE)
	}
	stmt.Write(LPAREN)

	selectSql, err := e.Select.ToSQL()
	if err != nil {
		panic(fmt.Errorf("ExpressionSelect: failed to convert select to SQL: %w", err))
	}
	stmt.WriteString(selectSql)
	stmt.Write(RPAREN, SPACE)
	return stmt.String()
}

func (e *ExpressionSelect) Joinable() bool {
	return false
}

type ExpressionCase struct {
	CaseExpression Expression
	WhenThenPairs  [][2]Expression
	ElseExpression Expression
}

func (e *ExpressionCase) IsExpression() {}
func (e *ExpressionCase) ToSQL() string {
	if len(e.WhenThenPairs) == 0 {
		panic("ExpressionCase: must contain at least 1 when-then pair")
	}

	stmt := newSQLBuilder()
	stmt.Write(SPACE, CASE, SPACE)
	if e.CaseExpression != nil {
		stmt.WriteString(e.CaseExpression.ToSQL())
		stmt.Write(SPACE)
	}

	for _, whenThen := range e.WhenThenPairs {
		stmt.Write(WHEN, SPACE)
		stmt.WriteString(whenThen[0].ToSQL())
		stmt.Write(SPACE, THEN, SPACE)
		stmt.WriteString(whenThen[1].ToSQL())
		stmt.Write(SPACE)
	}

	if e.ElseExpression != nil {
		stmt.Write(ELSE, SPACE)
		stmt.WriteString(e.ElseExpression.ToSQL())
		stmt.Write(SPACE)
	}

	stmt.Write(END, SPACE)
	return stmt.String()
}

func (e *ExpressionCase) Joinable() bool {
	return false
}
