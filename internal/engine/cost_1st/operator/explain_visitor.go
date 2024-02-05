package operator

type ExplainVisitor struct {
	//BaseOperatorVisitor
}

func NewExplainVisitor() *ExplainVisitor {
	return &ExplainVisitor{}
}

func (v *ExplainVisitor) VisitOperator(op Operator) any {
	return op.Accept(v)
}

func (v *ExplainVisitor) VisitLogicalScan(op *LogicalScanOperator) any {
	return op.String()
}

func (v *ExplainVisitor) VisitLogicalFilter(op *LogicalFilterOperator) any {
	return op.String()
}

func (v *ExplainVisitor) VisitLogicalLimit(op *LogicalLimitOperator) any {
	return op.String()
}

func (v *ExplainVisitor) VisitLogicalJoin(op *LogicalJoinOperator) any {
	return op.String()
}

func (v *ExplainVisitor) VisitLogicalSet(op *LogicalSetOperator) any {
	return op.String()
}

func (v *ExplainVisitor) VisitLogicalAggregate(op *LogicalAggregateOperator) any {
	return op.String()
}

func (v *ExplainVisitor) VisitLogicalTakeN(op *LogicalTakeNOperator) any {
	return op.String()
}

func (v *ExplainVisitor) VisitLogicalDistinct(op *LogicalDistinctOperator) any {
	return op.String()
}

func (v *ExplainVisitor) VisitLogicalProjection(op *LogicalProjectionOperator) any {
	return op.String()
}
