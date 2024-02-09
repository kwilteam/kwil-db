package logical_plan_test

import (
	"fmt"

	"github.com/kwilteam/kwil-db/internal/engine/cost/datasource"
	"github.com/kwilteam/kwil-db/internal/engine/cost/logical_plan"
)

func ExampleLogicalPlan_String_selection() {
	ds := datasource.NewMemDataSource(nil, nil)
	plan := logical_plan.Scan("users", ds)
	plan = logical_plan.Projection(plan, logical_plan.Column("", "username"), logical_plan.Column("", "age"))
	fmt.Println(logical_plan.Format(plan, 0))
	// Output:
	// Projection: username, age
	//   Scan: users; projection=[]
}

func ExampleLogicalPlan_DataFrame() {
	ds := datasource.NewMemDataSource(nil, nil)
	aop := logical_plan.NewDataFrame(logical_plan.Scan("users", ds))
	plan := aop.Filter(logical_plan.Eq(logical_plan.Column("", "age"), logical_plan.LiteralInt(20))).
		Aggregate([]logical_plan.LogicalExpr{logical_plan.Column("", "state")},
			[]logical_plan.AggregateExpr{logical_plan.Count(logical_plan.Column("", "username"))}).
		// the alias for aggregate result is bit weird
		Project(logical_plan.Column("", "state"), logical_plan.Alias(logical_plan.Count(logical_plan.Column("", "username")), "num")).
		LogicalPlan()

	fmt.Println(logical_plan.Format(plan, 0))
	// Output:
	// Projection: state, COUNT(username) AS num
	//   Aggregate: [state], [COUNT(username)]
	//     Selection: [age = 20]
	//       Scan: users; projection=[]
}
