package sqlparser

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/antlr/antlr4/runtime/Go/antlr/v4"
	"github.com/kwilteam/kwil-db/parse/sql/tree"
	"github.com/kwilteam/sql-grammar-go/sqlgrammar"
)

// KFSqliteVisitor is visitor that visit Antlr parsed tree and returns the AST.
type KFSqliteVisitor struct {
	sqlgrammar.BaseSQLParserVisitor

	trace bool
}

type KFSqliteVisitorOption func(*KFSqliteVisitor)

func KFVisitorWithTrace(on bool) KFSqliteVisitorOption {
	return func(l *KFSqliteVisitor) {
		l.trace = on
	}
}

var _ sqlgrammar.SQLParserVisitor = &KFSqliteVisitor{}

func NewKFSqliteVisitor(opts ...KFSqliteVisitorOption) *KFSqliteVisitor {
	k := &KFSqliteVisitor{}
	for _, opt := range opts {
		opt(k)
	}
	return k
}

// VisitCommon_table_expression is called when visiting a common_table_expression, return *tree.CTE
func (v *KFSqliteVisitor) VisitCommon_table_expression(ctx *sqlgrammar.Common_table_expressionContext) interface{} {
	cte := tree.CTE{}

	// cte_table_name
	cteTableCtx := ctx.Cte_table_name()
	cte.Table = extractSQLName(cteTableCtx.Table_name().GetText())
	if len(cteTableCtx.AllColumn_name()) > 0 {
		cte.Columns = make([]string, len(cteTableCtx.AllColumn_name()))
		for i, colNameCtx := range cteTableCtx.AllColumn_name() {
			cte.Columns[i] = extractSQLName(colNameCtx.GetText())
		}
	}

	selectStmtCoreCtx := ctx.Select_stmt_core()
	cte.Select = v.Visit(selectStmtCoreCtx).(*tree.SelectStmt)
	return &cte
}

// VisitCommon_table_stmt is called when visiting a common_table_stmt, return []*tree.CTE.
func (v *KFSqliteVisitor) VisitCommon_table_stmt(ctx *sqlgrammar.Common_table_stmtContext) interface{} {
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

func getInsertType(ctx *sqlgrammar.Insert_stmtContext) tree.InsertType {
	return tree.InsertTypeInsert
}

// NOTE: VisitExpr() is dispatched in various VisitXXX_expr methods.
// Calling `v.Visit(ctx.Expr()).(tree.Expression)` will dispatch to the correct
// VisitXXX_expr method.
// func (v *KFSqliteVisitor) VisitExpr(ctx *sqlgrammar.ExprContext) interface{} {
// 	//return v.visitExpr(ctx)
// }

// VisitType_cast is called when visiting a type_cast, return tree.TypeCastType
func (v *KFSqliteVisitor) VisitType_cast(ctx *sqlgrammar.Type_castContext) interface{} {
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

// VisitLiteral is called when visiting a literal, return *tree.ExpressionLiteral
func (v *KFSqliteVisitor) VisitLiteral(ctx *sqlgrammar.LiteralContext) interface{} {
	// all literal values are string
	text := ctx.GetText()
	if strings.EqualFold(text, "null") {
		text = "NULL"
	}
	return text
}

// VisitLiteral_expr is called when visiting a literal_expr, return *tree.ExpressionLiteral
func (v *KFSqliteVisitor) VisitLiteral_expr(ctx *sqlgrammar.Literal_exprContext) interface{} {
	// all literal values are string
	expr := &tree.ExpressionLiteral{
		Value: v.Visit(ctx.Literal()).(string),
	}
	if ctx.Type_cast() != nil {
		expr.TypeCast = v.Visit(ctx.Type_cast()).(tree.TypeCastType)
	}
	return expr
}

// VisitVariable_expr is called when visiting a variable_expr, return *tree.ExpressionBindParameter
func (v *KFSqliteVisitor) VisitVariable_expr(ctx *sqlgrammar.Variable_exprContext) interface{} {
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
func (v *KFSqliteVisitor) VisitColumn_ref(ctx *sqlgrammar.Column_refContext) interface{} {
	expr := &tree.ExpressionColumn{}
	if ctx.Table_name() != nil {
		expr.Table = extractSQLName(ctx.Table_name().GetText())
	}
	expr.Column = extractSQLName(ctx.Column_name().GetText())
	return expr
}

// VisitColumn_expr is called when visiting a column_expr, return *tree.ExpressionColumn
func (v *KFSqliteVisitor) VisitColumn_expr(ctx *sqlgrammar.Column_exprContext) interface{} {
	expr := v.Visit(ctx.Column_ref()).(*tree.ExpressionColumn)
	if ctx.Type_cast() != nil {
		expr.TypeCast = v.Visit(ctx.Type_cast()).(tree.TypeCastType)
	}
	return expr
}

// VistUnary_expr is called when visiting a unary_expr, return *tree.ExpressionUnary
func (v *KFSqliteVisitor) VisitUnary_expr(ctx *sqlgrammar.Unary_exprContext) interface{} {
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

func (v *KFSqliteVisitor) getCollateType(collationName string) tree.CollationType {
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
func (v *KFSqliteVisitor) VisitCollate_expr(ctx *sqlgrammar.Collate_exprContext) interface{} {
	expr := v.Visit(ctx.Expr()).(tree.Expression)
	collationName := extractSQLName(ctx.Collation_name().GetText())
	return &tree.ExpressionCollate{
		Expression: expr,
		Collation:  v.getCollateType(collationName),
	}
}

// VisitParenthesized_expr is called when visiting a parenthesized_expr, return *tree.Expression
func (v *KFSqliteVisitor) VisitParenthesized_expr(ctx *sqlgrammar.Parenthesized_exprContext) interface{} {
	var typeCast tree.TypeCastType
	if ctx.Type_cast() != nil {
		typeCast = v.Visit(ctx.Type_cast()).(tree.TypeCastType)
	}

	expr := v.Visit(ctx.Expr()).(tree.Expression)
	switch e := expr.(type) {
	case *tree.ExpressionLiteral:
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

func (v *KFSqliteVisitor) VisitSubquery(ctx *sqlgrammar.SubqueryContext) interface{} {
	return v.Visit(ctx.Select_stmt_core()).(*tree.SelectStmt)
}

// VisitSubquery_expr is called when visiting a subquery_expr, return *tree.ExpressionSelect
func (v *KFSqliteVisitor) VisitSubquery_expr(ctx *sqlgrammar.Subquery_exprContext) interface{} {
	stmt := v.Visit(ctx.Subquery()).(*tree.SelectStmt)
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
func (v *KFSqliteVisitor) VisitWhen_clause(ctx *sqlgrammar.When_clauseContext) interface{} {
	var when = [2]tree.Expression{}
	when[0] = v.Visit(ctx.GetCondition()).(tree.Expression)
	when[1] = v.Visit(ctx.GetResult()).(tree.Expression)
	return when
}

// VisitCase_expr is called when visiting a case_expr, return *tree.ExpressionCase
func (v *KFSqliteVisitor) VisitCase_expr(ctx *sqlgrammar.Case_exprContext) interface{} {
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
func (v *KFSqliteVisitor) VisitFunction_call(ctx *sqlgrammar.Function_callContext) interface{} {
	expr := &tree.ExpressionFunction{
		Inputs: make([]tree.Expression, len(ctx.AllExpr())),
	}
	funcName := extractSQLName(ctx.Function_name().GetText())

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
func (v *KFSqliteVisitor) VisitFunction_expr(ctx *sqlgrammar.Function_exprContext) interface{} {
	expr := v.Visit(ctx.Function_call()).(*tree.ExpressionFunction)
	if ctx.Type_cast() != nil {
		expr.TypeCast = v.Visit(ctx.Type_cast()).(tree.TypeCastType)
	}
	return expr
}

// VisitExpr_list_expr is called when visiting a expr_list_expr, return *tree.ExpressionList
func (v *KFSqliteVisitor) VisitExpr_list_expr(ctx *sqlgrammar.Expr_list_exprContext) interface{} {
	return v.Visit(ctx.Expr_list()).(*tree.ExpressionList)
}

// VisitArithmetic_expr is called when visiting a arithmetic_expr, return *tree.ExpressionArithmetic
func (v *KFSqliteVisitor) VisitArithmetic_expr(ctx *sqlgrammar.Arithmetic_exprContext) interface{} {
	expr := &tree.ExpressionArithmetic{}
	expr.Left = v.Visit(ctx.GetLeft()).(tree.Expression)
	expr.Right = v.Visit(ctx.GetRight()).(tree.Expression)

	switch {
	case ctx.PLUS() != nil:
		expr.Operator = tree.ArithmeticOperatorAdd
	case ctx.MINUS() != nil:
		expr.Operator = tree.ArithmeticOperatorSubtract
	case ctx.STAR() != nil:
		expr.Operator = tree.ArithmeticOperatorMultiply
	case ctx.DIV() != nil:
		expr.Operator = tree.ArithmeticOperatorDivide
	case ctx.MOD() != nil:
		expr.Operator = tree.ArithmeticOperatorModulus
	default:
		panic(fmt.Sprintf("unknown arithmetic operator %s", ctx.GetText()))
	}

	return expr
}

// VisitIn_subquery_expr is called when visiting a in_suquery_expr, return *tree.ExpressionBinaryComparison
func (v *KFSqliteVisitor) VisitIn_subquery_expr(ctx *sqlgrammar.In_subquery_exprContext) interface{} {
	expr := &tree.ExpressionBinaryComparison{
		Left:     v.Visit(ctx.GetElem()).(tree.Expression),
		Operator: tree.ComparisonOperatorIn,
	}
	if ctx.NOT_() != nil {
		expr.Operator = tree.ComparisonOperatorNotIn
	}
	sub := v.Visit(ctx.Subquery()).(*tree.SelectStmt)
	expr.Right = &tree.ExpressionSelect{Select: sub}
	return expr
}

// VisitExpr_list is called when visiting a expr_list, return *tree.ExpressionList
func (v *KFSqliteVisitor) VisitExpr_list(ctx *sqlgrammar.Expr_listContext) interface{} {
	exprCount := len(ctx.AllExpr())
	exprs := make([]tree.Expression, exprCount)
	for i, exprCtx := range ctx.AllExpr() {
		exprs[i] = v.Visit(exprCtx).(tree.Expression)
	}
	return &tree.ExpressionList{Expressions: exprs}
}

// VisitIn_list_expr is called when visiting a in_list_expr, return *tree.ExpressionBinaryComparison
func (v *KFSqliteVisitor) VisitIn_list_expr(ctx *sqlgrammar.In_list_exprContext) interface{} {
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
func (v *KFSqliteVisitor) VisitBetween_expr(ctx *sqlgrammar.Between_exprContext) interface{} {
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
func (v *KFSqliteVisitor) VisitLike_expr(ctx *sqlgrammar.Like_exprContext) interface{} {
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
func (v *KFSqliteVisitor) VisitComparisonOperator(ctx *sqlgrammar.ComparisonOperatorContext) interface{} {
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
func (v *KFSqliteVisitor) VisitComparison_expr(ctx *sqlgrammar.Comparison_exprContext) interface{} {
	expr := &tree.ExpressionBinaryComparison{
		Left:     v.Visit(ctx.GetLeft()).(tree.Expression),
		Operator: v.Visit(ctx.ComparisonOperator()).(tree.ComparisonOperator),
		Right:    v.Visit(ctx.GetRight()).(tree.Expression),
	}

	return expr
}

// VisitBollean_value is called when visiting a boolean_value, return *tree.ExpressionLiteral
func (v *KFSqliteVisitor) VisitBoolean_value(ctx *sqlgrammar.Boolean_valueContext) interface{} {
	return &tree.ExpressionLiteral{
		Value: ctx.GetText(),
	}
}

// VisitIs_expr is called when visiting a is_expr, return *tree.ExpressionIs
func (v *KFSqliteVisitor) VisitIs_expr(ctx *sqlgrammar.Is_exprContext) interface{} {
	expr := &tree.ExpressionIs{
		Left: v.Visit(ctx.Expr(0)).(tree.Expression),
	}
	if ctx.NOT_() != nil {
		expr.Not = true
	}

	switch {
	case ctx.NULL_() != nil:
		expr.Right = &tree.ExpressionLiteral{Value: "NULL"}
	case ctx.Boolean_value() != nil:
		expr.Right = v.Visit(ctx.Boolean_value()).(tree.Expression)
	case ctx.DISTINCT_() != nil:
		expr.Right = v.Visit(ctx.Expr(1)).(tree.Expression)
		expr.Distinct = true
	default:
		panic(fmt.Sprintf("unknown IS expression %s", ctx.GetText()))
	}
	return expr
}

// VisitNull_expr is called when visiting a null_expr, return *tree.ExpressionIs
func (v *KFSqliteVisitor) VisitNull_expr(ctx *sqlgrammar.Null_exprContext) interface{} {
	expr := &tree.ExpressionIs{
		Left:  v.Visit(ctx.Expr()).(tree.Expression),
		Right: &tree.ExpressionLiteral{Value: "NULL"},
	}
	if ctx.NOTNULL_() != nil {
		expr.Not = true
	}
	return expr
}

// VisitLogical_not_expr is called when visiting a logical_not_expr, return *tree.ExpressionUnary
func (v *KFSqliteVisitor) VisitLogical_not_expr(ctx *sqlgrammar.Logical_not_exprContext) interface{} {
	return &tree.ExpressionUnary{
		Operator: tree.UnaryOperatorNot,
		Operand:  v.Visit(ctx.Expr()).(tree.Expression),
	}
}

// VisitLogical_binary_expr is called when visiting a logical_binary_expr, return *tree.ExpressionBinaryLogical
func (v *KFSqliteVisitor) VisitLogical_binary_expr(ctx *sqlgrammar.Logical_binary_exprContext) interface{} {
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
func (v *KFSqliteVisitor) VisitValues_clause(ctx *sqlgrammar.Values_clauseContext) interface{} {
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
func (v *KFSqliteVisitor) VisitUpsert_clause(ctx *sqlgrammar.Upsert_clauseContext) interface{} {
	clause := tree.Upsert{
		Type: tree.UpsertTypeDoNothing,
	}

	// conflict target
	conflictTarget := tree.ConflictTarget{}
	allIndexedColumnCtx := ctx.AllIndexed_column()
	indexedColumns := make([]string, len(allIndexedColumnCtx))
	for i, indexedColumnCtx := range allIndexedColumnCtx {
		indexedColumns[i] = extractSQLName(indexedColumnCtx.Column_name().GetText())
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
func (v *KFSqliteVisitor) VisitUpsert_update(ctx *sqlgrammar.Upsert_updateContext) interface{} {
	clause := tree.UpdateSetClause{}
	if ctx.Column_name_list() != nil {
		clause.Columns = v.Visit(ctx.Column_name_list()).([]string)
	} else {
		clause.Columns = []string{extractSQLName(ctx.Column_name().GetText())}
	}

	clause.Expression = v.Visit(ctx.Expr()).(tree.Expression)
	return &clause
}

// VisitColumn_name_list is called when visiting a column_name_list, return []string
func (v *KFSqliteVisitor) VisitColumn_name_list(ctx *sqlgrammar.Column_name_listContext) interface{} {
	names := make([]string, len(ctx.AllColumn_name()))
	for i, nameCtx := range ctx.AllColumn_name() {
		names[i] = extractSQLName(nameCtx.GetText())
	}
	return names
}

// VisitColumn_name is called when visiting a column_name, return string
func (v *KFSqliteVisitor) VisitColumn_name(ctx *sqlgrammar.Column_nameContext) interface{} {
	return extractSQLName(ctx.GetText())
}

// VisitReturning_clause is called when visiting a returning_clause, return *tree.ReturningClause
func (v *KFSqliteVisitor) VisitReturning_clause(ctx *sqlgrammar.Returning_clauseContext) interface{} {
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
			clause.Returned[i].Alias = extractSQLName(columnCtx.Column_alias().GetText())
		}

	}
	return &clause
}

// VisitUpdate_set_subclause is called when visiting a column_assign_subclause, return *tree.UpdateSetClause
func (v *KFSqliteVisitor) VisitUpdate_set_subclause(ctx *sqlgrammar.Update_set_subclauseContext) interface{} {
	result := tree.UpdateSetClause{}

	if ctx.Column_name_list() != nil {
		result.Columns = v.Visit(ctx.Column_name_list()).([]string)
	} else {
		result.Columns = []string{extractSQLName(ctx.Column_name().GetText())}
	}

	result.Expression = v.Visit(ctx.Expr()).(tree.Expression)
	return &result
}

// VisitQualified_table_name is called when visiting a qualified_table_name, return *tree.QualifiedTableName
func (v *KFSqliteVisitor) VisitQualified_table_name(ctx *sqlgrammar.Qualified_table_nameContext) interface{} {
	result := tree.QualifiedTableName{}

	result.TableName = extractSQLName(ctx.Table_name().GetText())

	if ctx.Table_alias() != nil {
		result.TableAlias = extractSQLName(ctx.Table_alias().GetText())
	}

	return &result
}

// VisitUpdate_stmt is called when visiting a update_stmt, return *tree.Update
func (v *KFSqliteVisitor) VisitUpdate_stmt(ctx *sqlgrammar.Update_stmtContext) interface{} {
	t := tree.Update{}
	var updateStmt tree.UpdateStmt

	if ctx.Common_table_stmt() != nil {
		t.CTE = v.Visit(ctx.Common_table_stmt()).([]*tree.CTE)
	}

	updateStmt.QualifiedTableName = v.Visit(ctx.Qualified_table_name()).(*tree.QualifiedTableName)

	updateStmt.UpdateSetClause = make([]*tree.UpdateSetClause, len(ctx.AllUpdate_set_subclause()))
	for i, subclauseCtx := range ctx.AllUpdate_set_subclause() {
		updateStmt.UpdateSetClause[i] = v.Visit(subclauseCtx).(*tree.UpdateSetClause)
	}

	if ctx.FROM_() != nil {
		fromClause := tree.FromClause{
			JoinClause: &tree.JoinClause{},
		}

		if ctx.Join_clause() != nil {
			fromClause.JoinClause = v.Visit(ctx.Join_clause()).(*tree.JoinClause)
		} else {
			// table_or_subquery
			fromClause.JoinClause.TableOrSubquery = v.Visit(ctx.Table_or_subquery()).(tree.TableOrSubquery)
		}

		updateStmt.From = &fromClause
	}

	if ctx.WHERE_() != nil {
		updateStmt.Where = v.Visit(ctx.Expr()).(tree.Expression)
	}

	if ctx.Returning_clause() != nil {
		updateStmt.Returning = v.Visit(ctx.Returning_clause()).(*tree.ReturningClause)
	}

	t.UpdateStmt = &updateStmt
	return &t
}

func (v *KFSqliteVisitor) VisitInsert_stmt(ctx *sqlgrammar.Insert_stmtContext) interface{} {
	t := tree.Insert{}
	var insertStmt tree.InsertStmt

	if ctx.Common_table_stmt() != nil {
		t.CTE = v.Visit(ctx.Common_table_stmt()).([]*tree.CTE)
	}

	insertStmt.InsertType = getInsertType(ctx)
	insertStmt.Table = extractSQLName(ctx.Table_name().GetText())
	if ctx.Table_alias() != nil {
		insertStmt.TableAlias = extractSQLName(ctx.Table_alias().GetText())
	}

	allColumnNameCtx := ctx.AllColumn_name()
	if len(allColumnNameCtx) > 0 {
		insertStmt.Columns = make([]string, len(allColumnNameCtx))
		for i, nc := range allColumnNameCtx {
			insertStmt.Columns[i] = extractSQLName(nc.GetText())
		}
	}

	insertStmt.Values = v.Visit(ctx.Values_clause()).([][]tree.Expression)
	if ctx.Upsert_clause() != nil {
		insertStmt.Upsert = v.Visit(ctx.Upsert_clause()).(*tree.Upsert)
	}
	if ctx.Returning_clause() != nil {
		insertStmt.ReturningClause = v.Visit(ctx.Returning_clause()).(*tree.ReturningClause)
	}

	t.InsertStmt = &insertStmt
	return &t
}

// VisitCompound_operator is called when visiting a compound_operator, return *tree.CompoundOperator
func (v *KFSqliteVisitor) VisitCompound_operator(ctx *sqlgrammar.Compound_operatorContext) interface{} {
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
func (v *KFSqliteVisitor) VisitOrdering_term(ctx *sqlgrammar.Ordering_termContext) interface{} {
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
func (v *KFSqliteVisitor) VisitOrder_by_stmt(ctx *sqlgrammar.Order_by_stmtContext) interface{} {
	count := len(ctx.AllOrdering_term())
	result := tree.OrderBy{OrderingTerms: make([]*tree.OrderingTerm, count)}

	for i, orderingTermCtx := range ctx.AllOrdering_term() {
		result.OrderingTerms[i] = v.Visit(orderingTermCtx).(*tree.OrderingTerm)
	}

	return &result
}

// VisitLimit_stmt is called when visiting a limit_stmt, return *tree.Limit
func (v *KFSqliteVisitor) VisitLimit_stmt(ctx *sqlgrammar.Limit_stmtContext) interface{} {
	result := tree.Limit{
		Expression: v.Visit(ctx.Expr(0)).(tree.Expression),
	}

	if ctx.OFFSET_() != nil {
		result.Offset = v.Visit(ctx.Expr(1)).(tree.Expression)
	}

	return &result
}

// VisitTable_or_subquery is called when visiting a table_or_subquery, return tree.TableOrSubquery
func (v *KFSqliteVisitor) VisitTable_or_subquery(ctx *sqlgrammar.Table_or_subqueryContext) interface{} {
	switch {
	case ctx.Table_name() != nil:
		t := tree.TableOrSubqueryTable{
			Name: extractSQLName(ctx.Table_name().GetText()),
		}
		if ctx.Table_alias() != nil {
			t.Alias = extractSQLName(ctx.Table_alias().GetText())
		}
		return &t
	case ctx.Select_stmt_core() != nil:
		t := tree.TableOrSubquerySelect{
			Select: v.Visit(ctx.Select_stmt_core()).(*tree.SelectStmt),
		}
		if ctx.Table_alias() != nil {
			t.Alias = extractSQLName(ctx.Table_alias().GetText())
		}
		return &t
	}
	return nil
}

// VisitJoin_operator is called when visiting a join_operator, return *tree.JoinOperator
func (v *KFSqliteVisitor) VisitJoin_operator(ctx *sqlgrammar.Join_operatorContext) interface{} {
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

// VisitJoin_clause is called when visiting a join_clause, return *tree.JoinClause
func (v *KFSqliteVisitor) VisitJoin_clause(ctx *sqlgrammar.Join_clauseContext) interface{} {
	clause := tree.JoinClause{}

	// just table_or_subquery
	clause.TableOrSubquery = v.Visit(ctx.Table_or_subquery(0)).(tree.TableOrSubquery)
	if len(ctx.AllTable_or_subquery()) == 1 {
		return &clause
	}

	// with joins
	joins := make([]*tree.JoinPredicate, len(ctx.AllJoin_operator()))
	for i, subCtx := range ctx.AllJoin_operator() {
		jp := tree.JoinPredicate{}
		jp.JoinOperator = v.Visit(subCtx).(*tree.JoinOperator)
		jp.Table = v.Visit(ctx.Table_or_subquery(i + 1)).(tree.TableOrSubquery)
		jp.Constraint = v.Visit(ctx.Join_constraint(i).Expr()).(tree.Expression)
		joins[i] = &jp
	}
	clause.Joins = joins

	return &clause
}

// VisitResult_column is called when visiting a result_column, return tree.ResultColumn
func (v *KFSqliteVisitor) VisitResult_column(ctx *sqlgrammar.Result_columnContext) interface{} {
	switch {
	// table_name need to be checked first
	case ctx.Table_name() != nil:
		return &tree.ResultColumnTable{
			TableName: extractSQLName(ctx.Table_name().GetText()),
		}
	case ctx.STAR() != nil:
		return &tree.ResultColumnStar{}
	case ctx.Expr() != nil:
		r := &tree.ResultColumnExpression{
			Expression: v.Visit(ctx.Expr()).(tree.Expression),
		}
		if ctx.Column_alias() != nil {
			r.Alias = extractSQLName(ctx.Column_alias().GetText())
		}
		return r
	}

	return nil
}

// VisitDelete_stmt is called when visiting a delete_stmt, return *tree.Delete
func (v *KFSqliteVisitor) VisitDelete_stmt(ctx *sqlgrammar.Delete_stmtContext) interface{} {
	t := tree.Delete{}

	if ctx.Common_table_stmt() != nil {
		t.CTE = v.Visit(ctx.Common_table_stmt()).([]*tree.CTE)
	}

	stmt := tree.DeleteStmt{}
	stmt.QualifiedTableName = v.Visit(ctx.Qualified_table_name()).(*tree.QualifiedTableName)

	if ctx.WHERE_() != nil {
		stmt.Where = v.Visit(ctx.Expr()).(tree.Expression)
	}

	if ctx.Returning_clause() != nil {
		stmt.Returning = v.Visit(ctx.Returning_clause()).(*tree.ReturningClause)
	}

	t.DeleteStmt = &stmt
	return &t
}

// VisitSelect_core is called when visiting a select_core, return *tree.SelectCore
func (v *KFSqliteVisitor) VisitSelect_core(ctx *sqlgrammar.Select_coreContext) interface{} {
	t := tree.SelectCore{
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
		fromClause := tree.FromClause{
			JoinClause: &tree.JoinClause{},
		}

		if ctx.Join_clause() != nil {
			fromClause.JoinClause = v.Visit(ctx.Join_clause()).(*tree.JoinClause)
		} else {
			// table_or_subquery
			fromClause.JoinClause.TableOrSubquery = v.Visit(ctx.Table_or_subquery()).(tree.TableOrSubquery)

			// with comma(cartesian) join
			//if len(ctx.AllTable_or_subquery()) == 1 {
			//	fromClause.JoinClause.TableOrSubquery = v.Visit(ctx.Table_or_subquery(0)).(tree.TableOrSubquery)
			//} else {
			//	//tos := make([]tree.TableOrSubquery, len(ctx.AllTable_or_subquery()))
			//	//
			//	//for i, tableOrSubqueryCtx := range ctx.AllTable_or_subquery() {
			//	//	tos[i] = v.Visit(tableOrSubqueryCtx).(tree.TableOrSubquery)
			//	//}
			//	//
			//	//fromClause.JoinClause.TableOrSubquery = &tree.TableOrSubqueryList{
			//	//	TableOrSubqueries: tos,
			//	//}
			//	panic("not support comma(cartesian) join")
			//}
		}

		t.From = &fromClause
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

// VisitSelect_stmt_core is called when visiting a select_stmt_core, return *tree.SelectStmt
func (v *KFSqliteVisitor) VisitSelect_stmt_core(ctx *sqlgrammar.Select_stmt_coreContext) interface{} {
	t := tree.SelectStmt{}
	selectCores := make([]*tree.SelectCore, len(ctx.AllSelect_core()))

	// first select_core
	selectCores[0] = v.Visit(ctx.Select_core(0)).(*tree.SelectCore)

	// rest select_core
	for i, selectCoreCtx := range ctx.AllSelect_core()[1:] {
		compoundOperator := v.Visit(ctx.Compound_operator(i)).(*tree.CompoundOperator)
		core := v.Visit(selectCoreCtx).(*tree.SelectCore)
		core.Compound = compoundOperator
		selectCores[i+1] = core
	}

	t.SelectCores = selectCores

	if ctx.Order_by_stmt() != nil {
		t.OrderBy = v.Visit(ctx.Order_by_stmt()).(*tree.OrderBy)
	}

	if ctx.Limit_stmt() != nil {
		t.Limit = v.Visit(ctx.Limit_stmt()).(*tree.Limit)
	}

	return &t
}

// VisitSelect_stmt is called when visiting a select_stmt, return *tree.Select
func (v *KFSqliteVisitor) VisitSelect_stmt(ctx *sqlgrammar.Select_stmtContext) interface{} {
	t := tree.Select{}

	if ctx.Common_table_stmt() != nil {
		t.CTE = v.Visit(ctx.Common_table_stmt()).([]*tree.CTE)
	}

	t.SelectStmt = v.Visit(ctx.Select_stmt_core()).(*tree.SelectStmt)
	return &t
}

func (v *KFSqliteVisitor) VisitSql_stmt_list(ctx *sqlgrammar.Sql_stmt_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *KFSqliteVisitor) VisitSql_stmt(ctx *sqlgrammar.Sql_stmtContext) interface{} {
	// Sql_stmtContext will only have one stmt
	return v.VisitChildren(ctx).([]tree.Ast)[0]
}

// VisitStatements is called first by Visitor.Visit
func (v *KFSqliteVisitor) VisitStatements(ctx *sqlgrammar.StatementsContext) interface{} {
	// ParseContext will only have one Sql_stmt_listContext
	sqlStmtListContext := ctx.Sql_stmt_list(0)
	return v.VisitChildren(sqlStmtListContext).([]tree.Ast)
}

// Visit dispatch to the visit method of the ctx
// e.g. if the tree is a ParseContext, then dispatch call VisitParse.
// Overwrite is needed,
// refer to https://github.com/antlr/antlr4/pull/1841#issuecomment-576791512
func (v *KFSqliteVisitor) Visit(parseTree antlr.ParseTree) interface{} {
	if v.trace {
		fmt.Printf("visit tree: %v, %s\n", reflect.TypeOf(parseTree), parseTree.GetText())
	}
	return parseTree.Accept(v)
}

// VisitChildren visits the children of the specified node.
// Overwrite is needed,
// refer to https://github.com/antlr/antlr4/pull/1841#issuecomment-576791512
// calling function need to convert the result to asts
func (v *KFSqliteVisitor) VisitChildren(node antlr.RuleNode) interface{} {
	var result []tree.Ast
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
		childResult := v.Visit(c).(tree.Ast)
		result = append(result, childResult)
	}
	return result
}

func (v *KFSqliteVisitor) shouldVisitNextChild(node antlr.Tree, currentResult interface{}) bool {
	if _, ok := node.(antlr.TerminalNode); ok {
		return false
	}

	return true
}

// extractSQLName remove surrounding lexical token(pair) of an identifier(name).
// Those tokens are: `"` and `[` `]` and "`".
// In sqlparser identifiers are used for: table name, table alias name, column name,
// column alias name, collation name, index name, function name.
func extractSQLName(name string) string {
	// remove surrounding token pairs
	if len(name) > 1 {
		if name[0] == '"' && name[len(name)-1] == '"' {
			name = name[1 : len(name)-1]
		}

		if name[0] == '[' && name[len(name)-1] == ']' {
			name = name[1 : len(name)-1]
		}

		if name[0] == '`' && name[len(name)-1] == '`' {
			name = name[1 : len(name)-1]
		}
	}

	return name
}
