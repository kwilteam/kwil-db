package sqlparser

import (
	"encoding/hex"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/antlr4-go/antlr/v4"

	"github.com/kwilteam/kwil-db/parse/internal/util"
	"github.com/kwilteam/kwil-db/parse/sql/grammar"
	"github.com/kwilteam/kwil-db/parse/sql/tree"
)

// astBuilder is a visitor that visits Antlr parsed tree and builds sql AST.
type astBuilder struct {
	*grammar.BaseSQLParserVisitor

	trace    bool
	trackPos bool
}

type astBuilderOption func(*astBuilder)

func astBuilderWithTrace(on bool) astBuilderOption {
	return func(l *astBuilder) {
		l.trace = on
	}
}

func astBuilderWithPos(on bool) astBuilderOption {
	return func(l *astBuilder) {
		l.trackPos = on
	}
}

var _ grammar.SQLParserVisitor = &astBuilder{}

func newAstBuilder(opts ...astBuilderOption) *astBuilder {
	k := &astBuilder{}

	for _, opt := range opts {
		opt(k)
	}

	return k
}

func (v *astBuilder) getPos(ctx antlr.ParserRuleContext) *tree.Position {
	if !v.trackPos {
		return nil
	}

	return &tree.Position{
		StartLine:   ctx.GetStart().GetLine(),
		StartColumn: ctx.GetStart().GetColumn(),
		EndLine:     ctx.GetStop().GetLine(),
		EndColumn:   ctx.GetStop().GetColumn(),
	}
}

// VisitCommon_table_expression is called when visiting a common_table_expression, return *tree.CTE
func (v *astBuilder) VisitCommon_table_expression(ctx *grammar.Common_table_expressionContext) interface{} {
	cte := tree.CTE{}

	// cte_table_name
	cteTableCtx := ctx.Cte_table_name()
	cte.Table = util.ExtractSQLName(cteTableCtx.Table_name().GetText())
	if len(cteTableCtx.AllColumn_name()) > 0 {
		cte.Columns = make([]string, len(cteTableCtx.AllColumn_name()))
		for i, colNameCtx := range cteTableCtx.AllColumn_name() {
			cte.Columns[i] = util.ExtractSQLName(colNameCtx.GetText())
		}
	}

	cte.Select = v.Visit(ctx.Select_core()).(*tree.SelectCore)
	return &cte
}

// VisitCommon_table_stmt is called when visiting a common_table_stmt, return []*tree.CTE.
func (v *astBuilder) VisitCommon_table_stmt(ctx *grammar.Common_table_stmtContext) interface{} {
	if ctx == nil {
		return nil
	}

	cteCount := len(ctx.AllCommon_table_expression())
	ctes := make([]*tree.CTE, cteCount)
	for i := 0; i < cteCount; i++ {
		cte := v.Visit(ctx.Common_table_expression(i)).(*tree.CTE)
		ctes[i] = cte
	}
	return ctes
}

func getInsertType(ctx *grammar.Insert_coreContext) tree.InsertType {
	return tree.InsertTypeInsert
}

// NOTE: VisitExpr() is dispatched in various VisitXXX_expr methods.
// Calling `v.Visit(ctx.Expr()).(tree.Expression)` will dispatch to the correct
// VisitXXX_expr method.
// func (v *KFSqliteVisitor) VisitExpr(ctx *grammar.ExprContext) interface{} {
// 	//return v.visitExpr(ctx)
// }

// VisitType_cast is called when visiting a type_cast, return tree.TypeCastType
func (v *astBuilder) VisitType_cast(ctx *grammar.Type_castContext) interface{} {
	if ctx != nil {
		typeCastRaw := ctx.Cast_type().GetText()
		if typeCastRaw[0] == '`' || typeCastRaw[0] == '"' || typeCastRaw[0] == '[' {
			// NOTE: typeCast is an IDENTIFIER, so it could be wrapped with ` or " or [ ]
			panic(fmt.Sprintf("type cast should not be wrapped in  %c", typeCastRaw[0]))
		}

		// NOTE: typeCast is case-insensitive
		switch strings.ToLower(typeCastRaw) {
		case "int":
			return tree.TypeCastInt
		case "text":
			return tree.TypeCastText
		default:
			// NOTE: we probably should move all semantic checks to analysis phase
			panic(fmt.Sprintf("unknown type cast %s", typeCastRaw))
		}
	} else {
		return ""
	}
}

// VisitLiteral_expr is called when visiting a literal_expr, return *tree.ExpressionLiteral
func (v *astBuilder) VisitText_literal_expr(ctx *grammar.Text_literal_exprContext) interface{} {
	// all literal values are string
	val := ctx.TEXT_LITERAL().GetText()
	if !strings.HasPrefix(val, "'") || !strings.HasSuffix(val, "'") {
		panic(fmt.Sprintf("invalid text literal %s", val))
	}

	expr := &tree.ExpressionTextLiteral{
		Value: val[1 : len(val)-1],
	}
	if ctx.Type_cast() != nil {
		expr.TypeCast = v.Visit(ctx.Type_cast()).(tree.TypeCastType)
	}
	return expr
}

func (v *astBuilder) VisitNumeric_literal_expr(ctx *grammar.Numeric_literal_exprContext) interface{} {
	t := ctx.NUMERIC_LITERAL().GetText()
	val, err := strconv.ParseInt(t, 10, 64)
	if err != nil {
		panic(fmt.Sprintf("failed to parse numeric literal %s: %v", t, err))
	}

	expr := &tree.ExpressionNumericLiteral{
		Value: val,
	}
	if ctx.Type_cast() != nil {
		expr.TypeCast = v.Visit(ctx.Type_cast()).(tree.TypeCastType)
	}
	return expr
}

func (v *astBuilder) VisitBoolean_literal_expr(ctx *grammar.Boolean_literal_exprContext) interface{} {
	b := ctx.BOOLEAN_LITERAL().GetText()
	boolVal, err := strconv.ParseBool(b)
	if err != nil {
		panic(fmt.Sprintf("failed to parse boolean literal %s: %v", b, err))
	}

	expr := &tree.ExpressionBooleanLiteral{
		Value: boolVal,
	}
	if ctx.Type_cast() != nil {
		expr.TypeCast = v.Visit(ctx.Type_cast()).(tree.TypeCastType)
	}
	return expr
}

func (v *astBuilder) VisitNull_literal_expr(ctx *grammar.Null_literal_exprContext) interface{} {
	expr := &tree.ExpressionNullLiteral{}
	if ctx.Type_cast() != nil {
		expr.TypeCast = v.Visit(ctx.Type_cast()).(tree.TypeCastType)
	}
	return expr
}

func (v *astBuilder) VisitBlob_literal_expr(ctx *grammar.Blob_literal_exprContext) interface{} {
	t := ctx.BLOB_LITERAL().GetText()

	// trim 0x prefix
	if !strings.HasPrefix(t, "0x") {
		panic(fmt.Sprintf("invalid blob literal %s", t))
	}
	t = t[2:]

	decoded, err := hex.DecodeString(t)
	if err != nil {
		panic(fmt.Sprintf("failed to decode blob literal %s: %v", t, err))
	}

	expr := &tree.ExpressionBlobLiteral{
		Value: decoded,
	}
	if ctx.Type_cast() != nil {
		expr.TypeCast = v.Visit(ctx.Type_cast()).(tree.TypeCastType)
	}
	return expr
}

// VisitVariable_expr is called when visiting a variable_expr, return *tree.ExpressionBindParameter
func (v *astBuilder) VisitVariable_expr(ctx *grammar.Variable_exprContext) interface{} {
	expr := &tree.ExpressionBindParameter{
		Parameter: ctx.Variable().GetText(),
	}
	if ctx.Type_cast() != nil {
		expr.TypeCast = v.Visit(ctx.Type_cast()).(tree.TypeCastType)
	}
	return expr
}

// VisitColumn_ref is called when visiting a column_ref, return *tree.ExpressionColumn, without
// type cast info.
func (v *astBuilder) VisitColumn_ref(ctx *grammar.Column_refContext) interface{} {
	expr := &tree.ExpressionColumn{}
	if ctx.Table_name() != nil {
		expr.Table = util.ExtractSQLName(ctx.Table_name().GetText())
	}
	expr.Column = util.ExtractSQLName(ctx.Column_name().GetText())
	return expr
}

// VisitColumn_expr is called when visiting a column_expr, return *tree.ExpressionColumn
func (v *astBuilder) VisitColumn_expr(ctx *grammar.Column_exprContext) interface{} {
	expr := v.Visit(ctx.Column_ref()).(*tree.ExpressionColumn)
	if ctx.Type_cast() != nil {
		expr.TypeCast = v.Visit(ctx.Type_cast()).(tree.TypeCastType)
	}
	return expr
}

// VistUnary_expr is called when visiting a unary_expr, return *tree.ExpressionUnary
func (v *astBuilder) VisitUnary_expr(ctx *grammar.Unary_exprContext) interface{} {
	expr := &tree.ExpressionUnary{}
	switch {
	case ctx.MINUS() != nil:
		expr.Operator = tree.UnaryOperatorMinus
	case ctx.PLUS() != nil:
		expr.Operator = tree.UnaryOperatorPlus
	default:
		panic(fmt.Sprintf("unknown unary operator %s", ctx.GetText()))
	}
	expr.Operand = v.Visit(ctx.Expr()).(tree.Expression)
	return expr
}

func (v *astBuilder) getCollateType(collationName string) tree.CollationType {
	// case insensitive
	switch strings.ToLower(collationName) {
	case "nocase":
		return tree.CollationTypeNoCase
	default:
		// NOTE: this is a semantic error
		panic(fmt.Sprintf("unknown collation type %s", collationName))
	}
}

// VisitCollate_expr is called when visiting a collate_expr, return *tree.ExpressionCollate
func (v *astBuilder) VisitCollate_expr(ctx *grammar.Collate_exprContext) interface{} {
	expr := v.Visit(ctx.Expr()).(tree.Expression)
	collationName := util.ExtractSQLName(ctx.Collation_name().GetText())
	return &tree.ExpressionCollate{
		Expression: expr,
		Collation:  v.getCollateType(collationName),
	}
}

// VisitParenthesized_expr is called when visiting a parenthesized_expr, return *tree.Expression
func (v *astBuilder) VisitParenthesized_expr(ctx *grammar.Parenthesized_exprContext) interface{} {
	var typeCast tree.TypeCastType
	if ctx.Type_cast() != nil {
		typeCast = v.Visit(ctx.Type_cast()).(tree.TypeCastType)
	}

	expr := v.Visit(ctx.Expr()).(tree.Expression)
	switch e := expr.(type) {
	case *tree.ExpressionTextLiteral:
		e.Wrapped = true
		e.TypeCast = typeCast
	case *tree.ExpressionNumericLiteral:
		e.Wrapped = true
		e.TypeCast = typeCast
	case *tree.ExpressionBooleanLiteral:
		e.Wrapped = true
		e.TypeCast = typeCast
	case *tree.ExpressionNullLiteral:
		e.Wrapped = true
		e.TypeCast = typeCast
	case *tree.ExpressionBlobLiteral:
		e.Wrapped = true
		e.TypeCast = typeCast
	case *tree.ExpressionBindParameter:
		e.Wrapped = true
		e.TypeCast = typeCast
	case *tree.ExpressionColumn:
		e.Wrapped = true
		e.TypeCast = typeCast
	case *tree.ExpressionUnary:
		e.Wrapped = true
		e.TypeCast = typeCast
	case *tree.ExpressionArithmetic:
		e.Wrapped = true
		e.TypeCast = typeCast
	case *tree.ExpressionBinaryComparison:
		e.Wrapped = true
		e.TypeCast = typeCast
	case *tree.ExpressionFunction:
		e.Wrapped = true
		e.TypeCast = typeCast
	case *tree.ExpressionList:
		e.Wrapped = true
		e.TypeCast = typeCast
	case *tree.ExpressionCollate:
		e.Wrapped = true
		e.TypeCast = typeCast
	case *tree.ExpressionStringCompare:
		e.Wrapped = true
		e.TypeCast = typeCast
	case *tree.ExpressionIs:
		e.Wrapped = true
		e.TypeCast = typeCast
	case *tree.ExpressionBetween:
		e.Wrapped = true
		e.TypeCast = typeCast
	case *tree.ExpressionSelect:
		e.Wrapped = true
		e.TypeCast = typeCast
	case *tree.ExpressionCase:
		e.Wrapped = true
		// typeCast does not apply on case expression
	}
	return expr
}

func (v *astBuilder) VisitSubquery(ctx *grammar.SubqueryContext) interface{} {
	return v.Visit(ctx.Select_core()).(*tree.SelectCore)
}

// VisitSubquery_expr is called when visiting a subquery_expr, return *tree.ExpressionSelect
func (v *astBuilder) VisitSubquery_expr(ctx *grammar.Subquery_exprContext) interface{} {
	stmt := v.Visit(ctx.Subquery()).(*tree.SelectCore)
	expr := &tree.ExpressionSelect{
		Select: stmt,
	}
	if ctx.EXISTS_() != nil {
		expr.IsExists = true
	}
	if ctx.NOT_() != nil {
		expr.IsNot = true
	}
	return expr
}

// VisitWhen_clause is called when visiting a when_clause, return [2]*tree.Expression
func (v *astBuilder) VisitWhen_clause(ctx *grammar.When_clauseContext) interface{} {
	var when = [2]tree.Expression{}
	when[0] = v.Visit(ctx.GetCondition()).(tree.Expression)
	when[1] = v.Visit(ctx.GetResult()).(tree.Expression)
	return when
}

// VisitCase_expr is called when visiting a case_expr, return *tree.ExpressionCase
func (v *astBuilder) VisitCase_expr(ctx *grammar.Case_exprContext) interface{} {
	expr := &tree.ExpressionCase{}
	if ctx.GetCase_clause() != nil {
		expr.CaseExpression = v.Visit(ctx.GetCase_clause()).(tree.Expression)
	}
	if ctx.GetElse_clause() != nil {
		expr.ElseExpression = v.Visit(ctx.GetElse_clause()).(tree.Expression)
	}

	for _, whenCtx := range ctx.AllWhen_clause() {
		expr.WhenThenPairs = append(expr.WhenThenPairs,
			v.Visit(whenCtx).([2]tree.Expression))
	}

	return expr
}

// VisitFunction_call is called when visiting a function_call, return *tree.ExpressionFunction
func (v *astBuilder) VisitFunction_call(ctx *grammar.Function_callContext) interface{} {
	expr := &tree.ExpressionFunction{
		Inputs: make([]tree.Expression, len(ctx.AllExpr())),
	}
	funcName := util.ExtractSQLName(ctx.Function_name().GetText())

	f, ok := tree.SQLFunctions[strings.ToLower(funcName)]
	if !ok {
		panic(fmt.Sprintf("unsupported function '%s'", funcName))
	}
	expr.Function = f

	if ctx.DISTINCT_() != nil {
		expr.Distinct = true
	}

	for i, e := range ctx.AllExpr() {
		expr.Inputs[i] = v.Visit(e).(tree.Expression)
	}

	return expr
}

// VisitFunction_expr is called when visiting a function_expr, return *tree.ExpressionFunction
func (v *astBuilder) VisitFunction_expr(ctx *grammar.Function_exprContext) interface{} {
	expr := v.Visit(ctx.Function_call()).(*tree.ExpressionFunction)
	if ctx.Type_cast() != nil {
		expr.TypeCast = v.Visit(ctx.Type_cast()).(tree.TypeCastType)
	}
	return expr
}

// VisitExpr_list_expr is called when visiting a expr_list_expr, return *tree.ExpressionList
func (v *astBuilder) VisitExpr_list_expr(ctx *grammar.Expr_list_exprContext) interface{} {
	return v.Visit(ctx.Expr_list()).(*tree.ExpressionList)
}

// VisitArithmetic_expr is called when visiting a arithmetic_expr, return *tree.ExpressionArithmetic
func (v *astBuilder) VisitArithmetic_expr(ctx *grammar.Arithmetic_exprContext) interface{} {
	expr := &tree.ExpressionArithmetic{}
	expr.Left = v.Visit(ctx.GetLeft()).(tree.Expression)
	expr.Right = v.Visit(ctx.GetRight()).(tree.Expression)

	switch {
	case ctx.STAR() != nil:
		expr.Operator = tree.ArithmeticOperatorMultiply
	case ctx.DIV() != nil:
		expr.Operator = tree.ArithmeticOperatorDivide
	case ctx.MOD() != nil:
		expr.Operator = tree.ArithmeticOperatorModulus
	case ctx.PLUS() != nil:
		expr.Operator = tree.ArithmeticOperatorAdd
	case ctx.MINUS() != nil:
		expr.Operator = tree.ArithmeticOperatorSubtract
	default:
		panic(fmt.Sprintf("unknown arithmetic operator %s", ctx.GetText()))
	}

	return expr
}

// VisitIn_subquery_expr is called when visiting a in_suquery_expr, return *tree.ExpressionBinaryComparison
func (v *astBuilder) VisitIn_subquery_expr(ctx *grammar.In_subquery_exprContext) interface{} {
	expr := &tree.ExpressionBinaryComparison{
		Left:     v.Visit(ctx.GetElem()).(tree.Expression),
		Operator: tree.ComparisonOperatorIn,
	}
	if ctx.NOT_() != nil {
		expr.Operator = tree.ComparisonOperatorNotIn
	}
	sub := v.Visit(ctx.Subquery()).(*tree.SelectCore)
	expr.Right = &tree.ExpressionSelect{Select: sub}
	return expr
}

// VisitExpr_list is called when visiting a expr_list, return *tree.ExpressionList
func (v *astBuilder) VisitExpr_list(ctx *grammar.Expr_listContext) interface{} {
	exprCount := len(ctx.AllExpr())
	exprs := make([]tree.Expression, exprCount)
	for i, exprCtx := range ctx.AllExpr() {
		exprs[i] = v.Visit(exprCtx).(tree.Expression)
	}
	return &tree.ExpressionList{Expressions: exprs}
}

// VisitIn_list_expr is called when visiting a in_list_expr, return *tree.ExpressionBinaryComparison
func (v *astBuilder) VisitIn_list_expr(ctx *grammar.In_list_exprContext) interface{} {
	expr := &tree.ExpressionBinaryComparison{
		Left:     v.Visit(ctx.GetElem()).(tree.Expression),
		Operator: tree.ComparisonOperatorIn,
	}
	if ctx.NOT_() != nil {
		expr.Operator = tree.ComparisonOperatorNotIn
	}
	expr.Right = v.Visit(ctx.Expr_list()).(*tree.ExpressionList)
	return expr
}

// VisitBetween_expr is called when visiting a between_expr, return *tree.ExpressionBetween
func (v *astBuilder) VisitBetween_expr(ctx *grammar.Between_exprContext) interface{} {
	expr := &tree.ExpressionBetween{
		Expression: v.Visit(ctx.GetElem()).(tree.Expression),
		Left:       v.Visit(ctx.GetLow()).(tree.Expression),
		Right:      v.Visit(ctx.GetHigh()).(tree.Expression),
	}
	if ctx.NOT_() != nil {
		expr.NotBetween = true
	}
	return expr
}

// VisitLike_expr is called when visiting a like_expr, return *tree.ExpressionStringCompare
func (v *astBuilder) VisitLike_expr(ctx *grammar.Like_exprContext) interface{} {
	expr := &tree.ExpressionStringCompare{
		Left:     v.Visit(ctx.GetElem()).(tree.Expression),
		Operator: tree.StringOperatorLike,
		Right:    v.Visit(ctx.GetPattern()).(tree.Expression),
	}
	if ctx.NOT_() != nil {
		expr.Operator = tree.StringOperatorNotLike
	}
	if ctx.ESCAPE_() != nil {
		expr.Escape = v.Visit(ctx.GetEscape()).(tree.Expression)
	}
	return expr
}

// VisitComparisonOperator is called when visiting a comparisonOpertor, return tree.ComparisonOperator
func (v *astBuilder) VisitComparisonOperator(ctx *grammar.ComparisonOperatorContext) interface{} {
	switch {
	case ctx.LT() != nil:
		return tree.ComparisonOperatorLessThan
	case ctx.LT_EQ() != nil:
		return tree.ComparisonOperatorLessThanOrEqual
	case ctx.GT() != nil:
		return tree.ComparisonOperatorGreaterThan
	case ctx.GT_EQ() != nil:
		return tree.ComparisonOperatorGreaterThanOrEqual
	case ctx.ASSIGN() != nil:
		return tree.ComparisonOperatorEqual
	case ctx.NOT_EQ1() != nil:
		return tree.ComparisonOperatorNotEqual
	case ctx.NOT_EQ2() != nil:
		return tree.ComparisonOperatorNotEqualDiamond
	default:
		panic(fmt.Sprintf("unknown comparison operator %s", ctx.GetText()))
	}
}

// VisitComparison_expr is called when visiting a comparison_expr, return *tree.ExpressionBinaryComparison
func (v *astBuilder) VisitComparison_expr(ctx *grammar.Comparison_exprContext) interface{} {
	expr := &tree.ExpressionBinaryComparison{
		Left:     v.Visit(ctx.GetLeft()).(tree.Expression),
		Operator: v.Visit(ctx.ComparisonOperator()).(tree.ComparisonOperator),
		Right:    v.Visit(ctx.GetRight()).(tree.Expression),
	}

	return expr
}

// VisitIs_expr is called when visiting a is_expr, return *tree.ExpressionIs
func (v *astBuilder) VisitIs_expr(ctx *grammar.Is_exprContext) interface{} {
	expr := &tree.ExpressionIs{
		Left: v.Visit(ctx.Expr(0)).(tree.Expression),
	}
	if ctx.NOT_() != nil {
		expr.Not = true
	}

	switch {
	case ctx.NULL_LITERAL() != nil:
		expr.Right = &tree.ExpressionNullLiteral{}
	case ctx.BOOLEAN_LITERAL() != nil:
		tf := ctx.BOOLEAN_LITERAL().GetText()
		var b bool
		if strings.EqualFold(tf, "true") {
			b = true
		} else if strings.EqualFold(tf, "false") {
			b = false
		} else {
			panic(fmt.Sprintf("unknown boolean literal %s", tf))
		}

		expr.Right = &tree.ExpressionBooleanLiteral{
			Value: b,
		}
	case ctx.DISTINCT_() != nil:
		expr.Right = v.Visit(ctx.Expr(1)).(tree.Expression)
		expr.Distinct = true
	default:
		panic(fmt.Sprintf("unknown IS expression %s", ctx.GetText()))
	}
	return expr
}

// VisitNull_expr is called when visiting a null_expr, return *tree.ExpressionIs
func (v *astBuilder) VisitNull_expr(ctx *grammar.Null_exprContext) interface{} {
	expr := &tree.ExpressionIs{
		Left:  v.Visit(ctx.Expr()).(tree.Expression),
		Right: &tree.ExpressionNullLiteral{},
	}
	if ctx.NOTNULL_() != nil {
		expr.Not = true
	}
	return expr
}

// VisitLogical_not_expr is called when visiting a logical_not_expr, return *tree.ExpressionUnary
func (v *astBuilder) VisitLogical_not_expr(ctx *grammar.Logical_not_exprContext) interface{} {
	return &tree.ExpressionUnary{
		Operator: tree.UnaryOperatorNot,
		Operand:  v.Visit(ctx.Expr()).(tree.Expression),
	}
}

// VisitLogical_binary_expr is called when visiting a logical_binary_expr, return *tree.ExpressionBinaryLogical
func (v *astBuilder) VisitLogical_binary_expr(ctx *grammar.Logical_binary_exprContext) interface{} {
	expr := &tree.ExpressionBinaryComparison{
		Left:  v.Visit(ctx.GetLeft()).(tree.Expression),
		Right: v.Visit(ctx.GetRight()).(tree.Expression),
	}

	switch {
	case ctx.AND_() != nil:
		expr.Operator = tree.LogicalOperatorAnd
	case ctx.OR_() != nil:
		expr.Operator = tree.LogicalOperatorOr
	default:
		panic(fmt.Sprintf("unknown logical operator %s", ctx.GetText()))
	}

	return expr
}

// VisitValues_clause is called when visiting a values_clause, return [][]tree.Expression
func (v *astBuilder) VisitValues_clause(ctx *grammar.Values_clauseContext) interface{} {
	if ctx == nil {
		return nil
	}

	allValueRowCtx := ctx.AllValue_row()
	rows := make([][]tree.Expression, len(allValueRowCtx))
	for i, valueRowCtx := range allValueRowCtx {
		allExprCtx := valueRowCtx.AllExpr()
		exprs := make([]tree.Expression, len(allExprCtx))
		for j, exprCtx := range allExprCtx {
			exprs[j] = v.Visit(exprCtx).(tree.Expression)
		}
		rows[i] = exprs
	}
	return rows
}

// VisitUpsert_clause is called when visiting a upsert_clause, return *tree.Upsert
func (v *astBuilder) VisitUpsert_clause(ctx *grammar.Upsert_clauseContext) interface{} {
	clause := tree.Upsert{
		Type: tree.UpsertTypeDoNothing,
	}

	// conflict target
	conflictTarget := tree.ConflictTarget{}
	allIndexedColumnCtx := ctx.AllIndexed_column()
	indexedColumns := make([]string, len(allIndexedColumnCtx))
	for i, indexedColumnCtx := range allIndexedColumnCtx {
		indexedColumns[i] = util.ExtractSQLName(indexedColumnCtx.Column_name().GetText())
	}
	conflictTarget.IndexedColumns = indexedColumns

	if ctx.GetTarget_expr() != nil {
		conflictTarget.Where = v.Visit(ctx.GetTarget_expr()).(tree.Expression)
	}

	if len(allIndexedColumnCtx) != 0 {
		clause.ConflictTarget = &conflictTarget
	}

	if ctx.NOTHING_() != nil {
		return &clause
	}

	// conflict update
	clause.Type = tree.UpsertTypeDoUpdate
	updateCount := len(ctx.AllUpsert_update())
	updates := make([]*tree.UpdateSetClause, updateCount)
	for i, updateCtx := range ctx.AllUpsert_update() {
		updates[i] = v.Visit(updateCtx).(*tree.UpdateSetClause)
	}

	clause.Updates = updates

	if ctx.GetUpdate_expr() != nil {
		clause.Where = v.Visit(ctx.GetUpdate_expr()).(tree.Expression)
	}
	return &clause
}

// VisitUpsert_update is called when visiting a upsert_update, return *tree.UpdateSetClause
func (v *astBuilder) VisitUpsert_update(ctx *grammar.Upsert_updateContext) interface{} {
	clause := tree.UpdateSetClause{}
	if ctx.Column_name_list() != nil {
		clause.Columns = v.Visit(ctx.Column_name_list()).([]string)
	} else {
		clause.Columns = []string{util.ExtractSQLName(ctx.Column_name().GetText())}
	}

	clause.Expression = v.Visit(ctx.Expr()).(tree.Expression)
	return &clause
}

// VisitColumn_name_list is called when visiting a column_name_list, return []string
func (v *astBuilder) VisitColumn_name_list(ctx *grammar.Column_name_listContext) interface{} {
	names := make([]string, len(ctx.AllColumn_name()))
	for i, nameCtx := range ctx.AllColumn_name() {
		names[i] = util.ExtractSQLName(nameCtx.GetText())
	}
	return names
}

// VisitColumn_name is called when visiting a column_name, return string
func (v *astBuilder) VisitColumn_name(ctx *grammar.Column_nameContext) interface{} {
	return util.ExtractSQLName(ctx.GetText())
}

// VisitReturning_clause is called when visiting a returning_clause, return *tree.ReturningClause
func (v *astBuilder) VisitReturning_clause(ctx *grammar.Returning_clauseContext) interface{} {
	clause := tree.ReturningClause{}
	clause.Returned = make([]*tree.ReturningClauseColumn, len(ctx.AllReturning_clause_result_column()))
	for i, columnCtx := range ctx.AllReturning_clause_result_column() {
		if columnCtx.STAR() != nil {
			clause.Returned[i] = &tree.ReturningClauseColumn{
				All: true,
			}
			continue
		}

		clause.Returned[i] = &tree.ReturningClauseColumn{
			Expression: v.Visit(columnCtx.Expr()).(tree.Expression),
		}

		if columnCtx.Column_alias() != nil {
			clause.Returned[i].Alias = util.ExtractSQLName(columnCtx.Column_alias().GetText())
		}

	}
	return &clause
}

// VisitUpdate_set_subclause is called when visiting a column_assign_subclause, return *tree.UpdateSetClause
func (v *astBuilder) VisitUpdate_set_subclause(ctx *grammar.Update_set_subclauseContext) interface{} {
	result := tree.UpdateSetClause{}

	if ctx.Column_name_list() != nil {
		result.Columns = v.Visit(ctx.Column_name_list()).([]string)
	} else {
		result.Columns = []string{util.ExtractSQLName(ctx.Column_name().GetText())}
	}

	result.Expression = v.Visit(ctx.Expr()).(tree.Expression)
	return &result
}

// VisitQualified_table_name is called when visiting a qualified_table_name, return *tree.QualifiedTableName
func (v *astBuilder) VisitQualified_table_name(ctx *grammar.Qualified_table_nameContext) interface{} {
	result := tree.QualifiedTableName{}

	result.TableName = util.ExtractSQLName(ctx.Table_name().GetText())

	if ctx.Table_alias() != nil {
		result.TableAlias = util.ExtractSQLName(ctx.Table_alias().GetText())
	}

	return &result
}

// VisitUpdate_core is called when visiting a update_core, return *tree.UpdateCore
func (v *astBuilder) VisitUpdate_core(ctx *grammar.Update_coreContext) interface{} {
	var updateStmt tree.UpdateCore

	updateStmt.QualifiedTableName = v.Visit(ctx.Qualified_table_name()).(*tree.QualifiedTableName)

	updateStmt.UpdateSetClause = make([]*tree.UpdateSetClause, len(ctx.AllUpdate_set_subclause()))
	for i, subclauseCtx := range ctx.AllUpdate_set_subclause() {
		updateStmt.UpdateSetClause[i] = v.Visit(subclauseCtx).(*tree.UpdateSetClause)
	}

	if ctx.FROM_() != nil {
		updateStmt.From = v.Visit(ctx.Relation()).(tree.Relation)
	}

	if ctx.WHERE_() != nil {
		updateStmt.Where = v.Visit(ctx.Expr()).(tree.Expression)
	}

	if ctx.Returning_clause() != nil {
		updateStmt.Returning = v.Visit(ctx.Returning_clause()).(*tree.ReturningClause)
	}

	return &updateStmt
}

// VisitUpdate_stmt is called when visiting a update_stmt, return *tree.UpdateStmt
func (v *astBuilder) VisitUpdate_stmt(ctx *grammar.Update_stmtContext) interface{} {
	t := tree.UpdateStmt{}

	if ctx.Common_table_stmt() != nil {
		t.CTE = v.Visit(ctx.Common_table_stmt()).([]*tree.CTE)
	}

	t.Core = v.Visit(ctx.Update_core()).(*tree.UpdateCore)
	return &t
}

func (v *astBuilder) VisitInsert_core(ctx *grammar.Insert_coreContext) interface{} {
	var insertStmt tree.InsertCore

	insertStmt.InsertType = getInsertType(ctx)
	insertStmt.Table = util.ExtractSQLName(ctx.Table_name().GetText())
	if ctx.Table_alias() != nil {
		insertStmt.TableAlias = util.ExtractSQLName(ctx.Table_alias().GetText())
	}

	allColumnNameCtx := ctx.AllColumn_name()
	if len(allColumnNameCtx) > 0 {
		insertStmt.Columns = make([]string, len(allColumnNameCtx))
		for i, nc := range allColumnNameCtx {
			insertStmt.Columns[i] = util.ExtractSQLName(nc.GetText())
		}
	}

	insertStmt.Values = v.Visit(ctx.Values_clause()).([][]tree.Expression)
	if ctx.Upsert_clause() != nil {
		insertStmt.Upsert = v.Visit(ctx.Upsert_clause()).(*tree.Upsert)
	}
	if ctx.Returning_clause() != nil {
		insertStmt.ReturningClause = v.Visit(ctx.Returning_clause()).(*tree.ReturningClause)
	}

	return &insertStmt
}

func (v *astBuilder) VisitInsert_stmt(ctx *grammar.Insert_stmtContext) interface{} {
	t := tree.InsertStmt{}

	if ctx.Common_table_stmt() != nil {
		t.CTE = v.Visit(ctx.Common_table_stmt()).([]*tree.CTE)
	}

	t.Core = v.Visit(ctx.Insert_core()).(*tree.InsertCore)
	return &t
}

// VisitCompound_operator is called when visiting a compound_operator, return *tree.CompoundOperator
func (v *astBuilder) VisitCompound_operator(ctx *grammar.Compound_operatorContext) interface{} {
	switch {
	case ctx.UNION_() != nil:
		if ctx.ALL_() != nil {
			return &tree.CompoundOperator{Operator: tree.CompoundOperatorTypeUnionAll}
		}
		return &tree.CompoundOperator{Operator: tree.CompoundOperatorTypeUnion}
	case ctx.INTERSECT_() != nil:
		return &tree.CompoundOperator{Operator: tree.CompoundOperatorTypeIntersect}
	case ctx.EXCEPT_() != nil:
		return &tree.CompoundOperator{Operator: tree.CompoundOperatorTypeExcept}
	}
	panic("unreachable")
}

// VisitOrdering_term is called when visiting a ordering_term, return *tree.OrderingTerm
func (v *astBuilder) VisitOrdering_term(ctx *grammar.Ordering_termContext) interface{} {
	result := tree.OrderingTerm{}
	result.Expression = v.Visit(ctx.Expr()).(tree.Expression)

	if ctx.Asc_desc() != nil {
		if ctx.Asc_desc().DESC_() != nil {
			result.OrderType = tree.OrderTypeDesc
		} else {
			result.OrderType = tree.OrderTypeAsc
		}
	}

	if ctx.NULLS_() != nil {
		if ctx.FIRST_() != nil {
			result.NullOrdering = tree.NullOrderingTypeFirst
		} else {
			result.NullOrdering = tree.NullOrderingTypeLast
		}
	}

	return &result
}

// VisitOrder_by_stmt is called when visiting a order_by_stmt, return *tree.OrderBy
func (v *astBuilder) VisitOrder_by_stmt(ctx *grammar.Order_by_stmtContext) interface{} {
	count := len(ctx.AllOrdering_term())
	result := tree.OrderBy{OrderingTerms: make([]*tree.OrderingTerm, count)}

	for i, orderingTermCtx := range ctx.AllOrdering_term() {
		result.OrderingTerms[i] = v.Visit(orderingTermCtx).(*tree.OrderingTerm)
	}

	return &result
}

// VisitLimit_stmt is called when visiting a limit_stmt, return *tree.Limit
func (v *astBuilder) VisitLimit_stmt(ctx *grammar.Limit_stmtContext) interface{} {
	result := tree.Limit{
		Expression: v.Visit(ctx.Expr(0)).(tree.Expression),
	}

	// LIMIT row_count OFFSET offset;
	// IS SAME AS
	// LIMIT offset, row_count;
	// TODO: in the tree we should just use one or the other, not both.

	if ctx.OFFSET_() != nil {
		result.Offset = v.Visit(ctx.Expr(1)).(tree.Expression)
	}

	return &result
}

// VisitTable_or_subquery is called when visiting a table_or_subquery, return tree.TableOrSubquery
func (v *astBuilder) VisitTable_or_subquery(ctx *grammar.Table_or_subqueryContext) interface{} {
	switch {
	case ctx.Table_name() != nil:
		t := tree.RelationTable{
			Name: util.ExtractSQLName(ctx.Table_name().GetText()),
		}
		if ctx.Table_alias() != nil {
			t.Alias = util.ExtractSQLName(ctx.Table_alias().GetText())
		}
		return &t
	case ctx.Select_core() != nil:
		t := tree.RelationSubquery{
			Select: v.Visit(ctx.Select_core()).(*tree.SelectCore),
		}
		if ctx.Table_alias() != nil {
			t.Alias = util.ExtractSQLName(ctx.Table_alias().GetText())
		}
		return &t
	default:
		panic("unsupported table_or_subquery type")
	}
}

// VisitJoin_operator is called when visiting a join_operator, return *tree.JoinOperator
func (v *astBuilder) VisitJoin_operator(ctx *grammar.Join_operatorContext) interface{} {
	jp := tree.JoinOperator{
		JoinType: tree.JoinTypeJoin,
	}

	if ctx.INNER_() != nil {
		jp.JoinType = tree.JoinTypeInner
		return &jp
	}

	switch {
	case ctx.LEFT_() != nil:
		jp.JoinType = tree.JoinTypeLeft
	case ctx.RIGHT_() != nil:
		jp.JoinType = tree.JoinTypeRight
	case ctx.FULL_() != nil:
		jp.JoinType = tree.JoinTypeFull
	}

	if ctx.OUTER_() != nil {
		jp.Outer = true
	}

	return &jp
}

// VisitJoin_relation is called when visiting a join_relation, return *tree.JoinPredicate
func (v *astBuilder) VisitJoin_relation(ctx *grammar.Join_relationContext) interface{} {
	jp := tree.JoinPredicate{}
	jp.JoinOperator = v.Visit(ctx.Join_operator()).(*tree.JoinOperator)
	jp.Table = v.Visit(ctx.GetRight_relation()).(tree.Relation)
	jp.Constraint = v.Visit(ctx.Join_constraint().Expr()).(tree.Expression)
	return &jp
}

// VisitResult_column is called when visiting a result_column, return tree.ResultColumn
func (v *astBuilder) VisitResult_column(ctx *grammar.Result_columnContext) interface{} {
	switch {
	// table_name need to be checked first
	case ctx.Table_name() != nil:
		return &tree.ResultColumnTable{
			TableName: util.ExtractSQLName(ctx.Table_name().GetText()),
		}
	case ctx.STAR() != nil:
		return &tree.ResultColumnStar{}
	case ctx.Expr() != nil:
		r := &tree.ResultColumnExpression{
			Expression: v.Visit(ctx.Expr()).(tree.Expression),
		}
		if ctx.Column_alias() != nil {
			r.Alias = util.ExtractSQLName(ctx.Column_alias().GetText())
		}
		return r
	}

	return nil
}

func (v *astBuilder) VisitDelete_core(ctx *grammar.Delete_coreContext) interface{} {
	var deleteStmt tree.DeleteCore
	deleteStmt.QualifiedTableName = v.Visit(ctx.Qualified_table_name()).(*tree.QualifiedTableName)

	if ctx.WHERE_() != nil {
		deleteStmt.Where = v.Visit(ctx.Expr()).(tree.Expression)
	}

	if ctx.Returning_clause() != nil {
		deleteStmt.Returning = v.Visit(ctx.Returning_clause()).(*tree.ReturningClause)
	}

	return &deleteStmt
}

// VisitDelete_stmt is called when visiting a delete_stmt, return *tree.DeleteStmt
func (v *astBuilder) VisitDelete_stmt(ctx *grammar.Delete_stmtContext) interface{} {
	t := tree.DeleteStmt{}

	if ctx.Common_table_stmt() != nil {
		t.CTE = v.Visit(ctx.Common_table_stmt()).([]*tree.CTE)
	}

	t.Core = v.Visit(ctx.Delete_core()).(*tree.DeleteCore)
	return &t
}

// VisitSimple_select is called when visiting a Simple_select, return *tree.SelectCore
func (v *astBuilder) VisitSimple_select(ctx *grammar.Simple_selectContext) interface{} {
	t := tree.SimpleSelect{
		SelectType: tree.SelectTypeAll,
	}

	if ctx.DISTINCT_() != nil {
		t.SelectType = tree.SelectTypeDistinct
	}

	//NOTE: Columns will be changed in SelectCore
	//assume all columns are * or table.* or table.column
	t.Columns = make([]tree.ResultColumn, len(ctx.AllResult_column()))
	for i, columnCtx := range ctx.AllResult_column() {
		t.Columns[i] = v.Visit(columnCtx).(tree.ResultColumn)
	}

	if ctx.FROM_() != nil {
		t.From = v.Visit(ctx.Relation()).(tree.Relation)
	}

	if ctx.GetWhereExpr() != nil {
		t.Where = v.Visit(ctx.GetWhereExpr()).(tree.Expression)
	}

	if ctx.GROUP_() != nil {
		exprs := make([]tree.Expression, len(ctx.GetGroupByExpr()))
		for i, exprCtx := range ctx.GetGroupByExpr() {
			exprs[i] = v.Visit(exprCtx).(tree.Expression)
		}

		groupBy := &tree.GroupBy{
			Expressions: exprs,
		}

		if ctx.HAVING_() != nil {
			groupBy.Having = v.Visit(ctx.GetHavingExpr()).(tree.Expression)
		}

		t.GroupBy = groupBy
	}

	return &t
}

// VisitRelation is called when visiting a relation, return tree.Relation
func (v *astBuilder) VisitRelation(ctx *grammar.RelationContext) interface{} {
	left := v.Visit(ctx.Table_or_subquery()).(tree.Relation)

	if len(ctx.AllJoin_relation()) > 0 {
		rel := tree.RelationJoin{
			Relation: left,
			Joins:    make([]*tree.JoinPredicate, len(ctx.AllJoin_relation())),
		}
		// join relations
		for i, joinRelationCtx := range ctx.AllJoin_relation() {
			rel.Joins[i] = v.Visit(joinRelationCtx).(*tree.JoinPredicate)
		}
		return &rel
	} else {
		// table or subquery relation
		return left
	}
}

// VisitSelect_core is called when visiting a select_stmt_core, return *tree.SelectCore
func (v *astBuilder) VisitSelect_core(ctx *grammar.Select_coreContext) interface{} {
	t := tree.SelectCore{}
	selectCores := make([]*tree.SimpleSelect, len(ctx.AllSimple_select()))

	// first Simple_select
	selectCores[0] = v.Visit(ctx.Simple_select(0)).(*tree.SimpleSelect)

	// rest Simple_select
	for i, selectCoreCtx := range ctx.AllSimple_select()[1:] {
		compoundOperator := v.Visit(ctx.Compound_operator(i)).(*tree.CompoundOperator)
		core := v.Visit(selectCoreCtx).(*tree.SimpleSelect)
		core.Compound = compoundOperator
		selectCores[i+1] = core
	}

	t.SimpleSelects = selectCores

	if ctx.Order_by_stmt() != nil {
		t.OrderBy = v.Visit(ctx.Order_by_stmt()).(*tree.OrderBy)
	}

	if ctx.Limit_stmt() != nil {
		t.Limit = v.Visit(ctx.Limit_stmt()).(*tree.Limit)
	}

	return &t
}

// VisitSelect_stmt is called when visiting a select_stmt, return *tree.SelectStmt
func (v *astBuilder) VisitSelect_stmt(ctx *grammar.Select_stmtContext) interface{} {
	t := tree.SelectStmt{}

	if ctx.Common_table_stmt() != nil {
		t.CTE = v.Visit(ctx.Common_table_stmt()).([]*tree.CTE)
	}

	t.Stmt = v.Visit(ctx.Select_core()).(*tree.SelectCore)
	return &t
}

func (v *astBuilder) VisitSql_stmt_list(ctx *grammar.Sql_stmt_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *astBuilder) VisitSql_stmt(ctx *grammar.Sql_stmtContext) interface{} {
	// Sql_stmtContext will only have one stmt
	return v.VisitChildren(ctx).([]tree.AstNode)[0]
}

// VisitStatements is called when visiting a statements, return []tree.AstNode
func (v *astBuilder) VisitStatements(ctx *grammar.StatementsContext) interface{} {
	// ParseContext will only have one Sql_stmt_listContext
	sqlStmtListContext := ctx.Sql_stmt_list(0)
	return v.VisitChildren(sqlStmtListContext).([]tree.AstNode)
}

// Visit dispatch to the visit method of the ctx
// e.g. if the tree is a ParseContext, then dispatch call VisitParse.
// Overwrite is needed,
// refer to https://github.com/antlr/antlr4/pull/1841#issuecomment-576791512
func (v *astBuilder) Visit(parseTree antlr.ParseTree) interface{} {
	if v.trace {
		fmt.Printf("visit tree: %v, %s\n", reflect.TypeOf(parseTree), parseTree.GetText())
	}
	return parseTree.Accept(v)
}

// VisitChildren visits the children of the specified node.
// Overwrite is needed,
// refer to https://github.com/antlr/antlr4/pull/1841#issuecomment-576791512
// calling function need to convert the result to asts
func (v *astBuilder) VisitChildren(node antlr.RuleNode) interface{} {
	var result []tree.AstNode
	n := node.GetChildCount()
	for i := 0; i < n; i++ {
		child := node.GetChild(i)
		if !v.shouldVisitNextChild(child, result) {
			if v.trace {
				fmt.Printf("should not visit next child: %v,\n", reflect.TypeOf(child))
			}
			break
		}
		c := child.(antlr.ParseTree)
		childResult := v.Visit(c).(tree.AstNode)
		result = append(result, childResult)
	}
	return result
}

func (v *astBuilder) shouldVisitNextChild(node antlr.Tree, currentResult interface{}) bool {
	if _, ok := node.(antlr.TerminalNode); ok {
		return false
	}

	return true
}
