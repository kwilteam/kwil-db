package query_planner

type PlannerContext struct {
	CurrentSchema string // the current postgres schema we are working on
	// cte's here
} // ???

func NewPlannerContext() *PlannerContext {
	return &PlannerContext{}
}
