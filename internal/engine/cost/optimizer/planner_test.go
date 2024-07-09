package optimizer

import (
	"fmt"

	"github.com/kwilteam/kwil-db/internal/engine/cost/internal/testkit"
	lp "github.com/kwilteam/kwil-db/internal/engine/cost/logical_plan"
	"github.com/kwilteam/kwil-db/internal/engine/cost/virtual_plan"
)

func Example_QueryPlanner_CreateVirtualPlan() {
	catalog := testkit.InitMockCatalog()
	dataSrc, err := catalog.GetDataSource(stubTable)
	if err != nil {
		panic(err)
	}

	df := lp.NewDataFrame(
		lp.ScanPlan(stubTable, dataSrc, nil))

	plan := df.
		Filter(lp.Eq(lp.Column(stubTable, "age"),
			lp.LiteralNumeric(20))).
		Project(lp.Column(stubTable, "state"),
			lp.Column(stubTable, "username"),
		).
		LogicalPlan()

	fmt.Println(lp.Format(plan, 0))

	r := &ProjectionRule{}
	got := r.Transform(plan)

	fmt.Printf("---After optimization---\n\n")
	fmt.Println(lp.Format(got, 0))

	qp := NewPlanner()
	vp := qp.ToPlan(got)
	fmt.Printf("---Got virtual plan---\n\n")
	fmt.Println(virtual_plan.Format(vp, 0))

	// Output:
	// Projection: users.state, users.username
	//   Filter: users.age = 20
	//     Scan: users
	//
	// ---After optimization---
	//
	// Projection: users.state, users.username
	//   Filter: users.age = 20
	//     Scan: users; projection=[age, state, username]
	//
	// ---Got virtual plan---
	//
	// VProjection: [state@1 username@2]
	//   VFilter: age@0 = 20
	//     VSeqScan: schema=[id/int64, username/string, age/int64, state/string, wallet/string], projection=[age state username]
}
