package cost_1st

import (
	"github.com/kwilteam/kwil-db/internal/engine/cost/plan"
)

type Calculator struct {
}

func (c *Calculator) Estimate(stat plan.Statistic) float64 {
	return 0
}

//func GenCostCalculator(stmt string, schema *types.Schema, info plan.SchemaGetter) (*plan.CostEstimate, error) {
//	ctx := plan.NewPlannerContext(schema, info)
//
//	s, err := sqlparser.Parse(stmt)
//	if err != nil {
//		return nil, err
//	}
//
//	planner := plan.NewStmtPlanner()
//	p, err := planner.Plan(s, ctx)
//	if err != nil {
//		return nil, err
//	}
//
//	return nil, nil
//}
