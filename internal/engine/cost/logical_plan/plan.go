package logical_plan

import (
	"bytes"
	"fmt"
	"github.com/kwilteam/kwil-db/internal/engine/cost/datatypes"
)

type LogicalPlan interface {
	fmt.Stringer

	// Schema returns the schema of the data that will be produced by this LogicalPlan.
	Schema() *datatypes.Schema

	Inputs() []LogicalPlan

	// Exprs returns all expressions in the current logical plan node.
	// This does not include expressions in plan's inputs
	Exprs() []LogicalExpr

	//Accept(visitor LogicalOperatorVisitor) any
}

func Format(plan LogicalPlan, indent int) string {
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

type LogicalOperatorVisitor interface {
	Visit(LogicalPlan) any
	VisitScanOp(*ScanOp) any
	VisitProjectionOp(*ProjectionOp) any
	//VisitLogicalSubquery(*LogicalSubquery) any
	VisitSelectionOp(*FilterOp) any
	VisitJoinOp(*JoinOp) any
	VisitLimitOp(*LimitOp) any
	VisitAggregate(*AggregateOp) any
	VisitSort(*SortOp) any
	//VisitLogicalDistinct(*LogicalDistinct) any
	//VisitLogicalSet(*LogicalSet) any
}

type baseLogicalOperatorVisitor struct{}

func (v *baseLogicalOperatorVisitor) Visit(plan LogicalPlan) any {
	//return plan.Accept(v)
	return nil
}

func (v *baseLogicalOperatorVisitor) VisitScanOp(op *ScanOp) any {
	return nil
}

func (v *baseLogicalOperatorVisitor) VisitProjectionOp(op *ProjectionOp) any {
	return nil
}

func (v *baseLogicalOperatorVisitor) VisitSelectionOp(op *FilterOp) any {
	return nil
}

func (v *baseLogicalOperatorVisitor) VisitJoinOp(op *JoinOp) any {
	return nil
}

func (v *baseLogicalOperatorVisitor) VisitLimitOp(op *LimitOp) any {
	return nil
}

func (v *baseLogicalOperatorVisitor) VisitAggregate(op *AggregateOp) any {
	return nil
}

func (v *baseLogicalOperatorVisitor) VisitSort(op *SortOp) any {
	return nil
}

var _ LogicalOperatorVisitor = &baseLogicalOperatorVisitor{}

//func Explain(p LogicalPlan) string {
//	return explainWithPrefix(p, "", "")
//}
//
//func explainWithPrefix(p LogicalPlan, titlePrefix string, bodyPrefix string) string {
//	var msg bytes.Buffer
//	msg.WriteString(titlePrefix)
//
//	ov := NewExplainVisitor()
//	msg.WriteString(p.Accept(ov).(string))
//	msg.WriteString("\n")
//
//	for _, child := range p.Inputs() {
//		msg.WriteString(explainWithPrefix(
//			child,
//			bodyPrefix+"->  ",
//			bodyPrefix+"      "))
//	}
//	return msg.String()
//
//}
