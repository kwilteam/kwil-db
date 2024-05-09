package clean

import (
	"errors"
	"strings"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/parse/metadata"
	"github.com/kwilteam/kwil-db/parse/sql/tree"
	parseTypes "github.com/kwilteam/kwil-db/parse/types"
	"github.com/kwilteam/kwil-db/parse/util"
)

// TODO: the statement cleaner should also check for table / column existence
func NewStatementCleaner(schema *types.Schema, errorListener parseTypes.NativeErrorListener) *StatementCleaner {
	return &StatementCleaner{
		AstListener: tree.NewBaseListener(),
		schema:      schema,
		errs:        errorListener,
	}
}

var _ tree.AstListener = &StatementCleaner{}

type StatementCleaner struct {
	tree.AstListener
	schema *types.Schema
	errs   parseTypes.NativeErrorListener
}

// err is a helper function to send errors to the error listener.
// It will ignore nil errors. If err1 is not nil but err2 is, it will
// not send an error. It is meant to be used to wrap errors, with the
// first error being the type of error, and the second error giving
// more information
func (s *StatementCleaner) err(err1, err2 error, getNode getNoder) {
	if err1 == nil {
		panic("internal api misuse: err1 cannot be nil")
	}

	if err2 == nil {
		return
	}

	s.errs.NodeErr(getNode.GetNode(), parseTypes.ParseErrorTypeSemantic, errors.Join(err1, err2))
}

type getNoder interface {
	GetNode() *parseTypes.Node
}

// EnterConflictTarget checks that the indexed column names are valid identifiers
func (s *StatementCleaner) EnterConflictTarget(node *tree.ConflictTarget) (err error) {
	node.IndexedColumns, err = cleanIdentifiers(node.IndexedColumns)
	s.err(ErrInvalidIdentifier, err, node)
	return nil
}

// EnterCTE checks that the table name and column names are valid identifiers
func (s *StatementCleaner) EnterCTE(node *tree.CTE) (err error) {
	node.Table, err = cleanIdentifier(node.Table)
	if err != nil {
		s.err(ErrInvalidIdentifier, err, node)
		return nil
	}
	node.Columns, err = cleanIdentifiers(node.Columns)
	s.err(ErrInvalidIdentifier, err, node)
	return nil
}

// EnterDelete does nothing
func (s *StatementCleaner) EnterDeleteStmt(node *tree.DeleteStmt) (err error) {
	return nil
}

// EnterDeleteStmt does nothing
func (s *StatementCleaner) EnterDeleteCore(node *tree.DeleteCore) (err error) {
	return nil
}

// EnterExpressionBindParameter checks that the bind parameter is a valid bind parameter
func (s *StatementCleaner) EnterExpressionBindParameter(node *tree.ExpressionBindParameter) (err error) {
	if !strings.HasPrefix(node.Parameter, "$") && !strings.HasPrefix(node.Parameter, "@") {
		s.err(ErrInvalidBindParameter, errors.New("bind parameter must start with $ or @"), node)
		return nil
	}

	node.Parameter = strings.ToLower(node.Parameter)
	return nil
}

// EnterExpressionColumn checks that the table and column names are valid identifiers
func (s *StatementCleaner) EnterExpressionColumn(node *tree.ExpressionColumn) (err error) {
	if node.Table != "" {
		node.Table, err = cleanIdentifier(node.Table)
		if err != nil {
			s.err(ErrInvalidIdentifier, err, node)
			return nil
		}
	}

	node.Column, err = cleanIdentifier(node.Column)
	s.err(ErrInvalidIdentifier, err, node)
	return nil
}

// EnterExpressionUnary checks that the operator is a valid operator
func (s *StatementCleaner) EnterExpressionUnary(node *tree.ExpressionUnary) (err error) {
	s.err(ErrInvalidUnaryOperator, node.Operator.Valid(), node)
	return nil
}

// EnterExpressionBinary checks that the operator is a valid operator
func (s *StatementCleaner) EnterExpressionBinaryComparison(node *tree.ExpressionBinaryComparison) (err error) {
	s.err(ErrInvalidBinaryOperator, node.Operator.Valid(), node)
	return nil
}

// EnterExpressionFunction lowers the function name and checks that it is a valid function
func (s *StatementCleaner) EnterExpressionFunction(node *tree.ExpressionFunction) (err error) {
	node.Function = strings.ToLower(node.Function)
	// this can either be a procedure or a function call.
	// it cannot be a foreign procedure, as those are parsed differently

	_, ok := metadata.Functions[node.Function]
	if !ok {
		// check if it's a procedure
		if _, ok := s.schema.FindProcedure(node.Function); ok {
			return nil
		}

		_, _, err := util.FindProcOrForeign(s.schema, node.Function)
		if err != nil {
			return err
		}
	}

	return nil
}

// EnterExpressionList does nothing
func (s *StatementCleaner) EnterExpressionList(node *tree.ExpressionList) (err error) {
	return nil
}

// EnterExpressionCollate checks that the collation is a valid collation
func (s *StatementCleaner) EnterExpressionCollate(node *tree.ExpressionCollate) (err error) {
	if node.Collation.Empty() {
		s.err(ErrInvalidCollation, errors.New("collation cannot be empty"), node)
		return nil
	}

	err = node.Collation.Valid()
	if err != nil {
		s.err(ErrInvalidCollation, err, node)
		return nil
	}

	return nil
}

// EnterExpressionStringCompare checks that the operator is a valid operator
func (s *StatementCleaner) EnterExpressionStringCompare(node *tree.ExpressionStringCompare) (err error) {
	s.err(ErrInvalidStringComparisonOperator, node.Operator.Valid(), node)
	return nil
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
		s.err(ErrInvalidIdentifier, errors.New("cannot negate non-EXISTS select expression"), node)
		return nil
	}

	return nil
}

// EnterExpressionCase does nothing
func (s *StatementCleaner) EnterExpressionCase(node *tree.ExpressionCase) (err error) {
	return nil
}

// EnterExpressionArithmetic checks the validity of the operator
func (s *StatementCleaner) EnterExpressionArithmetic(node *tree.ExpressionArithmetic) (err error) {
	s.err(ErrInvalidArithmeticOperator, node.Operator.Valid(), node)
	return nil
}

// EnterGroupBy does nothing
func (s *StatementCleaner) EnterGroupBy(node *tree.GroupBy) (err error) {
	return nil
}

// EnterInsert does nothing
func (s *StatementCleaner) EnterInsertStmt(node *tree.InsertStmt) (err error) {
	return nil
}

// EnterInsertStmt cleans the insert type, table, table alias, and columns
func (s *StatementCleaner) EnterInsertCore(node *tree.InsertCore) (err error) {
	err = node.InsertType.Valid()
	if err != nil {
		s.err(ErrInvalidInsertType, err, node)
		return nil
	}

	node.Table, err = cleanIdentifier(node.Table)
	if err != nil {
		s.err(ErrInvalidIdentifier, err, node)
		return nil
	}

	_, found := s.schema.FindTable(node.Table)
	if !found {
		s.err(ErrTableNotFound, ErrTableNotFound, node)
		return nil
	}

	if node.TableAlias != "" {
		node.TableAlias, err = cleanIdentifier(node.TableAlias)
		if err != nil {
			s.err(ErrInvalidIdentifier, err, node)
			return nil
		}
	}

	node.Columns, err = cleanIdentifiers(node.Columns)
	s.err(ErrInvalidIdentifier, err, node)
	return nil
}

func (s *StatementCleaner) EnterRelationFunction(node *tree.RelationFunction) (err error) {
	// check the alias is a valid identifier
	if node.Alias != "" {
		node.Alias, err = cleanIdentifier(node.Alias)
		if err != nil {
			s.err(ErrInvalidIdentifier, err, node)
			return nil
		}
	}
	return nil
}

// EnterJoinConstraint does nothing
func (s *StatementCleaner) EnterJoinPredicate(node *tree.JoinPredicate) (err error) {
	return nil
}

// EnterJoinOperator validates the join operator
func (s *StatementCleaner) EnterJoinOperator(node *tree.JoinOperator) (err error) {
	s.err(ErrInvalidJoinOperator, node.Valid(), node)
	return nil
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
		s.err(ErrInvalidOrderType, err, node)
		return nil
	}

	if err = node.NullOrdering.Valid(); err != nil {
		s.err(ErrInvalidNullOrderType, err, node)
		return nil
	}

	return nil
}

// EnterQualifiedTableName checks the table name and alias and indexed by column
func (s *StatementCleaner) EnterQualifiedTableName(node *tree.QualifiedTableName) (err error) {
	node.TableName, err = cleanIdentifier(node.TableName)
	if err != nil {
		s.err(ErrInvalidIdentifier, err, node)
		return nil
	}

	// we do not check for table existence here since it can reference a cte

	if node.TableAlias != "" {
		node.TableAlias, err = cleanIdentifier(node.TableAlias)
		if err != nil {
			s.err(ErrInvalidIdentifier, err, node)
			return nil
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
			s.err(ErrInvalidIdentifier, err, node)
			return nil
		}
	}

	return nil
}

// EnterResultColumnTable checks the table name
func (s *StatementCleaner) EnterResultColumnTable(node *tree.ResultColumnTable) (err error) {
	node.TableName, err = cleanIdentifier(node.TableName)
	if err != nil {
		s.err(ErrInvalidIdentifier, err, node)
		return nil
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
		s.err(ErrInvalidReturningClause, errors.New("all and expression cannot be set at the same time"), node)
		return nil
	}

	if node.Alias != "" && node.Expression == nil {
		s.err(ErrInvalidReturningClause, errors.New("alias cannot be set without an expression"), node)
		return nil
	}

	return nil
}

// EnterSelect does nothing
func (s *StatementCleaner) EnterSelectStmt(node *tree.SelectStmt) (err error) {
	return nil
}

// EnterSelectCore validates the select type
func (s *StatementCleaner) EnterSimpleSelect(node *tree.SimpleSelect) (err error) {
	s.err(ErrInvalidSelectType, node.SelectType.Valid(), node)
	return nil
}

// EnterSelectStmt checks that, for each SelectCore besides the last, a compound operator is provided
func (s *StatementCleaner) EnterSelectCore(node *tree.SelectCore) (err error) {
	for _, core := range node.SimpleSelects[1:] {
		if core.Compound == nil {
			s.err(ErrInvalidCompoundOperator, errors.New("compound operator must be provided for all SelectCores except the first"), node)
			return nil
		}
	}

	return nil
}

// EnterCompoundOperator validates the compound operator
func (s *StatementCleaner) EnterCompoundOperator(node *tree.CompoundOperator) (err error) {
	s.err(ErrInvalidCompoundOperator, node.Operator.Valid(), node)
	return nil
}

// EnterRelationTable checks the table name and alias
func (s *StatementCleaner) EnterRelationTable(node *tree.RelationTable) (err error) {
	node.Name, err = cleanIdentifier(node.Name)
	if err != nil {
		s.err(ErrInvalidIdentifier, err, node)
		return nil
	}

	// we do not check for table existence here since it can reference a cte

	if node.Alias != "" {
		node.Alias, err = cleanIdentifier(node.Alias)
		if err != nil {
			s.err(ErrInvalidIdentifier, err, node)
			return nil
		}
	}

	return nil
}

// EnterRelationSubquery checks the alias
func (s *StatementCleaner) EnterRelationSubquery(node *tree.RelationSubquery) (err error) {
	if node.Alias != "" {
		node.Alias, err = cleanIdentifier(node.Alias)
		if err != nil {
			s.err(ErrInvalidIdentifier, err, node)
			return nil
		}
	}

	return nil
}

// EnterRelationJoin does nothing
func (s *StatementCleaner) EnterRelationJoin(node *tree.RelationJoin) (err error) {
	return nil
}

// EnterUpdateSetClause checks the column names
func (s *StatementCleaner) EnterUpdateSetClause(node *tree.UpdateSetClause) (err error) {
	for i, column := range node.Columns {
		node.Columns[i], err = cleanIdentifier(column)
		if err != nil {
			s.err(ErrInvalidIdentifier, err, node)
			return nil
		}
	}

	return nil
}

// EnterUpdate does nothing
func (s *StatementCleaner) EnterUpdateStmt(node *tree.UpdateStmt) (err error) {
	return nil
}

// EnterUpsert validates the upsert type
func (s *StatementCleaner) EnterUpsert(node *tree.Upsert) (err error) {
	s.err(ErrInvalidUpsertType, node.Type.Valid(), node)
	return nil
}
