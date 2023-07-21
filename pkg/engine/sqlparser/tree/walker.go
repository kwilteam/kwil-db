package tree

import (
	"errors"
	"reflect"
)

func run(errs ...error) error {
	return errors.Join(errs...)
}

type Accepter interface {
	Accept(Walker) error
}

func isNil(input interface{}) bool {
	if input == nil {
		return true
	}
	kind := reflect.ValueOf(input).Kind()
	switch kind {
	case reflect.Ptr, reflect.Map, reflect.Slice, reflect.Chan:
		return reflect.ValueOf(input).IsNil()
	default:
		return false
	}
}

func accept(v Walker, a Accepter) error {
	if isNil(a) {
		return nil
	}

	return a.Accept(v)
}

func acceptMany[T Accepter](v Walker, as []T) error {
	for _, a := range as {
		err := accept(v, a)
		if err != nil {
			return err
		}
	}

	return nil
}

type Walker interface {
	EnterAggregateFunc(*AggregateFunc) error
	ExitAggregateFunc(*AggregateFunc) error
	EnterConflictTarget(*ConflictTarget) error
	ExitConflictTarget(*ConflictTarget) error
	EnterCTE(*CTE) error
	ExitCTE(*CTE) error
	EnterDateTimeFunc(*DateTimeFunction) error
	ExitDateTimeFunc(*DateTimeFunction) error
	EnterDelete(*Delete) error
	ExitDelete(*Delete) error
	EnterDeleteStmt(*DeleteStmt) error
	ExitDeleteStmt(*DeleteStmt) error
	EnterExpressionLiteral(*ExpressionLiteral) error
	ExitExpressionLiteral(*ExpressionLiteral) error
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
	EnterExpressionIsNull(*ExpressionIsNull) error
	ExitExpressionIsNull(*ExpressionIsNull) error
	EnterExpressionDistinct(*ExpressionDistinct) error
	ExitExpressionDistinct(*ExpressionDistinct) error
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
	EnterInsert(*Insert) error
	ExitInsert(*Insert) error
	EnterInsertStmt(*InsertStmt) error
	ExitInsertStmt(*InsertStmt) error
	EnterJoinClause(*JoinClause) error
	ExitJoinClause(*JoinClause) error
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
	EnterSelect(*Select) error
	ExitSelect(*Select) error
	EnterSelectCore(*SelectCore) error
	ExitSelectCore(*SelectCore) error
	EnterSelectStmt(*SelectStmt) error
	ExitSelectStmt(*SelectStmt) error
	EnterFromClause(*FromClause) error
	ExitFromClause(*FromClause) error
	EnterCompoundOperator(*CompoundOperator) error
	ExitCompoundOperator(*CompoundOperator) error
	EnterTableOrSubqueryTable(*TableOrSubqueryTable) error
	ExitTableOrSubqueryTable(*TableOrSubqueryTable) error
	EnterTableOrSubquerySelect(*TableOrSubquerySelect) error
	ExitTableOrSubquerySelect(*TableOrSubquerySelect) error
	EnterTableOrSubqueryList(*TableOrSubqueryList) error
	ExitTableOrSubqueryList(*TableOrSubqueryList) error
	EnterTableOrSubqueryJoin(*TableOrSubqueryJoin) error
	ExitTableOrSubqueryJoin(*TableOrSubqueryJoin) error
	EnterUpdateSetClause(*UpdateSetClause) error
	ExitUpdateSetClause(*UpdateSetClause) error
	EnterUpdate(*Update) error
	ExitUpdate(*Update) error
	EnterUpdateStmt(*UpdateStmt) error
	ExitUpdateStmt(*UpdateStmt) error
	EnterUpsert(*Upsert) error
	ExitUpsert(*Upsert) error
}

type BaseWalker struct{}

func NewBaseWalker() Walker {
	return &BaseWalker{}
}

func (b *BaseWalker) EnterAggregateFunc(p0 *AggregateFunc) error {
	return nil
}

func (b *BaseWalker) ExitAggregateFunc(p0 *AggregateFunc) error {
	return nil
}

func (b *BaseWalker) EnterCTE(p0 *CTE) error {
	return nil
}

func (b *BaseWalker) ExitCTE(p0 *CTE) error {
	return nil
}

func (b *BaseWalker) EnterCompoundOperator(p0 *CompoundOperator) error {
	return nil
}

func (b *BaseWalker) ExitCompoundOperator(p0 *CompoundOperator) error {
	return nil
}

func (b *BaseWalker) EnterConflictTarget(p0 *ConflictTarget) error {
	return nil
}

func (b *BaseWalker) ExitConflictTarget(p0 *ConflictTarget) error {
	return nil
}

func (b *BaseWalker) EnterDateTimeFunc(p0 *DateTimeFunction) error {
	return nil
}

func (b *BaseWalker) ExitDateTimeFunc(p0 *DateTimeFunction) error {
	return nil
}

func (b *BaseWalker) EnterDelete(p0 *Delete) error {
	return nil
}

func (b *BaseWalker) ExitDelete(p0 *Delete) error {
	return nil
}

func (b *BaseWalker) EnterDeleteStmt(p0 *DeleteStmt) error {
	return nil
}

func (b *BaseWalker) ExitDeleteStmt(p0 *DeleteStmt) error {
	return nil
}

func (b *BaseWalker) EnterExpressionArithmetic(p0 *ExpressionArithmetic) error {
	return nil
}

func (b *BaseWalker) ExitExpressionArithmetic(p0 *ExpressionArithmetic) error {
	return nil
}

func (b *BaseWalker) EnterExpressionBetween(p0 *ExpressionBetween) error {
	return nil
}

func (b *BaseWalker) ExitExpressionBetween(p0 *ExpressionBetween) error {
	return nil
}

func (b *BaseWalker) EnterExpressionBinaryComparison(p0 *ExpressionBinaryComparison) error {
	return nil
}

func (b *BaseWalker) ExitExpressionBinaryComparison(p0 *ExpressionBinaryComparison) error {
	return nil
}

func (b *BaseWalker) EnterExpressionBindParameter(p0 *ExpressionBindParameter) error {
	return nil
}

func (b *BaseWalker) ExitExpressionBindParameter(p0 *ExpressionBindParameter) error {
	return nil
}

func (b *BaseWalker) EnterExpressionCase(p0 *ExpressionCase) error {
	return nil
}

func (b *BaseWalker) ExitExpressionCase(p0 *ExpressionCase) error {
	return nil
}

func (b *BaseWalker) EnterExpressionCollate(p0 *ExpressionCollate) error {
	return nil
}

func (b *BaseWalker) ExitExpressionCollate(p0 *ExpressionCollate) error {
	return nil
}

func (b *BaseWalker) EnterExpressionColumn(p0 *ExpressionColumn) error {
	return nil
}

func (b *BaseWalker) ExitExpressionColumn(p0 *ExpressionColumn) error {
	return nil
}

func (b *BaseWalker) EnterExpressionDistinct(p0 *ExpressionDistinct) error {
	return nil
}

func (b *BaseWalker) ExitExpressionDistinct(p0 *ExpressionDistinct) error {
	return nil
}

func (b *BaseWalker) EnterExpressionFunction(p0 *ExpressionFunction) error {
	return nil
}

func (b *BaseWalker) ExitExpressionFunction(p0 *ExpressionFunction) error {
	return nil
}

func (b *BaseWalker) EnterExpressionIsNull(p0 *ExpressionIsNull) error {
	return nil
}

func (b *BaseWalker) ExitExpressionIsNull(p0 *ExpressionIsNull) error {
	return nil
}

func (b *BaseWalker) EnterExpressionList(p0 *ExpressionList) error {
	return nil
}

func (b *BaseWalker) ExitExpressionList(p0 *ExpressionList) error {
	return nil
}

func (b *BaseWalker) EnterExpressionLiteral(p0 *ExpressionLiteral) error {
	return nil
}

func (b *BaseWalker) ExitExpressionLiteral(p0 *ExpressionLiteral) error {
	return nil
}

func (b *BaseWalker) EnterExpressionSelect(p0 *ExpressionSelect) error {
	return nil
}

func (b *BaseWalker) ExitExpressionSelect(p0 *ExpressionSelect) error {
	return nil
}

func (b *BaseWalker) EnterExpressionStringCompare(p0 *ExpressionStringCompare) error {
	return nil
}

func (b *BaseWalker) ExitExpressionStringCompare(p0 *ExpressionStringCompare) error {
	return nil
}

func (b *BaseWalker) EnterExpressionUnary(p0 *ExpressionUnary) error {
	return nil
}

func (b *BaseWalker) ExitExpressionUnary(p0 *ExpressionUnary) error {
	return nil
}

func (b *BaseWalker) EnterFromClause(p0 *FromClause) error {
	return nil
}

func (b *BaseWalker) ExitFromClause(p0 *FromClause) error {
	return nil
}

func (b *BaseWalker) EnterGroupBy(p0 *GroupBy) error {
	return nil
}

func (b *BaseWalker) ExitGroupBy(p0 *GroupBy) error {
	return nil
}

func (b *BaseWalker) EnterInsert(p0 *Insert) error {
	return nil
}

func (b *BaseWalker) ExitInsert(p0 *Insert) error {
	return nil
}

func (b *BaseWalker) EnterInsertStmt(p0 *InsertStmt) error {
	return nil
}

func (b *BaseWalker) ExitInsertStmt(p0 *InsertStmt) error {
	return nil
}

func (b *BaseWalker) EnterJoinClause(p0 *JoinClause) error {
	return nil
}

func (b *BaseWalker) ExitJoinClause(p0 *JoinClause) error {
	return nil
}

func (b *BaseWalker) EnterJoinOperator(p0 *JoinOperator) error {
	return nil
}

func (b *BaseWalker) ExitJoinOperator(p0 *JoinOperator) error {
	return nil
}

func (b *BaseWalker) EnterJoinPredicate(p0 *JoinPredicate) error {
	return nil
}

func (b *BaseWalker) ExitJoinPredicate(p0 *JoinPredicate) error {
	return nil
}

func (b *BaseWalker) EnterLimit(p0 *Limit) error {
	return nil
}

func (b *BaseWalker) ExitLimit(p0 *Limit) error {
	return nil
}

func (b *BaseWalker) EnterOrderBy(p0 *OrderBy) error {
	return nil
}

func (b *BaseWalker) ExitOrderBy(p0 *OrderBy) error {
	return nil
}

func (b *BaseWalker) EnterOrderingTerm(p0 *OrderingTerm) error {
	return nil
}

func (b *BaseWalker) ExitOrderingTerm(p0 *OrderingTerm) error {
	return nil
}

func (b *BaseWalker) EnterQualifiedTableName(p0 *QualifiedTableName) error {
	return nil
}

func (b *BaseWalker) ExitQualifiedTableName(p0 *QualifiedTableName) error {
	return nil
}

func (b *BaseWalker) EnterResultColumnExpression(p0 *ResultColumnExpression) error {
	return nil
}

func (b *BaseWalker) ExitResultColumnExpression(p0 *ResultColumnExpression) error {
	return nil
}

func (b *BaseWalker) EnterResultColumnStar(p0 *ResultColumnStar) error {
	return nil
}

func (b *BaseWalker) ExitResultColumnStar(p0 *ResultColumnStar) error {
	return nil
}

func (b *BaseWalker) EnterResultColumnTable(p0 *ResultColumnTable) error {
	return nil
}

func (b *BaseWalker) ExitResultColumnTable(p0 *ResultColumnTable) error {
	return nil
}

func (b *BaseWalker) EnterReturningClause(p0 *ReturningClause) error {
	return nil
}

func (b *BaseWalker) ExitReturningClause(p0 *ReturningClause) error {
	return nil
}

func (b *BaseWalker) EnterReturningClauseColumn(p0 *ReturningClauseColumn) error {
	return nil
}

func (b *BaseWalker) ExitReturningClauseColumn(p0 *ReturningClauseColumn) error {
	return nil
}

func (b *BaseWalker) EnterScalarFunc(p0 *ScalarFunction) error {
	return nil
}

func (b *BaseWalker) ExitScalarFunc(p0 *ScalarFunction) error {
	return nil
}

func (b *BaseWalker) EnterSelect(p0 *Select) error {
	return nil
}

func (b *BaseWalker) ExitSelect(p0 *Select) error {
	return nil
}

func (b *BaseWalker) EnterSelectCore(p0 *SelectCore) error {
	return nil
}

func (b *BaseWalker) ExitSelectCore(p0 *SelectCore) error {
	return nil
}

func (b *BaseWalker) EnterSelectStmt(p0 *SelectStmt) error {
	return nil
}

func (b *BaseWalker) ExitSelectStmt(p0 *SelectStmt) error {
	return nil
}

func (b *BaseWalker) EnterTableOrSubqueryJoin(p0 *TableOrSubqueryJoin) error {
	return nil
}

func (b *BaseWalker) ExitTableOrSubqueryJoin(p0 *TableOrSubqueryJoin) error {
	return nil
}

func (b *BaseWalker) EnterTableOrSubqueryList(p0 *TableOrSubqueryList) error {
	return nil
}

func (b *BaseWalker) ExitTableOrSubqueryList(p0 *TableOrSubqueryList) error {
	return nil
}

func (b *BaseWalker) EnterTableOrSubquerySelect(p0 *TableOrSubquerySelect) error {
	return nil
}

func (b *BaseWalker) ExitTableOrSubquerySelect(p0 *TableOrSubquerySelect) error {
	return nil
}

func (b *BaseWalker) EnterTableOrSubqueryTable(p0 *TableOrSubqueryTable) error {
	return nil
}

func (b *BaseWalker) ExitTableOrSubqueryTable(p0 *TableOrSubqueryTable) error {
	return nil
}

func (b *BaseWalker) EnterUpdate(p0 *Update) error {
	return nil
}

func (b *BaseWalker) ExitUpdate(p0 *Update) error {
	return nil
}

func (b *BaseWalker) EnterUpdateSetClause(p0 *UpdateSetClause) error {
	return nil
}

func (b *BaseWalker) ExitUpdateSetClause(p0 *UpdateSetClause) error {
	return nil
}

func (b *BaseWalker) EnterUpdateStmt(p0 *UpdateStmt) error {
	return nil
}

func (b *BaseWalker) ExitUpdateStmt(p0 *UpdateStmt) error {
	return nil
}

func (b *BaseWalker) EnterUpsert(p0 *Upsert) error {
	return nil
}

func (b *BaseWalker) ExitUpsert(p0 *Upsert) error {
	return nil
}
