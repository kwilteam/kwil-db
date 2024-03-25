package tree

// AstVisitor defines visitor for AstNode.
type AstVisitor interface {
	Visit(AstNode) interface{}
	VisitAggregateFunc(*AggregateFunc) any
	VisitConflictTarget(*ConflictTarget) any
	VisitCompoundOperator(*CompoundOperator) any
	VisitCTE(*CTE) any
	VisitDeleteStmt(*DeleteStmt) any
	VisitDeleteCore(*DeleteCore) any
	VisitExpression(Expression) any
	VisitExpressionLiteral(*ExpressionLiteral) any
	VisitExpressionBindParameter(*ExpressionBindParameter) any
	VisitExpressionColumn(*ExpressionColumn) any
	VisitExpressionUnary(*ExpressionUnary) any
	VisitExpressionBinaryComparison(*ExpressionBinaryComparison) any
	VisitExpressionFunction(*ExpressionFunction) any
	VisitExpressionList(*ExpressionList) any
	VisitExpressionCollate(*ExpressionCollate) any
	VisitExpressionStringCompare(*ExpressionStringCompare) any
	VisitExpressionIs(*ExpressionIs) any
	VisitExpressionBetween(*ExpressionBetween) any
	VisitExpressionSelect(*ExpressionSelect) any
	VisitExpressionCase(*ExpressionCase) any
	VisitExpressionArithmetic(*ExpressionArithmetic) any
	VisitScalarFunc(*ScalarFunction) any
	VisitGroupBy(*GroupBy) any
	VisitInsertStmt(*InsertStmt) any
	VisitInsertCore(*InsertCore) any
	VisitJoinPredicate(*JoinPredicate) any
	VisitJoinOperator(*JoinOperator) any
	VisitLimit(*Limit) any
	VisitOrderBy(*OrderBy) any
	VisitOrderingTerm(*OrderingTerm) any
	VisitQualifiedTableName(*QualifiedTableName) any
	VisitRelation(Relation) any
	VisitRelationTable(*RelationTable) any
	VisitRelationSubquery(*RelationSubquery) any
	VisitRelationJoin(*RelationJoin) any
	VisitResultColumnStar(*ResultColumnStar) any
	VisitResultColumnExpression(*ResultColumnExpression) any
	VisitResultColumnTable(*ResultColumnTable) any
	VisitReturningClause(*ReturningClause) any
	VisitReturningClauseColumn(*ReturningClauseColumn) any
	VisitSelectStmt(*SelectStmt) any
	VisitSimpleSelect(*SimpleSelect) any
	VisitSelectCore(*SelectCore) any
	VisitUpdateSetClause(*UpdateSetClause) any
	VisitUpdateStmt(*UpdateStmt) any
	VisitUpdateCore(*UpdateCore) any
	VisitUpsert(*Upsert) any
}

// BaseAstVisitor implements AstVisitor interface, it can be embedded in
// other structs to provide default implementation for all methods.
type BaseAstVisitor struct {
}

func (v *BaseAstVisitor) Visit(node AstNode) interface{} {
	// dispatch to the concrete visit method
	return node.Accept(v)
}

func (v *BaseAstVisitor) VisitAggregateFunc(node *AggregateFunc) any {
	return nil
}

func (v *BaseAstVisitor) VisitConflictTarget(node *ConflictTarget) any {
	return nil
}

func (v *BaseAstVisitor) VisitCompoundOperator(node *CompoundOperator) any {
	return nil
}

func (v *BaseAstVisitor) VisitCTE(node *CTE) any {
	return nil
}

func (v *BaseAstVisitor) VisitDeleteStmt(node *DeleteStmt) any {
	return nil
}

func (v *BaseAstVisitor) VisitDeleteCore(node *DeleteCore) any {
	return nil
}

func (v *BaseAstVisitor) VisitExpression(node Expression) any {
	return v.Visit(node)
}

func (v *BaseAstVisitor) VisitExpressionLiteral(node *ExpressionLiteral) any {
	return nil
}

func (v *BaseAstVisitor) VisitExpressionBindParameter(node *ExpressionBindParameter) any {
	return nil
}

func (v *BaseAstVisitor) VisitExpressionColumn(node *ExpressionColumn) any {
	return nil
}

func (v *BaseAstVisitor) VisitExpressionUnary(node *ExpressionUnary) any {
	return nil
}

func (v *BaseAstVisitor) VisitExpressionBinaryComparison(node *ExpressionBinaryComparison) any {
	return nil
}

func (v *BaseAstVisitor) VisitExpressionFunction(node *ExpressionFunction) any {
	return nil
}

func (v *BaseAstVisitor) VisitExpressionList(node *ExpressionList) any {
	return nil
}

func (v *BaseAstVisitor) VisitExpressionCollate(node *ExpressionCollate) any {
	return nil
}

func (v *BaseAstVisitor) VisitExpressionStringCompare(node *ExpressionStringCompare) any {
	return nil
}

func (v *BaseAstVisitor) VisitExpressionIs(node *ExpressionIs) any {
	return nil
}

func (v *BaseAstVisitor) VisitExpressionBetween(node *ExpressionBetween) any {
	return nil
}

func (v *BaseAstVisitor) VisitExpressionSelect(node *ExpressionSelect) any {
	return nil
}

func (v *BaseAstVisitor) VisitExpressionCase(node *ExpressionCase) any {
	return nil
}

func (v *BaseAstVisitor) VisitExpressionArithmetic(node *ExpressionArithmetic) any {
	return nil
}

func (v *BaseAstVisitor) VisitScalarFunc(node *ScalarFunction) any {
	return nil
}

func (v *BaseAstVisitor) VisitGroupBy(node *GroupBy) any {
	return nil
}

func (v *BaseAstVisitor) VisitInsertStmt(node *InsertStmt) any {
	return nil
}

func (v *BaseAstVisitor) VisitInsertCore(node *InsertCore) any {
	return nil
}

func (v *BaseAstVisitor) VisitJoinPredicate(node *JoinPredicate) any {
	return nil
}

func (v *BaseAstVisitor) VisitJoinOperator(node *JoinOperator) any {
	return nil
}

func (v *BaseAstVisitor) VisitLimit(node *Limit) any {
	return nil
}

func (v *BaseAstVisitor) VisitOrderBy(node *OrderBy) any {
	return nil
}

func (v *BaseAstVisitor) VisitOrderingTerm(node *OrderingTerm) any {
	return nil
}

func (v *BaseAstVisitor) VisitQualifiedTableName(node *QualifiedTableName) any {
	return nil
}

func (v *BaseAstVisitor) VisitResultColumnStar(node *ResultColumnStar) any {
	return nil
}

func (v *BaseAstVisitor) VisitResultColumnExpression(node *ResultColumnExpression) any {
	return nil
}

func (v *BaseAstVisitor) VisitResultColumnTable(node *ResultColumnTable) any {
	return nil
}

func (v *BaseAstVisitor) VisitReturningClause(node *ReturningClause) any {
	return nil
}

func (v *BaseAstVisitor) VisitReturningClauseColumn(node *ReturningClauseColumn) any {
	return nil
}

func (v *BaseAstVisitor) VisitSelectStmt(node *SelectStmt) any {
	return nil
}

func (v *BaseAstVisitor) VisitSimpleSelect(node *SimpleSelect) any {
	return nil
}

func (v *BaseAstVisitor) VisitSelectCore(node *SelectCore) any {
	return nil
}

func (v *BaseAstVisitor) VisitRelation(node Relation) any {
	return v.Visit(node)
}

func (v *BaseAstVisitor) VisitRelationTable(node *RelationTable) any {
	return nil
}

func (v *BaseAstVisitor) VisitRelationSubquery(node *RelationSubquery) any {
	return nil
}

func (v *BaseAstVisitor) VisitRelationJoin(node *RelationJoin) any {
	return nil
}

func (v *BaseAstVisitor) VisitUpdateSetClause(node *UpdateSetClause) any {
	return nil
}

func (v *BaseAstVisitor) VisitUpdateStmt(node *UpdateStmt) any {
	return nil
}

func (v *BaseAstVisitor) VisitUpdateCore(node *UpdateCore) any {
	return nil
}

func (v *BaseAstVisitor) VisitUpsert(node *Upsert) any {
	return nil
}
