package algebra

import (
	"bytes"
	"fmt"
)

// Field represents a field in a schema.
type Field struct {
	Name string
	Type string
}

type schema struct {
	Fields []Field
}

func Schema(fields ...Field) *schema {
	return &schema{Fields: fields}
}

type LogicalPlan interface {
	fmt.Stringer

	// Schema returns the schema of the data that will be produced by this LogicalPlan.
	Schema() *schema

	Inputs() []LogicalPlan
}

type DataFrame interface {
	// Project applies a projection
	Project(expr []LogicalExpr) DataFrame

	// Filter applies a filter
	Filter(expr LogicalExpr) DataFrame

	// Aggregate appliex an aggregation
	Aggregate(groupBy []LogicalExpr, aggregateExpr []AggregateExpr) DataFrame

	// Schema returns the schema of the data that will be produced by this DataFrame.
	Schema() *schema

	// LogicalPlan returns the logical plan
	LogicalPlan() LogicalPlan
}

type DataFrameImpl struct {
	plan LogicalPlan
}

func (df *DataFrameImpl) Project(expr []LogicalExpr) DataFrame {
	return &DataFrameImpl{Projection(df.plan, expr)}
}

func (df *DataFrameImpl) Filter(expr LogicalExpr) DataFrame {
	return &DataFrameImpl{Selection(df.plan, expr)}
}

func (df *DataFrameImpl) Aggregate(groupBy []LogicalExpr, aggregateExpr []AggregateExpr) DataFrame {
	return &DataFrameImpl{Aggregate(df.plan, groupBy, aggregateExpr)}
}

func (df *DataFrameImpl) Schema() *schema {
	return df.plan.Schema()
}

func (df *DataFrameImpl) LogicalPlan() LogicalPlan {
	return df.plan
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

// use visitor to explain
//type LogicalVisitor interface {
//	Visit(LogicalPlan) any
//	VisitLogicalScan(*LogicalScan) any
//	VisitLogicalProjection(*LogicalProjection) any
//	VisitLogicalSubquery(*LogicalSubquery) any
//	VisitLogicalFilter(*LogicalFilter) any
//	VisitLogicalJoin(*LogicalJoin) any
//	VisitLogicalLimit(*LogicalLimit) any
//	VisitLogicalAggregate(*LogicalAggregate) any
//	VisitLogicalSort(*LogicalSort) any
//	VisitLogicalDistinct(*LogicalDistinct) any
//	VisitLogicalSet(*LogicalSet) any
//}

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
