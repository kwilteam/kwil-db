package order

import (
	"fmt"
	"sort"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/internal/engine"
	"github.com/kwilteam/kwil-db/internal/engine/sqlanalyzer/typing"
	"github.com/kwilteam/kwil-db/internal/engine/sqlanalyzer/utils"
	"github.com/kwilteam/kwil-db/internal/parse/sql/tree"
)

func NewOrderWalker(tables []*types.Table) tree.AstListener {
	// copy tables, since we will be modifying the tables slice to register CTEs
	tbls := make([]*types.Table, len(tables))
	copy(tbls, tables)

	return &orderingWalker{
		tables: tbls,
	}
}

// orderingWalker is the highest level walker to order a statement
type orderingWalker struct {
	tree.BaseListener

	tables []*types.Table // all tables in the schema
}

var _ tree.AstListener = &orderingWalker{}

// we need to register common table expressions as tables, so that we can order them.
func (o *orderingWalker) EnterCTE(node *tree.CTE) error {
	if len(node.Select.SimpleSelects) == 0 {
		return nil
	}

	res, err := typing.AnalyzeTypes(node.Select, o.tables, &typing.AnalyzeOptions{
		ArbitraryBinds: true,
	})
	if err != nil {
		return err
	}

	tbl, err := tableFromRelation(res, node.Table)
	if err != nil {
		return err
	}

	// make sure this does not have a conflicting name with another table
	_, err = findTable(o.tables, tbl.Name)
	if err == nil {
		return fmt.Errorf(`table or subquery with name or alias "%s" already exists`, tbl.Name)
	}

	o.tables = append(o.tables, tbl)

	return nil
}

// tableFromRelation will return a table from a relation.
// It will make all columns primary keys.
func tableFromRelation(rel *engine.Relation, name string) (*types.Table, error) {
	tbl := &types.Table{
		Name: name,
	}

	cols := make([]string, 0)
	err := rel.Loop(func(s string, a *engine.Attribute) error {
		tbl.Columns = append(tbl.Columns, &types.Column{
			Name: s,
			Type: a.Type,
		})

		cols = append(cols, s)
		return nil
	})
	if err != nil {
		return nil, err
	}

	tbl.Indexes = []*types.Index{
		{
			Name:    "_auto_primary_",
			Columns: cols,
			Type:    types.PRIMARY,
		},
	}

	return tbl, nil
}

// Register RelationSubquerys as tables, so that we can order them.
func (o *orderingWalker) EnterRelationSubquery(node *tree.RelationSubquery) error {
	if node.Select == nil {
		return fmt.Errorf("subquery select is nil")
	}
	if len(node.Select.SimpleSelects) == 0 {
		return fmt.Errorf("subquery select has no select cores")
	}

	name := node.Alias // can be ""

	res, err := typing.AnalyzeTypes(node.Select, o.tables, &typing.AnalyzeOptions{
		ArbitraryBinds: true,
	})
	if err != nil {
		return err
	}

	tbl, err := tableFromRelation(res, name)
	if err != nil {
		return err
	}

	// make sure this does not have a conflicting name with another table
	_, err = findTable(o.tables, tbl.Name)
	if err == nil {
		return fmt.Errorf(`table or subquery with name or alias "%s" already exists`, tbl.Name)
	}

	o.tables = append(o.tables, tbl)

	return nil
}

// put this on exit so we can search the whole statement for used tables
func (o *orderingWalker) ExitSelectCore(node *tree.SelectCore) error {
	var terms []*tree.OrderingTerm
	var err error
	switch len(node.SimpleSelects) {
	case 0:
		return fmt.Errorf("no select cores in select statement")
	case 1:
		terms, err = orderSimpleStatement(node.SimpleSelects[0], o.tables)
	default:
		terms, err = orderCompoundStatement(node.SimpleSelects, o.tables)
	}
	if err != nil {
		return fmt.Errorf("error ordering statement: %w", err)
	}

	if len(terms) == 0 {
		return nil
	}

	if node.OrderBy == nil {
		node.OrderBy = &tree.OrderBy{
			OrderingTerms: terms,
		}
	} else {
		node.OrderBy.OrderingTerms = append(node.OrderBy.OrderingTerms, terms...)
	}

	return nil
}

var ErrDistinctWithGroupBy = fmt.Errorf("select distinct with group by not supported")

// orderSimpleStatement will return the ordering required for a simple statement.
func orderSimpleStatement(stmt *tree.SimpleSelect, tables []*types.Table) ([]*tree.OrderingTerm, error) {
	// it is possible to not have any tables in a select
	// if so, no ordering is required
	if stmt.From == nil {
		return nil, nil
	}

	// if there is a group by clause, then we order by each term in the group by clause
	// if there is no group by clause, then we order by primary keys.
	// if there is no group by and an aggregate function is used, all other columns returned must
	// be aggregates, or else we throw an error. We do not need to order in this case, since an
	// aggregate with no group by will always return a simple row.

	if stmt.GroupBy != nil && len(stmt.GroupBy.Expressions) > 0 {
		// if it has a distinct, error as this is not allowed. We do not know how to order this.
		if stmt.SelectType == tree.SelectTypeDistinct {
			return nil, ErrDistinctWithGroupBy
		}

		// it has a group by, order by each of them
		columns := make([]*tree.OrderingTerm, len(stmt.GroupBy.Expressions))
		for i, expr := range stmt.GroupBy.Expressions {
			cols := utils.SearchResultColumns(expr)
			switch len(cols) {
			case 0:
				return nil, nil
			case 1:
				columns[i] = &tree.OrderingTerm{
					Expression: &tree.ExpressionColumn{
						Table:  cols[0].Table,
						Column: cols[0].Column,
					},
				}
			default:
				return nil, fmt.Errorf("multiple columns in a simple grouped term in a group by expression not supported")
			}
		}

		return columns, nil
	}

	// if we reach here, there is no group by clause.

	// we will now get a list of all tables that are renamed to the used aliases
	// this allows us to search for them by their alias, and not their real name.
	usedTables, err := utils.GetUsedTables(stmt.From)
	if err != nil {
		return nil, fmt.Errorf("error getting used tables: %w", err)
	}

	usedTblsFull := make([]*types.Table, len(usedTables))
	for i, tbl := range usedTables {
		newTable, err := findTable(tables, tbl.Name)
		if err != nil {
			return nil, fmt.Errorf("error finding table: %w", err)
		}

		copied := newTable.Copy() // copy since we will be modifying the table

		// set the alias to the table name if it exists
		if tbl.Alias != "" {
			copied.Name = tbl.Alias
		}

		usedTblsFull[i] = copied
	}

	if stmt.SelectType == tree.SelectTypeDistinct {
		return getReturnedColumnOrderingTerms(stmt.Columns, usedTblsFull)
	}

	sort.Slice(usedTblsFull, func(i, j int) bool {
		return usedTblsFull[i].Name < usedTblsFull[j].Name
	})

	// we first must check if there are any aggregate functions in the result columns.
	// if so, then all other columns must be aggregates, or else we throw an error.
	numberOfAggregates := 0
	for _, ret := range stmt.Columns {
		containsAggregate, err := containsAggregateFunc(ret)
		if err != nil {
			return nil, fmt.Errorf("error checking for aggregate function: %w", err)
		}

		if containsAggregate {
			numberOfAggregates++
		}
	}

	if numberOfAggregates > 0 {
		if numberOfAggregates != len(stmt.Columns) {
			return nil, fmt.Errorf("all columns must be aggregates if an aggregate function is used without a group by")
		}
		return nil, nil // order nothing in this case
	}

	orderingTerms := make([]*tree.OrderingTerm, 0)
	for _, tbl := range usedTblsFull {
		primaries, err := tbl.GetPrimaryKey()
		if err != nil {
			return nil, fmt.Errorf("error getting primary key: %w", err)
		}

		if len(primaries) == 0 {
			continue // can't happen I think
		}

		sort.Strings(primaries)

		for _, primary := range primaries {
			orderingTerms = append(orderingTerms, &tree.OrderingTerm{
				Expression: &tree.ExpressionColumn{
					Table:  tbl.Name,
					Column: primary,
				},
			})
		}
	}

	return orderingTerms, nil
}

// there can be two types of ordering: simple and compound statements.
// a simple statement is just a simple select statement, while a compound statement is a select statement with a union, intersect, etc
// each of the below functions will return the ordering that is required for the statement.

// containsAggregateFunc returns true if the result column contains an aggregate function.
func containsAggregateFunc(ret tree.ResultColumn) (bool, error) {
	containsAggregateFunc := false
	depth := 0 // depth tracks if we are in a subquery or not

	err := ret.Walk(&tree.ImplementedListener{
		FuncEnterExpressionFunction: func(p0 *tree.ExpressionFunction) error {
			if depth == 0 {
				def, ok := engine.Functions[p0.Function]
				if !ok {
					return fmt.Errorf("function %s not found", p0.Function)
				}

				containsAggregateFunc = def.IsAggregate
			}
			return nil
		},
		FuncEnterSelectCore: func(p0 *tree.SelectCore) error {
			depth++
			return nil
		},
		FuncExitSelectCore: func(p0 *tree.SelectCore) error {
			depth--
			return nil
		},
	})

	return containsAggregateFunc, err
}

var ErrGroupByWithCompoundStatement = fmt.Errorf("statements cannot have a group by clause with a compound SELECT statement")
var ErrCompoundStatementDifferentNumberOfColumns = fmt.Errorf("select cores have different number of result columns")

// orderCompoundStatement will return the ordering required for a compound statement.
// it will order each result column, in the order they are returned.
// if there is a group by clause in any of the select cores, we will return an error.
// using a group by with a compound statement is not yet supported because idk how
// to make it deterministic with postgres's ordering, and it is not a common use case.
func orderCompoundStatement(stmt []*tree.SimpleSelect, tables []*types.Table) ([]*tree.OrderingTerm, error) {
	if len(stmt) == 0 {
		return nil, fmt.Errorf("no select cores in compound statement")
	}

	expectedNumberOfColumns := len(stmt[0].Columns)

	// we need to ensure that all cores have the same amount of result columns.
	// if so, we will simply order the first one, and then return.
	// we also need to check there are no group by clauses in any of the select cores.
	for _, core := range stmt {
		contains, err := containsGroupBy(core)
		if err != nil {
			return nil, fmt.Errorf("error checking for group by: %w", err)
		}

		if contains {
			return nil, ErrGroupByWithCompoundStatement
		}

		if len(core.Columns) != expectedNumberOfColumns {
			return nil, ErrCompoundStatementDifferentNumberOfColumns
		}
	}

	// we will order the first select core, and then return.
	return getReturnedColumnOrderingTerms(stmt[0].Columns, tables)
}

// containsGroupBy will return true if the select core contains a group by clause.
func containsGroupBy(stmt *tree.SimpleSelect) (bool, error) {
	contains := false
	depth := 0

	err := stmt.Walk(&tree.ImplementedListener{
		FuncEnterGroupBy: func(p0 *tree.GroupBy) error {
			if depth == 0 {
				if len(p0.Expressions) > 0 {
					contains = true
				}
			}
			return nil
		},
		FuncEnterSelectCore: func(p0 *tree.SelectCore) error {
			depth++
			return nil
		},
		FuncExitSelectCore: func(p0 *tree.SelectCore) error {
			depth--
			return nil
		},
	})

	return contains, err
}

// getReturnedColumnOrderingTerms gets the ordering terms for the returned columns.
// it is used to order result columns for compound select statements or distinct statements.
func getReturnedColumnOrderingTerms(resultCols []tree.ResultColumn, tables []*types.Table) ([]*tree.OrderingTerm, error) {
	terms := []*tree.OrderingTerm{}

	for _, col := range resultCols {
		switch c := col.(type) {
		default:
			panic(fmt.Sprintf("unexpected result column type: %T", c))
		case *tree.ResultColumnExpression:
			// we simply use the expression as the ordering term
			// this works even for complex expressions, such as
			// "SELECT sum(a%4/(b+2)) FROM table"
			terms = append(terms, &tree.OrderingTerm{
				Expression: c.Expression,
			})
		case *tree.ResultColumnTable:
			tbl, err := findTable(tables, c.TableName)
			if err != nil {
				return nil, fmt.Errorf("error finding table: %w", err)
			}

			terms = append(terms, getTableOrderTerms(tbl)...)
		case *tree.ResultColumnStar:
			sort.Slice(tables, func(i, j int) bool {
				return tables[i].Name < tables[j].Name
			})

			for _, tbl := range tables {
				terms = append(terms, getTableOrderTerms(tbl)...)
			}
		}
	}

	return terms, nil
}

// getTableOrderTerms returns the ordering terms for a table.
// It is used to order result columns for compound select statements or distinct statements.
func getTableOrderTerms(tbl *types.Table) []*tree.OrderingTerm {
	terms := []*tree.OrderingTerm{}

	columns := tbl.Columns
	sort.Slice(columns, func(i, j int) bool {
		return columns[i].Name < columns[j].Name
	})

	for _, col := range columns {
		terms = append(terms, &tree.OrderingTerm{
			Expression: &tree.ExpressionColumn{
				Table:  tbl.Name,
				Column: col.Name,
			},
		})
	}

	return terms
}

// findTable will return the table with the given name, or an error if it does not exist.
func findTable(tables []*types.Table, name string) (*types.Table, error) {
	for _, tbl := range tables {
		if tbl.Name == name {
			return tbl, nil
		}
	}

	return nil, fmt.Errorf("table not found: %s", name)
}
