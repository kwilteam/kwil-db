package aggregate

import (
	"fmt"

	"github.com/kwilteam/kwil-db/internal/engine/sqlanalyzer/utils"
	"github.com/kwilteam/kwil-db/parse/sql/tree"
)

func (g *groupByAnalyzer) EnterSelectCore(s *tree.SelectCore) error {
	g.newContext()
	if len(s.Columns) == 0 {
		g.context.containsSelectAll = true
	}
	return nil
}

func (g *groupByAnalyzer) ExitSelectCore(s *tree.SelectCore) error {
	if s.GroupBy != nil {
		g.context.isAggregateStatement = true
	}

	resultsetColumns := []*tree.ExpressionColumn{}
	for _, returningClause := range g.context.returningClauses {
		if returningClause.containsAggregateFunc {
			g.context.isAggregateStatement = true
		}
		if returningClause.containsSelectAll {
			g.context.containsSelectAll = true
		}

		resultsetColumns = append(resultsetColumns, returningClause.bareColumns...)
	}

	if !g.context.isAggregateStatement {
		return nil
	}
	if g.context.containsSelectAll {
		return ErrAggregateQueryContainsSelectAll
	}

	for _, column := range resultsetColumns {
		if _, ok := g.context.groupedColumns[*column]; !ok {
			return fmt.Errorf("%w: column: %s", ErrResultSetContainsBareColumn, column.Column)
		}
	}

	havingColumns := utils.SearchResultColumns(g.context.havingClause)
	for _, column := range havingColumns {
		if _, ok := g.context.groupedColumns[*column]; !ok {
			return fmt.Errorf("%w: column: %s", ErrHavingClauseContainsUngroupedColumn, column.Column)
		}
	}

	g.oldContext()
	return nil
}

func (g *groupByAnalyzer) ExitGroupBy(gb *tree.GroupBy) error {
	for _, expr := range gb.Expressions {
		exprColumn, ok := expr.(*tree.ExpressionColumn)
		if !ok {
			return fmt.Errorf("%w: received type: %s", ErrGroupByContainsInvalidExpr, expr)
		}

		g.context.groupedColumns[*exprColumn] = struct{}{}
	}

	g.context.havingClause = gb.Having

	return nil
}
