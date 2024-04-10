package tree

import (
	"encoding/hex"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/kwilteam/kwil-db/core/types"
	sqlwriter "github.com/kwilteam/kwil-db/internal/parse/sql/tree/sql-writer"
)

type expressionBase struct{}

func (e *expressionBase) expression() {}

func (e *expressionBase) joinable() joinableStatus {
	return joinableStatusInvalid
}

func (e *expressionBase) ToSQL() string {
	panic("expressionBase: ToSQL() must be implemented by child")
}

func (e *expressionBase) Walk(w AstListener) error {
	return fmt.Errorf("expressionBase: Walk() must be implemented by child")
}

type Wrapped bool

type ExpressionTextLiteral struct {
	node
	expressionBase
	Wrapped
	Value    string
	TypeCast *types.DataType
}

func (e *ExpressionTextLiteral) Accept(v AstVisitor) any {
	return v.VisitExpressionTextLiteral(e)
}

func (e *ExpressionTextLiteral) Walk(w AstListener) error {
	return run(
		w.EnterExpressionTextLiteral(e),
		w.ExitExpressionTextLiteral(e),
	)
}

func (e *ExpressionTextLiteral) ToSQL() string {
	stmt := sqlwriter.NewWriter()

	if e.Wrapped {
		stmt.WrapParen()
	}

	stmt.WriteString(fmt.Sprintf("'%s'", e.Value))
	return castSuffix(stmt.String(), e.TypeCast)
}

func castSuffix(stmt string, typeCast *types.DataType) string {
	if typeCast != nil {
		stmt += "::"

		tp, err := typeCast.PGString()
		if err != nil {
			panic(err)
		}

		stmt += tp
	}

	return stmt
}

type ExpressionNumericLiteral struct {
	node
	expressionBase
	Wrapped
	Value    int64
	TypeCast *types.DataType
}

func (e *ExpressionNumericLiteral) Accept(v AstVisitor) any {
	return v.VisitExpressionNumericLiteral(e)
}

func (e *ExpressionNumericLiteral) Walk(w AstListener) error {
	return run(
		w.EnterExpressionNumericLiteral(e),
		w.ExitExpressionNumericLiteral(e),
	)
}

func (e *ExpressionNumericLiteral) ToSQL() string {
	stmt := sqlwriter.NewWriter()

	if e.Wrapped {
		stmt.WrapParen()
	}

	stmt.WriteString(strconv.FormatInt(e.Value, 10))
	return castSuffix(stmt.String(), e.TypeCast)
}

type ExpressionBooleanLiteral struct {
	node
	expressionBase
	Wrapped
	Value    bool
	TypeCast *types.DataType
}

func (e *ExpressionBooleanLiteral) Accept(v AstVisitor) any {
	return v.VisitExpressionBooleanLiteral(e)
}

func (e *ExpressionBooleanLiteral) Walk(w AstListener) error {
	return run(
		w.EnterExpressionBooleanLiteral(e),
		w.ExitExpressionBooleanLiteral(e),
	)
}

func (e *ExpressionBooleanLiteral) ToSQL() string {
	stmt := sqlwriter.NewWriter()

	if e.Wrapped {
		stmt.WrapParen()
	}

	stmt.WriteString(strconv.FormatBool(e.Value))
	return castSuffix(stmt.String(), e.TypeCast)
}

type ExpressionNullLiteral struct {
	node
	expressionBase
	Wrapped
	TypeCast *types.DataType
}

func (e *ExpressionNullLiteral) Accept(v AstVisitor) any {
	return v.VisitExpressionNullLiteral(e)
}

func (e *ExpressionNullLiteral) Walk(w AstListener) error {
	return run(
		w.EnterExpressionNullLiteral(e),
		w.ExitExpressionNullLiteral(e),
	)
}

func (e *ExpressionNullLiteral) ToSQL() string {
	stmt := sqlwriter.NewWriter()

	if e.Wrapped {
		stmt.WrapParen()
	}

	stmt.WriteString("NULL")
	return castSuffix(stmt.String(), e.TypeCast)
}

type ExpressionBlobLiteral struct {
	node
	expressionBase
	Wrapped
	Value    []byte
	TypeCast *types.DataType
}

func (e *ExpressionBlobLiteral) Accept(v AstVisitor) any {
	return v.VisitExpressionBlobLiteral(e)
}

func (e *ExpressionBlobLiteral) Walk(w AstListener) error {
	return run(
		w.EnterExpressionBlobLiteral(e),
		w.ExitExpressionBlobLiteral(e),
	)
}

func (e *ExpressionBlobLiteral) ToSQL() string {
	stmt := sqlwriter.NewWriter()

	if e.Wrapped {
		stmt.WrapParen()
	}

	stmt.WriteString(fmt.Sprintf("'\\%s'", hex.EncodeToString(e.Value)))

	return castSuffix(stmt.String(), e.TypeCast)
}

type ExpressionBindParameter struct {
	node

	expressionBase
	Wrapped
	Parameter string
	TypeCast  *types.DataType
}

func (e *ExpressionBindParameter) Accept(v AstVisitor) any {
	return v.VisitExpressionBindParameter(e)
}

func (e *ExpressionBindParameter) Walk(w AstListener) error {
	return run(
		w.EnterExpressionBindParameter(e),
		w.ExitExpressionBindParameter(e),
	)
}

func (e *ExpressionBindParameter) ToSQL() string {
	if e.Parameter == "" {
		panic("ExpressionBindParameter: bind parameter cannot be empty")
	}

	stmt := sqlwriter.NewWriter()

	if e.Wrapped {
		stmt.WrapParen()
	}

	stmt.WriteString(e.Parameter)

	return castSuffix(stmt.String(), e.TypeCast)
}

type ExpressionColumn struct {
	node

	expressionBase
	Wrapped
	Table    string
	Column   string
	TypeCast *types.DataType
}

func (e *ExpressionColumn) Accept(v AstVisitor) any {
	return v.VisitExpressionColumn(e)
}

func (e *ExpressionColumn) Walk(w AstListener) error {
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

	return castSuffix(stmt.String(), e.TypeCast)
}

type ExpressionUnary struct {
	node

	expressionBase
	Wrapped
	Operator UnaryOperator
	Operand  Expression

	TypeCast *types.DataType
	// NOTE: type cast only makes sense when wrapped,
}

func (e *ExpressionUnary) Accept(v AstVisitor) any {
	return v.VisitExpressionUnary(e)
}

func (e *ExpressionUnary) Walk(w AstListener) error {
	return run(
		w.EnterExpressionUnary(e),
		w.ExitExpressionUnary(e),
	)
}

func (e *ExpressionUnary) ToSQL() string {
	checkTypeWrap(e.TypeCast, bool(e.Wrapped), e)

	stmt := sqlwriter.NewWriter()

	if e.Wrapped {
		stmt.WrapParen()
	}

	stmt.WriteString(e.Operator.String())
	stmt.WriteString(e.Operand.ToSQL())

	return castSuffix(stmt.String(), e.TypeCast)
}

// checkTypeWrap checks if the type is nil and wrapped is false
func checkTypeWrap(typ *types.DataType, wrapped bool, node any) {
	typeOf := reflect.TypeOf(node)
	if typeOf.Kind() == reflect.Ptr {
		typeOf = typeOf.Elem()
	}
	if typeOf.Kind() != reflect.Struct {
		panic("checkTypeWrap: node must be a struct. this is a bug")
	}

	if typ != nil && !wrapped {
		panic(fmt.Sprintf("%s: expression with type cast must be wrapped in parentheses", typeOf.Name()))
	}
}

type ExpressionBinaryComparison struct {
	node

	expressionBase
	Wrapped
	Left     Expression
	Operator BinaryOperator
	Right    Expression

	TypeCast *types.DataType
	// NOTE: type cast only makes sense when wrapped,
}

func (e *ExpressionBinaryComparison) Accept(v AstVisitor) any {
	return v.VisitExpressionBinaryComparison(e)
}

func (e *ExpressionBinaryComparison) Walk(w AstListener) error {
	return run(
		w.EnterExpressionBinaryComparison(e),
		walk(w, e.Left),
		walk(w, e.Right),
		w.ExitExpressionBinaryComparison(e),
	)
}

func (e *ExpressionBinaryComparison) ToSQL() string {
	checkTypeWrap(e.TypeCast, bool(e.Wrapped), e)

	stmt := sqlwriter.NewWriter()

	if e.Wrapped {
		stmt.WrapParen()
	}

	stmt.WriteString(e.Left.ToSQL())
	stmt.WriteString(e.Operator.String())
	stmt.WriteString(e.Right.ToSQL())

	return castSuffix(stmt.String(), e.TypeCast)
}

type ExpressionFunction struct {
	node

	expressionBase
	Wrapped
	Function string
	Inputs   []Expression
	Star     bool // true if function is called with *
	Distinct bool

	TypeCast *types.DataType
}

func (e *ExpressionFunction) Accept(v AstVisitor) any {
	return v.VisitExpressionFunction(e)
}

func (e *ExpressionFunction) Walk(w AstListener) error {
	return run(
		w.EnterExpressionFunction(e),
		walkMany(w, e.Inputs),
		w.ExitExpressionFunction(e),
	)
}

func (e *ExpressionFunction) ToSQL() string {
	stmt := sqlwriter.NewWriter()

	if e.Wrapped {
		stmt.WrapParen()
	}

	stmt.WriteString(strings.ToLower(e.Function))
	stmt.Token.Lparen()

	if e.Star {
		stmt.Token.Asterisk()
	} else {
		if e.Distinct {
			stmt.Token.Distinct()
		}

		for i, input := range e.Inputs {
			if i > 0 {
				stmt.Token.Comma()
			}
			stmt.WriteString(input.ToSQL())
		}
	}

	stmt.Token.Rparen()

	return castSuffix(stmt.String(), e.TypeCast)
}

type ExpressionList struct {
	node

	expressionBase
	Wrapped
	Expressions []Expression

	TypeCast *types.DataType
}

func (e *ExpressionList) Accept(v AstVisitor) any {
	return v.VisitExpressionList(e)
}

func (e *ExpressionList) Walk(w AstListener) error {
	return run(
		w.EnterExpressionList(e),
		walkMany(w, e.Expressions),
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

	return castSuffix(stmt.String(), e.TypeCast)
}

type ExpressionCollate struct {
	node

	expressionBase
	Wrapped
	Expression Expression
	Collation  CollationType

	TypeCast *types.DataType
	// NOTE: type cast only makes sense when wrapped
}

func (e *ExpressionCollate) Accept(v AstVisitor) any {
	return v.VisitExpressionCollate(e)
}

func (e *ExpressionCollate) Walk(w AstListener) error {
	return run(
		w.EnterExpressionCollate(e),
		walk(w, e.Expression),
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
	checkTypeWrap(e.TypeCast, bool(e.Wrapped), e)

	stmt := sqlwriter.NewWriter()

	if e.Wrapped {
		stmt.WrapParen()
	}

	stmt.WriteString(e.Expression.ToSQL())
	stmt.Token.Collate()
	stmt.WriteString(e.Collation.String())

	return castSuffix(stmt.String(), e.TypeCast)
}

type ExpressionStringCompare struct {
	node

	expressionBase
	Wrapped
	Left     Expression
	Operator StringOperator
	Right    Expression
	Escape   Expression // can only be used with LIKE or NOT LIKE

	TypeCast *types.DataType
	// NOTE: type cast only makes sense when wrapped
}

func (e *ExpressionStringCompare) Accept(v AstVisitor) any {
	return v.VisitExpressionStringCompare(e)
}

func (e *ExpressionStringCompare) Walk(w AstListener) error {
	return run(
		w.EnterExpressionStringCompare(e),
		walk(w, e.Left),
		walk(w, e.Right),
		walk(w, e.Escape),
		w.ExitExpressionStringCompare(e),
	)
}

func (e *ExpressionStringCompare) ToSQL() string {
	if e.Left == nil {
		panic("ExpressionStringCompare: left expression cannot be nil")
	}
	if e.Right == nil {
		panic("ExpressionStringCompare: right expression cannot be nil")
	}
	checkTypeWrap(e.TypeCast, bool(e.Wrapped), e)

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

	return castSuffix(stmt.String(), e.TypeCast)
}

type ExpressionIs struct {
	node

	expressionBase
	Wrapped
	Left     Expression
	Distinct bool
	Not      bool
	Right    Expression

	TypeCast *types.DataType
	// NOTE: type cast only makes sense when wrapped
}

func (e *ExpressionIs) Accept(v AstVisitor) any {
	return v.VisitExpressionIs(e)
}

func (e *ExpressionIs) Walk(w AstListener) error {
	return run(
		w.EnterExpressionIs(e),
		walk(w, e.Left),
		walk(w, e.Right),
		w.ExitExpressionIs(e),
	)
}

func (e *ExpressionIs) ToSQL() string {
	if e.Left == nil {
		panic("ExpressionIs: left expression cannot be nil")
	}
	if e.Right == nil {
		panic("ExpressionIs: right expression cannot be nil")
	}
	checkTypeWrap(e.TypeCast, bool(e.Wrapped), e)

	stmt := sqlwriter.NewWriter()

	if e.Wrapped {
		stmt.WrapParen()
	}

	stmt.WriteString(e.Left.ToSQL())
	stmt.Token.Is()
	if e.Not {
		stmt.Token.Not()
	}
	if e.Distinct {
		stmt.Token.Distinct().From()
	}
	stmt.WriteString(e.Right.ToSQL())
	return castSuffix(stmt.String(), e.TypeCast)
}

type ExpressionBetween struct {
	node

	expressionBase
	Wrapped
	Expression Expression
	NotBetween bool
	Left       Expression
	Right      Expression

	TypeCast *types.DataType
	// NOTE: type cast only makes sense when wrapped
}

func (e *ExpressionBetween) Accept(v AstVisitor) any {
	return v.VisitExpressionBetween(e)
}

func (e *ExpressionBetween) Walk(w AstListener) error {
	return run(
		w.EnterExpressionBetween(e),
		walk(w, e.Expression),
		walk(w, e.Left),
		walk(w, e.Right),
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
	checkTypeWrap(e.TypeCast, bool(e.Wrapped), e)

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

	return castSuffix(stmt.String(), e.TypeCast)
}

type ExpressionSelect struct {
	node

	expressionBase
	Wrapped
	IsNot    bool
	IsExists bool
	Select   *SelectCore

	TypeCast *types.DataType
	// NOTE: type cast only makes sense when wrapped
}

func (e *ExpressionSelect) Accept(v AstVisitor) any {
	return v.VisitExpressionSelect(e)
}

func (e *ExpressionSelect) Walk(w AstListener) error {
	return run(
		w.EnterExpressionSelect(e),
		walk(w, e.Select),
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

	return castSuffix(stmt.String(), e.TypeCast)
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

	checkTypeWrap(e.TypeCast, bool(e.Wrapped), e)
}

type ExpressionCase struct {
	node

	expressionBase
	Wrapped
	CaseExpression Expression
	WhenThenPairs  [][2]Expression
	ElseExpression Expression

	TypeCast *types.DataType
	// NOTE: type cast does not apply to the whole case expression
}

func (e *ExpressionCase) Accept(v AstVisitor) any {
	return v.VisitExpressionCase(e)
}

func (e *ExpressionCase) Walk(w AstListener) error {
	return run(
		w.EnterExpressionCase(e),
		walk(w, e.CaseExpression),
		func() error {
			for _, whenThen := range e.WhenThenPairs {
				err := walk(w, whenThen[0])
				if err != nil {
					return err
				}
				err = walk(w, whenThen[1])
				if err != nil {
					return err
				}
			}
			return nil
		}(),
		walk(w, e.ElseExpression),
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

	TypeCast *types.DataType
	// NOTE: type cast only makes sense when wrapped
}

func (e *ExpressionArithmetic) Accept(v AstVisitor) any {
	return v.VisitExpressionArithmetic(e)
}

func (e *ExpressionArithmetic) Walk(w AstListener) error {
	return run(
		w.EnterExpressionArithmetic(e),
		walk(w, e.Left),
		walk(w, e.Right),
		w.ExitExpressionArithmetic(e),
	)
}

func (e *ExpressionArithmetic) ToSQL() string {
	checkTypeWrap(e.TypeCast, bool(e.Wrapped), e)

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

	return castSuffix(stmt.String(), e.TypeCast)
}
