package types

type CollationType uint8

const (
	CollationTypeBINARY CollationType = iota
	CollationTypeNOCASE
	CollationTypeRTRIM
)

type ExpressionType uint8

// I have included some examples of each type of expression below
// For full documentation, see https://www.sqlite.org/lang_expr.html
// Will update this when we have Kwil specific grammar documentation

const (
	ExpressionTypeLiteral        ExpressionType = iota // 1, 'hello', 1.0
	ExpressionTypeBindParameter                        // $1, $2
	ExpressionTypeComparison                           // expr1 BINARY_OPERATOR expr2
	ExpressionTypeFunction                             // function_name(expr1, expr2, ...) filter_clause? over_clause?
	ExpressionTypeExpressionList                       // (expr1, expr2, ...)
	ExpressionTypeCollate                              // expr COLLATE collation_name
	ExpressionTypeStringCompare                        // expr1 NOT? LIKE|REGEXP|MATCH expr2 ESCAPE? expr3 // ESCAPE can only be used with LIKE
	ExpressionTypeIsNull                               // expr ISNULL|NOTNULL|NOT NULL
	ExpressionTypeDistinct                             // expr IS NOT? {DISTINCT FROM}? expr2
	ExpressionTypeBetween                              // expr NOT? BETWEEN expr2 AND expr3
	ExpressionTypeIn                                   // expr NOT? IN (expr1, expr2, select_statement1, ...)
	ExpressionTypeSelect                               // NOT? EXISTS (select_statement1)
	ExpressionTypeCase                                 // CASE expr WHEN expr1 THEN expr2 WHEN expr3 THEN expr4 ELSE expr5 END
)

type EPLiteral struct{}
type EPBindParameter struct{}
type EPComparison struct{}
type EPFunction struct{}
type EPExpressionList struct{}
type EPCollate struct{}
type EPStringCompare struct{}
type EPIsNull struct{}
type EPDistinct struct{}
type EPBetween struct{}
type EPIn struct{}
type EPSelect struct{}
type EPCase struct{}

type InsertExpression interface {
	Insert() struct{}
}

func (e *EPLiteral) Insert() {
}

func (e *EPBindParameter) Insert() {
}

func (e *EPSelect) Insert() {
}
