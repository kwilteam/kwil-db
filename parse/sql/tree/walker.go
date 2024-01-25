package tree

import (
	"errors"
	"reflect"
)

func run(errs ...error) error {
	return errors.Join(errs...)
}

type Walker interface {
	// Walk accepts an AstWalker to walk through itself.
	// NOTE: 'Walk' is a bit misleading. Should we change to another name? like AcceptWalker?
	Walk(AstWalker) error
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

func accept(v AstWalker, a Walker) error {
	if isNil(a) {
		return nil
	}

	return a.Walk(v)
}

func acceptMany[T Walker](v AstWalker, as []T) error {
	for _, a := range as {
		err := accept(v, a)
		if err != nil {
			return err
		}
	}

	return nil
}

type AstWalker interface {
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

var _ AstWalker = &BaseWalker{}

func NewBaseWalker() AstWalker {
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

// ImplementedWalker implements the AstWalker interface.
// Unlike BaseWalker, it holds the methods to be implemented
// as functions in a struct.  This makes it easier to implement
// for small, one-off walkers.
type ImplementedWalker struct {
	FuncEnterAggregateFunc              func(p0 *AggregateFunc) error
	FuncExitAggregateFunc               func(p0 *AggregateFunc) error
	FuncEnterCTE                        func(p0 *CTE) error
	FuncExitCTE                         func(p0 *CTE) error
	FuncEnterCompoundOperator           func(p0 *CompoundOperator) error
	FuncExitCompoundOperator            func(p0 *CompoundOperator) error
	FuncEnterConflictTarget             func(p0 *ConflictTarget) error
	FuncExitConflictTarget              func(p0 *ConflictTarget) error
	FuncEnterDateTimeFunc               func(p0 *DateTimeFunction) error
	FuncExitDateTimeFunc                func(p0 *DateTimeFunction) error
	FuncEnterDelete                     func(p0 *Delete) error
	FuncExitDelete                      func(p0 *Delete) error
	FuncEnterDeleteStmt                 func(p0 *DeleteStmt) error
	FuncExitDeleteStmt                  func(p0 *DeleteStmt) error
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
	FuncEnterExpressionDistinct         func(p0 *ExpressionDistinct) error
	FuncExitExpressionDistinct          func(p0 *ExpressionDistinct) error
	FuncEnterExpressionFunction         func(p0 *ExpressionFunction) error
	FuncExitExpressionFunction          func(p0 *ExpressionFunction) error
	FuncEnterExpressionIsNull           func(p0 *ExpressionIsNull) error
	FuncExitExpressionIsNull            func(p0 *ExpressionIsNull) error
	FuncEnterExpressionList             func(p0 *ExpressionList) error
	FuncExitExpressionList              func(p0 *ExpressionList) error
	FuncEnterExpressionLiteral          func(p0 *ExpressionLiteral) error
	FuncExitExpressionLiteral           func(p0 *ExpressionLiteral) error
	FuncEnterExpressionSelect           func(p0 *ExpressionSelect) error
	FuncExitExpressionSelect            func(p0 *ExpressionSelect) error
	FuncEnterExpressionStringCompare    func(p0 *ExpressionStringCompare) error
	FuncExitExpressionStringCompare     func(p0 *ExpressionStringCompare) error
	FuncEnterExpressionUnary            func(p0 *ExpressionUnary) error
	FuncExitExpressionUnary             func(p0 *ExpressionUnary) error
	FuncEnterFromClause                 func(p0 *FromClause) error
	FuncExitFromClause                  func(p0 *FromClause) error
	FuncEnterGroupBy                    func(p0 *GroupBy) error
	FuncExitGroupBy                     func(p0 *GroupBy) error
	FuncEnterInsert                     func(p0 *Insert) error
	FuncExitInsert                      func(p0 *Insert) error
	FuncEnterInsertStmt                 func(p0 *InsertStmt) error
	FuncExitInsertStmt                  func(p0 *InsertStmt) error
	FuncEnterJoinClause                 func(p0 *JoinClause) error
	FuncExitJoinClause                  func(p0 *JoinClause) error
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
	FuncEnterSelect                     func(p0 *Select) error
	FuncExitSelect                      func(p0 *Select) error
	FuncEnterSelectCore                 func(p0 *SelectCore) error
	FuncExitSelectCore                  func(p0 *SelectCore) error
	FuncEnterSelectStmt                 func(p0 *SelectStmt) error
	FuncExitSelectStmt                  func(p0 *SelectStmt) error
	FuncEnterTableOrSubqueryJoin        func(p0 *TableOrSubqueryJoin) error
	FuncExitTableOrSubqueryJoin         func(p0 *TableOrSubqueryJoin) error
	FuncEnterTableOrSubqueryList        func(p0 *TableOrSubqueryList) error
	FuncExitTableOrSubqueryList         func(p0 *TableOrSubqueryList) error
	FuncEnterTableOrSubquerySelect      func(p0 *TableOrSubquerySelect) error
	FuncExitTableOrSubquerySelect       func(p0 *TableOrSubquerySelect) error
	FuncEnterTableOrSubqueryTable       func(p0 *TableOrSubqueryTable) error
	FuncExitTableOrSubqueryTable        func(p0 *TableOrSubqueryTable) error
	FuncEnterUpdate                     func(p0 *Update) error
	FuncExitUpdate                      func(p0 *Update) error
	FuncEnterUpdateSetClause            func(p0 *UpdateSetClause) error
	FuncExitUpdateSetClause             func(p0 *UpdateSetClause) error
	FuncEnterUpdateStmt                 func(p0 *UpdateStmt) error
	FuncExitUpdateStmt                  func(p0 *UpdateStmt) error
	FuncEnterUpsert                     func(p0 *Upsert) error
	FuncExitUpsert                      func(p0 *Upsert) error
}

var _ AstWalker = &ImplementedWalker{}

func (b *ImplementedWalker) EnterAggregateFunc(p0 *AggregateFunc) error {
	if b.FuncEnterAggregateFunc == nil {
		return nil
	}

	return b.FuncEnterAggregateFunc(p0)
}

func (b *ImplementedWalker) ExitAggregateFunc(p0 *AggregateFunc) error {
	if b.FuncExitAggregateFunc == nil {
		return nil
	}

	return b.FuncExitAggregateFunc(p0)
}

func (b *ImplementedWalker) EnterCTE(p0 *CTE) error {
	if b.FuncEnterCTE == nil {
		return nil
	}

	return b.FuncEnterCTE(p0)
}

func (b *ImplementedWalker) ExitCTE(p0 *CTE) error {
	if b.FuncExitCTE == nil {
		return nil
	}

	return b.FuncExitCTE(p0)
}

func (b *ImplementedWalker) EnterCompoundOperator(p0 *CompoundOperator) error {
	if b.FuncEnterCompoundOperator == nil {
		return nil
	}

	return b.FuncEnterCompoundOperator(p0)
}

func (b *ImplementedWalker) ExitCompoundOperator(p0 *CompoundOperator) error {
	if b.FuncExitCompoundOperator == nil {
		return nil
	}

	return b.FuncExitCompoundOperator(p0)
}

func (b *ImplementedWalker) EnterConflictTarget(p0 *ConflictTarget) error {
	if b.FuncEnterConflictTarget == nil {
		return nil
	}

	return b.FuncEnterConflictTarget(p0)
}

func (b *ImplementedWalker) ExitConflictTarget(p0 *ConflictTarget) error {
	if b.FuncExitConflictTarget == nil {
		return nil
	}

	return b.FuncExitConflictTarget(p0)
}

func (b *ImplementedWalker) EnterDateTimeFunc(p0 *DateTimeFunction) error {
	if b.FuncEnterDateTimeFunc == nil {
		return nil
	}

	return b.FuncEnterDateTimeFunc(p0)
}

func (b *ImplementedWalker) ExitDateTimeFunc(p0 *DateTimeFunction) error {
	if b.FuncExitDateTimeFunc == nil {
		return nil
	}

	return b.FuncExitDateTimeFunc(p0)
}

func (b *ImplementedWalker) EnterDelete(p0 *Delete) error {
	if b.FuncEnterDelete == nil {
		return nil
	}

	return b.FuncEnterDelete(p0)
}

func (b *ImplementedWalker) ExitDelete(p0 *Delete) error {
	if b.FuncExitDelete == nil {
		return nil
	}

	return b.FuncExitDelete(p0)
}

func (b *ImplementedWalker) EnterDeleteStmt(p0 *DeleteStmt) error {
	if b.FuncEnterDeleteStmt == nil {
		return nil
	}

	return b.FuncEnterDeleteStmt(p0)
}

func (b *ImplementedWalker) ExitDeleteStmt(p0 *DeleteStmt) error {
	if b.FuncExitDeleteStmt == nil {
		return nil
	}

	return b.FuncExitDeleteStmt(p0)
}

func (b *ImplementedWalker) EnterExpressionArithmetic(p0 *ExpressionArithmetic) error {
	if b.FuncEnterExpressionArithmetic == nil {
		return nil
	}

	return b.FuncEnterExpressionArithmetic(p0)
}

func (b *ImplementedWalker) ExitExpressionArithmetic(p0 *ExpressionArithmetic) error {
	if b.FuncExitExpressionArithmetic == nil {
		return nil
	}

	return b.FuncExitExpressionArithmetic(p0)
}

func (b *ImplementedWalker) EnterExpressionBetween(p0 *ExpressionBetween) error {
	if b.FuncEnterExpressionBetween == nil {
		return nil
	}

	return b.FuncEnterExpressionBetween(p0)
}

func (b *ImplementedWalker) ExitExpressionBetween(p0 *ExpressionBetween) error {
	if b.FuncExitExpressionBetween == nil {
		return nil
	}

	return b.FuncExitExpressionBetween(p0)
}

func (b *ImplementedWalker) EnterExpressionBinaryComparison(p0 *ExpressionBinaryComparison) error {
	if b.FuncEnterExpressionBinaryComparison == nil {
		return nil
	}

	return b.FuncEnterExpressionBinaryComparison(p0)
}

func (b *ImplementedWalker) ExitExpressionBinaryComparison(p0 *ExpressionBinaryComparison) error {
	if b.FuncExitExpressionBinaryComparison == nil {
		return nil
	}

	return b.FuncExitExpressionBinaryComparison(p0)
}

func (b *ImplementedWalker) EnterExpressionBindParameter(p0 *ExpressionBindParameter) error {
	if b.FuncEnterExpressionBindParameter == nil {
		return nil
	}

	return b.FuncEnterExpressionBindParameter(p0)
}

func (b *ImplementedWalker) ExitExpressionBindParameter(p0 *ExpressionBindParameter) error {
	if b.FuncExitExpressionBindParameter == nil {
		return nil
	}

	return b.FuncExitExpressionBindParameter(p0)
}

func (b *ImplementedWalker) EnterExpressionCase(p0 *ExpressionCase) error {
	if b.FuncEnterExpressionCase == nil {
		return nil
	}

	return b.FuncEnterExpressionCase(p0)
}

func (b *ImplementedWalker) ExitExpressionCase(p0 *ExpressionCase) error {
	if b.FuncExitExpressionCase == nil {
		return nil
	}

	return b.FuncExitExpressionCase(p0)
}

func (b *ImplementedWalker) EnterExpressionCollate(p0 *ExpressionCollate) error {
	if b.FuncEnterExpressionCollate == nil {
		return nil
	}

	return b.FuncEnterExpressionCollate(p0)
}

func (b *ImplementedWalker) ExitExpressionCollate(p0 *ExpressionCollate) error {
	if b.FuncExitExpressionCollate == nil {
		return nil
	}

	return b.FuncExitExpressionCollate(p0)
}

func (b *ImplementedWalker) EnterExpressionColumn(p0 *ExpressionColumn) error {
	if b.FuncEnterExpressionColumn == nil {
		return nil
	}

	return b.FuncEnterExpressionColumn(p0)
}

func (b *ImplementedWalker) ExitExpressionColumn(p0 *ExpressionColumn) error {
	if b.FuncExitExpressionColumn == nil {
		return nil
	}

	return b.FuncExitExpressionColumn(p0)
}

func (b *ImplementedWalker) EnterExpressionDistinct(p0 *ExpressionDistinct) error {
	if b.FuncEnterExpressionDistinct == nil {
		return nil
	}

	return b.FuncEnterExpressionDistinct(p0)
}

func (b *ImplementedWalker) ExitExpressionDistinct(p0 *ExpressionDistinct) error {
	if b.FuncExitExpressionDistinct == nil {
		return nil
	}

	return b.FuncExitExpressionDistinct(p0)
}

func (b *ImplementedWalker) EnterExpressionFunction(p0 *ExpressionFunction) error {
	if b.FuncEnterExpressionFunction == nil {
		return nil
	}

	return b.FuncEnterExpressionFunction(p0)
}

func (b *ImplementedWalker) ExitExpressionFunction(p0 *ExpressionFunction) error {
	if b.FuncExitExpressionFunction == nil {
		return nil
	}

	return b.FuncExitExpressionFunction(p0)
}

func (b *ImplementedWalker) EnterExpressionIsNull(p0 *ExpressionIsNull) error {
	if b.FuncEnterExpressionIsNull == nil {
		return nil
	}

	return b.FuncEnterExpressionIsNull(p0)
}

func (b *ImplementedWalker) ExitExpressionIsNull(p0 *ExpressionIsNull) error {
	if b.FuncExitExpressionIsNull == nil {
		return nil
	}

	return b.FuncExitExpressionIsNull(p0)
}

func (b *ImplementedWalker) EnterExpressionList(p0 *ExpressionList) error {
	if b.FuncEnterExpressionList == nil {
		return nil
	}

	return b.FuncEnterExpressionList(p0)
}

func (b *ImplementedWalker) ExitExpressionList(p0 *ExpressionList) error {
	if b.FuncExitExpressionList == nil {
		return nil
	}

	return b.FuncExitExpressionList(p0)
}

func (b *ImplementedWalker) EnterExpressionLiteral(p0 *ExpressionLiteral) error {
	if b.FuncEnterExpressionLiteral == nil {
		return nil
	}

	return b.FuncEnterExpressionLiteral(p0)
}

func (b *ImplementedWalker) ExitExpressionLiteral(p0 *ExpressionLiteral) error {
	if b.FuncExitExpressionLiteral == nil {
		return nil
	}

	return b.FuncExitExpressionLiteral(p0)
}

func (b *ImplementedWalker) EnterExpressionSelect(p0 *ExpressionSelect) error {
	if b.FuncEnterExpressionSelect == nil {
		return nil
	}

	return b.FuncEnterExpressionSelect(p0)
}

func (b *ImplementedWalker) ExitExpressionSelect(p0 *ExpressionSelect) error {
	if b.FuncExitExpressionSelect == nil {
		return nil
	}

	return b.FuncExitExpressionSelect(p0)
}

func (b *ImplementedWalker) EnterExpressionStringCompare(p0 *ExpressionStringCompare) error {
	if b.FuncEnterExpressionStringCompare == nil {
		return nil
	}

	return b.FuncEnterExpressionStringCompare(p0)
}

func (b *ImplementedWalker) ExitExpressionStringCompare(p0 *ExpressionStringCompare) error {
	if b.FuncExitExpressionStringCompare == nil {
		return nil
	}

	return b.FuncExitExpressionStringCompare(p0)
}

func (b *ImplementedWalker) EnterExpressionUnary(p0 *ExpressionUnary) error {
	if b.FuncEnterExpressionUnary == nil {
		return nil
	}

	return b.FuncEnterExpressionUnary(p0)
}

func (b *ImplementedWalker) ExitExpressionUnary(p0 *ExpressionUnary) error {
	if b.FuncExitExpressionUnary == nil {
		return nil
	}

	return b.FuncExitExpressionUnary(p0)
}

func (b *ImplementedWalker) EnterFromClause(p0 *FromClause) error {
	if b.FuncEnterFromClause == nil {
		return nil
	}

	return b.FuncEnterFromClause(p0)
}

func (b *ImplementedWalker) ExitFromClause(p0 *FromClause) error {
	if b.FuncExitFromClause == nil {
		return nil
	}

	return b.FuncExitFromClause(p0)
}

func (b *ImplementedWalker) EnterGroupBy(p0 *GroupBy) error {
	if b.FuncEnterGroupBy == nil {
		return nil
	}

	return b.FuncEnterGroupBy(p0)
}

func (b *ImplementedWalker) ExitGroupBy(p0 *GroupBy) error {
	if b.FuncExitGroupBy == nil {
		return nil
	}

	return b.FuncExitGroupBy(p0)
}

func (b *ImplementedWalker) EnterInsert(p0 *Insert) error {
	if b.FuncEnterInsert == nil {
		return nil
	}

	return b.FuncEnterInsert(p0)
}

func (b *ImplementedWalker) ExitInsert(p0 *Insert) error {
	if b.FuncExitInsert == nil {
		return nil
	}

	return b.FuncExitInsert(p0)
}

func (b *ImplementedWalker) EnterInsertStmt(p0 *InsertStmt) error {
	if b.FuncEnterInsertStmt == nil {
		return nil
	}

	return b.FuncEnterInsertStmt(p0)
}

func (b *ImplementedWalker) ExitInsertStmt(p0 *InsertStmt) error {
	if b.FuncExitInsertStmt == nil {
		return nil
	}

	return b.FuncExitInsertStmt(p0)
}

func (b *ImplementedWalker) EnterJoinClause(p0 *JoinClause) error {
	if b.FuncEnterJoinClause == nil {
		return nil
	}

	return b.FuncEnterJoinClause(p0)
}

func (b *ImplementedWalker) ExitJoinClause(p0 *JoinClause) error {
	if b.FuncExitJoinClause == nil {
		return nil
	}

	return b.FuncExitJoinClause(p0)
}

func (b *ImplementedWalker) EnterJoinOperator(p0 *JoinOperator) error {
	if b.FuncEnterJoinOperator == nil {
		return nil
	}

	return b.FuncEnterJoinOperator(p0)
}

func (b *ImplementedWalker) ExitJoinOperator(p0 *JoinOperator) error {
	if b.FuncExitJoinOperator == nil {
		return nil
	}

	return b.FuncExitJoinOperator(p0)
}

func (b *ImplementedWalker) EnterJoinPredicate(p0 *JoinPredicate) error {
	if b.FuncEnterJoinPredicate == nil {
		return nil
	}

	return b.FuncEnterJoinPredicate(p0)
}

func (b *ImplementedWalker) ExitJoinPredicate(p0 *JoinPredicate) error {
	if b.FuncExitJoinPredicate == nil {
		return nil
	}

	return b.FuncExitJoinPredicate(p0)
}

func (b *ImplementedWalker) EnterLimit(p0 *Limit) error {
	if b.FuncEnterLimit == nil {
		return nil
	}

	return b.FuncEnterLimit(p0)
}

func (b *ImplementedWalker) ExitLimit(p0 *Limit) error {
	if b.FuncExitLimit == nil {
		return nil
	}

	return b.FuncExitLimit(p0)
}

func (b *ImplementedWalker) EnterOrderBy(p0 *OrderBy) error {
	if b.FuncEnterOrderBy == nil {
		return nil
	}

	return b.FuncEnterOrderBy(p0)
}

func (b *ImplementedWalker) ExitOrderBy(p0 *OrderBy) error {
	if b.FuncExitOrderBy == nil {
		return nil
	}

	return b.FuncExitOrderBy(p0)
}

func (b *ImplementedWalker) EnterOrderingTerm(p0 *OrderingTerm) error {
	if b.FuncEnterOrderingTerm == nil {
		return nil
	}

	return b.FuncEnterOrderingTerm(p0)
}

func (b *ImplementedWalker) ExitOrderingTerm(p0 *OrderingTerm) error {
	if b.FuncExitOrderingTerm == nil {
		return nil
	}

	return b.FuncExitOrderingTerm(p0)
}

func (b *ImplementedWalker) EnterQualifiedTableName(p0 *QualifiedTableName) error {
	if b.FuncEnterQualifiedTableName == nil {
		return nil
	}

	return b.FuncEnterQualifiedTableName(p0)
}

func (b *ImplementedWalker) ExitQualifiedTableName(p0 *QualifiedTableName) error {
	if b.FuncExitQualifiedTableName == nil {
		return nil
	}

	return b.FuncExitQualifiedTableName(p0)
}

func (b *ImplementedWalker) EnterResultColumnExpression(p0 *ResultColumnExpression) error {
	if b.FuncEnterResultColumnExpression == nil {
		return nil
	}

	return b.FuncEnterResultColumnExpression(p0)
}

func (b *ImplementedWalker) ExitResultColumnExpression(p0 *ResultColumnExpression) error {
	if b.FuncExitResultColumnExpression == nil {
		return nil
	}

	return b.FuncExitResultColumnExpression(p0)
}

func (b *ImplementedWalker) EnterResultColumnStar(p0 *ResultColumnStar) error {
	if b.FuncEnterResultColumnStar == nil {
		return nil
	}

	return b.FuncEnterResultColumnStar(p0)
}

func (b *ImplementedWalker) ExitResultColumnStar(p0 *ResultColumnStar) error {
	if b.FuncExitResultColumnStar == nil {
		return nil
	}

	return b.FuncExitResultColumnStar(p0)
}

func (b *ImplementedWalker) EnterResultColumnTable(p0 *ResultColumnTable) error {
	if b.FuncEnterResultColumnTable == nil {
		return nil
	}

	return b.FuncEnterResultColumnTable(p0)
}

func (b *ImplementedWalker) ExitResultColumnTable(p0 *ResultColumnTable) error {
	if b.FuncExitResultColumnTable == nil {
		return nil
	}

	return b.FuncExitResultColumnTable(p0)
}

func (b *ImplementedWalker) EnterReturningClause(p0 *ReturningClause) error {
	if b.FuncEnterReturningClause == nil {
		return nil
	}

	return b.FuncEnterReturningClause(p0)
}

func (b *ImplementedWalker) ExitReturningClause(p0 *ReturningClause) error {
	if b.FuncExitReturningClause == nil {
		return nil
	}

	return b.FuncExitReturningClause(p0)
}

func (b *ImplementedWalker) EnterReturningClauseColumn(p0 *ReturningClauseColumn) error {
	if b.FuncEnterReturningClauseColumn == nil {
		return nil
	}

	return b.FuncEnterReturningClauseColumn(p0)
}

func (b *ImplementedWalker) ExitReturningClauseColumn(p0 *ReturningClauseColumn) error {
	if b.FuncExitReturningClauseColumn == nil {
		return nil
	}

	return b.FuncExitReturningClauseColumn(p0)
}

func (b *ImplementedWalker) EnterScalarFunc(p0 *ScalarFunction) error {
	if b.FuncEnterScalarFunc == nil {
		return nil
	}

	return b.FuncEnterScalarFunc(p0)
}

func (b *ImplementedWalker) ExitScalarFunc(p0 *ScalarFunction) error {
	if b.FuncExitScalarFunc == nil {
		return nil
	}

	return b.FuncExitScalarFunc(p0)
}

func (b *ImplementedWalker) EnterSelect(p0 *Select) error {
	if b.FuncEnterSelect == nil {
		return nil
	}

	return b.FuncEnterSelect(p0)
}

func (b *ImplementedWalker) ExitSelect(p0 *Select) error {
	if b.FuncExitSelect == nil {
		return nil
	}

	return b.FuncExitSelect(p0)
}

func (b *ImplementedWalker) EnterSelectCore(p0 *SelectCore) error {
	if b.FuncEnterSelectCore == nil {
		return nil
	}

	return b.FuncEnterSelectCore(p0)
}

func (b *ImplementedWalker) ExitSelectCore(p0 *SelectCore) error {
	if b.FuncExitSelectCore == nil {
		return nil
	}

	return b.FuncExitSelectCore(p0)
}

func (b *ImplementedWalker) EnterSelectStmt(p0 *SelectStmt) error {
	if b.FuncEnterSelectStmt == nil {
		return nil
	}

	return b.FuncEnterSelectStmt(p0)
}

func (b *ImplementedWalker) ExitSelectStmt(p0 *SelectStmt) error {
	if b.FuncExitSelectStmt == nil {
		return nil
	}

	return b.FuncExitSelectStmt(p0)
}

func (b *ImplementedWalker) EnterTableOrSubqueryJoin(p0 *TableOrSubqueryJoin) error {
	if b.FuncEnterTableOrSubqueryJoin == nil {
		return nil
	}

	return b.FuncEnterTableOrSubqueryJoin(p0)
}

func (b *ImplementedWalker) ExitTableOrSubqueryJoin(p0 *TableOrSubqueryJoin) error {
	if b.FuncExitTableOrSubqueryJoin == nil {
		return nil
	}

	return b.FuncExitTableOrSubqueryJoin(p0)
}

func (b *ImplementedWalker) EnterTableOrSubqueryList(p0 *TableOrSubqueryList) error {
	if b.FuncEnterTableOrSubqueryList == nil {
		return nil
	}

	return b.FuncEnterTableOrSubqueryList(p0)
}

func (b *ImplementedWalker) ExitTableOrSubqueryList(p0 *TableOrSubqueryList) error {
	if b.FuncExitTableOrSubqueryList == nil {
		return nil
	}

	return b.FuncExitTableOrSubqueryList(p0)
}

func (b *ImplementedWalker) EnterTableOrSubquerySelect(p0 *TableOrSubquerySelect) error {
	if b.FuncEnterTableOrSubquerySelect == nil {
		return nil
	}

	return b.FuncEnterTableOrSubquerySelect(p0)
}

func (b *ImplementedWalker) ExitTableOrSubquerySelect(p0 *TableOrSubquerySelect) error {
	if b.FuncExitTableOrSubquerySelect == nil {
		return nil
	}

	return b.FuncExitTableOrSubquerySelect(p0)
}

func (b *ImplementedWalker) EnterTableOrSubqueryTable(p0 *TableOrSubqueryTable) error {
	if b.FuncEnterTableOrSubqueryTable == nil {
		return nil
	}

	return b.FuncEnterTableOrSubqueryTable(p0)
}

func (b *ImplementedWalker) ExitTableOrSubqueryTable(p0 *TableOrSubqueryTable) error {
	if b.FuncExitTableOrSubqueryTable == nil {
		return nil
	}

	return b.FuncExitTableOrSubqueryTable(p0)
}

func (b *ImplementedWalker) EnterUpdate(p0 *Update) error {
	if b.FuncEnterUpdate == nil {
		return nil
	}

	return b.FuncEnterUpdate(p0)
}

func (b *ImplementedWalker) ExitUpdate(p0 *Update) error {
	if b.FuncExitUpdate == nil {
		return nil
	}

	return b.FuncExitUpdate(p0)
}

func (b *ImplementedWalker) EnterUpdateSetClause(p0 *UpdateSetClause) error {
	if b.FuncEnterUpdateSetClause == nil {
		return nil
	}

	return b.FuncEnterUpdateSetClause(p0)
}

func (b *ImplementedWalker) ExitUpdateSetClause(p0 *UpdateSetClause) error {
	if b.FuncExitUpdateSetClause == nil {
		return nil
	}

	return b.FuncExitUpdateSetClause(p0)
}

func (b *ImplementedWalker) EnterUpdateStmt(p0 *UpdateStmt) error {
	if b.FuncEnterUpdateStmt == nil {
		return nil
	}

	return b.FuncEnterUpdateStmt(p0)
}

func (b *ImplementedWalker) ExitUpdateStmt(p0 *UpdateStmt) error {
	if b.FuncExitUpdateStmt == nil {
		return nil
	}

	return b.FuncExitUpdateStmt(p0)
}

func (b *ImplementedWalker) EnterUpsert(p0 *Upsert) error {
	if b.FuncEnterUpsert == nil {
		return nil
	}

	return b.FuncEnterUpsert(p0)
}

func (b *ImplementedWalker) ExitUpsert(p0 *Upsert) error {
	if b.FuncExitUpsert == nil {
		return nil
	}

	return b.FuncExitUpsert(p0)
}
