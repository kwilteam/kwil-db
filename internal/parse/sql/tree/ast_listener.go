package tree

// AstListener defines the interface for walking through the AstNode.
type AstListener interface {
	EnterAggregateFunc(*AggregateFunc) error
	ExitAggregateFunc(*AggregateFunc) error
	EnterConflictTarget(*ConflictTarget) error
	ExitConflictTarget(*ConflictTarget) error
	EnterCTE(*CTE) error
	ExitCTE(*CTE) error
	EnterDeleteStmt(*DeleteStmt) error
	ExitDeleteStmt(*DeleteStmt) error
	EnterDeleteCore(*DeleteCore) error
	ExitDeleteCore(*DeleteCore) error
	EnterExpressionTextLiteral(*ExpressionTextLiteral) error
	ExitExpressionTextLiteral(*ExpressionTextLiteral) error
	EnterExpressionNumericLiteral(*ExpressionNumericLiteral) error
	ExitExpressionNumericLiteral(*ExpressionNumericLiteral) error
	EnterExpressionBooleanLiteral(*ExpressionBooleanLiteral) error
	ExitExpressionBooleanLiteral(*ExpressionBooleanLiteral) error
	EnterExpressionNullLiteral(*ExpressionNullLiteral) error
	ExitExpressionNullLiteral(*ExpressionNullLiteral) error
	EnterExpressionBlobLiteral(*ExpressionBlobLiteral) error
	ExitExpressionBlobLiteral(*ExpressionBlobLiteral) error
	EnterExpressionBindParameter(*ExpressionBindParameter) error
	ExitExpressionBindParameter(*ExpressionBindParameter) error
	EnterExpressionColumn(*ExpressionColumn) error
	ExitExpressionColumn(*ExpressionColumn) error
	EnterExpressionUnary(*ExpressionUnary) error
	ExitExpressionUnary(*ExpressionUnary) error
	EnterExpressionBinaryComparison(*ExpressionBinaryComparison) error
	ExitExpressionBinaryComparison(*ExpressionBinaryComparison) error
	EnterExpressionFunction(*ExpressionFunction) error
	ExitExpressionFunction(*ExpressionFunction) error
	EnterExpressionList(*ExpressionList) error
	ExitExpressionList(*ExpressionList) error
	EnterExpressionCollate(*ExpressionCollate) error
	ExitExpressionCollate(*ExpressionCollate) error
	EnterExpressionStringCompare(*ExpressionStringCompare) error
	ExitExpressionStringCompare(*ExpressionStringCompare) error
	EnterExpressionIs(*ExpressionIs) error
	ExitExpressionIs(*ExpressionIs) error
	EnterExpressionBetween(*ExpressionBetween) error
	ExitExpressionBetween(*ExpressionBetween) error
	EnterExpressionSelect(*ExpressionSelect) error
	ExitExpressionSelect(*ExpressionSelect) error
	EnterExpressionCase(*ExpressionCase) error
	ExitExpressionCase(*ExpressionCase) error
	EnterExpressionArithmetic(*ExpressionArithmetic) error
	ExitExpressionArithmetic(*ExpressionArithmetic) error
	EnterScalarFunc(*ScalarFunction) error
	ExitScalarFunc(*ScalarFunction) error
	EnterGroupBy(*GroupBy) error
	ExitGroupBy(*GroupBy) error
	EnterInsertStmt(*InsertStmt) error
	ExitInsertStmt(*InsertStmt) error
	EnterInsertCore(*InsertCore) error
	ExitInsertCore(*InsertCore) error
	EnterJoinPredicate(*JoinPredicate) error
	ExitJoinPredicate(*JoinPredicate) error
	EnterJoinOperator(*JoinOperator) error
	ExitJoinOperator(*JoinOperator) error
	EnterLimit(*Limit) error
	ExitLimit(*Limit) error
	EnterOrderBy(*OrderBy) error
	ExitOrderBy(*OrderBy) error
	EnterOrderingTerm(*OrderingTerm) error
	ExitOrderingTerm(*OrderingTerm) error
	EnterQualifiedTableName(*QualifiedTableName) error
	ExitQualifiedTableName(*QualifiedTableName) error
	EnterRelation(Relation) error
	ExitRelation(Relation) error
	EnterRelationTable(*RelationTable) error
	ExitRelationTable(*RelationTable) error
	EnterRelationSubquery(*RelationSubquery) error
	ExitRelationSubquery(*RelationSubquery) error
	EnterRelationJoin(*RelationJoin) error
	ExitRelationJoin(*RelationJoin) error
	EnterResultColumnStar(*ResultColumnStar) error
	ExitResultColumnStar(*ResultColumnStar) error
	EnterResultColumnExpression(*ResultColumnExpression) error
	ExitResultColumnExpression(*ResultColumnExpression) error
	EnterResultColumnTable(*ResultColumnTable) error
	ExitResultColumnTable(*ResultColumnTable) error
	EnterReturningClause(*ReturningClause) error
	ExitReturningClause(*ReturningClause) error
	EnterReturningClauseColumn(*ReturningClauseColumn) error
	ExitReturningClauseColumn(*ReturningClauseColumn) error
	EnterSelectStmt(*SelectStmt) error
	ExitSelectStmt(*SelectStmt) error
	EnterSimpleSelect(*SimpleSelect) error
	ExitSimpleSelect(*SimpleSelect) error
	EnterSelectCore(*SelectCore) error
	ExitSelectCore(*SelectCore) error
	EnterCompoundOperator(*CompoundOperator) error
	ExitCompoundOperator(*CompoundOperator) error
	EnterUpdateSetClause(*UpdateSetClause) error
	ExitUpdateSetClause(*UpdateSetClause) error
	EnterUpdateStmt(*UpdateStmt) error
	ExitUpdateStmt(*UpdateStmt) error
	EnterUpdateCore(*UpdateCore) error
	ExitUpdateCore(*UpdateCore) error
	EnterUpsert(*Upsert) error
	ExitUpsert(*Upsert) error
}

type BaseListener struct{}

var _ AstListener = &BaseListener{}

func NewBaseListener() AstListener {
	return &BaseListener{}
}

func (b *BaseListener) EnterAggregateFunc(p0 *AggregateFunc) error {
	return nil
}

func (b *BaseListener) ExitAggregateFunc(p0 *AggregateFunc) error {
	return nil
}

func (b *BaseListener) EnterCTE(p0 *CTE) error {
	return nil
}

func (b *BaseListener) ExitCTE(p0 *CTE) error {
	return nil
}

func (b *BaseListener) EnterCompoundOperator(p0 *CompoundOperator) error {
	return nil
}

func (b *BaseListener) ExitCompoundOperator(p0 *CompoundOperator) error {
	return nil
}

func (b *BaseListener) EnterConflictTarget(p0 *ConflictTarget) error {
	return nil
}

func (b *BaseListener) ExitConflictTarget(p0 *ConflictTarget) error {
	return nil
}

func (b *BaseListener) EnterDeleteStmt(p0 *DeleteStmt) error {
	return nil
}

func (b *BaseListener) ExitDeleteStmt(p0 *DeleteStmt) error {
	return nil
}

func (b *BaseListener) EnterDeleteCore(p0 *DeleteCore) error {
	return nil
}

func (b *BaseListener) ExitDeleteCore(p0 *DeleteCore) error {
	return nil
}

func (b *BaseListener) EnterExpressionArithmetic(p0 *ExpressionArithmetic) error {
	return nil
}

func (b *BaseListener) ExitExpressionArithmetic(p0 *ExpressionArithmetic) error {
	return nil
}

func (b *BaseListener) EnterExpressionBetween(p0 *ExpressionBetween) error {
	return nil
}

func (b *BaseListener) ExitExpressionBetween(p0 *ExpressionBetween) error {
	return nil
}

func (b *BaseListener) EnterExpressionBinaryComparison(p0 *ExpressionBinaryComparison) error {
	return nil
}

func (b *BaseListener) ExitExpressionBinaryComparison(p0 *ExpressionBinaryComparison) error {
	return nil
}

func (b *BaseListener) EnterExpressionBindParameter(p0 *ExpressionBindParameter) error {
	return nil
}

func (b *BaseListener) ExitExpressionBindParameter(p0 *ExpressionBindParameter) error {
	return nil
}

func (b *BaseListener) EnterExpressionCase(p0 *ExpressionCase) error {
	return nil
}

func (b *BaseListener) ExitExpressionCase(p0 *ExpressionCase) error {
	return nil
}

func (b *BaseListener) EnterExpressionCollate(p0 *ExpressionCollate) error {
	return nil
}

func (b *BaseListener) ExitExpressionCollate(p0 *ExpressionCollate) error {
	return nil
}

func (b *BaseListener) EnterExpressionColumn(p0 *ExpressionColumn) error {
	return nil
}

func (b *BaseListener) ExitExpressionColumn(p0 *ExpressionColumn) error {
	return nil
}

func (b *BaseListener) EnterExpressionFunction(p0 *ExpressionFunction) error {
	return nil
}

func (b *BaseListener) ExitExpressionFunction(p0 *ExpressionFunction) error {
	return nil
}

func (b *BaseListener) EnterExpressionIs(p0 *ExpressionIs) error {
	return nil
}

func (b *BaseListener) ExitExpressionIs(p0 *ExpressionIs) error {
	return nil
}

func (b *BaseListener) EnterExpressionList(p0 *ExpressionList) error {
	return nil
}

func (b *BaseListener) ExitExpressionList(p0 *ExpressionList) error {
	return nil
}

func (b *BaseListener) EnterExpressionTextLiteral(p0 *ExpressionTextLiteral) error {
	return nil
}

func (b *BaseListener) ExitExpressionTextLiteral(p0 *ExpressionTextLiteral) error {
	return nil
}

func (b *BaseListener) EnterExpressionNumericLiteral(p0 *ExpressionNumericLiteral) error {
	return nil
}

func (b *BaseListener) ExitExpressionNumericLiteral(p0 *ExpressionNumericLiteral) error {
	return nil
}

func (b *BaseListener) EnterExpressionBooleanLiteral(p0 *ExpressionBooleanLiteral) error {
	return nil
}

func (b *BaseListener) ExitExpressionBooleanLiteral(p0 *ExpressionBooleanLiteral) error {
	return nil
}

func (b *BaseListener) EnterExpressionNullLiteral(p0 *ExpressionNullLiteral) error {
	return nil
}

func (b *BaseListener) ExitExpressionNullLiteral(p0 *ExpressionNullLiteral) error {
	return nil
}

func (b *BaseListener) EnterExpressionBlobLiteral(p0 *ExpressionBlobLiteral) error {
	return nil
}

func (b *BaseListener) ExitExpressionBlobLiteral(p0 *ExpressionBlobLiteral) error {
	return nil
}

func (b *BaseListener) EnterExpressionSelect(p0 *ExpressionSelect) error {
	return nil
}

func (b *BaseListener) ExitExpressionSelect(p0 *ExpressionSelect) error {
	return nil
}

func (b *BaseListener) EnterExpressionStringCompare(p0 *ExpressionStringCompare) error {
	return nil
}

func (b *BaseListener) ExitExpressionStringCompare(p0 *ExpressionStringCompare) error {
	return nil
}

func (b *BaseListener) EnterExpressionUnary(p0 *ExpressionUnary) error {
	return nil
}

func (b *BaseListener) ExitExpressionUnary(p0 *ExpressionUnary) error {
	return nil
}

func (b *BaseListener) EnterGroupBy(p0 *GroupBy) error {
	return nil
}

func (b *BaseListener) ExitGroupBy(p0 *GroupBy) error {
	return nil
}

func (b *BaseListener) EnterInsertStmt(p0 *InsertStmt) error {
	return nil
}

func (b *BaseListener) ExitInsertStmt(p0 *InsertStmt) error {
	return nil
}

func (b *BaseListener) EnterInsertCore(p0 *InsertCore) error {
	return nil
}

func (b *BaseListener) ExitInsertCore(p0 *InsertCore) error {
	return nil
}

func (b *BaseListener) EnterJoinOperator(p0 *JoinOperator) error {
	return nil
}

func (b *BaseListener) ExitJoinOperator(p0 *JoinOperator) error {
	return nil
}

func (b *BaseListener) EnterJoinPredicate(p0 *JoinPredicate) error {
	return nil
}

func (b *BaseListener) ExitJoinPredicate(p0 *JoinPredicate) error {
	return nil
}

func (b *BaseListener) EnterLimit(p0 *Limit) error {
	return nil
}

func (b *BaseListener) ExitLimit(p0 *Limit) error {
	return nil
}

func (b *BaseListener) EnterOrderBy(p0 *OrderBy) error {
	return nil
}

func (b *BaseListener) ExitOrderBy(p0 *OrderBy) error {
	return nil
}

func (b *BaseListener) EnterOrderingTerm(p0 *OrderingTerm) error {
	return nil
}

func (b *BaseListener) ExitOrderingTerm(p0 *OrderingTerm) error {
	return nil
}

func (b *BaseListener) EnterQualifiedTableName(p0 *QualifiedTableName) error {
	return nil
}

func (b *BaseListener) ExitQualifiedTableName(p0 *QualifiedTableName) error {
	return nil
}

func (b *BaseListener) EnterRelation(p0 Relation) error {
	return nil
}

func (b *BaseListener) ExitRelation(p0 Relation) error {
	return nil
}

func (b *BaseListener) EnterRelationJoin(p0 *RelationJoin) error {
	return nil
}

func (b *BaseListener) ExitRelationJoin(p0 *RelationJoin) error {
	return nil
}

func (b *BaseListener) EnterRelationSubquery(p0 *RelationSubquery) error {
	return nil
}

func (b *BaseListener) ExitRelationSubquery(p0 *RelationSubquery) error {
	return nil
}

func (b *BaseListener) EnterRelationTable(p0 *RelationTable) error {
	return nil
}

func (b *BaseListener) ExitRelationTable(p0 *RelationTable) error {
	return nil
}

func (b *BaseListener) EnterResultColumnExpression(p0 *ResultColumnExpression) error {
	return nil
}

func (b *BaseListener) ExitResultColumnExpression(p0 *ResultColumnExpression) error {
	return nil
}

func (b *BaseListener) EnterResultColumnStar(p0 *ResultColumnStar) error {
	return nil
}

func (b *BaseListener) ExitResultColumnStar(p0 *ResultColumnStar) error {
	return nil
}

func (b *BaseListener) EnterResultColumnTable(p0 *ResultColumnTable) error {
	return nil
}

func (b *BaseListener) ExitResultColumnTable(p0 *ResultColumnTable) error {
	return nil
}

func (b *BaseListener) EnterReturningClause(p0 *ReturningClause) error {
	return nil
}

func (b *BaseListener) ExitReturningClause(p0 *ReturningClause) error {
	return nil
}

func (b *BaseListener) EnterReturningClauseColumn(p0 *ReturningClauseColumn) error {
	return nil
}

func (b *BaseListener) ExitReturningClauseColumn(p0 *ReturningClauseColumn) error {
	return nil
}

func (b *BaseListener) EnterScalarFunc(p0 *ScalarFunction) error {
	return nil
}

func (b *BaseListener) ExitScalarFunc(p0 *ScalarFunction) error {
	return nil
}

func (b *BaseListener) EnterSelectStmt(p0 *SelectStmt) error {
	return nil
}

func (b *BaseListener) ExitSelectStmt(p0 *SelectStmt) error {
	return nil
}

func (b *BaseListener) EnterSimpleSelect(p0 *SimpleSelect) error {
	return nil
}

func (b *BaseListener) ExitSimpleSelect(p0 *SimpleSelect) error {
	return nil
}

func (b *BaseListener) EnterSelectCore(p0 *SelectCore) error {
	return nil
}

func (b *BaseListener) ExitSelectCore(p0 *SelectCore) error {
	return nil
}

func (b *BaseListener) EnterUpdateStmt(p0 *UpdateStmt) error {
	return nil
}

func (b *BaseListener) ExitUpdateStmt(p0 *UpdateStmt) error {
	return nil
}

func (b *BaseListener) EnterUpdateSetClause(p0 *UpdateSetClause) error {
	return nil
}

func (b *BaseListener) ExitUpdateSetClause(p0 *UpdateSetClause) error {
	return nil
}

func (b *BaseListener) EnterUpdateCore(p0 *UpdateCore) error {
	return nil
}

func (b *BaseListener) ExitUpdateCore(p0 *UpdateCore) error {
	return nil
}

func (b *BaseListener) EnterUpsert(p0 *Upsert) error {
	return nil
}

func (b *BaseListener) ExitUpsert(p0 *Upsert) error {
	return nil
}

// ImplementedListener implements the AstListener interface.
// Unlike BaseListener, it holds the methods to be implemented
// as functions in a struct.  This makes it easier to implement
// for small, one-off walkers.
type ImplementedListener struct {
	FuncEnterAggregateFunc              func(p0 *AggregateFunc) error
	FuncExitAggregateFunc               func(p0 *AggregateFunc) error
	FuncEnterCTE                        func(p0 *CTE) error
	FuncExitCTE                         func(p0 *CTE) error
	FuncEnterCompoundOperator           func(p0 *CompoundOperator) error
	FuncExitCompoundOperator            func(p0 *CompoundOperator) error
	FuncEnterConflictTarget             func(p0 *ConflictTarget) error
	FuncExitConflictTarget              func(p0 *ConflictTarget) error
	FuncEnterDeleteStmt                 func(p0 *DeleteStmt) error
	FuncExitDeleteStmt                  func(p0 *DeleteStmt) error
	FuncEnterDeleteCore                 func(p0 *DeleteCore) error
	FuncExitDeleteCore                  func(p0 *DeleteCore) error
	FuncEnterExpressionArithmetic       func(p0 *ExpressionArithmetic) error
	FuncExitExpressionArithmetic        func(p0 *ExpressionArithmetic) error
	FuncEnterExpressionBetween          func(p0 *ExpressionBetween) error
	FuncExitExpressionBetween           func(p0 *ExpressionBetween) error
	FuncEnterExpressionBinaryComparison func(p0 *ExpressionBinaryComparison) error
	FuncExitExpressionBinaryComparison  func(p0 *ExpressionBinaryComparison) error
	FuncEnterExpressionBindParameter    func(p0 *ExpressionBindParameter) error
	FuncExitExpressionBindParameter     func(p0 *ExpressionBindParameter) error
	FuncEnterExpressionCase             func(p0 *ExpressionCase) error
	FuncExitExpressionCase              func(p0 *ExpressionCase) error
	FuncEnterExpressionCollate          func(p0 *ExpressionCollate) error
	FuncExitExpressionCollate           func(p0 *ExpressionCollate) error
	FuncEnterExpressionColumn           func(p0 *ExpressionColumn) error
	FuncExitExpressionColumn            func(p0 *ExpressionColumn) error
	FuncEnterExpressionFunction         func(p0 *ExpressionFunction) error
	FuncExitExpressionFunction          func(p0 *ExpressionFunction) error
	FuncEnterExpressionIs               func(p0 *ExpressionIs) error
	FuncExitExpressionIs                func(p0 *ExpressionIs) error
	FuncEnterExpressionList             func(p0 *ExpressionList) error
	FuncExitExpressionList              func(p0 *ExpressionList) error
	FuncEnterExpressionTextLiteral      func(p0 *ExpressionTextLiteral) error
	FuncExitExpressionTextLiteral       func(p0 *ExpressionTextLiteral) error
	FuncEnterExpressionNumericLiteral   func(p0 *ExpressionNumericLiteral) error
	FuncExitExpressionNumericLiteral    func(p0 *ExpressionNumericLiteral) error
	FuncEnterExpressionBooleanLiteral   func(p0 *ExpressionBooleanLiteral) error
	FuncExitExpressionBooleanLiteral    func(p0 *ExpressionBooleanLiteral) error
	FuncEnterExpressionNullLiteral      func(p0 *ExpressionNullLiteral) error
	FuncExitExpressionNullLiteral       func(p0 *ExpressionNullLiteral) error
	FuncEnterExpressionBlobLiteral      func(p0 *ExpressionBlobLiteral) error
	FuncExitExpressionBlobLiteral       func(p0 *ExpressionBlobLiteral) error
	FuncEnterExpressionSelect           func(p0 *ExpressionSelect) error
	FuncExitExpressionSelect            func(p0 *ExpressionSelect) error
	FuncEnterExpressionStringCompare    func(p0 *ExpressionStringCompare) error
	FuncExitExpressionStringCompare     func(p0 *ExpressionStringCompare) error
	FuncEnterExpressionUnary            func(p0 *ExpressionUnary) error
	FuncExitExpressionUnary             func(p0 *ExpressionUnary) error
	FuncEnterGroupBy                    func(p0 *GroupBy) error
	FuncExitGroupBy                     func(p0 *GroupBy) error
	FuncEnterInsertStmt                 func(p0 *InsertStmt) error
	FuncExitInsertStmt                  func(p0 *InsertStmt) error
	FuncEnterInsertCore                 func(p0 *InsertCore) error
	FuncExitInsertCore                  func(p0 *InsertCore) error
	FuncEnterJoinOperator               func(p0 *JoinOperator) error
	FuncExitJoinOperator                func(p0 *JoinOperator) error
	FuncEnterJoinPredicate              func(p0 *JoinPredicate) error
	FuncExitJoinPredicate               func(p0 *JoinPredicate) error
	FuncEnterLimit                      func(p0 *Limit) error
	FuncExitLimit                       func(p0 *Limit) error
	FuncEnterOrderBy                    func(p0 *OrderBy) error
	FuncExitOrderBy                     func(p0 *OrderBy) error
	FuncEnterOrderingTerm               func(p0 *OrderingTerm) error
	FuncExitOrderingTerm                func(p0 *OrderingTerm) error
	FuncEnterQualifiedTableName         func(p0 *QualifiedTableName) error
	FuncExitQualifiedTableName          func(p0 *QualifiedTableName) error
	FuncEnterRelation                   func(p0 Relation) error
	FuncExitRelation                    func(p0 Relation) error
	FuncEnterRelationJoin               func(p0 *RelationJoin) error
	FuncExitRelationJoin                func(p0 *RelationJoin) error
	FuncEnterRelationSubquery           func(p0 *RelationSubquery) error
	FuncExitRelationSubquery            func(p0 *RelationSubquery) error
	FuncEnterRelationTable              func(p0 *RelationTable) error
	FuncExitRelationTable               func(p0 *RelationTable) error
	FuncEnterResultColumnExpression     func(p0 *ResultColumnExpression) error
	FuncExitResultColumnExpression      func(p0 *ResultColumnExpression) error
	FuncEnterResultColumnStar           func(p0 *ResultColumnStar) error
	FuncExitResultColumnStar            func(p0 *ResultColumnStar) error
	FuncEnterResultColumnTable          func(p0 *ResultColumnTable) error
	FuncExitResultColumnTable           func(p0 *ResultColumnTable) error
	FuncEnterReturningClause            func(p0 *ReturningClause) error
	FuncExitReturningClause             func(p0 *ReturningClause) error
	FuncEnterReturningClauseColumn      func(p0 *ReturningClauseColumn) error
	FuncExitReturningClauseColumn       func(p0 *ReturningClauseColumn) error
	FuncEnterScalarFunc                 func(p0 *ScalarFunction) error
	FuncExitScalarFunc                  func(p0 *ScalarFunction) error
	FuncEnterSelectStmt                 func(p0 *SelectStmt) error
	FuncExitSelectStmt                  func(p0 *SelectStmt) error
	FuncEnterSimpleSelect               func(p0 *SimpleSelect) error
	FuncExitSimpleSelect                func(p0 *SimpleSelect) error
	FuncEnterSelectCore                 func(p0 *SelectCore) error
	FuncExitSelectCore                  func(p0 *SelectCore) error
	FuncEnterUpdateStmt                 func(p0 *UpdateStmt) error
	FuncExitUpdateStmt                  func(p0 *UpdateStmt) error
	FuncEnterUpdateSetClause            func(p0 *UpdateSetClause) error
	FuncExitUpdateSetClause             func(p0 *UpdateSetClause) error
	FuncEnterUpdateCore                 func(p0 *UpdateCore) error
	FuncExitUpdateCore                  func(p0 *UpdateCore) error
	FuncEnterUpsert                     func(p0 *Upsert) error
	FuncExitUpsert                      func(p0 *Upsert) error
}

var _ AstListener = &ImplementedListener{}

func (b *ImplementedListener) EnterAggregateFunc(p0 *AggregateFunc) error {
	if b.FuncEnterAggregateFunc == nil {
		return nil
	}

	return b.FuncEnterAggregateFunc(p0)
}

func (b *ImplementedListener) ExitAggregateFunc(p0 *AggregateFunc) error {
	if b.FuncExitAggregateFunc == nil {
		return nil
	}

	return b.FuncExitAggregateFunc(p0)
}

func (b *ImplementedListener) EnterCTE(p0 *CTE) error {
	if b.FuncEnterCTE == nil {
		return nil
	}

	return b.FuncEnterCTE(p0)
}

func (b *ImplementedListener) ExitCTE(p0 *CTE) error {
	if b.FuncExitCTE == nil {
		return nil
	}

	return b.FuncExitCTE(p0)
}

func (b *ImplementedListener) EnterCompoundOperator(p0 *CompoundOperator) error {
	if b.FuncEnterCompoundOperator == nil {
		return nil
	}

	return b.FuncEnterCompoundOperator(p0)
}

func (b *ImplementedListener) ExitCompoundOperator(p0 *CompoundOperator) error {
	if b.FuncExitCompoundOperator == nil {
		return nil
	}

	return b.FuncExitCompoundOperator(p0)
}

func (b *ImplementedListener) EnterConflictTarget(p0 *ConflictTarget) error {
	if b.FuncEnterConflictTarget == nil {
		return nil
	}

	return b.FuncEnterConflictTarget(p0)
}

func (b *ImplementedListener) ExitConflictTarget(p0 *ConflictTarget) error {
	if b.FuncExitConflictTarget == nil {
		return nil
	}

	return b.FuncExitConflictTarget(p0)
}

func (b *ImplementedListener) EnterDeleteStmt(p0 *DeleteStmt) error {
	if b.FuncEnterDeleteStmt == nil {
		return nil
	}

	return b.FuncEnterDeleteStmt(p0)
}

func (b *ImplementedListener) ExitDeleteStmt(p0 *DeleteStmt) error {
	if b.FuncExitDeleteStmt == nil {
		return nil
	}

	return b.FuncExitDeleteStmt(p0)
}

func (b *ImplementedListener) EnterDeleteCore(p0 *DeleteCore) error {
	if b.FuncEnterDeleteCore == nil {
		return nil
	}

	return b.FuncEnterDeleteCore(p0)
}

func (b *ImplementedListener) ExitDeleteCore(p0 *DeleteCore) error {
	if b.FuncExitDeleteCore == nil {
		return nil
	}

	return b.FuncExitDeleteCore(p0)
}

func (b *ImplementedListener) EnterExpressionArithmetic(p0 *ExpressionArithmetic) error {
	if b.FuncEnterExpressionArithmetic == nil {
		return nil
	}

	return b.FuncEnterExpressionArithmetic(p0)
}

func (b *ImplementedListener) ExitExpressionArithmetic(p0 *ExpressionArithmetic) error {
	if b.FuncExitExpressionArithmetic == nil {
		return nil
	}

	return b.FuncExitExpressionArithmetic(p0)
}

func (b *ImplementedListener) EnterExpressionBetween(p0 *ExpressionBetween) error {
	if b.FuncEnterExpressionBetween == nil {
		return nil
	}

	return b.FuncEnterExpressionBetween(p0)
}

func (b *ImplementedListener) ExitExpressionBetween(p0 *ExpressionBetween) error {
	if b.FuncExitExpressionBetween == nil {
		return nil
	}

	return b.FuncExitExpressionBetween(p0)
}

func (b *ImplementedListener) EnterExpressionBinaryComparison(p0 *ExpressionBinaryComparison) error {
	if b.FuncEnterExpressionBinaryComparison == nil {
		return nil
	}

	return b.FuncEnterExpressionBinaryComparison(p0)
}

func (b *ImplementedListener) ExitExpressionBinaryComparison(p0 *ExpressionBinaryComparison) error {
	if b.FuncExitExpressionBinaryComparison == nil {
		return nil
	}

	return b.FuncExitExpressionBinaryComparison(p0)
}

func (b *ImplementedListener) EnterExpressionBindParameter(p0 *ExpressionBindParameter) error {
	if b.FuncEnterExpressionBindParameter == nil {
		return nil
	}

	return b.FuncEnterExpressionBindParameter(p0)
}

func (b *ImplementedListener) ExitExpressionBindParameter(p0 *ExpressionBindParameter) error {
	if b.FuncExitExpressionBindParameter == nil {
		return nil
	}

	return b.FuncExitExpressionBindParameter(p0)
}

func (b *ImplementedListener) EnterExpressionCase(p0 *ExpressionCase) error {
	if b.FuncEnterExpressionCase == nil {
		return nil
	}

	return b.FuncEnterExpressionCase(p0)
}

func (b *ImplementedListener) ExitExpressionCase(p0 *ExpressionCase) error {
	if b.FuncExitExpressionCase == nil {
		return nil
	}

	return b.FuncExitExpressionCase(p0)
}

func (b *ImplementedListener) EnterExpressionCollate(p0 *ExpressionCollate) error {
	if b.FuncEnterExpressionCollate == nil {
		return nil
	}

	return b.FuncEnterExpressionCollate(p0)
}

func (b *ImplementedListener) ExitExpressionCollate(p0 *ExpressionCollate) error {
	if b.FuncExitExpressionCollate == nil {
		return nil
	}

	return b.FuncExitExpressionCollate(p0)
}

func (b *ImplementedListener) EnterExpressionColumn(p0 *ExpressionColumn) error {
	if b.FuncEnterExpressionColumn == nil {
		return nil
	}

	return b.FuncEnterExpressionColumn(p0)
}

func (b *ImplementedListener) ExitExpressionColumn(p0 *ExpressionColumn) error {
	if b.FuncExitExpressionColumn == nil {
		return nil
	}

	return b.FuncExitExpressionColumn(p0)
}

func (b *ImplementedListener) EnterExpressionFunction(p0 *ExpressionFunction) error {
	if b.FuncEnterExpressionFunction == nil {
		return nil
	}

	return b.FuncEnterExpressionFunction(p0)
}

func (b *ImplementedListener) ExitExpressionFunction(p0 *ExpressionFunction) error {
	if b.FuncExitExpressionFunction == nil {
		return nil
	}

	return b.FuncExitExpressionFunction(p0)
}

func (b *ImplementedListener) EnterExpressionIs(p0 *ExpressionIs) error {
	if b.FuncEnterExpressionIs == nil {
		return nil
	}

	return b.FuncEnterExpressionIs(p0)
}

func (b *ImplementedListener) ExitExpressionIs(p0 *ExpressionIs) error {
	if b.FuncExitExpressionIs == nil {
		return nil
	}

	return b.FuncExitExpressionIs(p0)
}

func (b *ImplementedListener) EnterExpressionList(p0 *ExpressionList) error {
	if b.FuncEnterExpressionList == nil {
		return nil
	}

	return b.FuncEnterExpressionList(p0)
}

func (b *ImplementedListener) ExitExpressionList(p0 *ExpressionList) error {
	if b.FuncExitExpressionList == nil {
		return nil
	}

	return b.FuncExitExpressionList(p0)
}

func (b *ImplementedListener) EnterExpressionTextLiteral(p0 *ExpressionTextLiteral) error {
	if b.FuncEnterExpressionTextLiteral == nil {
		return nil
	}

	return b.FuncEnterExpressionTextLiteral(p0)
}

func (b *ImplementedListener) ExitExpressionTextLiteral(p0 *ExpressionTextLiteral) error {
	if b.FuncExitExpressionTextLiteral == nil {
		return nil
	}

	return b.FuncExitExpressionTextLiteral(p0)
}

func (b *ImplementedListener) EnterExpressionNumericLiteral(p0 *ExpressionNumericLiteral) error {
	if b.FuncEnterExpressionNumericLiteral == nil {
		return nil
	}

	return b.FuncEnterExpressionNumericLiteral(p0)
}

func (b *ImplementedListener) ExitExpressionNumericLiteral(p0 *ExpressionNumericLiteral) error {
	if b.FuncExitExpressionNumericLiteral == nil {
		return nil
	}

	return b.FuncExitExpressionNumericLiteral(p0)
}

func (b *ImplementedListener) EnterExpressionBooleanLiteral(p0 *ExpressionBooleanLiteral) error {
	if b.FuncEnterExpressionBooleanLiteral == nil {
		return nil
	}

	return b.FuncEnterExpressionBooleanLiteral(p0)
}

func (b *ImplementedListener) ExitExpressionBooleanLiteral(p0 *ExpressionBooleanLiteral) error {
	if b.FuncExitExpressionBooleanLiteral == nil {
		return nil
	}

	return b.FuncExitExpressionBooleanLiteral(p0)
}

func (b *ImplementedListener) EnterExpressionNullLiteral(p0 *ExpressionNullLiteral) error {
	if b.FuncEnterExpressionNullLiteral == nil {
		return nil
	}

	return b.FuncEnterExpressionNullLiteral(p0)
}

func (b *ImplementedListener) ExitExpressionNullLiteral(p0 *ExpressionNullLiteral) error {
	if b.FuncExitExpressionNullLiteral == nil {
		return nil
	}

	return b.FuncExitExpressionNullLiteral(p0)
}

func (b *ImplementedListener) EnterExpressionBlobLiteral(p0 *ExpressionBlobLiteral) error {
	if b.FuncEnterExpressionBlobLiteral == nil {
		return nil
	}

	return b.FuncEnterExpressionBlobLiteral(p0)
}

func (b *ImplementedListener) ExitExpressionBlobLiteral(p0 *ExpressionBlobLiteral) error {
	if b.FuncExitExpressionBlobLiteral == nil {
		return nil
	}

	return b.FuncExitExpressionBlobLiteral(p0)
}

func (b *ImplementedListener) EnterExpressionSelect(p0 *ExpressionSelect) error {
	if b.FuncEnterExpressionSelect == nil {
		return nil
	}

	return b.FuncEnterExpressionSelect(p0)
}

func (b *ImplementedListener) ExitExpressionSelect(p0 *ExpressionSelect) error {
	if b.FuncExitExpressionSelect == nil {
		return nil
	}

	return b.FuncExitExpressionSelect(p0)
}

func (b *ImplementedListener) EnterExpressionStringCompare(p0 *ExpressionStringCompare) error {
	if b.FuncEnterExpressionStringCompare == nil {
		return nil
	}

	return b.FuncEnterExpressionStringCompare(p0)
}

func (b *ImplementedListener) ExitExpressionStringCompare(p0 *ExpressionStringCompare) error {
	if b.FuncExitExpressionStringCompare == nil {
		return nil
	}

	return b.FuncExitExpressionStringCompare(p0)
}

func (b *ImplementedListener) EnterExpressionUnary(p0 *ExpressionUnary) error {
	if b.FuncEnterExpressionUnary == nil {
		return nil
	}

	return b.FuncEnterExpressionUnary(p0)
}

func (b *ImplementedListener) ExitExpressionUnary(p0 *ExpressionUnary) error {
	if b.FuncExitExpressionUnary == nil {
		return nil
	}

	return b.FuncExitExpressionUnary(p0)
}

func (b *ImplementedListener) EnterGroupBy(p0 *GroupBy) error {
	if b.FuncEnterGroupBy == nil {
		return nil
	}

	return b.FuncEnterGroupBy(p0)
}

func (b *ImplementedListener) ExitGroupBy(p0 *GroupBy) error {
	if b.FuncExitGroupBy == nil {
		return nil
	}

	return b.FuncExitGroupBy(p0)
}

func (b *ImplementedListener) EnterInsertStmt(p0 *InsertStmt) error {
	if b.FuncEnterInsertStmt == nil {
		return nil
	}

	return b.FuncEnterInsertStmt(p0)
}

func (b *ImplementedListener) ExitInsertStmt(p0 *InsertStmt) error {
	if b.FuncExitInsertStmt == nil {
		return nil
	}

	return b.FuncExitInsertStmt(p0)
}

func (b *ImplementedListener) EnterInsertCore(p0 *InsertCore) error {
	if b.FuncEnterInsertCore == nil {
		return nil
	}

	return b.FuncEnterInsertCore(p0)
}

func (b *ImplementedListener) ExitInsertCore(p0 *InsertCore) error {
	if b.FuncExitInsertCore == nil {
		return nil
	}

	return b.FuncExitInsertCore(p0)
}

func (b *ImplementedListener) EnterJoinOperator(p0 *JoinOperator) error {
	if b.FuncEnterJoinOperator == nil {
		return nil
	}

	return b.FuncEnterJoinOperator(p0)
}

func (b *ImplementedListener) ExitJoinOperator(p0 *JoinOperator) error {
	if b.FuncExitJoinOperator == nil {
		return nil
	}

	return b.FuncExitJoinOperator(p0)
}

func (b *ImplementedListener) EnterJoinPredicate(p0 *JoinPredicate) error {
	if b.FuncEnterJoinPredicate == nil {
		return nil
	}

	return b.FuncEnterJoinPredicate(p0)
}

func (b *ImplementedListener) ExitJoinPredicate(p0 *JoinPredicate) error {
	if b.FuncExitJoinPredicate == nil {
		return nil
	}

	return b.FuncExitJoinPredicate(p0)
}

func (b *ImplementedListener) EnterLimit(p0 *Limit) error {
	if b.FuncEnterLimit == nil {
		return nil
	}

	return b.FuncEnterLimit(p0)
}

func (b *ImplementedListener) ExitLimit(p0 *Limit) error {
	if b.FuncExitLimit == nil {
		return nil
	}

	return b.FuncExitLimit(p0)
}

func (b *ImplementedListener) EnterOrderBy(p0 *OrderBy) error {
	if b.FuncEnterOrderBy == nil {
		return nil
	}

	return b.FuncEnterOrderBy(p0)
}

func (b *ImplementedListener) ExitOrderBy(p0 *OrderBy) error {
	if b.FuncExitOrderBy == nil {
		return nil
	}

	return b.FuncExitOrderBy(p0)
}

func (b *ImplementedListener) EnterOrderingTerm(p0 *OrderingTerm) error {
	if b.FuncEnterOrderingTerm == nil {
		return nil
	}

	return b.FuncEnterOrderingTerm(p0)
}

func (b *ImplementedListener) ExitOrderingTerm(p0 *OrderingTerm) error {
	if b.FuncExitOrderingTerm == nil {
		return nil
	}

	return b.FuncExitOrderingTerm(p0)
}

func (b *ImplementedListener) EnterQualifiedTableName(p0 *QualifiedTableName) error {
	if b.FuncEnterQualifiedTableName == nil {
		return nil
	}

	return b.FuncEnterQualifiedTableName(p0)
}

func (b *ImplementedListener) ExitQualifiedTableName(p0 *QualifiedTableName) error {
	if b.FuncExitQualifiedTableName == nil {
		return nil
	}

	return b.FuncExitQualifiedTableName(p0)
}

func (b *ImplementedListener) EnterRelation(p0 Relation) error {
	if b.FuncEnterRelation == nil {
		return nil
	}

	return b.FuncEnterRelation(p0)
}

func (b *ImplementedListener) ExitRelation(p0 Relation) error {
	if b.FuncExitRelation == nil {
		return nil
	}

	return b.FuncExitRelation(p0)
}

func (b *ImplementedListener) EnterRelationJoin(p0 *RelationJoin) error {
	if b.FuncEnterRelationJoin == nil {
		return nil
	}

	return b.FuncEnterRelationJoin(p0)
}

func (b *ImplementedListener) ExitRelationJoin(p0 *RelationJoin) error {
	if b.FuncExitRelationJoin == nil {
		return nil
	}

	return b.FuncExitRelationJoin(p0)
}

func (b *ImplementedListener) EnterRelationSubquery(p0 *RelationSubquery) error {
	if b.FuncEnterRelationSubquery == nil {
		return nil
	}

	return b.FuncEnterRelationSubquery(p0)
}

func (b *ImplementedListener) ExitRelationSubquery(p0 *RelationSubquery) error {
	if b.FuncExitRelationSubquery == nil {
		return nil
	}

	return b.FuncExitRelationSubquery(p0)
}

func (b *ImplementedListener) EnterRelationTable(p0 *RelationTable) error {
	if b.FuncEnterRelationTable == nil {
		return nil
	}

	return b.FuncEnterRelationTable(p0)
}

func (b *ImplementedListener) ExitRelationTable(p0 *RelationTable) error {
	if b.FuncExitRelationTable == nil {
		return nil
	}

	return b.FuncExitRelationTable(p0)
}

func (b *ImplementedListener) EnterResultColumnExpression(p0 *ResultColumnExpression) error {
	if b.FuncEnterResultColumnExpression == nil {
		return nil
	}

	return b.FuncEnterResultColumnExpression(p0)
}

func (b *ImplementedListener) ExitResultColumnExpression(p0 *ResultColumnExpression) error {
	if b.FuncExitResultColumnExpression == nil {
		return nil
	}

	return b.FuncExitResultColumnExpression(p0)
}

func (b *ImplementedListener) EnterResultColumnStar(p0 *ResultColumnStar) error {
	if b.FuncEnterResultColumnStar == nil {
		return nil
	}

	return b.FuncEnterResultColumnStar(p0)
}

func (b *ImplementedListener) ExitResultColumnStar(p0 *ResultColumnStar) error {
	if b.FuncExitResultColumnStar == nil {
		return nil
	}

	return b.FuncExitResultColumnStar(p0)
}

func (b *ImplementedListener) EnterResultColumnTable(p0 *ResultColumnTable) error {
	if b.FuncEnterResultColumnTable == nil {
		return nil
	}

	return b.FuncEnterResultColumnTable(p0)
}

func (b *ImplementedListener) ExitResultColumnTable(p0 *ResultColumnTable) error {
	if b.FuncExitResultColumnTable == nil {
		return nil
	}

	return b.FuncExitResultColumnTable(p0)
}

func (b *ImplementedListener) EnterReturningClause(p0 *ReturningClause) error {
	if b.FuncEnterReturningClause == nil {
		return nil
	}

	return b.FuncEnterReturningClause(p0)
}

func (b *ImplementedListener) ExitReturningClause(p0 *ReturningClause) error {
	if b.FuncExitReturningClause == nil {
		return nil
	}

	return b.FuncExitReturningClause(p0)
}

func (b *ImplementedListener) EnterReturningClauseColumn(p0 *ReturningClauseColumn) error {
	if b.FuncEnterReturningClauseColumn == nil {
		return nil
	}

	return b.FuncEnterReturningClauseColumn(p0)
}

func (b *ImplementedListener) ExitReturningClauseColumn(p0 *ReturningClauseColumn) error {
	if b.FuncExitReturningClauseColumn == nil {
		return nil
	}

	return b.FuncExitReturningClauseColumn(p0)
}

func (b *ImplementedListener) EnterScalarFunc(p0 *ScalarFunction) error {
	if b.FuncEnterScalarFunc == nil {
		return nil
	}

	return b.FuncEnterScalarFunc(p0)
}

func (b *ImplementedListener) ExitScalarFunc(p0 *ScalarFunction) error {
	if b.FuncExitScalarFunc == nil {
		return nil
	}

	return b.FuncExitScalarFunc(p0)
}

func (b *ImplementedListener) EnterSelectStmt(p0 *SelectStmt) error {
	if b.FuncEnterSelectStmt == nil {
		return nil
	}

	return b.FuncEnterSelectStmt(p0)
}

func (b *ImplementedListener) ExitSelectStmt(p0 *SelectStmt) error {
	if b.FuncExitSelectStmt == nil {
		return nil
	}

	return b.FuncExitSelectStmt(p0)
}

func (b *ImplementedListener) EnterSimpleSelect(p0 *SimpleSelect) error {
	if b.FuncEnterSimpleSelect == nil {
		return nil
	}

	return b.FuncEnterSimpleSelect(p0)
}

func (b *ImplementedListener) ExitSimpleSelect(p0 *SimpleSelect) error {
	if b.FuncExitSimpleSelect == nil {
		return nil
	}

	return b.FuncExitSimpleSelect(p0)
}

func (b *ImplementedListener) EnterSelectCore(p0 *SelectCore) error {
	if b.FuncEnterSelectCore == nil {
		return nil
	}

	return b.FuncEnterSelectCore(p0)
}

func (b *ImplementedListener) ExitSelectCore(p0 *SelectCore) error {
	if b.FuncExitSelectCore == nil {
		return nil
	}

	return b.FuncExitSelectCore(p0)
}

func (b *ImplementedListener) EnterUpdateStmt(p0 *UpdateStmt) error {
	if b.FuncEnterUpdateStmt == nil {
		return nil
	}

	return b.FuncEnterUpdateStmt(p0)
}

func (b *ImplementedListener) ExitUpdateStmt(p0 *UpdateStmt) error {
	if b.FuncExitUpdateStmt == nil {
		return nil
	}

	return b.FuncExitUpdateStmt(p0)
}

func (b *ImplementedListener) EnterUpdateSetClause(p0 *UpdateSetClause) error {
	if b.FuncEnterUpdateSetClause == nil {
		return nil
	}

	return b.FuncEnterUpdateSetClause(p0)
}

func (b *ImplementedListener) ExitUpdateSetClause(p0 *UpdateSetClause) error {
	if b.FuncExitUpdateSetClause == nil {
		return nil
	}

	return b.FuncExitUpdateSetClause(p0)
}

func (b *ImplementedListener) EnterUpdateCore(p0 *UpdateCore) error {
	if b.FuncEnterUpdateCore == nil {
		return nil
	}

	return b.FuncEnterUpdateCore(p0)
}

func (b *ImplementedListener) ExitUpdateCore(p0 *UpdateCore) error {
	if b.FuncExitUpdateCore == nil {
		return nil
	}

	return b.FuncExitUpdateCore(p0)
}

func (b *ImplementedListener) EnterUpsert(p0 *Upsert) error {
	if b.FuncEnterUpsert == nil {
		return nil
	}

	return b.FuncEnterUpsert(p0)
}

func (b *ImplementedListener) ExitUpsert(p0 *Upsert) error {
	if b.FuncExitUpsert == nil {
		return nil
	}

	return b.FuncExitUpsert(p0)
}
