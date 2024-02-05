package cost_2nd

type ExplainVisitor struct {
	//BaseOperatorVisitor
}

func (e *ExplainVisitor) Visit(p LogicalPlan) any {
	return p.Accept(e)
}

func (e *ExplainVisitor) VisitLogicalScan(p *LogicalScan) any {
	return p.String()
}

func (e *ExplainVisitor) VisitLogicalProjection(p *LogicalProjection) any {
	return p.String()
}

func (e *ExplainVisitor) VisitLogicalSubquery(p *LogicalSubquery) any {
	return p.String()
}

func (e *ExplainVisitor) VisitLogicalFilter(p *LogicalFilter) any {
	return p.String()
}

func (e *ExplainVisitor) VisitLogicalJoin(p *LogicalJoin) any {
	return p.String()
}

func (e *ExplainVisitor) VisitLogicalLimit(p *LogicalLimit) any {
	return p.String()
}

func (e *ExplainVisitor) VisitLogicalAggregate(p *LogicalAggregate) any {
	return p.String()
}

func (e *ExplainVisitor) VisitLogicalSort(p *LogicalSort) any {
	return p.String()
}

func (e *ExplainVisitor) VisitLogicalDistinct(p *LogicalDistinct) any {
	return p.String()
}

func (e *ExplainVisitor) VisitLogicalSet(p *LogicalSet) any {
	return p.String()
}

func NewExplainVisitor() *ExplainVisitor {
	return &ExplainVisitor{}
}

var _ LogicalVisitor = &ExplainVisitor{}
