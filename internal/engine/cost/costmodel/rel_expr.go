package costmodel

import (
	"bytes"
	"fmt"
	"github.com/kwilteam/kwil-db/internal/engine/cost/datatypes"
	"github.com/kwilteam/kwil-db/internal/engine/cost/logical_plan"
)

const (
	SeqAccessCostPerRow = 1 // sequential access disk cost
	RandAccessCost      = 3 // random access disk cost, i.e., index scan
)

// RelExpr is a wrapper of a logical plan,it's used for cost estimation.
// It tracks the statistics and cost from bottom to top.
// NOTE: this is simplified version of LogicalRel in memo package.
type RelExpr struct {
	logical_plan.LogicalPlan

	stat   *datatypes.Statistics // current node's statistics
	cost   int64
	inputs []*RelExpr
}

func (r *RelExpr) Inputs() []*RelExpr {
	return r.inputs
}

func (r *RelExpr) String() string {
	return fmt.Sprintf("%s, Stat: (%s), Cost: %d",
		logical_plan.PlanString(r.LogicalPlan), r.stat, r.cost)
}

// reorderColStat reorders the columns in the statistics according to the schema.
// Schema can be changed by the projection/join, so we need to reorder the columns in
// the statistics.
func reorderColStat(oldStat *datatypes.Statistics, schema *datatypes.Schema) *datatypes.Statistics {

}

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
		stat = inputs[0].stat

	case *logical_plan.FilterOp:
		stat = inputs[0].stat
		// with filter, we can make uniformity assumption to simplify the cost model
		exprs := p.Exprs()
		fields := make([]datatypes.Field, len(exprs))
		for i, expr := range exprs {
			fields[i] = expr.Resolve()
		}

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

//func EstimateCost(plan *RelExpr) int64 {
//	cost := int64(0)
//	// bottom-up
//	for _, child := range plan.Inputs() {
//		cost += EstimateCost(child)
//	}
//
//	// estimate current node's cost
//	switch plan.LogicalPlan.(type) {
//	case *logical_plan.ScanOp:
//		// TODO: index scan
//		cost += SeqAccessCost
//	}
//	return cost
//}
//

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
