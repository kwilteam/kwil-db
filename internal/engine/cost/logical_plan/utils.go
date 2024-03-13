package logical_plan

import (
	"fmt"
	ds "github.com/kwilteam/kwil-db/internal/engine/cost/datasource"
)

// SplitConjunction splits the given expression into a list of expressions.
// A conjunction is an expression connected by AND operators.
// If the given expression is not a conjunction, it returns a list with the
// given expression.
func SplitConjunction(expr LogicalExpr) []LogicalExpr {
	return splitConjunction(expr, []LogicalExpr{})
}

// splitConjunction recursively splits the given conjunction into seen array.
func splitConjunction(expr LogicalExpr, seen []LogicalExpr) []LogicalExpr {
	switch e := expr.(type) {
	case *AliasExpr:
		return splitConjunction(e.Expr, seen)
	case BoolBinaryExpr:
		if e.Op() == "AND" {
			seen = splitConjunction(e.L(), seen)
			return splitConjunction(e.R(), seen)
		} else {
			seen = append(seen, e)
			return seen
		}
	default:
		seen = append(seen, e)
		return seen
	}
}

// Conjunction returns a new expression that is a conjunction of the given
// expressions.
func Conjunction(exprs ...LogicalExpr) (expr LogicalExpr) {
	expr = exprs[0]
	for i := 1; i < len(exprs); i++ {
		expr = And(expr, exprs[i])
	}
	return
}

// ExtractColumns extracts the columns from the expression.
// It keeps track of the columns that have been seen in the 'seen' map.
func ExtractColumns(expr LogicalExpr,
	schema *ds.Schema, seen map[string]bool) {
	switch e := expr.(type) {
	case *LiteralStringExpr:
	case *LiteralIntExpr:
	case *AliasExpr:
		ExtractColumns(e.Expr, schema, seen)
	case UnaryExpr:
		ExtractColumns(e.E(), schema, seen)
	case AggregateExpr:
		ExtractColumns(e.E(), schema, seen)
	case BinaryExpr:
		ExtractColumns(e.L(), schema, seen)
		ExtractColumns(e.R(), schema, seen)
	case *ColumnExpr:
		seen[e.Name] = true
	case *ColumnIdxExpr:
		seen[schema.Fields[e.Idx].Name] = true
	default:
		panic(fmt.Sprintf("unknown expression type %T", e))
	}
}
