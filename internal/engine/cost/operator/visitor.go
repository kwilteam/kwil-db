package operator

type VisitorResult interface {
	string
}

type Visitor interface {
	VisitOperator(op Operator) any
	VisitLogicalScan(op *LogicalScanOperator) any
	VisitLogicalFilter(op *LogicalFilterOperator) any
	VisitLogicalLimit(op *LogicalLimitOperator) any
	VisitLogicalJoin(op *LogicalJoinOperator) any
	VisitLogicalSet(op *LogicalSetOperator) any
	VisitLogicalAggregate(op *LogicalAggregateOperator) any
	VisitLogicalTakeN(op *LogicalTakeNOperator) any
	VisitLogicalDistinct(op *LogicalDistinctOperator) any
	VisitLogicalProjection(op *LogicalProjectionOperator) any
}

//
//type BaseOperatorVisitor struct {
//}
//
//func (v *BaseOperatorVisitor) visitOperator(op Operator) any {
//	return nil
//}
//
//func (v *BaseOperatorVisitor) VisitLogicalScan(op *LogicalScanOperator) any {
//	return v.visitOperator(op)
//}
