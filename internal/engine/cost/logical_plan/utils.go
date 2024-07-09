package logical_plan

import (
	"fmt"

	ds "github.com/kwilteam/kwil-db/internal/engine/cost/datatypes"
	pt "github.com/kwilteam/kwil-db/internal/engine/cost/plantree"
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
// Expressions that terminate with a ColumnExpr or ColumnIdxExpr will make an
// entry in the seen map.
func ExtractColumns(expr LogicalExpr, schema *ds.Schema, seen map[string]bool) {
	switch e := expr.(type) {
	case *LiteralTextExpr:
	case *LiteralNumericExpr:
	case *AliasExpr:
		ExtractColumns(e.Expr, schema, seen)
	case UnaryExpr: // e.g. NOT (expr)
		ExtractColumns(e.E(), schema, seen)
	case AggregateExpr:
		ExtractColumns(e.E(), schema, seen)
	case BinaryExpr: // e.g. AND, OR, LT, etc.
		ExtractColumns(e.L(), schema, seen)
		ExtractColumns(e.R(), schema, seen)
	case *ColumnExpr:
		seen[e.Name] = true
	case *ColumnIdxExpr:
		seen[schema.Fields[e.Idx].Name] = true
	case *SortExpression:
		ExtractColumns(e.Expr, schema, seen)
	default:
		panic(fmt.Sprintf("ExtractColumns: unknown expression type %T", e))
	}
}

// NormalizeColumn qualifies a column with schema from a given logical plan.
// i.e. This creates a new ColumnExpr with the Relation field set.
// func NormalizeColumn(plan LogicalPlan, column *ColumnExpr) *ColumnExpr {
// 	return column.QualifyWithSchemas(plan.Schema())
// }

// NormalizeExpr normalizes the given expression with the given logical plan.
// That is, if the expression is a ColumnExpr, it uses the plan's Schema to set
// the Relation field.
func NormalizeExpr(expr LogicalExpr, plan LogicalPlan) LogicalExpr {
	return pt.TransformPostOrder(expr, func(n pt.TreeNode) pt.TreeNode {
		if c, ok := n.(*ColumnExpr); ok {
			return c.QualifyWithSchemas(plan.Schema()) // NormalizeColumn(plan, c)
		}
		return n
	}).(LogicalExpr)

	//return expr.TransformUp(func(n pt.TreeNode) pt.TreeNode {
	//	if c, ok := n.(*ColumnExpr); ok {
	//		return NormalizeColumn(plan, c)
	//	}
	//	return n
	//}).(LogicalExpr)
}

func NormalizeExprs(exprs []LogicalExpr, plan LogicalPlan) []LogicalExpr {
	normalized := make([]LogicalExpr, len(exprs))
	for i, e := range exprs {
		normalized[i] = NormalizeExpr(e, plan)
	}
	return normalized
}

func ColumnFromDefToExpr(column *ds.ColumnDef) *ColumnExpr {
	return Column(column.Relation, column.Name)
}

func ColumnFromExprToDef(column *ColumnExpr) *ds.ColumnDef {
	return &ds.ColumnDef{
		Relation: column.Relation,
		Name:     column.Name,
	}
}

func exprListToNodeList(exprs []LogicalExpr) []pt.TreeNode {
	nodes := make([]pt.TreeNode, len(exprs))
	for i, e := range exprs {
		nodes[i] = e
	}
	return nodes
}

func exprListToFields(exprs []LogicalExpr, schema *ds.Schema) []ds.Field {
	fields := make([]ds.Field, len(exprs))
	for i, e := range exprs {
		switch t := e.(type) {
		case *ColumnExpr:
			fields[i] = ds.Field{
				Name:     t.Name,
				Type:     inferExprType(t, schema),
				Nullable: inferNullable(t, schema),
				Rel:      t.Relation,
			}
		case *AliasExpr:
			fields[i] = ds.Field{
				Name:     t.Alias,
				Type:     inferExprType(t, schema),
				Nullable: inferNullable(t, schema),
				Rel:      t.Relation,
			}
		default:
			fields[i] = ds.Field{
				Name:     t.String(),
				Type:     inferExprType(t, schema),
				Nullable: inferNullable(t, schema),
			}
		}
	}
	return fields
}

// inferExprType returns the type of the expression, based on the schema.
// For example, if `col + 1` should return int
func inferExprType(expr LogicalExpr, schema *ds.Schema) string {
	switch e := expr.(type) {
	case *ColumnExpr:
		return schema.FieldFromColumn(ColumnFromExprToDef(e)).Type
	case *AliasExpr:
		panic("implement me")
	default:
		panic(fmt.Sprintf("inferExprType: unknown expression type %T", e))
	}
}

func inferNullable(expr LogicalExpr, schema *ds.Schema) bool {
	switch e := expr.(type) {
	case *ColumnExpr:
		panic("implement me")
	default:
		panic(fmt.Sprintf("inferNUllable: unknown expression type %T", e))
	}
}

// PpList returns a string representation of the given list.
func PpList[T any](l []T) string {
	if len(l) == 0 {
		return ""
	}

	str := ""
	for i, e := range l {
		str += fmt.Sprintf("%v", e)
		if i < len(l)-1 {
			str += ", "
		}
	}
	return str
}
