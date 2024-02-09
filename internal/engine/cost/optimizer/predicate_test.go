package optimizer

import (
	"fmt"
	"github.com/kwilteam/kwil-db/internal/engine/cost/datasource"
	"github.com/kwilteam/kwil-db/internal/engine/cost/logical_plan"
)

func ExamplePredicateRule_optimize_pushDown() {
	ds := datasource.NewMemDataSource(nil, nil)
	aop := logical_plan.NewDataFrame(logical_plan.Scan("users", ds, nil))
	plan := aop.
		Filter(logical_plan.Eq(logical_plan.Column("", "age"),
			logical_plan.LiteralInt(20))).
		//Project(logical_plan.Column("", "state"),
		//	logical_plan.Alias(logical_plan.Column("", "username"), "name")).
		LogicalPlan()

	fmt.Println(logical_plan.Format(plan, 0))

	r := &PredicateRule{}
	got := r.Optimize(plan)

	fmt.Printf("---After optimization---\n\n")
	fmt.Println(logical_plan.Format(got, 0))

	// Output:
	// Selection: age = 20
	//   Scan: users; projection=[]
	//
	// ---After optimization---
	//
	// Selection: age = 20
	//   Scan: users; selection=[age = 20]; projection=[]
}

func ExamplePredicateRule_optimize_pushDown_with_nested_selection() {
	ds := datasource.NewMemDataSource(nil, nil)
	aop := logical_plan.NewDataFrame(logical_plan.Scan("users", ds, nil))
	plan := aop.
		Filter(logical_plan.Gt(logical_plan.Column("", "age"),
			logical_plan.LiteralInt(20))).
		Filter(logical_plan.Lt(logical_plan.Column("", "age"),
			logical_plan.LiteralInt(30))).
		LogicalPlan()

	fmt.Println(logical_plan.Format(plan, 0))

	r := &PredicateRule{}
	got := r.Optimize(plan)

	fmt.Printf("---After optimization---\n\n")
	fmt.Println(logical_plan.Format(got, 0))

	// Output:
	// Selection: age < 30
	//   Selection: age > 20
	//     Scan: users; projection=[]
	//
	// ---After optimization---
	//
	// Selection: age < 30 AND age > 20
	//   Scan: users; selection=age < 30, age > 20; projection=[]
}
