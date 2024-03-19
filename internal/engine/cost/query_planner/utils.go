package query_planner

import (
	"fmt"
	dt "github.com/kwilteam/kwil-db/internal/engine/cost/datatypes"
	lp "github.com/kwilteam/kwil-db/internal/engine/cost/logical_plan"
	pt "github.com/kwilteam/kwil-db/internal/engine/cost/plantree"
	"slices"
	"strings"
)

func expandStar(schema *dt.Schema) []lp.LogicalExpr {
	fmt.Println("=-------", schema)
	// is there columns to skip?
	var exprs []lp.LogicalExpr
	for _, field := range schema.Fields {
		// TODO: better way to get column expr?
		exprs = append(exprs, &lp.ColumnExpr{Relation: field.Relation(), Name: field.Name})
	}
	return exprs
}

func expandQualifiedStar(schema *dt.Schema, table string) []lp.LogicalExpr {
	panic("not implemented")
}

// qualifyExpr returns a new expression qualified with the given relation.
// It won't change the original expression if it's not ColumnExpr.
// func qualifyExpr(expr lp.LogicalExpr, seen map[string] lp.LogicalExpr, schemas ...*dt.Schema) lp.LogicalExpr {
func qualifyExpr(expr lp.LogicalExpr, schemas ...*dt.Schema) lp.LogicalExpr {
	c, ok := expr.(*lp.ColumnExpr)
	if !ok {
		return expr
	}

	//// TODO: make all lp.LogicalExpr to implement pt.Node ?
	//return c.TransformUp(func(n pt.Node) pt.Node {
	//	if c, ok := n.(*lp.ColumnExpr); ok {
	//		c.QualifyWithSchema(seen, schemas...)
	//	}
	//	return n
	//
	//}).(*lp.ColumnExpr)

	return c.QualifyWithSchemas(schemas...)
}

// extractColumnsFromFilterExpr extracts the columns are references by the filter expression.
// It keeps track of the columns that have been seen in the 'seen' map.
// TODO: use visitor
func extractColumnsFromFilterExpr(expr lp.LogicalExpr, seen map[*lp.ColumnExpr]bool) {
	switch e := expr.(type) {
	case *lp.LiteralStringExpr:
	case *lp.LiteralIntExpr:
	case *lp.AliasExpr:
		extractColumnsFromFilterExpr(e.Expr, seen)
	case lp.UnaryExpr:
		extractColumnsFromFilterExpr(e.E(), seen)
	case lp.AggregateExpr:
		extractColumnsFromFilterExpr(e.E(), seen)
	case lp.BinaryExpr:
		extractColumnsFromFilterExpr(e.L(), seen)
		extractColumnsFromFilterExpr(e.R(), seen)
	case *lp.ColumnExpr:
		seen[e] = true
	//case *.ColumnIdxExpr:
	//	seen[input.Schema().Fields[e.Idx].Name] = true
	default:
		panic(fmt.Sprintf("unknown expression type %T", e))
	}
}

// extractAliases extracts the mapping of alias to its expression
func extractAliases(exprs []lp.LogicalExpr) map[string]lp.LogicalExpr {
	aliases := make(map[string]lp.LogicalExpr)
	for _, expr := range exprs {
		if e, ok := expr.(*lp.AliasExpr); ok {
			aliases[e.Alias] = e.Expr
		}
	}
	return aliases
}

func cloneAliases(aliases map[string]lp.LogicalExpr) map[string]lp.LogicalExpr {
	clone := make(map[string]lp.LogicalExpr)
	for k, v := range aliases {
		clone[k] = v
	}
	return clone
}

// resolveAliases resolves the expr to its un-aliased expression.
// It's used to resolve the alias in the select list to the actual expression.
func resolveAlias(expr lp.LogicalExpr, aliases map[string]lp.LogicalExpr) lp.LogicalExpr {
	e := expr.TransformUp(func(n pt.TreeNode) pt.TreeNode {
		if c, ok := n.(*lp.ColumnExpr); ok {
			if e, ok := aliases[c.Name]; ok {
				return e
			} else {
				return c
			}
		}
		// otherwise, return the original node
		return n
	})

	//_, e := pt.PostOrderApply(expr, func(n pt.TreeNode) (bool, any) {
	//	if e, ok := n.(*lp.ColumnExpr); ok {
	//		if e.Relation == nil {
	//			return true, nil
	//		}
	//
	//		if aliasExpr, ok := aliases[e.Name]; ok {
	//			return true, aliasExpr
	//		} else {
	//			return true, e
	//		}
	//	} else {
	//		return true, n
	//	}
	//})

	return e.(lp.LogicalExpr)
}

func extractAggrExprs(exprs []lp.LogicalExpr) []lp.LogicalExpr {
	var aggrExprs []lp.LogicalExpr
	for _, expr := range exprs {
		if e, ok := expr.(lp.AggregateExpr); ok {
			aggrExprs = append(aggrExprs, e)
		}
	}
	return aggrExprs
}

// allReferredColumns returns all the columns that are referenced by the expression.
func allReferredColumns(exprs []lp.LogicalExpr) []*lp.ColumnExpr {
	var columns []*lp.ColumnExpr
	for _, expr := range exprs {
		pt.PreOrderApply(expr, func(n pt.TreeNode) (bool, any) {
			if c, ok := n.(*lp.ColumnExpr); ok {
				columns = append(columns, c)
			}
			return true, nil
		})
	}
	return columns
}

// ensureSchemaSatifiesExprs ensures that the schema contains all the columns
// referenced by the expression.
func ensureSchemaSatifiesExprs(schema *dt.Schema, exprs []lp.LogicalExpr) error {
	referredCols := allReferredColumns(exprs)

	for _, col := range referredCols {
		if !schema.ContainsColumn(col.Relation, col.Name) {
			return fmt.Errorf("column %s not found in schema", col.Name)
		}
	}

	return nil
}

// rebaseExprs builds the expression on top of the base expressions.
// This is useful in the context of a query like:
// SELECT a + b < 1 ... GROUP BY a + b
func rebaseExprs(expr lp.LogicalExpr, baseExprs []lp.LogicalExpr, plan lp.LogicalPlan) lp.LogicalExpr {
	return expr.TransformDown(func(n pt.TreeNode) pt.TreeNode {
		contains := slices.ContainsFunc(baseExprs, func(e lp.LogicalExpr) bool {
			// TODO: String() may not work
			return e.String() == n.String()
		})

		if contains {
			return exprAsColumn(n.(lp.LogicalExpr), plan)
		} else {
			return n
		}
	}).(lp.LogicalExpr)
}

// checkExprsProjectFromColumns checks if the expression can be projected from the columns.
func checkExprsProjectFromColumns(exprs []lp.LogicalExpr, columns []lp.LogicalExpr) error {
	for _, col := range columns {
		if _, ok := col.(*lp.ColumnExpr); !ok {
			return fmt.Errorf("expression %s is not a column", col.String())
		}
	}

	colExprs := allReferredColumns(exprs)
	for _, col := range colExprs {
		if err := checkExprProjectFromColumns(col, columns); err != nil {
			return err
		}
	}
	return nil
}

func checkExprProjectFromColumns(expr lp.LogicalExpr, columns []lp.LogicalExpr) error {
	valid := slices.ContainsFunc(columns, func(c lp.LogicalExpr) bool {
		return c.String() == expr.String()
	})

	if !valid {
		return fmt.Errorf(
			"expression %s cannot be resolved from available columns: %s",
			expr.String(), columns)
	} else {
		return nil
	}
}

func exprAsColumn(expr lp.LogicalExpr, plan lp.LogicalPlan) *lp.ColumnExpr {
	if c, ok := expr.(*lp.ColumnExpr); ok {
		colDef := lp.ColumnFromExprToDef(c)
		field := plan.Schema().FieldFromColumn(colDef)
		return lp.ColumnFromDefToExpr(field.QualifiedColumn())
	} else {
		// use the expression as the column name
		// TODO: String() may not work
		return lp.ColumnUnqualified(expr.String())
	}
}

type TableRefName string

func (t TableRefName) String() string {
	return string(t)
}

func (t TableRefName) Segments() []string {
	return strings.Split(string(t), ".")
}

func relationNameToTableRef(relationName string) (*dt.TableRef, error) {
	tr := TableRefName(relationName)
	segments := tr.Segments()
	switch len(segments) {
	case 1:
		return &dt.TableRef{Table: segments[0]}, nil
	case 2:
		return &dt.TableRef{Schema: segments[0], Table: segments[1]}, nil
	default:
		return nil, fmt.Errorf("invalid relation name: %s", relationName)
	}
}
