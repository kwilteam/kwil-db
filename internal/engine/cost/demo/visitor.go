package demo

import "fmt"

type VisitorResult interface {
	string
}

type Visitor interface {
	VisitLogicalScan(op *LogicalScanOperator) any
	VisitLogicalFilter(op *LogicalFilterOperator) any
	VisitLogicalLimit(op *LogicalLimitOperator) any
}

type BaseOperatorVisitor struct {
}

func (v *BaseOperatorVisitor) visitOperator(op Operator) any {
	return nil
}

func (v *BaseOperatorVisitor) VisitLogicalScan(op *LogicalScanOperator) any {
	return v.visitOperator(op)
}

// debugOperatorVisitor is a visitor that prints the operator visited.
type debugOperatorVisitor struct {
	BaseOperatorVisitor
}

func NewDebugOperatorVisitor() *debugOperatorVisitor {
	return &debugOperatorVisitor{}
}

func (v *debugOperatorVisitor) VisitLogicalScan(op *LogicalScanOperator) any {
	return fmt.Sprintf("LogicalScanOpeartor table=%s", op.Table)
}

func (v *debugOperatorVisitor) VisitLogicalFilter(op *LogicalFilterOperator) any {
	return fmt.Sprintf("LogicalFilterOperator op=%s", op.String())
}

func (v *debugOperatorVisitor) VisitLogicalLimit(op *LogicalLimitOperator) any {
	return fmt.Sprintf("LogicalLimitOperator op=%s", op.String())
}
