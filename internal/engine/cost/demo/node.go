package demo

import "fmt"

type OperatorType uint16

const (
	unknownOp OperatorType = iota
	// Logical operator
	Logical
	LogicalSeqScan
	LogicalFilter
	LogicalLimit
	// Physical operator
	//Physical
)

var operatorName = [...]string{
	unknownOp: "unknown",
	// Logical operator
	Logical:        "Logical",
	LogicalSeqScan: "LogicalSeqScan",
	LogicalFilter:  "LogicalFilter",
	LogicalLimit:   "LogicalLimit",
	// Physical operator
}

type AnyVisitor struct {
	V Visitor
}

func (a *AnyVisitor) Visit(op Operator) any {
	return nil
}

type Operator interface {
	fmt.Stringer
	Accept(AnyVisitor) any
	//acceptRelation(Visitor, rel) any
	OpType() OperatorType
}

type baseOperatorNode struct {
	opType OperatorType
}

func (n *baseOperatorNode) Accept(v AnyVisitor) any {
	return nil
}

func (n *baseOperatorNode) OpType() OperatorType {
	return n.opType
}

func (n *baseOperatorNode) String() string {
	return operatorName[n.opType]
}

// LogicalScanOperator represents a logical scan operator.
type LogicalScanOperator struct {
	baseOperatorNode

	// TODO: getter/setter
	Table  string
	Cols   []string
	Filter string
}

func NewLogicalScanOperator(opType OperatorType,
	table string,
	cols []string) *LogicalScanOperator {
	return &LogicalScanOperator{
		baseOperatorNode: baseOperatorNode{
			opType: opType,
		},
		Table: table,
		Cols:  cols,
	}
}

func (n *LogicalScanOperator) Accept(v AnyVisitor) any {
	return v.V.VisitLogicalScan(n)
}

// LogicalFilterOperator represents a logical filter operator.
type LogicalFilterOperator struct {
	baseOperatorNode
}

func NewLogicalFilterOperator(opType OperatorType) *LogicalFilterOperator {
	return &LogicalFilterOperator{
		baseOperatorNode: baseOperatorNode{
			opType: opType,
		},
	}
}

func (n *LogicalFilterOperator) Accept(v AnyVisitor) any {
	return v.V.VisitLogicalFilter(n)
}

// LogicalLimitOperator represents a logical limit operator.
type LogicalLimitOperator struct {
	baseOperatorNode
}

func NewLogicalLimitOperator(opType OperatorType) *LogicalLimitOperator {
	return &LogicalLimitOperator{
		baseOperatorNode: baseOperatorNode{
			opType: opType,
		},
	}
}

func (n *LogicalLimitOperator) Accept(v AnyVisitor) any {
	return v.V.VisitLogicalLimit(n)
}
