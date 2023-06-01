package sql_parser

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/antlr/antlr4/runtime/Go/antlr/v4"
	"github.com/kwilteam/kwil-db/pkg/engine/tree"
	grammar "github.com/kwilteam/kwil-db/pkg/sql_parser/sql_grammar"
)

// KFSqliteVisitor is visitor that visit Antlr parsed tree and returns the AST.
type KFSqliteVisitor struct {
	grammar.BaseSQLiteParserVisitor
	ErrorHandler

	trace bool
}

type KFSqliteVisitorOption func(*KFSqliteVisitor)

func KFVisitorWithTrace() KFSqliteVisitorOption {
	return func(l *KFSqliteVisitor) {
		l.trace = true
	}
}

var _ grammar.SQLiteParserVisitor = &KFSqliteVisitor{}

func NewKFSqliteVisitor(eh *ErrorHandler, opts ...KFSqliteVisitorOption) *KFSqliteVisitor {
	k := &KFSqliteVisitor{
		ErrorHandler: *eh,
		trace:        false,
	}
	for _, opt := range opts {
		opt(k)
	}
	return k
}

// VisitCommon_table_expression is called when visiting a common_table_expression, return *tree.CTE
func (v *KFSqliteVisitor) VisitCommon_table_expression(ctx *grammar.Common_table_expressionContext) interface{} {
	cte := tree.CTE{}

	// cte_table_name
	cteTableCtx := ctx.Cte_table_name()
	cte.Table = cteTableCtx.Table_name().GetText()
	if len(cteTableCtx.AllColumn_name()) > 0 {
		cte.Columns = make([]string, len(cteTableCtx.AllColumn_name()))
		for i, colNameCtx := range cteTableCtx.AllColumn_name() {
			cte.Columns[i] = colNameCtx.GetText()
		}
	}

	selectStmtCoreCtx := ctx.Select_stmt_core()
	cte.Select = v.Visit(selectStmtCoreCtx).(*tree.SelectStmt)
	return &cte
}

// VisitCommon_table_stmt is called when visiting a common_table_stmt, return []*tree.CTE.
func (v *KFSqliteVisitor) VisitCommon_table_stmt(ctx *grammar.Common_table_stmtContext) interface{} {
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

func getInsertType(ctx *grammar.Insert_stmtContext) tree.InsertType {
	var insertType tree.InsertType
	if ctx.OR_() != nil {
		switch {
		case ctx.REPLACE_() != nil:
			insertType = tree.InsertTypeInsertOrReplace
		}
	} else {
		if ctx.REPLACE_() != nil {
			insertType = tree.InsertTypeReplace
		} else {
			insertType = tree.InsertTypeInsert
		}
	}

	return insertType
}

func (v *KFSqliteVisitor) visitExprList(exprList []grammar.IExprContext) *tree.ExpressionList {
	exprCount := len(exprList)
	exprs := make([]tree.Expression, exprCount)
	for i, exprCtx := range exprList {
		exprs[i] = v.visitExpr(exprCtx)
	}
	return &tree.ExpressionList{Expressions: exprs}
}

// VisitExpr is called when visiting an expression, return tree.Expression.
func (v *KFSqliteVisitor) VisitExpr(ctx *grammar.ExprContext) interface{} {
	return v.visitExpr(ctx)
}

func (v *KFSqliteVisitor) getCollateType(collationName string) tree.CollationType {
	switch strings.ToLower(collationName) {
	case "binary":
		return tree.CollationTypeBinary
	case "nocase":
		return tree.CollationTypeNoCase
	case "rtrim":
		return tree.CollationTypeRTrim
	default:
		panic(fmt.Sprintf("unknown collation type %s", collationName))
	}
}

func (v *KFSqliteVisitor) visitExpr(ctx grammar.IExprContext) tree.Expression {
	if ctx == nil {
		return nil
	}

	// order is important, map to expr definition in Antlr grammar(not exactly)
	switch {
	// primary expressions
	case ctx.Literal_value() != nil:
		return &tree.ExpressionLiteral{Value: ctx.Literal_value().GetText()}
	case ctx.BIND_PARAMETER() != nil:
		return &tree.ExpressionBindParameter{Parameter: ctx.BIND_PARAMETER().GetText()}
	case ctx.Table_name() != nil || ctx.Column_name() != nil:
		expr := &tree.ExpressionColumn{}
		if ctx.Table_name() != nil {
			expr.Table = ctx.Table_name().GetText()
		}
		if ctx.Column_name() != nil {
			expr.Column = ctx.Column_name().GetText()
		}
		return expr
	case ctx.Select_stmt_core() != nil && ctx.IN_() == nil:
		// select_stmt_core not in IN
		stmt := v.Visit(ctx.Select_stmt_core()).(*tree.SelectStmt)
		expr := &tree.ExpressionSelect{
			IsNot:    false,
			IsExists: false,
			Select:   stmt,
		}
		if ctx.NOT_() != nil {
			expr.IsNot = true
		}
		if ctx.EXISTS_() != nil {
			expr.IsExists = true
		}
		return expr
	case ctx.GetElevate_expr() != nil:
		expr := v.visitExpr(ctx.GetElevate_expr())
		switch t := expr.(type) {
		case *tree.ExpressionLiteral:
			t.Wrapped = true
		case *tree.ExpressionBindParameter:
			t.Wrapped = true
		case *tree.ExpressionColumn:
			t.Wrapped = true
		case *tree.ExpressionUnary:
			t.Wrapped = true
		case *tree.ExpressionBinaryComparison:
			t.Wrapped = true
		case *tree.ExpressionFunction:
			t.Wrapped = true
		case *tree.ExpressionList:
			t.Wrapped = true
		case *tree.ExpressionCollate:
			t.Wrapped = true
		case *tree.ExpressionStringCompare:
			t.Wrapped = true
		case *tree.ExpressionIsNull:
			t.Wrapped = true
		case *tree.ExpressionDistinct:
			t.Wrapped = true
		case *tree.ExpressionBetween:
			t.Wrapped = true
		case *tree.ExpressionSelect:
			t.Wrapped = true
		case *tree.ExpressionCase:
			t.Wrapped = true
		default:
			panic(fmt.Sprintf("unknown expression type %T", expr))
		}
		return expr
	// unary operators
	case ctx.MINUS() != nil && ctx.GetUnary_expr() != nil:
		return &tree.ExpressionUnary{
			Operator: tree.UnaryOperatorMinus,
			Operand:  v.visitExpr(ctx.GetUnary_expr()),
		}
	case ctx.PLUS() != nil && ctx.GetUnary_expr() != nil:
		return &tree.ExpressionUnary{
			Operator: tree.UnaryOperatorPlus,
			Operand:  v.visitExpr(ctx.GetUnary_expr()),
		}
	case ctx.TILDE() != nil && ctx.GetUnary_expr() != nil:
		return &tree.ExpressionUnary{
			Operator: tree.UnaryOperatorBitNot,
			Operand:  v.visitExpr(ctx.GetUnary_expr()),
		}
	case ctx.COLLATE_() != nil:
		// collation_name is any_name
		collationName := ctx.Collation_name().GetText()
		return &tree.ExpressionCollate{
			Expression: v.visitExpr(ctx.Expr(0)),
			Collation:  v.getCollateType(collationName),
		}
	// binary opertors
	//case ctx.PIPE2() != nil: // TODO: ??
	//	return &tree.ExpressionBinaryComparison{
	//		Left:     v.visitExpr(ctx.Expr(0)),
	//		Right:    v.visitExpr(ctx.Expr(1)),
	//		Operator: tree.BitwiseOperatorBitwiseOr,
	//	}
	case ctx.STAR() != nil:
		return &tree.ExpressionArithmetic{
			Left:     v.visitExpr(ctx.Expr(0)),
			Right:    v.visitExpr(ctx.Expr(1)),
			Operator: tree.ArithmeticOperatorMultiply,
		}
	case ctx.DIV() != nil:
		return &tree.ExpressionArithmetic{
			Left:     v.visitExpr(ctx.Expr(0)),
			Right:    v.visitExpr(ctx.Expr(1)),
			Operator: tree.ArithmeticOperatorDivide,
		}
	case ctx.MOD() != nil:
		return &tree.ExpressionArithmetic{
			Left:     v.visitExpr(ctx.Expr(0)),
			Right:    v.visitExpr(ctx.Expr(1)),
			Operator: tree.ArithmeticOperatorModulus,
		}
	case ctx.PLUS() != nil:
		return &tree.ExpressionArithmetic{
			Left:     v.visitExpr(ctx.Expr(0)),
			Right:    v.visitExpr(ctx.Expr(1)),
			Operator: tree.ArithmeticOperatorAdd,
		}
	case ctx.MINUS() != nil:
		return &tree.ExpressionArithmetic{
			Left:     v.visitExpr(ctx.Expr(0)),
			Right:    v.visitExpr(ctx.Expr(1)),
			Operator: tree.ArithmeticOperatorSubtract,
		}
	case ctx.LT2() != nil:
		return &tree.ExpressionBinaryComparison{
			Left:     v.visitExpr(ctx.Expr(0)),
			Right:    v.visitExpr(ctx.Expr(1)),
			Operator: tree.BitwiseOperatorLeftShift,
		}
	case ctx.GT2() != nil:
		return &tree.ExpressionBinaryComparison{
			Left:     v.visitExpr(ctx.Expr(0)),
			Right:    v.visitExpr(ctx.Expr(1)),
			Operator: tree.BitwiseOperatorRightShift,
		}
	case ctx.AMP() != nil:
		return &tree.ExpressionBinaryComparison{
			Left:     v.visitExpr(ctx.Expr(0)),
			Right:    v.visitExpr(ctx.Expr(1)),
			Operator: tree.BitwiseOperatorAnd,
		}
	case ctx.PIPE() != nil:
		return &tree.ExpressionBinaryComparison{
			Left:     v.visitExpr(ctx.Expr(0)),
			Right:    v.visitExpr(ctx.Expr(1)),
			Operator: tree.BitwiseOperatorOr,
		}
	case ctx.LT() != nil:
		return &tree.ExpressionBinaryComparison{
			Left:     v.visitExpr(ctx.Expr(0)),
			Right:    v.visitExpr(ctx.Expr(1)),
			Operator: tree.ComparisonOperatorLessThan,
		}
	case ctx.LT_EQ() != nil:
		return &tree.ExpressionBinaryComparison{
			Left:     v.visitExpr(ctx.Expr(0)),
			Right:    v.visitExpr(ctx.Expr(1)),
			Operator: tree.ComparisonOperatorLessThanOrEqual,
		}
	case ctx.GT() != nil:
		return &tree.ExpressionBinaryComparison{
			Left:     v.visitExpr(ctx.Expr(0)),
			Right:    v.visitExpr(ctx.Expr(1)),
			Operator: tree.ComparisonOperatorGreaterThan,
		}
	case ctx.GT_EQ() != nil:
		return &tree.ExpressionBinaryComparison{
			Left:     v.visitExpr(ctx.Expr(0)),
			Right:    v.visitExpr(ctx.Expr(1)),
			Operator: tree.ComparisonOperatorGreaterThanOrEqual,
		}
	case ctx.ASSIGN() != nil:
		return &tree.ExpressionBinaryComparison{
			Left:     v.visitExpr(ctx.Expr(0)),
			Right:    v.visitExpr(ctx.Expr(1)),
			Operator: tree.ComparisonOperatorEqual,
		}
	//case ctx.EQ() != nil:
	//	return &tree.ExpressionBinaryComparison{
	//		Left:     v.visitExpr(ctx.Expr(0)),
	//		Right:    v.visitExpr(ctx.Expr(1)), // TODO: equal
	//		Operator: tree.ComparisonOperatorEqual,
	//	}
	case ctx.NOT_EQ1() != nil:
		return &tree.ExpressionBinaryComparison{
			Left:     v.visitExpr(ctx.Expr(0)),
			Right:    v.visitExpr(ctx.Expr(1)),
			Operator: tree.ComparisonOperatorNotEqual,
		}
	//case ctx.NOT_EQ2() != nil:
	//	return &tree.ExpressionBinaryComparison{
	//		Left:     v.visitExpr(ctx.Expr(0)),
	//		Right:    v.visitExpr(ctx.Expr(1)),
	//		Operator: tree.ComparisonOperatorNotEqual2,
	//	}
	case ctx.IS_() != nil:
		if ctx.DISTINCT_() == nil {
			// binary comparison
			expr := &tree.ExpressionBinaryComparison{
				Left:     v.visitExpr(ctx.Expr(0)),
				Right:    v.visitExpr(ctx.Expr(1)),
				Operator: tree.ComparisonOperatorIs,
			}
			if ctx.NOT_() != nil {
				expr.Operator = tree.ComparisonOperatorIsNot
			}
			return expr
		}

		// distinct comparison
		expr := &tree.ExpressionDistinct{
			Left:  v.visitExpr(ctx.Expr(0)),
			Right: v.visitExpr(ctx.Expr(1)),
		}
		if ctx.NOT_() != nil {
			expr.IsNot = true
		}
		return expr
	case ctx.IN_() != nil:
		expr := &tree.ExpressionBinaryComparison{
			Left:     v.visitExpr(ctx.Expr(0)),
			Operator: tree.ComparisonOperatorIn,
		}

		if ctx.NOT_() != nil {
			expr.Operator = tree.ComparisonOperatorNotIn
		}

		if ctx.OPEN_PAR() != nil {
			// in follows by expr list
			exprCount := len(ctx.AllExpr())
			exprs := make([]tree.Expression, exprCount-1)
			for i, e := range ctx.AllExpr()[1:] {
				exprs[i] = v.visitExpr(e)
			}
			expr.Right = &tree.ExpressionList{
				Expressions: exprs,
			}
		} else {
			// in follows by expr(potentially expr list)
			expr.Right = v.visitExpr(ctx.Expr(1))
		}
		return expr
	// string comparison
	case ctx.LIKE_() != nil:
		expr := &tree.ExpressionStringCompare{
			Left:     v.visitExpr(ctx.Expr(0)),
			Operator: tree.StringOperatorLike,
			Right:    v.visitExpr(ctx.Expr(1)),
		}
		if ctx.NOT_() != nil {
			expr.Operator = tree.StringOperatorNotLike
		}
		if ctx.ESCAPE_() != nil {
			expr.Escape = v.visitExpr(ctx.Expr(2))
		}
		return expr
	case ctx.REGEXP_() != nil:
		expr := &tree.ExpressionStringCompare{
			Left:     v.visitExpr(ctx.Expr(0)),
			Operator: tree.StringOperatorRegexp,
			Right:    v.visitExpr(ctx.Expr(1)),
		}
		if ctx.NOT_() != nil {
			expr.Operator = tree.StringOperatorNotRegexp
		}
		return expr
	case ctx.MATCH_() != nil:
		expr := &tree.ExpressionStringCompare{
			Left:     v.visitExpr(ctx.Expr(0)),
			Operator: tree.StringOperatorMatch,
			Right:    v.visitExpr(ctx.Expr(1)),
		}
		if ctx.NOT_() != nil {
			expr.Operator = tree.StringOperatorNotMatch
		}
		return expr
	case ctx.GLOB_() != nil:
		expr := &tree.ExpressionStringCompare{
			Left:     v.visitExpr(ctx.Expr(0)),
			Operator: tree.StringOperatorGlob,
			Right:    v.visitExpr(ctx.Expr(1)),
		}
		if ctx.NOT_() != nil {
			expr.Operator = tree.StringOperatorNotGlob
		}
		return expr
	case ctx.BETWEEN_() != nil:
		expr := &tree.ExpressionBetween{
			Expression: v.visitExpr(ctx.Expr(0)),
			Left:       v.visitExpr(ctx.Expr(1)),
			Right:      v.visitExpr(ctx.Expr(2)),
		}
		if ctx.NOT_() != nil {
			expr.NotBetween = true
		}
		return expr
	// null
	case ctx.ISNULL_() != nil:
		return &tree.ExpressionIsNull{
			Expression: v.visitExpr(ctx.Expr(0)),
			IsNull:     true,
		}
	case ctx.NOTNULL_() != nil:
		return &tree.ExpressionIsNull{
			Expression: v.visitExpr(ctx.Expr(0)),
			IsNull:     false,
		}
	case ctx.NULL_() != nil && ctx.NOT_() != nil:
		return &tree.ExpressionIsNull{
			Expression: v.visitExpr(ctx.Expr(0)),
			IsNull:     false,
		}
	// not unary op
	case ctx.NOT_() != nil && ctx.GetUnary_expr() != nil:
		return &tree.ExpressionUnary{
			Operator: tree.UnaryOperatorNot,
			Operand:  v.visitExpr(ctx.GetUnary_expr()),
		}
	case ctx.AND_() != nil:
		return &tree.ExpressionBinaryComparison{
			Left:     v.visitExpr(ctx.Expr(0)),
			Operator: tree.LogicalOperatorAnd,
			Right:    v.visitExpr(ctx.Expr(1)),
		}
	case ctx.OR_() != nil:
		return &tree.ExpressionBinaryComparison{
			Left:     v.visitExpr(ctx.Expr(0)),
			Operator: tree.LogicalOperatorOr,
			Right:    v.visitExpr(ctx.Expr(1)),
		}
	case ctx.GetExpr_list() != nil:
		return v.visitExprList(ctx.AllExpr())
	case ctx.Function_name() != nil:
		expr := &tree.ExpressionFunction{
			Inputs: make([]tree.Expression, len(ctx.AllExpr())),
		}
		funcName := ctx.Function_name().GetText()

		f, ok := tree.SQLFunctions[strings.ToLower(funcName)]
		if !ok {
			panic(fmt.Sprintf("unsupported function '%s'", funcName))
		}
		expr.Function = f

		for i, e := range ctx.AllExpr() {
			expr.Inputs[i] = v.visitExpr(e)
		}
		return expr
	case ctx.CASE_() != nil:
		whenExprCount := len(ctx.GetWhen_expr())
		expr := &tree.ExpressionCase{
			WhenThenPairs: make([][2]tree.Expression, whenExprCount),
		}
		for i := 0; i < whenExprCount; i++ {
			expr.WhenThenPairs[i][0] = v.visitExpr(ctx.GetWhen_expr()[i])
			expr.WhenThenPairs[i][1] = v.visitExpr(ctx.GetThen_expr()[i])
		}

		if ctx.GetCase_expr() != nil {
			expr.CaseExpression = v.visitExpr(ctx.GetCase_expr())
		}

		if ctx.GetElse_expr() != nil {
			expr.ElseExpression = v.visitExpr(ctx.GetElse_expr())
		}
		return expr
	default:
		panic(fmt.Sprintf("cannot recognize expr '%s'", ctx.GetText()))
	}
}

// VisitValues_clause is called when visiting a values_clause, return [][]tree.Expression
func (v *KFSqliteVisitor) VisitValues_clause(ctx *grammar.Values_clauseContext) interface{} {
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
func (v *KFSqliteVisitor) VisitUpsert_clause(ctx *grammar.Upsert_clauseContext) interface{} {
	clause := tree.Upsert{
		Type: tree.UpsertTypeDoNothing,
	}

	// conflict target
	conflictTarget := tree.ConflictTarget{}
	allIndexedColumnCtx := ctx.AllIndexed_column()
	indexedColumns := make([]string, len(allIndexedColumnCtx))
	for i, indexedColumnCtx := range allIndexedColumnCtx {
		indexedColumns[i] = indexedColumnCtx.Column_name().GetText()
	}
	conflictTarget.IndexedColumns = indexedColumns

	if ctx.GetTarget_expr() != nil {
		conflictTarget.Where = v.visitExpr(ctx.GetTarget_expr())
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
		clause.Where = v.visitExpr(ctx.GetUpdate_expr())
	}
	return &clause
}

// VisitUpsert_update is called when visiting a upsert_update, return *tree.UpdateSetClause
func (v *KFSqliteVisitor) VisitUpsert_update(ctx *grammar.Upsert_updateContext) interface{} {
	clause := tree.UpdateSetClause{}
	if ctx.Column_name_list() != nil {
		clause.Columns = v.Visit(ctx.Column_name_list()).([]string)
	} else {
		clause.Columns = []string{ctx.Column_name().GetText()}
	}

	clause.Expression = v.Visit(ctx.Expr()).(tree.Expression)
	return &clause
}

// VisitColumn_name_list is called when visiting a column_name_list, return []string
func (v *KFSqliteVisitor) VisitColumn_name_list(ctx *grammar.Column_name_listContext) interface{} {
	names := make([]string, len(ctx.AllColumn_name()))
	for i, nameCtx := range ctx.AllColumn_name() {
		names[i] = nameCtx.GetText()
	}
	return names
}

// VisitColumn_name is called when visiting a column_name, return string
func (v *KFSqliteVisitor) VisitColumn_name(ctx *grammar.Column_nameContext) interface{} {
	return ctx.GetText()
}

// VisitReturning_clause is called when visiting a returning_clause, return *tree.ReturningClause
func (v *KFSqliteVisitor) VisitReturning_clause(ctx *grammar.Returning_clauseContext) interface{} {
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
			clause.Returned[i].Alias = columnCtx.Column_alias().GetText()
		}

	}
	return &clause
}

// VisitUpdate_set_subclause is called when visiting a column_assign_subclause, return *tree.UpdateSetClause
func (v *KFSqliteVisitor) VisitUpdate_set_subclause(ctx *grammar.Update_set_subclauseContext) interface{} {
	result := tree.UpdateSetClause{}

	if ctx.Column_name_list() != nil {
		result.Columns = v.Visit(ctx.Column_name_list()).([]string)
	} else {
		result.Columns = []string{ctx.Column_name().GetText()}
	}

	result.Expression = v.Visit(ctx.Expr()).(tree.Expression)
	return &result
}

// VisitQualified_table_name is called when visiting a qualified_table_name, return *tree.QualifiedTableName
func (v *KFSqliteVisitor) VisitQualified_table_name(ctx *grammar.Qualified_table_nameContext) interface{} {
	result := tree.QualifiedTableName{}

	result.TableName = ctx.Table_name().GetText()

	if ctx.Alias() != nil {
		result.TableAlias = ctx.Alias().GetText()
	}

	if ctx.INDEXED_() != nil {
		if ctx.NOT_() != nil {
			result.NotIndexed = true
		} else {
			result.IndexedBy = ctx.Index_name().GetText()
		}
	}

	return &result
}

// VisitUpdate_stmt is called when visiting a update_stmt, return *tree.Update
func (v *KFSqliteVisitor) VisitUpdate_stmt(ctx *grammar.Update_stmtContext) interface{} {
	t := tree.Update{}
	var updateStmt tree.UpdateStmt

	if ctx.Common_table_stmt() != nil {
		t.CTE = v.Visit(ctx.Common_table_stmt()).([]*tree.CTE)
	}

	if ctx.OR_() != nil {
		switch {
		case ctx.ROLLBACK_() != nil:
			updateStmt.Or = tree.UpdateOrRollback
		case ctx.ABORT_() != nil:
			updateStmt.Or = tree.UpdateOrAbort
		case ctx.REPLACE_() != nil:
			updateStmt.Or = tree.UpdateOrReplace
		case ctx.FAIL_() != nil:
			updateStmt.Or = tree.UpdateOrFail
		case ctx.IGNORE_() != nil:
			updateStmt.Or = tree.UpdateOrIgnore
		}
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

func (v *KFSqliteVisitor) VisitInsert_stmt(ctx *grammar.Insert_stmtContext) interface{} {
	t := tree.Insert{}
	var insertStmt tree.InsertStmt

	if ctx.Common_table_stmt() != nil {
		t.CTE = v.Visit(ctx.Common_table_stmt()).([]*tree.CTE)
	}

	insertStmt.InsertType = getInsertType(ctx)
	insertStmt.Table = ctx.Table_name().GetText()
	if ctx.Table_alias() != nil {
		insertStmt.TableAlias = ctx.Table_alias().GetText()
	}

	allColumnNameCtx := ctx.AllColumn_name()
	if len(allColumnNameCtx) > 0 {
		insertStmt.Columns = make([]string, len(allColumnNameCtx))
		for i, nc := range allColumnNameCtx {
			insertStmt.Columns[i] = nc.GetText()
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
func (v *KFSqliteVisitor) VisitCompound_operator(ctx *grammar.Compound_operatorContext) interface{} {
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
func (v *KFSqliteVisitor) VisitOrdering_term(ctx *grammar.Ordering_termContext) interface{} {
	result := tree.OrderingTerm{}

	// @yaiba NOTE: antlr will treat expr as a `expr collate collation_name` expression if COLLATE is present
	// then `COLLATE_()` will be in ctx.Expr() ctx
	// then the returned expression will be tree.ExpressionCollate
	if ctx.Expr().COLLATE_() != nil {
		collateExpr := v.Visit(ctx.Expr()).(tree.Expression)
		e, ok := collateExpr.(*tree.ExpressionCollate)
		if ok {
			result.Expression = e.Expression
			result.Collation = e.Collation
		} else {
			panic("parse COLLATE failed in ordering_term")
		}
	} else {
		result.Expression = v.Visit(ctx.Expr()).(tree.Expression)
	}

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
func (v *KFSqliteVisitor) VisitOrder_by_stmt(ctx *grammar.Order_by_stmtContext) interface{} {
	count := len(ctx.AllOrdering_term())
	result := tree.OrderBy{OrderingTerms: make([]*tree.OrderingTerm, count)}

	for i, orderingTermCtx := range ctx.AllOrdering_term() {
		result.OrderingTerms[i] = v.Visit(orderingTermCtx).(*tree.OrderingTerm)
	}

	return &result
}

// VisitLimit_stmt is called when visiting a limit_stmt, return *tree.Limit
func (v *KFSqliteVisitor) VisitLimit_stmt(ctx *grammar.Limit_stmtContext) interface{} {
	result := tree.Limit{
		Expression: v.Visit(ctx.Expr(0)).(tree.Expression),
	}

	if ctx.OFFSET_() != nil {
		result.Offset = v.Visit(ctx.Expr(1)).(tree.Expression)
	}

	if ctx.COMMA() != nil {
		result.SecondExpression = v.Visit(ctx.Expr(1)).(tree.Expression)
	}

	return &result
}

// VisitTable_or_subquery is called when visiting a table_or_subquery, return tree.TableOrSubquery
func (v *KFSqliteVisitor) VisitTable_or_subquery(ctx *grammar.Table_or_subqueryContext) interface{} {
	switch {
	case ctx.Table_name() != nil:
		t := tree.TableOrSubqueryTable{
			Name: ctx.Table_name().GetText(),
		}
		if ctx.Table_alias() != nil {
			t.Alias = ctx.Table_alias().GetText()
		}
		return &t
	case len(ctx.AllTable_or_subquery()) != 0:
		ts := make([]tree.TableOrSubquery, len(ctx.AllTable_or_subquery()))
		for i, subCtx := range ctx.AllTable_or_subquery() {
			ts[i] = v.Visit(subCtx).(tree.TableOrSubquery)
		}
		return &tree.TableOrSubqueryList{
			TableOrSubqueries: ts,
		}
	case ctx.Select_stmt_core() != nil:
		t := tree.TableOrSubquerySelect{
			Select: v.Visit(ctx.Select_stmt_core()).(*tree.SelectStmt),
		}
		if ctx.Table_alias() != nil {
			t.Alias = ctx.Table_alias().GetText()
		}
		return &t
	}
	return nil
}

// VisitJoin_operator is called when visiting a join_operator, return *tree.JoinOperator
func (v *KFSqliteVisitor) VisitJoin_operator(ctx *grammar.Join_operatorContext) interface{} {
	jp := tree.JoinOperator{
		JoinType: tree.JoinTypeJoin,
	}

	if ctx.NATURAL_() != nil {
		jp.Natural = true
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
func (v *KFSqliteVisitor) VisitJoin_clause(ctx *grammar.Join_clauseContext) interface{} {
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
func (v *KFSqliteVisitor) VisitResult_column(ctx *grammar.Result_columnContext) interface{} {
	switch {
	// table_name need to be checked first
	case ctx.Table_name() != nil:
		return &tree.ResultColumnTable{
			TableName: ctx.Table_name().GetText(),
		}
	case ctx.STAR() != nil:
		return &tree.ResultColumnStar{}
	case ctx.Expr() != nil:
		r := &tree.ResultColumnExpression{
			Expression: v.Visit(ctx.Expr()).(tree.Expression),
		}
		if ctx.Column_alias() != nil {
			r.Alias = ctx.Column_alias().GetText()
		}
		return r
	}

	return nil
}

// VisitDelete_stmt is called when visiting a delete_stmt, return *tree.Delete
func (v *KFSqliteVisitor) VisitDelete_stmt(ctx *grammar.Delete_stmtContext) interface{} {
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
func (v *KFSqliteVisitor) VisitSelect_core(ctx *grammar.Select_coreContext) interface{} {
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
func (v *KFSqliteVisitor) VisitSelect_stmt_core(ctx *grammar.Select_stmt_coreContext) interface{} {
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
func (v *KFSqliteVisitor) VisitSelect_stmt(ctx *grammar.Select_stmtContext) interface{} {
	t := tree.Select{}

	if ctx.Common_table_stmt() != nil {
		t.CTE = v.Visit(ctx.Common_table_stmt()).([]*tree.CTE)
	}

	t.SelectStmt = v.Visit(ctx.Select_stmt_core()).(*tree.SelectStmt)
	return &t
}

func (v *KFSqliteVisitor) VisitSql_stmt_list(ctx *grammar.Sql_stmt_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *KFSqliteVisitor) VisitSql_stmt(ctx *grammar.Sql_stmtContext) interface{} {
	// Sql_stmtContext will only have one stmt
	return v.VisitChildren(ctx).([]tree.Ast)[0]
}

// VisitParse is called first by Visitor.Visit
func (v *KFSqliteVisitor) VisitParse(ctx *grammar.ParseContext) interface{} {
	// ParseContext will only have one Sql_stmt_listContext
	sqlStmtListContext := ctx.Sql_stmt_list(0)
	return v.VisitChildren(sqlStmtListContext).([]tree.Ast)
}

// Visit dispatch to the visit method of the ctx
// e.g. if the tree is a ParseContext, then dispatch call VisitParse.
// Overwrite is needed,
// refer to https://github.com/antlr/antlr4/pull/1841#issuecomment-576791512
func (v *KFSqliteVisitor) Visit(tree antlr.ParseTree) interface{} {
	//if tree == nil {
	//	return nil
	//}
	if v.trace {
		fmt.Printf("visit tree: %v, %s\n", reflect.TypeOf(tree), tree.GetText())
	}
	return tree.Accept(v)
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
