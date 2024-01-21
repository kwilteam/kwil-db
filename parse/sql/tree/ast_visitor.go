package tree

// AstVisitor defines visitor for AstNode.
type AstVisitor interface {
	Visit(AstNode) interface{}
	VisitAggregateFunc(*AggregateFunc) any
	VisitConflictTarget(*ConflictTarget) any
	VisitCompoundOperator(*CompoundOperator) any
	VisitCTE(*CTE) any
	VisitDateTimeFunc(*DateTimeFunction) any
	VisitDelete(*Delete) any
	VisitDeleteStmt(*DeleteStmt) any
	VisitExpressionLiteral(*ExpressionLiteral) any
	VisitExpressionBindParameter(*ExpressionBindParameter) any
	VisitExpressionColumn(*ExpressionColumn) any
	VisitExpressionUnary(*ExpressionUnary) any
	VisitExpressionBinaryComparison(*ExpressionBinaryComparison) any
	VisitExpressionFunction(*ExpressionFunction) any
	VisitExpressionList(*ExpressionList) any
	VisitExpressionCollate(*ExpressionCollate) any
	VisitExpressionStringCompare(*ExpressionStringCompare) any
	VisitExpressionIsNull(*ExpressionIsNull) any
	VisitExpressionDistinct(*ExpressionDistinct) any
	VisitExpressionBetween(*ExpressionBetween) any
	VisitExpressionSelect(*ExpressionSelect) any
	VisitExpressionCase(*ExpressionCase) any
	VisitExpressionArithmetic(*ExpressionArithmetic) any
	VisitScalarFunc(*ScalarFunction) any
	VisitGroupBy(*GroupBy) any
	VisitInsert(*Insert) any
	VisitInsertStmt(*InsertStmt) any
	VisitJoinClause(*JoinClause) any
	VisitJoinPredicate(*JoinPredicate) any
	VisitJoinOperator(*JoinOperator) any
	VisitLimit(*Limit) any
	VisitOrderBy(*OrderBy) any
	VisitOrderingTerm(*OrderingTerm) any
	VisitQualifiedTableName(*QualifiedTableName) any
	VisitResultColumnStar(*ResultColumnStar) any
	VisitResultColumnExpression(*ResultColumnExpression) any
	VisitResultColumnTable(*ResultColumnTable) any
	VisitReturningClause(*ReturningClause) any
	VisitReturningClauseColumn(*ReturningClauseColumn) any
	VisitSelect(*Select) any
	VisitSelectCore(*SelectCore) any
	VisitSelectStmt(*SelectStmt) any
	VisitFromClause(*FromClause) any
	VisitTableOrSubqueryTable(*TableOrSubqueryTable) any
	VisitTableOrSubquerySelect(*TableOrSubquerySelect) any
	VisitTableOrSubqueryList(*TableOrSubqueryList) any
	VisitTableOrSubqueryJoin(*TableOrSubqueryJoin) any
	VisitUpdateSetClause(*UpdateSetClause) any
	VisitUpdate(*Update) any
	VisitUpdateStmt(*UpdateStmt) any
	VisitUpsert(*Upsert) any
}

// BaseAstNodeVisitor implements AstVisitor interface, it can be embedded in
// other structs to provide default implementation for all methods.
type BaseAstNodeVisitor struct {
}

func (v *BaseAstNodeVisitor) Visit(node AstNode) interface{} {
	// dispatch to the concrete visit method
	return node.Accept(v)
}

func (v *BaseAstNodeVisitor) VisitAggregateFunc(node *AggregateFunc) any {
	return nil
}

func (v *BaseAstNodeVisitor) VisitConflictTarget(node *ConflictTarget) any {
	return nil
}

func (v *BaseAstNodeVisitor) VisitCompoundOperator(node *CompoundOperator) any {
	return nil
}

func (v *BaseAstNodeVisitor) VisitCTE(node *CTE) any {
	return nil
}

func (v *BaseAstNodeVisitor) VisitDateTimeFunc(node *DateTimeFunction) any {
	return nil
}

func (v *BaseAstNodeVisitor) VisitDelete(node *Delete) any {
	return nil
}

func (v *BaseAstNodeVisitor) VisitDeleteStmt(node *DeleteStmt) any {
	return nil
}

func (v *BaseAstNodeVisitor) VisitExpressionLiteral(node *ExpressionLiteral) any {
	return nil
}

func (v *BaseAstNodeVisitor) VisitExpressionBindParameter(node *ExpressionBindParameter) any {
	return nil
}

func (v *BaseAstNodeVisitor) VisitExpressionColumn(node *ExpressionColumn) any {
	return nil
}

func (v *BaseAstNodeVisitor) VisitExpressionUnary(node *ExpressionUnary) any {
	return nil
}

func (v *BaseAstNodeVisitor) VisitExpressionBinaryComparison(node *ExpressionBinaryComparison) any {
	return nil
}

func (v *BaseAstNodeVisitor) VisitExpressionFunction(node *ExpressionFunction) any {
	return nil
}

func (v *BaseAstNodeVisitor) VisitExpressionList(node *ExpressionList) any {
	return nil
}

func (v *BaseAstNodeVisitor) VisitExpressionCollate(node *ExpressionCollate) any {
	return nil
}

func (v *BaseAstNodeVisitor) VisitExpressionStringCompare(node *ExpressionStringCompare) any {
	return nil
}

func (v *BaseAstNodeVisitor) VisitExpressionIsNull(node *ExpressionIsNull) any {
	return nil
}

func (v *BaseAstNodeVisitor) VisitExpressionDistinct(node *ExpressionDistinct) any {
	return nil
}

func (v *BaseAstNodeVisitor) VisitExpressionBetween(node *ExpressionBetween) any {
	return nil
}

func (v *BaseAstNodeVisitor) VisitExpressionSelect(node *ExpressionSelect) any {
	return nil
}

func (v *BaseAstNodeVisitor) VisitExpressionCase(node *ExpressionCase) any {
	return nil
}

func (v *BaseAstNodeVisitor) VisitExpressionArithmetic(node *ExpressionArithmetic) any {
	return nil
}

func (v *BaseAstNodeVisitor) VisitScalarFunc(node *ScalarFunction) any {
	return nil
}

func (v *BaseAstNodeVisitor) VisitGroupBy(node *GroupBy) any {
	return nil
}

func (v *BaseAstNodeVisitor) VisitInsert(node *Insert) any {
	return nil
}

func (v *BaseAstNodeVisitor) VisitInsertStmt(node *InsertStmt) any {
	return nil
}

func (v *BaseAstNodeVisitor) VisitJoinClause(node *JoinClause) any {
	return nil
}

func (v *BaseAstNodeVisitor) VisitJoinPredicate(node *JoinPredicate) any {
	return nil
}

func (v *BaseAstNodeVisitor) VisitJoinOperator(node *JoinOperator) any {
	return nil
}

func (v *BaseAstNodeVisitor) VisitLimit(node *Limit) any {
	return nil
}

func (v *BaseAstNodeVisitor) VisitOrderBy(node *OrderBy) any {
	return nil
}

func (v *BaseAstNodeVisitor) VisitOrderingTerm(node *OrderingTerm) any {
	return nil
}

func (v *BaseAstNodeVisitor) VisitQualifiedTableName(node *QualifiedTableName) any {
	return nil
}

func (v *BaseAstNodeVisitor) VisitResultColumnStar(node *ResultColumnStar) any {
	return nil
}

func (v *BaseAstNodeVisitor) VisitResultColumnExpression(node *ResultColumnExpression) any {
	return nil
}

func (v *BaseAstNodeVisitor) VisitResultColumnTable(node *ResultColumnTable) any {
	return nil
}

func (v *BaseAstNodeVisitor) VisitReturningClause(node *ReturningClause) any {
	return nil
}

func (v *BaseAstNodeVisitor) VisitReturningClauseColumn(node *ReturningClauseColumn) any {
	return nil
}

func (v *BaseAstNodeVisitor) VisitSelect(node *Select) any {
	return nil
}

func (v *BaseAstNodeVisitor) VisitSelectCore(node *SelectCore) any {
	return nil
}

func (v *BaseAstNodeVisitor) VisitSelectStmt(node *SelectStmt) any {
	return nil
}

func (v *BaseAstNodeVisitor) VisitFromClause(node *FromClause) any {
	return nil
}

func (v *BaseAstNodeVisitor) VisitTableOrSubqueryTable(node *TableOrSubqueryTable) any {
	return nil
}

func (v *BaseAstNodeVisitor) VisitTableOrSubquerySelect(node *TableOrSubquerySelect) any {
	return nil
}

func (v *BaseAstNodeVisitor) VisitTableOrSubqueryList(node *TableOrSubqueryList) any {
	return nil
}

func (v *BaseAstNodeVisitor) VisitTableOrSubqueryJoin(node *TableOrSubqueryJoin) any {
	return nil
}

func (v *BaseAstNodeVisitor) VisitUpdateSetClause(node *UpdateSetClause) any {
	return nil
}

func (v *BaseAstNodeVisitor) VisitUpdate(node *Update) any {
	return nil
}

func (v *BaseAstNodeVisitor) VisitUpdateStmt(node *UpdateStmt) any {
	return nil
}

func (v *BaseAstNodeVisitor) VisitUpsert(node *Upsert) any {
	return nil
}
