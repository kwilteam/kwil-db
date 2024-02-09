package virtual_plan

import (
	"fmt"

	"github.com/kwilteam/kwil-db/internal/engine/cost/logical_plan"
	"github.com/kwilteam/kwil-db/internal/engine/cost/optimizer"
)

func Example_QueryPlanner_CreateVirtualPlan() {
	ctx := NewExecutionContext()
	df := ctx.csv("users", "../testdata/users.csv")
	plan := df.
		Filter(logical_plan.Eq(logical_plan.Column("", "age"),
			logical_plan.LiteralInt(20))).
		Project(logical_plan.Column("", "state"),
			logical_plan.Column("", "username"),
		).
		LogicalPlan()

	fmt.Println(logical_plan.Format(plan, 0))

	r := &optimizer.ProjectionRule{}
	got := r.Optimize(plan)

	fmt.Printf("---After optimization---\n\n")
	fmt.Println(logical_plan.Format(got, 0))

	qp := NewQueryPlanner()
	vp := qp.CreateVirtualPlan(got)
	fmt.Printf("---Got virtual plan---\n\n")
	fmt.Println(Format(vp, 0))

	// Output:
	// Projection: state, username
	//   Selection: [age = 20]
	//     Scan: users; projection=[]
	//
	// ---After optimization---
	//
	// Projection: state, username
	//   Selection: [age = 20]
	//     Scan: users; projection=[age state username]
	//
	// ---Got virtual plan---
	//
	// VProjection: [state@1 username@2]
	//   VSelection: age@0 = 20
	//     VScan: schema=[id/int, username/string, age/int, state/string, wallet/string], projection=[age state username]
}
