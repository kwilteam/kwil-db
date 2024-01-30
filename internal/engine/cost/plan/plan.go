package plan

import "github.com/kwilteam/kwil-db/internal/engine/types"

type Statistic struct{}

type CostEstimate interface {
	Estimate(Statistic) float64
}

//
//type Plan interface {
//	CostEstimate
//
//	PlanType() operator.OperatorType
//}
//
//type LogicalPlan struct {
//	root *OperationBuilder
//}
//
//func NewLogicalPlan(root *OperationBuilder) *LogicalPlan {
//	return &LogicalPlan{
//		root: root,
//	}
//}
//
//func (p *LogicalPlan) PlanType() operator.OperatorType {
//	return p.root.op.OpType()
//}
//
//func (p *LogicalPlan) Estimate(statistic Statistic) float64 {
//	return 0
//}

type field struct {
	Name     string
	Type     string
	Nullable bool
}

type schema struct {
	cols []*types.Column
	keys []*types.Index
}

func newSchema(cols ...*types.Column) *schema {
	return &schema{
		cols: cols,
	}
}

type Plan interface {
	Schema() *schema
	SetSchema(*schema)
}

type basePlan struct {
	schema *schema
}

func (p *basePlan) Schema() *schema {
	return p.schema
}

func (p *basePlan) SetSchema(s *schema) {
	p.schema = s
}

type LogicalPlan interface {
	Plan
	Children() []LogicalPlan
}
