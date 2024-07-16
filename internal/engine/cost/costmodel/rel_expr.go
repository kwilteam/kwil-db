package costmodel

import (
	"bytes"
	"fmt"

	"github.com/kwilteam/kwil-db/internal/engine/cost/datatypes"
	"github.com/kwilteam/kwil-db/internal/engine/cost/logical_plan"
)

const (
	SeqScanRowCost   = 100
	IndexScanRowCost = 5

	SeqAccessCost  = 1 // sequential access disk cost
	RandAccessCost = 3 // random access disk cost, i.e., index scan (? if the index doesn't also store the projected or filtered data ?)

	ProjectionCost = 2 // cost for returning the data?
	FilterEqCost   = 2
)

// RelExpr is a wrapper of a logical plan,it's used for cost estimation.
// It tracks the statistics and cost from bottom to top.
// NOTE: this is simplified version of LogicalRel in memo package.
type RelExpr struct {
	logical_plan.LogicalPlan

	stat   *datatypes.Statistics // current node's statistics
	cost   int64                 // ??? remove?
	inputs []*RelExpr            // LogicalPlan.Inputs() each converted into a RelExpr
}

func (r *RelExpr) Inputs() []*RelExpr {
	return r.inputs
}

func (r *RelExpr) String() string {
	return fmt.Sprintf("%s, Stat: (%s), Cost: %d",
		r.LogicalPlan, r.stat, r.cost)
}

//// reorderColStat reorders the columns in the statistics according to the schema.
//// Schema can be changed by the projection/join, so we need to reorder the columns in
//// the statistics.
//func reorderColStat(oldStat *datatypes.Statistics, schema *datatypes.Schema) *datatypes.Statistics {
//
//}

// BuildRelExpr builds a RelExpr from a logical plan, also build the statistics.
// TODO: using iterator to traverse the plan tree.
func BuildRelExpr(plan logical_plan.LogicalPlan) *RelExpr {
	inputs := make([]*RelExpr, len(plan.Inputs()))
	for i, input := range plan.Inputs() {
		inputs[i] = BuildRelExpr(input)
	}

	var stat *datatypes.Statistics

	switch p := plan.(type) {
	case *logical_plan.ScanOp:
		stat = p.DataSource().Statistics()

	case *logical_plan.ProjectionOp:
		stat = inputs[0].stat // up

	case *logical_plan.FilterOp:
		stat = inputs[0].stat // up
		// with filter, we can make uniformity assumption to simplify the cost model
		exprs := p.Exprs()
		fields := make([]datatypes.Field, len(exprs))
		for i, expr := range exprs {
			fields[i] = expr.Resolve(plan.Schema())
		}

	// case *logical_plan.AggregateOp:
	// case *logical_plan.BagOp:

	default:
		stat = datatypes.NewEmptyStatistics()
	}

	return &RelExpr{
		LogicalPlan: plan,
		cost:        0,
		inputs:      inputs,
		stat:        stat,
	}
}

func Format(plan *RelExpr, indent int) string {
	var msg bytes.Buffer
	for i := 0; i < indent; i++ {
		msg.WriteString(" ")
	}
	msg.WriteString(plan.String())
	msg.WriteString("\n")
	for _, child := range plan.Inputs() {
		msg.WriteString(Format(child, indent+2))
	}
	return msg.String()
}

func EstimateCost(plan *RelExpr) int64 {
	var cost int64
	// bottom-up
	for _, child := range plan.Inputs() {
		cost += EstimateCost(child)
	}

	// estimate current node's cost
	switch p := plan.LogicalPlan.(type) {
	case *logical_plan.ScanOp:
		rows := plan.stat.RowCount // all the way from the scan
		// TODO: index scan
		plan.cost = SeqScanRowCost * rows // set plan.cost for printing?
		cost += plan.cost

		// if pushdown ran, ScanOp will have filter and/or projections
		// ... so reduce the cost??? how?

		// TODO: other cases based on Cost() methods of types in virtual_plan/operator.go

	case *logical_plan.ProjectionOp:
		// p.Exprs() ??? what about the cost of an expression like Add/+
		plan.cost = ProjectionCost * int64(len(p.Exprs()))
		cost += plan.cost

	case *logical_plan.FilterOp:
		// Cost of the filter depends on type and number of operations applied
		// in the expressions.
		exp := p.Exprs()[0] // FilterOp has one, which may nest others via logical

		plan.cost = ExprCost(exp, p)
		cost += plan.cost

		// now how does filter selectivity get applied to RowCount up in ScanOp???

		// also, if we want selectivity, we can't have arg placeholders like $1,
		// we need an actual value. Do we need to rewrite the AST with literals
		// substituted for the arguments first?
		// Maybe let's allow variables but make that result in no cost reduction,
		// (assuming it would filter out nothing i.e. high selectivity).

	case *logical_plan.AggregateOp:
	case *logical_plan.BagOp:
	case *logical_plan.DistinctOp:
	case *logical_plan.JoinOp:
	case *logical_plan.LimitOp:
	case *logical_plan.NoRelationOp:
	case *logical_plan.SortOp:
	case *logical_plan.SubqueryOp:
	}
	return cost
}

// ExprCost returns the cost for an expression.  The idea is that evaluation of
// the expression is not free e.g. arithmetic. So stringing together a massive
// formula or logical expressions isn't for free.
func ExprCost(expr logical_plan.LogicalExpr, input logical_plan.LogicalPlan) int64 {
	switch e := expr.(type) {
	case *logical_plan.LiteralNumericExpr:
		return 0
	case *logical_plan.LiteralTextExpr, *logical_plan.LiteralBoolExpr,
		*logical_plan.LiteralNullExpr, *logical_plan.LiteralBlobExpr:
		return 0
	case logical_plan.AggregateExpr: // e.g. SUM / MIN / AVG / COUNT
		// are these free if already used in a group by?
		return ExprCost(e.E(), input) + 8 // hmm, the operation has cost, but reduces data returned...

	case *logical_plan.AliasExpr:
		return ExprCost(e.Expr, input)

	case *logical_plan.SortExpression: // pseudo-expression, part of filter/order plan
		return ExprCost(e.Expr, input)

	case *logical_plan.ColumnExpr:
		for _, field := range input.Schema().Fields {
			if field.Name == e.Name {
				return 1 // base on field.Type?
			}
		}
		return 0 // panic(fmt.Sprintf("field %s not found", e.Name)) // need projection with sort / order by?
	case *logical_plan.ColumnIdxExpr:
		return 1 // ? input.Schema().Fields[e.Idx].Type

	case logical_plan.UnaryExpr: // e.g. NOT / + (positive) / - (negate)
		return ExprCost(e.E(), input) // all ops free, cost only for the targeted expression

	case logical_plan.BinaryExpr: // *boolBinaryExpr, *arithmeticBinaryExpr
		cost := ExprCost(e.L(), input) + ExprCost(e.R(), input)
		switch e.Op() {
		case "AND", "OR":
			return 0 + cost
		case "=", "!=", ">", "<", ">=", "<=":
			return 0 + cost
		case "+", "-":
			return 1 + cost
		case "*", "/":
			return 2 + cost
		default:
			panic(fmt.Sprintf("unknown binary operator %s", e.Op()))
		}

	default:
		return 0
	}
}

//// EstimateCost estimates the cost of a logical plan.
//// It uses iterator to traverse the plan tree.
//func EstimateCost(plan *RelExpr) int64 {
//	stack := []*RelExpr{plan}
//	cost := int64(0)
//
//	for len(stack) > 0 {
//		// Pop a node from the stack
//		n := len(stack) - 1
//		node := stack[n]
//		stack = stack[:n]
//
//		// Estimate current node's cost
//		switch p := node.LogicalPlan.(type) {
//		case *logical_plan.ScanOp:
//			// TODO: index scan
//			cost += p.
//		}
//
//		// Push all children onto the stack
//		for _, child := range node.Inputs() {
//			stack = append(stack, child)
//		}
//	}
//
//	return cost
//}
