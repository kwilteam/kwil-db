package sql_parser

import (
	"github.com/kwilteam/kwil-db/pkg/engine/tree"
	"github.com/kwilteam/kwil-db/pkg/sql_parser/sqlite"
)

type KFSqliteVisitor struct {
	*sqlite.BaseSQLiteParserVisitor

	actionCtx ActionContext
	dbCtx     DatabaseContext

	trace bool
}

var _ sqlite.SQLiteParserVisitor = &KFSqliteVisitor{}

func NewKFSqliteVisitor() *KFSqliteVisitor {
	k := &KFSqliteVisitor{
		BaseSQLiteParserVisitor: nil,
		actionCtx:               nil,
		dbCtx:                   DatabaseContext{},
		trace:                   false,
	}
	return k
}

type cteTableName struct {
	table   string
	columns []string
}

type withClause struct {
	tableName cteTableName
}

func (k *KFSqliteVisitor) visitCteTableName(ctx sqlite.ICte_table_nameContext) (tableName cteTableName) {
	tableName.table = ctx.Table_name().GetText()
	colNameCtxs := ctx.AllColumn_name()
	for _, colName := range colNameCtxs {
		tableName.columns = append(tableName.columns, colName.GetText())
	}

	return tableName
}

func (k *KFSqliteVisitor) visitCommonTableExpression(ctx sqlite.ICommon_table_expressionContext) (t *tree.CTE) {
	cteTableCtx := ctx.Cte_table_name()
	tableName := k.visitCteTableName(cteTableCtx)
	t.Table = tableName.table
	t.Columns = tableName.columns

	selectCtx := ctx.Select_stmt()
	selectStmt := k.visitSelectStmt(selectCtx)
	t.Select = &selectStmt
	return t
}

func (k *KFSqliteVisitor) visitWithClause(ctx sqlite.IWith_clauseContext) []*tree.CTE {
	cteCount := len(ctx.AllCommon_table_expression())
	ctes := make([]*tree.CTE, cteCount)
	for i := 0; i < cteCount; i++ {
		cteCtx := ctx.Common_table_expression(i)
		cte := k.visitCommonTableExpression(cteCtx)
		ctes[i] = cte
	}
	return ctes
}

func (k *KFSqliteVisitor) visitSelectStmt(ctx sqlite.ISelect_stmtContext) tree.SelectStmt {
	t := tree.SelectStmt{}
	return t
}

func getInsertType(ctx *sqlite.Insert_stmtContext) tree.InsertType {
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

func (k *KFSqliteVisitor) visitExpr(ctx sqlite.IExprContext) tree.Expression {
	switch {
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
	case ctx.Function_name() != nil:

	// binary opertors
	//case ctx.PIPE2() != nil: // TODO: ??
	//	return &tree.ExpressionBinaryComparison{
	//		Left:     k.visitExpr(ctx.Expr(0)),
	//		Right:    k.visitExpr(ctx.Expr(1)),
	//		Operator: tree.BitwiseOperatorBitwiseOr,
	//	}
	case ctx.STAR() != nil:
		return &tree.ExpressionBinaryComparison{
			Left:     k.visitExpr(ctx.Expr(0)),
			Right:    k.visitExpr(ctx.Expr(1)),
			Operator: tree.ArithmeticOperatorMultiply,
		}
	case ctx.DIV() != nil:
		return &tree.ExpressionBinaryComparison{
			Left:     k.visitExpr(ctx.Expr(0)),
			Right:    k.visitExpr(ctx.Expr(1)),
			Operator: tree.ArithmeticOperatorDivide,
		}
	case ctx.MOD() != nil:
		return &tree.ExpressionBinaryComparison{
			Left:     k.visitExpr(ctx.Expr(0)),
			Right:    k.visitExpr(ctx.Expr(1)),
			Operator: tree.ArithmeticOperatorModulus,
		}
	case ctx.PLUS() != nil:
		return &tree.ExpressionBinaryComparison{
			Left:     k.visitExpr(ctx.Expr(0)),
			Right:    k.visitExpr(ctx.Expr(1)),
			Operator: tree.ArithmeticOperatorAdd,
		}
	case ctx.MINUS() != nil:
		return &tree.ExpressionBinaryComparison{
			Left:     k.visitExpr(ctx.Expr(0)),
			Right:    k.visitExpr(ctx.Expr(1)),
			Operator: tree.ArithmeticOperatorSubtract,
		}
	case ctx.LT2() != nil:
		return &tree.ExpressionBinaryComparison{
			Left:     k.visitExpr(ctx.Expr(0)),
			Right:    k.visitExpr(ctx.Expr(1)),
			Operator: tree.BitwiseOperatorLeftShift,
		}
	case ctx.GT2() != nil:
		return &tree.ExpressionBinaryComparison{
			Left:     k.visitExpr(ctx.Expr(0)),
			Right:    k.visitExpr(ctx.Expr(1)),
			Operator: tree.BitwiseOperatorRightShift,
		}
	case ctx.AMP() != nil:
		return &tree.ExpressionBinaryComparison{
			Left:     k.visitExpr(ctx.Expr(0)),
			Right:    k.visitExpr(ctx.Expr(1)),
			Operator: tree.BitwiseOperatorAnd,
		}
	case ctx.PIPE() != nil:
		return &tree.ExpressionBinaryComparison{
			Left:     k.visitExpr(ctx.Expr(0)),
			Right:    k.visitExpr(ctx.Expr(1)),
			Operator: tree.BitwiseOperatorOr,
		}
	case ctx.LT() != nil:
		return &tree.ExpressionBinaryComparison{
			Left:     k.visitExpr(ctx.Expr(0)),
			Right:    k.visitExpr(ctx.Expr(1)),
			Operator: tree.ComparisonOperatorLessThan,
		}
	case ctx.LT_EQ() != nil:
		return &tree.ExpressionBinaryComparison{
			Left:     k.visitExpr(ctx.Expr(0)),
			Right:    k.visitExpr(ctx.Expr(1)),
			Operator: tree.ComparisonOperatorLessThanOrEqual,
		}
	case ctx.GT() != nil:
		return &tree.ExpressionBinaryComparison{
			Left:     k.visitExpr(ctx.Expr(0)),
			Right:    k.visitExpr(ctx.Expr(1)),
			Operator: tree.ComparisonOperatorGreaterThan,
		}
	case ctx.GT_EQ() != nil:
		return &tree.ExpressionBinaryComparison{
			Left:     k.visitExpr(ctx.Expr(0)),
			Right:    k.visitExpr(ctx.Expr(1)),
			Operator: tree.ComparisonOperatorGreaterThanOrEqual,
		}
	//case ctx.ASSIGN() != nil:
	//	return &tree.ExpressionBinaryComparison{
	//		Left:     k.visitExpr(ctx.Expr(0)),
	//		Right:    k.visitExpr(ctx.Expr(1)),
	//		Operator: tree.ComparisonOperatorAssign, // TODO: assign
	//	}
	case ctx.EQ() != nil:
		return &tree.ExpressionBinaryComparison{
			Left:     k.visitExpr(ctx.Expr(0)),
			Right:    k.visitExpr(ctx.Expr(1)), // TODO: equal
			Operator: tree.ComparisonOperatorEqual,
		}
	case ctx.NOT_EQ1() != nil:
		return &tree.ExpressionBinaryComparison{
			Left:     k.visitExpr(ctx.Expr(0)),
			Right:    k.visitExpr(ctx.Expr(1)),
			Operator: tree.ComparisonOperatorNotEqual,
		}
	//case ctx.NOT_EQ2() != nil:
	//	return &tree.ExpressionBinaryComparison{
	//		Left:     k.visitExpr(ctx.Expr(0)),
	//		Right:    k.visitExpr(ctx.Expr(1)),
	//		Operator: tree.ComparisonOperatorNotEqual2,
	//	}
	case ctx.IS_() != nil:
		e := &tree.ExpressionBinaryComparison{
			Left:     k.visitExpr(ctx.Expr(0)),
			Right:    k.visitExpr(ctx.Expr(1)),
			Operator: tree.ComparisonOperatorIs,
		}
		if ctx.NOT_() != nil {
			e.Operator = tree.ComparisonOperatorIsNot
		}
		return e
	case ctx.IN_() != nil:
		return &tree.ExpressionBinaryComparison{
			Left:     k.visitExpr(ctx.Expr(0)),
			Right:    k.visitExpr(ctx.Expr(1)),
			Operator: tree.ComparisonOperatorIn,
		}
		//case ctx.LIKE_() != nil:
		//	e := &tree.ExpressionBinaryComparison{
		//		Left:     k.visitExpr(ctx.Expr(0)),
		//		Right:    k.visitExpr(ctx.Expr(1)),
		//		Operator: tree.ComparisonOperatorLike,
		//	}
		//case ctx.MATCH_()	!= nil:
		//	e := &tree.ExpressionBinaryComparison{
		//		Left:     k.visitExpr(ctx.Expr(0)),
		//		Right:    k.visitExpr(ctx.Expr(1)),
		//		Operator: tree.ComparisonOperatorMatch,
		//	}
		//case ctx.REGEXP_() != nil:
		//	e := &tree.ExpressionBinaryComparison{
		//		Left:     k.visitExpr(ctx.Expr(0)),
		//		Right:    k.visitExpr(ctx.Expr(1)),
		//		Operator: tree.ComparisonOperatorRegexp,
		//	}

	}
	return nil
}

func any(item ...interface{}) (truthiness bool) {
	for _, n := range item {
		if n != nil {
			return true
		}
	}
	return false
}

func anyWithIndex(item ...interface{}) (truthiness bool, index int) {
	for i, n := range item {
		if n != nil {
			return true, i
		}
	}
	return false, -1
}

func ifExprBinaryOp(ctx sqlite.IExprContext) bool {
	binOps := []interface{}{
		ctx.PIPE2(),
		ctx.STAR(), ctx.DIV(), ctx.MOD(),
		ctx.PLUS(), ctx.MINUS(),
		ctx.LT2(), ctx.GT2(), ctx.AMP(), ctx.PIPE(),
		ctx.LT(), ctx.LT_EQ(), ctx.GT(), ctx.GT_EQ(),
		ctx.ASSIGN(), ctx.EQ(), ctx.NOT_EQ1(), ctx.NOT_EQ2(), ctx.IS_(), //ctx.NOT_(),
		ctx.IN_(), ctx.LIKE_(), ctx.MATCH_(), ctx.REGEXP_(),
		ctx.AND_(),
		ctx.OR_()}
	return any(binOps)
}

func (k *KFSqliteVisitor) visitBinaryOp(ctx)

func (k *KFSqliteVisitor) visitValuesClause(ctx sqlite.IValues_clauseContext) [][]tree.Expression {
	allValueRowCtx := ctx.AllValue_row()
	rows := make([][]tree.Expression, len(allValueRowCtx))
	for i, valueRowCtx := range allValueRowCtx {
		allExprCtx := valueRowCtx.AllExpr()
		exprs := make([]tree.Expression, len(allExprCtx))
		for j, exprCtx := range allExprCtx {

		}
	}
}

func a() {

}

func (k *KFSqliteVisitor) VisitInsert_stmt(ctx *sqlite.Insert_stmtContext) interface{} {
	t := tree.Insert{}
	var insertStmt *tree.InsertStmt

	withClauseCtx := ctx.With_clause()
	if withClauseCtx != nil {
		t.CTE = k.visitWithClause(withClauseCtx)
	}

	insertStmt.InsertType = getInsertType(ctx)
	insertStmt.Table = ctx.Table_name().GetText()
	if ctx.Table_alias() != nil {
		insertStmt.TableAlias = ctx.Table_alias().GetText()
	}

	allColumnNameCtx := ctx.AllColumn_name()
	if len(allColumnNameCtx) > 0 {
		for _, nc := range allColumnNameCtx {
			insertStmt.Columns = append(insertStmt.Columns, nc.GetText())
		}
	}

	insertStmt.Values = k.visitValuesClause(ctx.Values_clause())

	return t
}

func (k *KFSqliteVisitor) VisitSelect_stmt(ctx *sqlite.Select_stmtContext) interface{} {
	stmt := k.visitSelectStmt(ctx)
	return stmt
}
