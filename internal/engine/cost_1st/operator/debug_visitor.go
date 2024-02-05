package operator

import (
	"bytes"
)

type debugOperatorVisitor struct {
	//BaseOperatorVisitor
}

func NewDebugOperatorVisitor() *debugOperatorVisitor {
	return &debugOperatorVisitor{}
}

func (v *debugOperatorVisitor) VisitOperator(op Operator) any {
	return op.Accept(v)
}

func (v *debugOperatorVisitor) VisitLogicalScan(op *LogicalScanOperator) any {
	return op.String()
	//return fmt.Sprintf("LogicalScan table=%s\n", op.Table)
}

func (v *debugOperatorVisitor) VisitLogicalFilter(op *LogicalFilterOperator) any {
	//return fmt.Sprintf("LogicalFilter op=%s\n", op.String())
	return op.String()
}

func (v *debugOperatorVisitor) VisitLogicalLimit(op *LogicalLimitOperator) any {
	//return fmt.Sprintln("LogicalLimit")
	return op.String()
}

func (v *debugOperatorVisitor) VisitLogicalJoin(op *LogicalJoinOperator) any {
	//var msg bytes.Buffer
	//msg.WriteString(op.String())
	//return msg.String()
	return op.String()
}

func (v *debugOperatorVisitor) VisitLogicalSet(op *LogicalSetOperator) any {
	//return fmt.Sprintln("LogicalSet")
	return op.String()
}

func (v *debugOperatorVisitor) VisitLogicalAggregate(op *LogicalAggregateOperator) any {
	var msg bytes.Buffer

	msg.WriteString("LogicalAggregate: ")

	//for col, fn := range op.colsAggrFuncs {
	//	msg.WriteString(fmt.Sprintf("%s => %s", col.ToSQL(), fn.FunctionName))
	//}

	msg.WriteString("\n")

	return msg.String()
}

func (v *debugOperatorVisitor) VisitLogicalTakeN(op *LogicalTakeNOperator) any {
	//return fmt.Sprintln("LogicalTakeN")
	return op.String()
}

func (v *debugOperatorVisitor) VisitLogicalDistinct(op *LogicalDistinctOperator) any {
	//return fmt.Sprintln("LogicalDistinct")
	return op.String()
}

func (v *debugOperatorVisitor) VisitLogicalProjection(op *LogicalProjectionOperator) any {
	//return fmt.Sprintln("LogicalProjection")
	return op.String()
}
