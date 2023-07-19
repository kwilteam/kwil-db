package tree

type accepter interface {
	Accept(Visitor) error
}

func accept(v Visitor, a accepter) error {
	if a == nil {
		return nil
	}

	return a.Accept(v)
}

func acceptMany[T accepter](v Visitor, as []T) error {
	for _, a := range as {
		err := accept(v, a)
		if err != nil {
			return err
		}
	}

	return nil
}

type Visitor interface {
	VisitAggregateFunc(*AggregateFunc) error
	VisitConflictTarget(*ConflictTarget) error
	VisitCTE(*CTE) error
	VisitDateTimeFunc(*DateTimeFunction) error
	VisitDelete(*Delete) error
	VisitDeleteStmt(*DeleteStmt) error
	VisitExpressionLiteral(*ExpressionLiteral) error
	VisitExpressionBindParameter(*ExpressionBindParameter) error
	VisitExpressionColumn(*ExpressionColumn) error
	VisitExpressionUnary(*ExpressionUnary) error
	VisitExpressionBinaryComparison(*ExpressionBinaryComparison) error
	VisitExpressionFunction(*ExpressionFunction) error
	VisitExpressionList(*ExpressionList) error
	VisitExpressionCollate(*ExpressionCollate) error
	VisitExpressionStringCompare(*ExpressionStringCompare) error
	VisitExpressionIsNull(*ExpressionIsNull) error
	VisitExpressionDistinct(*ExpressionDistinct) error
	VisitExpressionBetween(*ExpressionBetween) error
	VisitExpressionSelect(*ExpressionSelect) error
	VisitExpressionCase(*ExpressionCase) error
	VisitExpressionArithmetic(*ExpressionArithmetic) error
	VisitScalarFunc(*ScalarFunction) error
	VisitGroupBy(*GroupBy) error
	VisitInsert(*Insert) error
	VisitInsertStmt(*InsertStmt) error
	VisitJoinClause(*JoinClause) error
	VisitJoinPredicate(*JoinPredicate) error
	VisitJoinOperator(*JoinOperator) error
	VisitLimit(*Limit) error
	VisitOrderBy(*OrderBy) error
	VisitOrderingTerm(*OrderingTerm) error
	VisitQualifiedTableName(*QualifiedTableName) error
	VisitResultColumnStar(*ResultColumnStar) error
	VisitResultColumnExpression(*ResultColumnExpression) error
	VisitResultColumnTable(*ResultColumnTable) error
	VisitReturningClause(*ReturningClause) error
	VisitReturningClauseColumn(*ReturningClauseColumn) error
	VisitSelect(*Select) error
	VisitSelectCore(*SelectCore) error
	VisitSelectStmt(*SelectStmt) error
	VisitFromClause(*FromClause) error
	VisitCompoundOperator(*CompoundOperator) error
	VisitTableOrSubqueryTable(*TableOrSubqueryTable) error
	VisitTableOrSubquerySelect(*TableOrSubquerySelect) error
	VisitTableOrSubqueryList(*TableOrSubqueryList) error
	VisitTableOrSubqueryJoin(*TableOrSubqueryJoin) error
	VisitUpdateSetClause(*UpdateSetClause) error
	VisitUpdate(*Update) error
	VisitUpdateStmt(*UpdateStmt) error
	VisitUpsert(*Upsert) error
}

func NewBaseVisitor() *BaseVisitor {
	return &BaseVisitor{}
}

type BaseVisitor struct {
}

func (b *BaseVisitor) VisitAggregateFunc(p0 *AggregateFunc) error {
	return nil
}

func (b *BaseVisitor) VisitCTE(p0 *CTE) error {
	return nil
}

func (b *BaseVisitor) VisitCompoundOperator(p0 *CompoundOperator) error {
	return nil
}

func (b *BaseVisitor) VisitConflictTarget(p0 *ConflictTarget) error {
	return nil
}

func (b *BaseVisitor) VisitDateTimeFunc(p0 *DateTimeFunction) error {
	return nil
}

func (b *BaseVisitor) VisitDelete(p0 *Delete) error {
	return nil
}

func (b *BaseVisitor) VisitDeleteStmt(p0 *DeleteStmt) error {
	return nil
}

func (b *BaseVisitor) VisitExpressionArithmetic(p0 *ExpressionArithmetic) error {
	return nil
}

func (b *BaseVisitor) VisitExpressionBetween(p0 *ExpressionBetween) error {
	return nil
}

func (b *BaseVisitor) VisitExpressionBinaryComparison(p0 *ExpressionBinaryComparison) error {
	return nil
}

func (b *BaseVisitor) VisitExpressionBindParameter(p0 *ExpressionBindParameter) error {
	return nil
}

func (b *BaseVisitor) VisitExpressionCase(p0 *ExpressionCase) error {
	return nil
}

func (b *BaseVisitor) VisitExpressionCollate(p0 *ExpressionCollate) error {
	return nil
}

func (b *BaseVisitor) VisitExpressionColumn(p0 *ExpressionColumn) error {
	return nil
}

func (b *BaseVisitor) VisitExpressionDistinct(p0 *ExpressionDistinct) error {
	return nil
}

func (b *BaseVisitor) VisitExpressionFunction(p0 *ExpressionFunction) error {
	return nil
}

func (b *BaseVisitor) VisitExpressionIsNull(p0 *ExpressionIsNull) error {
	return nil
}

func (b *BaseVisitor) VisitExpressionList(p0 *ExpressionList) error {
	return nil
}

func (b *BaseVisitor) VisitExpressionLiteral(p0 *ExpressionLiteral) error {
	return nil
}

func (b *BaseVisitor) VisitExpressionSelect(p0 *ExpressionSelect) error {
	return nil
}

func (b *BaseVisitor) VisitExpressionStringCompare(p0 *ExpressionStringCompare) error {
	return nil
}

func (b *BaseVisitor) VisitExpressionUnary(p0 *ExpressionUnary) error {
	return nil
}

func (b *BaseVisitor) VisitFromClause(p0 *FromClause) error {
	return nil
}

func (b *BaseVisitor) VisitGroupBy(p0 *GroupBy) error {
	return nil
}

func (b *BaseVisitor) VisitInsert(p0 *Insert) error {
	return nil
}

func (b *BaseVisitor) VisitInsertStmt(p0 *InsertStmt) error {
	return nil
}

func (b *BaseVisitor) VisitJoinClause(p0 *JoinClause) error {
	return nil
}

func (b *BaseVisitor) VisitJoinOperator(p0 *JoinOperator) error {
	return nil
}

func (b *BaseVisitor) VisitJoinPredicate(p0 *JoinPredicate) error {
	return nil
}

func (b *BaseVisitor) VisitLimit(p0 *Limit) error {
	return nil
}

func (b *BaseVisitor) VisitOrderBy(p0 *OrderBy) error {
	return nil
}

func (b *BaseVisitor) VisitOrderingTerm(p0 *OrderingTerm) error {
	return nil
}

func (b *BaseVisitor) VisitQualifiedTableName(p0 *QualifiedTableName) error {
	return nil
}

func (b *BaseVisitor) VisitResultColumnExpression(p0 *ResultColumnExpression) error {
	return nil
}

func (b *BaseVisitor) VisitResultColumnStar(p0 *ResultColumnStar) error {
	return nil
}

func (b *BaseVisitor) VisitResultColumnTable(p0 *ResultColumnTable) error {
	return nil
}

func (b *BaseVisitor) VisitReturningClause(p0 *ReturningClause) error {
	return nil
}

func (b *BaseVisitor) VisitReturningClauseColumn(p0 *ReturningClauseColumn) error {
	return nil
}

func (b *BaseVisitor) VisitScalarFunc(p0 *ScalarFunction) error {
	return nil
}

func (b *BaseVisitor) VisitSelect(p0 *Select) error {
	return nil
}

func (b *BaseVisitor) VisitSelectCore(p0 *SelectCore) error {
	return nil
}

func (b *BaseVisitor) VisitSelectStmt(p0 *SelectStmt) error {
	return nil
}

func (b *BaseVisitor) VisitTableOrSubqueryJoin(p0 *TableOrSubqueryJoin) error {
	return nil
}

func (b *BaseVisitor) VisitTableOrSubqueryList(p0 *TableOrSubqueryList) error {
	return nil
}

func (b *BaseVisitor) VisitTableOrSubquerySelect(p0 *TableOrSubquerySelect) error {
	return nil
}

func (b *BaseVisitor) VisitTableOrSubqueryTable(p0 *TableOrSubqueryTable) error {
	return nil
}

func (b *BaseVisitor) VisitUpdate(p0 *Update) error {
	return nil
}

func (b *BaseVisitor) VisitUpdateSetClause(p0 *UpdateSetClause) error {
	return nil
}

func (b *BaseVisitor) VisitUpdateStmt(p0 *UpdateStmt) error {
	return nil
}

func (b *BaseVisitor) VisitUpsert(p0 *Upsert) error {
	return nil
}

type ExtendableVisitor interface {
	Visitor
	AddVisitor(Visitor)
}

type extendableVisitor struct {
	decorators []Visitor
}

func NewExtendableVisitor(decorators ...Visitor) ExtendableVisitor {
	return &extendableVisitor{
		decorators: decorators,
	}
}

func (b *extendableVisitor) runDecorators(fn func(visitor Visitor) error) error {
	for _, decorator := range b.decorators {
		if err := fn(decorator); err != nil {
			return err
		}
	}
	return nil
}

func (b *extendableVisitor) VisitAggregateFunc(p0 *AggregateFunc) error {
	return b.runDecorators(func(v Visitor) error {
		return v.VisitAggregateFunc(p0)
	})
}

func (b *extendableVisitor) VisitCTE(p0 *CTE) error {
	return b.runDecorators(func(v Visitor) error {
		return v.VisitCTE(p0)
	})
}

func (b *extendableVisitor) VisitCompoundOperator(p0 *CompoundOperator) error {
	return b.runDecorators(func(v Visitor) error {
		return v.VisitCompoundOperator(p0)
	})
}

func (b *extendableVisitor) VisitConflictTarget(p0 *ConflictTarget) error {
	return b.runDecorators(func(v Visitor) error {
		return v.VisitConflictTarget(p0)
	})
}

func (b *extendableVisitor) VisitDateTimeFunc(p0 *DateTimeFunction) error {
	return b.runDecorators(func(v Visitor) error {
		return v.VisitDateTimeFunc(p0)
	})
}

func (b *extendableVisitor) VisitDelete(p0 *Delete) error {
	return b.runDecorators(func(v Visitor) error {
		return v.VisitDelete(p0)
	})
}

func (b *extendableVisitor) VisitDeleteStmt(p0 *DeleteStmt) error {
	return b.runDecorators(func(v Visitor) error {
		return v.VisitDeleteStmt(p0)
	})
}

func (b *extendableVisitor) VisitExpressionArithmetic(p0 *ExpressionArithmetic) error {
	return b.runDecorators(func(v Visitor) error {
		return v.VisitExpressionArithmetic(p0)
	})
}

func (b *extendableVisitor) VisitExpressionBetween(p0 *ExpressionBetween) error {
	return b.runDecorators(func(v Visitor) error {
		return v.VisitExpressionBetween(p0)
	})
}

func (b *extendableVisitor) VisitExpressionBinaryComparison(p0 *ExpressionBinaryComparison) error {
	return b.runDecorators(func(v Visitor) error {
		return v.VisitExpressionBinaryComparison(p0)
	})
}

func (b *extendableVisitor) VisitExpressionBindParameter(p0 *ExpressionBindParameter) error {
	return b.runDecorators(func(v Visitor) error {
		return v.VisitExpressionBindParameter(p0)
	})
}

func (b *extendableVisitor) VisitExpressionCase(p0 *ExpressionCase) error {
	return b.runDecorators(func(v Visitor) error {
		return v.VisitExpressionCase(p0)
	})
}

func (b *extendableVisitor) VisitExpressionCollate(p0 *ExpressionCollate) error {
	return b.runDecorators(func(v Visitor) error {
		return v.VisitExpressionCollate(p0)
	})
}

func (b *extendableVisitor) VisitExpressionColumn(p0 *ExpressionColumn) error {
	return b.runDecorators(func(v Visitor) error {
		return v.VisitExpressionColumn(p0)
	})
}

func (b *extendableVisitor) VisitExpressionDistinct(p0 *ExpressionDistinct) error {
	return b.runDecorators(func(v Visitor) error {
		return v.VisitExpressionDistinct(p0)
	})
}

func (b *extendableVisitor) VisitExpressionFunction(p0 *ExpressionFunction) error {
	return b.runDecorators(func(v Visitor) error {
		return v.VisitExpressionFunction(p0)
	})
}

func (b *extendableVisitor) VisitExpressionIsNull(p0 *ExpressionIsNull) error {
	return b.runDecorators(func(v Visitor) error {
		return v.VisitExpressionIsNull(p0)
	})
}

func (b *extendableVisitor) VisitExpressionList(p0 *ExpressionList) error {
	return b.runDecorators(func(v Visitor) error {
		return v.VisitExpressionList(p0)
	})
}

func (b *extendableVisitor) VisitExpressionLiteral(p0 *ExpressionLiteral) error {
	return b.runDecorators(func(v Visitor) error {
		return v.VisitExpressionLiteral(p0)
	})
}

func (b *extendableVisitor) VisitExpressionSelect(p0 *ExpressionSelect) error {
	return b.runDecorators(func(v Visitor) error {
		return v.VisitExpressionSelect(p0)
	})
}

func (b *extendableVisitor) VisitExpressionStringCompare(p0 *ExpressionStringCompare) error {
	return b.runDecorators(func(v Visitor) error {
		return v.VisitExpressionStringCompare(p0)
	})
}

func (b *extendableVisitor) VisitExpressionUnary(p0 *ExpressionUnary) error {
	return b.runDecorators(func(v Visitor) error {
		return v.VisitExpressionUnary(p0)
	})
}

func (b *extendableVisitor) VisitFromClause(p0 *FromClause) error {
	return b.runDecorators(func(v Visitor) error {
		return v.VisitFromClause(p0)
	})
}

func (b *extendableVisitor) VisitGroupBy(p0 *GroupBy) error {
	return b.runDecorators(func(v Visitor) error {
		return v.VisitGroupBy(p0)
	})
}

func (b *extendableVisitor) VisitInsert(p0 *Insert) error {
	return b.runDecorators(func(v Visitor) error {
		return v.VisitInsert(p0)
	})
}

func (b *extendableVisitor) VisitInsertStmt(p0 *InsertStmt) error {
	return b.runDecorators(func(v Visitor) error {
		return v.VisitInsertStmt(p0)
	})
}

func (b *extendableVisitor) VisitJoinClause(p0 *JoinClause) error {
	return b.runDecorators(func(v Visitor) error {
		return v.VisitJoinClause(p0)
	})
}

func (b *extendableVisitor) VisitJoinOperator(p0 *JoinOperator) error {
	return b.runDecorators(func(v Visitor) error {
		return v.VisitJoinOperator(p0)
	})
}

func (b *extendableVisitor) VisitJoinPredicate(p0 *JoinPredicate) error {
	return b.runDecorators(func(v Visitor) error {
		return v.VisitJoinPredicate(p0)
	})
}

func (b *extendableVisitor) VisitLimit(p0 *Limit) error {
	return b.runDecorators(func(v Visitor) error {
		return v.VisitLimit(p0)
	})
}

func (b *extendableVisitor) VisitOrderBy(p0 *OrderBy) error {
	return b.runDecorators(func(v Visitor) error {
		return v.VisitOrderBy(p0)
	})
}

func (b *extendableVisitor) VisitOrderingTerm(p0 *OrderingTerm) error {
	return b.runDecorators(func(v Visitor) error {
		return v.VisitOrderingTerm(p0)
	})
}

func (b *extendableVisitor) VisitQualifiedTableName(p0 *QualifiedTableName) error {
	return b.runDecorators(func(v Visitor) error {
		return v.VisitQualifiedTableName(p0)
	})
}

func (b *extendableVisitor) VisitResultColumnExpression(p0 *ResultColumnExpression) error {
	return b.runDecorators(func(v Visitor) error {
		return v.VisitResultColumnExpression(p0)
	})
}

func (b *extendableVisitor) VisitResultColumnStar(p0 *ResultColumnStar) error {
	return b.runDecorators(func(v Visitor) error {
		return v.VisitResultColumnStar(p0)
	})
}

func (b *extendableVisitor) VisitResultColumnTable(p0 *ResultColumnTable) error {
	return b.runDecorators(func(v Visitor) error {
		return v.VisitResultColumnTable(p0)
	})
}

func (b *extendableVisitor) VisitReturningClause(p0 *ReturningClause) error {
	return b.runDecorators(func(v Visitor) error {
		return v.VisitReturningClause(p0)
	})
}

func (b *extendableVisitor) VisitReturningClauseColumn(p0 *ReturningClauseColumn) error {
	return b.runDecorators(func(v Visitor) error {
		return v.VisitReturningClauseColumn(p0)
	})
}

func (b *extendableVisitor) VisitScalarFunc(p0 *ScalarFunction) error {
	return b.runDecorators(func(v Visitor) error {
		return v.VisitScalarFunc(p0)
	})
}

func (b *extendableVisitor) VisitSelect(p0 *Select) error {
	return b.runDecorators(func(v Visitor) error {
		return v.VisitSelect(p0)
	})
}

func (b *extendableVisitor) VisitSelectCore(p0 *SelectCore) error {
	return b.runDecorators(func(v Visitor) error {
		return v.VisitSelectCore(p0)
	})
}

func (b *extendableVisitor) VisitSelectStmt(p0 *SelectStmt) error {
	return b.runDecorators(func(v Visitor) error {
		return v.VisitSelectStmt(p0)
	})
}

func (b *extendableVisitor) VisitTableOrSubqueryJoin(p0 *TableOrSubqueryJoin) error {
	return b.runDecorators(func(v Visitor) error {
		return v.VisitTableOrSubqueryJoin(p0)
	})
}

func (b *extendableVisitor) VisitTableOrSubqueryList(p0 *TableOrSubqueryList) error {
	return b.runDecorators(func(v Visitor) error {
		return v.VisitTableOrSubqueryList(p0)
	})
}

func (b *extendableVisitor) VisitTableOrSubquerySelect(p0 *TableOrSubquerySelect) error {
	return b.runDecorators(func(v Visitor) error {
		return v.VisitTableOrSubquerySelect(p0)
	})
}

func (b *extendableVisitor) VisitTableOrSubqueryTable(p0 *TableOrSubqueryTable) error {
	return b.runDecorators(func(v Visitor) error {
		return v.VisitTableOrSubqueryTable(p0)
	})
}

func (b *extendableVisitor) VisitUpdate(p0 *Update) error {
	return b.runDecorators(func(v Visitor) error {
		return v.VisitUpdate(p0)
	})
}

func (b *extendableVisitor) VisitUpdateSetClause(p0 *UpdateSetClause) error {
	return b.runDecorators(func(v Visitor) error {
		return v.VisitUpdateSetClause(p0)
	})
}

func (b *extendableVisitor) VisitUpdateStmt(p0 *UpdateStmt) error {
	return b.runDecorators(func(v Visitor) error {
		return v.VisitUpdateStmt(p0)
	})
}

func (b *extendableVisitor) VisitUpsert(p0 *Upsert) error {
	return b.runDecorators(func(v Visitor) error {
		return v.VisitUpsert(p0)
	})
}

func (b *extendableVisitor) AddVisitor(v Visitor) {
	b.decorators = append(b.decorators, v)
}
