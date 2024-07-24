package query_planner

import (
	"fmt"

	dt "github.com/kwilteam/kwil-db/internal/engine/cost/datatypes"
	lp "github.com/kwilteam/kwil-db/internal/engine/cost/logical_plan"
	tree "github.com/kwilteam/kwil-db/parse"
)

/*
	This file is mostly my (brennan) attempt at understanding the construction of the
	logical query plan, and making it match more closely to our new SQL AST
*/

type queryPlanner2 struct {
	catalog Catalog
}

func (q *queryPlanner2) buildSelectStmt(node *tree.SelectStatement, ctx *PlannerContext) lp.LogicalPlan {
}

// buildSelectCore builds a logical plan for a simple select statement.
// The order of building is:
// 1. from (combining any joins into single source plan)
// 2. where
// 3. group by(can use reference from select)
// 4. having(can use reference from select)
// 5. select
// 6. distinct
// 7. order by, done in buildSelect
// 8. limit, done in buildSelect
func (q *queryPlanner2) buildSelectCore(node *tree.SelectCore, ctx *PlannerContext) lp.LogicalPlan {
	// building the relations from FROM and JOIN
	var plan lp.LogicalPlan
	if node.From == nil {
		plan = lp.NoSource()
	} else {
		plan = q.buildTable(node.From, ctx)
	}

}

// buildTable builds a logical plan for a table.
// it is used for the parse.Table interface, which specifies a tables
// used in a SELECT statement.
func (q *queryPlanner2) buildTable(node tree.Table, ctx *PlannerContext) lp.LogicalPlan {
	switch t := node.(type) {
	case *tree.RelationTable:
		tableRef := &dt.TableRef{
			Namespace: ctx.CurrentSchema,
			Table:     t.Table,
		}

		schemaProvider, err := q.catalog.GetDataSource(tableRef)
		if err != nil {
			panic(err)
		}

		if t.Alias != "" {
			tableRef.Table = t.Alias
		}

		return lp.ScanPlan(tableRef, schemaProvider, nil)
	case *tree.RelationSubquery:
		plan := q.buildSelectStmt(t.Subquery, ctx)
		lp.ScanPlan()
	case *tree.RelationFunctionCall:
		panic("not implemented")
	default:
		panic(fmt.Sprintf("unexpected table type %T", t))
	}
}
