package tree

import (
	"fmt"
	"math"
	"reflect"

	sqlwriter "github.com/kwilteam/kwil-db/pkg/engine2/tree/sql-writer"

	"github.com/cstockton/go-conv"
)

type Expression interface {
	isExpression() // private function to prevent external packages from implementing this interface
	ToSQL() string
	Joinable() bool
}

/*
// Literal, BindParameter, ExpressionSelect
type InsertExpression interface {
	Expression
	// Insert doesn't do anything, but it makes it clear to package consumers what can be used in an INSERT statement
	Insert()
}*/

type ExpressionLiteral struct {
	Value interface{}
}

func (e *ExpressionLiteral) Insert()       {}
func (e *ExpressionLiteral) isExpression() {}
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
func (e *ExpressionBindParameter) isExpression() {}
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
func (e *ExpressionColumn) isExpression() {}
func (e *ExpressionColumn) ToSQL() string {
	stmt := sqlwriter.NewWriter()
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

func (e *ExpressionColumn) Joinable() bool {
	return true
}

type ExpressionBinaryComparison struct {
	Left     Expression
	Operator BinaryOperator
	Right    Expression
}

func (e *ExpressionBinaryComparison) isExpression() {}
func (e *ExpressionBinaryComparison) ToSQL() string {
	stmt := sqlwriter.NewWriter()
	stmt.WriteString(e.Left.ToSQL())
	stmt.WriteString(e.Operator.String())
	stmt.WriteString(e.Right.ToSQL())
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

func (e *ExpressionFunction) isExpression() {}
func (e *ExpressionFunction) Insert()       {}
func (e *ExpressionFunction) ToSQL() string {
	stmt := sqlwriter.NewWriter()
	stmt.WriteString(e.Function.String(e.Inputs))
	return stmt.String()
}

func (e *ExpressionFunction) Joinable() bool {
	return false
}

type ExpressionList struct {
	Expressions []Expression
}

func (e *ExpressionList) isExpression() {}
func (e *ExpressionList) Insert()       {}
func (e *ExpressionList) ToSQL() string {
	stmt := sqlwriter.NewWriter()

	if len(e.Expressions) == 0 {
		panic("ExpressionExpressionList: expressions cannot be empty")
	}

	stmt.WriteParenList(len(e.Expressions), func(i int) {
		stmt.WriteString(e.Expressions[i].ToSQL())
	})

	return stmt.String()
}

func (e *ExpressionList) Joinable() bool {
	return false
}

type ExpressionCollate struct {
	Expression Expression
	Collation  CollationType
}

func (e *ExpressionCollate) isExpression() {}
func (e *ExpressionCollate) ToSQL() string {
	if e.Expression == nil {
		panic("ExpressionCollate: expression cannot be nil")
	}
	if e.Collation == "" {
		panic("ExpressionCollate: collation name cannot be empty")
	}

	stmt := sqlwriter.NewWriter()
	stmt.WriteString(e.Expression.ToSQL())
	stmt.Token.Collate()
	stmt.WriteString(e.Collation.String())
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

func (e *ExpressionStringCompare) isExpression() {}
func (e *ExpressionStringCompare) ToSQL() string {
	stmt := sqlwriter.NewWriter()
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

func (e *ExpressionStringCompare) Joinable() bool {
	return false
}

type ExpressionIsNull struct {
	Expression Expression
	IsNull     bool
}

func (e *ExpressionIsNull) isExpression() {}
func (e *ExpressionIsNull) ToSQL() string {
	if e.Expression == nil {
		panic("ExpressionIsNull: expression cannot be nil")
	}

	stmt := sqlwriter.NewWriter()
	stmt.WriteString(e.Expression.ToSQL())
	if e.IsNull {
		stmt.Token.Is().Null()
	} else {
		stmt.Token.Is().Not().Null()
	}
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

func (e *ExpressionDistinct) isExpression() {}
func (e *ExpressionDistinct) ToSQL() string {
	if e.Left == nil {
		panic("ExpressionDistinct: left expression cannot be nil")
	}
	if e.Right == nil {
		panic("ExpressionDistinct: right expression cannot be nil")
	}

	stmt := sqlwriter.NewWriter()
	stmt.WriteString(e.Left.ToSQL())
	stmt.Token.Is()
	if e.IsNot {
		stmt.Token.Not()
	}
	if e.Distinct {
		stmt.Token.Distinct().From()
	}
	stmt.WriteString(e.Right.ToSQL())
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

func (e *ExpressionBetween) isExpression() {}
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

func (e *ExpressionBetween) Joinable() bool {
	return e.Expression.Joinable()
}

type ExpressionIn struct {
	Expression    Expression
	NotIn         bool
	InExpressions []Expression
}

func (e *ExpressionIn) isExpression() {}
func (e *ExpressionIn) ToSQL() string {
	if e.Expression == nil {
		panic("ExpressionIn: expression cannot be nil")
	}
	if len(e.InExpressions) == 0 {
		panic("ExpressionIn: expressions cannot be empty")
	}

	stmt := sqlwriter.NewWriter()
	stmt.WriteString(e.Expression.ToSQL())
	if e.NotIn {
		stmt.Token.Not()
	}
	stmt.Token.In()

	stmt.WriteParenList(len(e.InExpressions), func(i int) {
		stmt.WriteString(e.InExpressions[i].ToSQL())
	})

	return stmt.String()
}

func (e *ExpressionIn) Joinable() bool {
	return false
}

type ExpressionSelect struct {
	IsNot    bool
	IsExists bool
	Select   *SelectStmt
}

func (e *ExpressionSelect) Insert()       {}
func (e *ExpressionSelect) isExpression() {}
func (e *ExpressionSelect) ToSQL() string {
	e.check()

	stmt := sqlwriter.NewWriter()
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

func (e *ExpressionSelect) Joinable() bool {
	return false
}

type ExpressionCase struct {
	CaseExpression Expression
	WhenThenPairs  [][2]Expression
	ElseExpression Expression
}

func (e *ExpressionCase) isExpression() {}
func (e *ExpressionCase) ToSQL() string {
	if len(e.WhenThenPairs) == 0 {
		panic("ExpressionCase: must contain at least 1 when-then pair")
	}

	stmt := sqlwriter.NewWriter()
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

func (e *ExpressionCase) Joinable() bool {
	return false
}
