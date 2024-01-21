package clean

import (
	"errors"
	"strings"

	"github.com/kwilteam/kwil-db/parse/sql/tree"
)

// TODO: the statement cleaner should also check for table / column existence
func NewStatementCleaner() *StatementCleaner {
	return &StatementCleaner{
		AstListener: tree.NewBaseListener(),
	}
}

var _ tree.AstListener = &StatementCleaner{}

type StatementCleaner struct {
	tree.AstListener
}

// EnterAggregateFunc checks that the function name is a valid identifier
func (s *StatementCleaner) EnterAggregateFunc(node *tree.AggregateFunc) (err error) {
	node.FunctionName, err = cleanIdentifier(node.FunctionName)
	return wrapErr(ErrInvalidIdentifier, err)
}

// EnterConflictTarget checks that the indexed column names are valid identifiers
func (s *StatementCleaner) EnterConflictTarget(node *tree.ConflictTarget) (err error) {
	node.IndexedColumns, err = cleanIdentifiers(node.IndexedColumns)
	return wrapErr(ErrInvalidIdentifier, err)
}

// EnterCTE checks that the table name and column names are valid identifiers
func (s *StatementCleaner) EnterCTE(node *tree.CTE) (err error) {
	node.Table, err = cleanIdentifier(node.Table)
	if err != nil {
		return wrapErr(ErrInvalidIdentifier, err)
	}
	node.Columns, err = cleanIdentifiers(node.Columns)
	return wrapErr(ErrInvalidIdentifier, err)
}

// EnterDelete does nothing
func (s *StatementCleaner) EnterDelete(node *tree.Delete) (err error) {
	return nil
}

// EnterDeleteStmt does nothing
func (s *StatementCleaner) EnterDeleteStmt(node *tree.DeleteStmt) (err error) {
	return nil
}

// EnterExpressionLiteral checks that the literal is a valid literal
func (s *StatementCleaner) EnterExpressionLiteral(node *tree.ExpressionLiteral) (err error) {
	return wrapErr(ErrInvalidLiteral, checkLiteral(node.Value))
}

// EnterExpressionBindParameter checks that the bind parameter is a valid bind parameter
func (s *StatementCleaner) EnterExpressionBindParameter(node *tree.ExpressionBindParameter) (err error) {
	if !strings.HasPrefix(node.Parameter, "$") && !strings.HasPrefix(node.Parameter, "@") {
		return wrapErr(ErrInvalidBindParameter, errors.New("bind parameter must start with $ or @"))
	}

	node.Parameter = strings.ToLower(node.Parameter)
	return nil
}

// EnterExpressionColumn checks that the table and column names are valid identifiers
func (s *StatementCleaner) EnterExpressionColumn(node *tree.ExpressionColumn) (err error) {
	if node.Table != "" {
		node.Table, err = cleanIdentifier(node.Table)
		if err != nil {
			return wrapErr(ErrInvalidIdentifier, err)
		}
	}

	node.Column, err = cleanIdentifier(node.Column)
	return wrapErr(ErrInvalidIdentifier, err)
}

// EnterExpressionUnary checks that the operator is a valid operator
func (s *StatementCleaner) EnterExpressionUnary(node *tree.ExpressionUnary) (err error) {
	return wrapErr(ErrInvalidUnaryOperator, node.Operator.Valid())
}

// EnterExpressionBinary checks that the operator is a valid operator
func (s *StatementCleaner) EnterExpressionBinaryComparison(node *tree.ExpressionBinaryComparison) (err error) {
	return wrapErr(ErrInvalidBinaryOperator, node.Operator.Valid())
}

// EnterExpressionFunction does nothing, since the function implementation is visited separately
func (s *StatementCleaner) EnterExpressionFunction(node *tree.ExpressionFunction) (err error) {
	return nil
}

// EnterExpressionList does nothing
func (s *StatementCleaner) EnterExpressionList(node *tree.ExpressionList) (err error) {
	return nil
}

// EnterExpressionCollate checks that the collation is a valid collation
func (s *StatementCleaner) EnterExpressionCollate(node *tree.ExpressionCollate) (err error) {
	if node.Collation.Empty() {
		return wrapErr(ErrInvalidCollation, errors.New("collation cannot be empty"))
	}

	err = node.Collation.Valid()
	if err != nil {
		return wrapErr(ErrInvalidCollation, err)
	}

	return nil
}

// EnterExpressionStringCompare checks that the operator is a valid operator
func (s *StatementCleaner) EnterExpressionStringCompare(node *tree.ExpressionStringCompare) (err error) {
	return wrapErr(ErrInvalidStringComparisonOperator, node.Operator.Valid())
}

// EnterExpressionIs does nothing
func (s *StatementCleaner) EnterExpressionIs(node *tree.ExpressionIs) (err error) {
	return nil
}

// EnterExpressionBetween does nothing
func (s *StatementCleaner) EnterExpressionBetween(node *tree.ExpressionBetween) (err error) {
	return nil
}

// EnterExpressionExists checks that you can only negate EXISTS
func (s *StatementCleaner) EnterExpressionSelect(node *tree.ExpressionSelect) (err error) {
	if node.IsNot && !node.IsExists {
		return wrapErr(ErrInvalidIdentifier, errors.New("cannot negate non-EXISTS select expression"))
	}

	return nil
}

// EnterExpressionCase does nothing
func (s *StatementCleaner) EnterExpressionCase(node *tree.ExpressionCase) (err error) {
	return nil
}

// EnterExpressionArithmetic checks the validity of the operator
func (s *StatementCleaner) EnterExpressionArithmetic(node *tree.ExpressionArithmetic) (err error) {
	return wrapErr(ErrInvalidArithmeticOperator, node.Operator.Valid())
}

// EnterScalarFunc checks that the function name is a valid identifier and is a scalar function
func (s *StatementCleaner) EnterScalarFunc(node *tree.ScalarFunction) (err error) {
	node.FunctionName, err = cleanIdentifier(node.FunctionName)
	return wrapErr(ErrInvalidIdentifier, err)
}

// EnterGroupBy does nothing
func (s *StatementCleaner) EnterGroupBy(node *tree.GroupBy) (err error) {
	return nil
}

// EnterInsert does nothing
func (s *StatementCleaner) EnterInsert(node *tree.Insert) (err error) {
	return nil
}

// EnterInsertStmt cleans the insert type, table, table alias, and columns
func (s *StatementCleaner) EnterInsertStmt(node *tree.InsertStmt) (err error) {
	err = node.InsertType.Valid()
	if err != nil {
		return wrapErr(ErrInvalidInsertType, err)
	}

	node.Table, err = cleanIdentifier(node.Table)
	if err != nil {
		return wrapErr(ErrInvalidIdentifier, err)
	}

	if node.TableAlias != "" {
		node.TableAlias, err = cleanIdentifier(node.TableAlias)
		if err != nil {
			return wrapErr(ErrInvalidIdentifier, err)
		}
	}

	node.Columns, err = cleanIdentifiers(node.Columns)
	return wrapErr(ErrInvalidIdentifier, err)
}

// EnterJoinClause does nothing
func (s *StatementCleaner) EnterJoinClause(node *tree.JoinClause) (err error) {
	return nil
}

// EnterJoinConstraint does nothing
func (s *StatementCleaner) EnterJoinPredicate(node *tree.JoinPredicate) (err error) {
	return nil
}

// EnterJoinOperator validates the join operator
func (s *StatementCleaner) EnterJoinOperator(node *tree.JoinOperator) (err error) {
	return wrapErr(ErrInvalidJoinOperator, node.Valid())
}

// EnterLimit does nothing
func (s *StatementCleaner) EnterLimit(node *tree.Limit) (err error) {
	return nil
}

// EnterOrderBy does nothing
func (s *StatementCleaner) EnterOrderBy(node *tree.OrderBy) (err error) {
	return nil
}

// EnterOrderingTerm validates the order type and null order type
func (s *StatementCleaner) EnterOrderingTerm(node *tree.OrderingTerm) (err error) {
	// ordertype and nullorderingtype are both valid as empty, so we don't need to check for that
	if err = node.OrderType.Valid(); err != nil {
		return wrapErr(ErrInvalidOrderType, err)
	}

	if err = node.NullOrdering.Valid(); err != nil {
		return wrapErr(ErrInvalidNullOrderType, err)
	}

	return nil
}

// EnterQualifiedTableName checks the table name and alias and indexed by column
func (s *StatementCleaner) EnterQualifiedTableName(node *tree.QualifiedTableName) (err error) {
	node.TableName, err = cleanIdentifier(node.TableName)
	if err != nil {
		return wrapErr(ErrInvalidIdentifier, err)
	}

	if node.TableAlias != "" {
		node.TableAlias, err = cleanIdentifier(node.TableAlias)
		if err != nil {
			return wrapErr(ErrInvalidIdentifier, err)
		}
	}

	return nil
}

// EnterResultColumnStar does nothing
func (s *StatementCleaner) EnterResultColumnStar(node *tree.ResultColumnStar) (err error) {
	return nil
}

// EnterResultColumnExpression checks the alias if it exists
func (s *StatementCleaner) EnterResultColumnExpression(node *tree.ResultColumnExpression) (err error) {
	if node.Alias != "" {
		node.Alias, err = cleanIdentifier(node.Alias)
		if err != nil {
			return wrapErr(ErrInvalidIdentifier, err)
		}
	}

	return nil
}

// EnterResultColumnTable checks the table name
func (s *StatementCleaner) EnterResultColumnTable(node *tree.ResultColumnTable) (err error) {
	node.TableName, err = cleanIdentifier(node.TableName)
	if err != nil {
		return wrapErr(ErrInvalidIdentifier, err)
	}

	return nil
}

// EnterReturningClause does nothing
func (s *StatementCleaner) EnterReturningClause(node *tree.ReturningClause) (err error) {
	return nil
}

// EnterReturningClauseColumn checks that either all is selected, or that an expression is used.  An alias can
// only be used if an expression is used.
func (s *StatementCleaner) EnterReturningClauseColumn(node *tree.ReturningClauseColumn) (err error) {
	if node.All && node.Expression != nil {
		return wrapErr(ErrInvalidReturningClause, errors.New("all and expression cannot be set at the same time"))
	}

	if node.Alias != "" && node.Expression == nil {
		return wrapErr(ErrInvalidReturningClause, errors.New("alias cannot be set without an expression"))
	}

	return nil
}

// EnterSelect does nothing
func (s *StatementCleaner) EnterSelect(node *tree.Select) (err error) {
	return nil
}

// EnterSelectCore validates the select type
func (s *StatementCleaner) EnterSelectCore(node *tree.SelectCore) (err error) {
	return wrapErr(ErrInvalidSelectType, node.SelectType.Valid())
}

// EnterSelectStmt checks that, for each SelectCore besides the last, a compound operator is provided
func (s *StatementCleaner) EnterSelectStmt(node *tree.SelectStmt) (err error) {
	for _, core := range node.SelectCores[:len(node.SelectCores)-1] {
		if core.Compound == nil {
			return wrapErr(ErrInvalidCompoundOperator, errors.New("compound operator must be provided for all SelectCores except the last"))
		}
	}

	return nil
}

// EnterFromClause does nothing
func (s *StatementCleaner) EnterFromClause(node *tree.FromClause) (err error) {
	return nil
}

// EnterCompoundOperator validates the compound operator
func (s *StatementCleaner) EnterCompoundOperator(node *tree.CompoundOperator) (err error) {
	return wrapErr(ErrInvalidCompoundOperator, node.Operator.Valid())
}

// EnterTableOrSubquery checks the table name and alias
func (s *StatementCleaner) EnterTableOrSubqueryTable(node *tree.TableOrSubqueryTable) (err error) {
	node.Name, err = cleanIdentifier(node.Name)
	if err != nil {
		return wrapErr(ErrInvalidIdentifier, err)
	}

	if node.Alias != "" {
		node.Alias, err = cleanIdentifier(node.Alias)
		if err != nil {
			return wrapErr(ErrInvalidIdentifier, err)
		}
	}

	return nil
}

// EnterTableOrSubquerySelect checks the alias
func (s *StatementCleaner) EnterTableOrSubquerySelect(node *tree.TableOrSubquerySelect) (err error) {
	if node.Alias != "" {
		node.Alias, err = cleanIdentifier(node.Alias)
		if err != nil {
			return wrapErr(ErrInvalidIdentifier, err)
		}
	}

	return nil
}

// EnterTableOrSubqueryList does nothing
func (s *StatementCleaner) EnterTableOrSubqueryList(node *tree.TableOrSubqueryList) (err error) {
	return nil
}

// EnterTableOrSubqueryJoin does nothing
func (s *StatementCleaner) EnterTableOrSubqueryJoin(node *tree.TableOrSubqueryJoin) (err error) {
	return nil
}

// EnterUpdateSetClause checks the column names
func (s *StatementCleaner) EnterUpdateSetClause(node *tree.UpdateSetClause) (err error) {
	for i, column := range node.Columns {
		node.Columns[i], err = cleanIdentifier(column)
		if err != nil {
			return wrapErr(ErrInvalidIdentifier, err)
		}
	}

	return nil
}

// EnterUpdate does nothing
func (s *StatementCleaner) EnterUpdate(node *tree.Update) (err error) {
	return nil
}

// EnterUpsert validates the upsert type
func (s *StatementCleaner) EnterUpsert(node *tree.Upsert) (err error) {
	return wrapErr(ErrInvalidUpsertType, node.Type.Valid())
}
