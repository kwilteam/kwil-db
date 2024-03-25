package optimizer

import (
	"fmt"
	"github.com/kwilteam/kwil-db/internal/engine/cost/datasource"
	dt "github.com/kwilteam/kwil-db/internal/engine/cost/datatypes"
	"github.com/kwilteam/kwil-db/internal/engine/cost/logical_plan"
)

func ExamplePredicateRule_optimize_pushDown() {
	stubUserData, _ := datasource.NewCSVDataSource("../testdata/users.csv")

	//ds := datasource.NewMemDataSource(nil, nil)
	tUser := &dt.TableRef{Table: "users"}
	aop := logical_plan.NewDataFrame(logical_plan.Scan(tUser, stubUserData, nil))
	plan := aop.
		Filter(logical_plan.Eq(logical_plan.ColumnUnqualified("age"),
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
	// Filter: age = 20
	//   Scan: users; projection=[]
	//
	// ---After optimization---
	//
	// Filter: age = 20
	//   Scan: users; filter=[age = 20]; projection=[]
}

func ExamplePredicateRule_optimize_pushDown_with_nested_selection() {
	stubUserData, _ := datasource.NewCSVDataSource("../testdata/users.csv")
	tUser := &dt.TableRef{Table: "users"}
	aop := logical_plan.NewDataFrame(logical_plan.Scan(tUser, stubUserData, nil))
	plan := aop.
		Filter(logical_plan.Gt(logical_plan.ColumnUnqualified("age"),
			logical_plan.LiteralInt(20))).
		Filter(logical_plan.Lt(logical_plan.ColumnUnqualified("age"),
			logical_plan.LiteralInt(30))).
		LogicalPlan()

	fmt.Println(logical_plan.Format(plan, 0))

	r := &PredicateRule{}
	got := r.Optimize(plan)

	fmt.Printf("---After optimization---\n\n")
	fmt.Println(logical_plan.Format(got, 0))

	// Output:
	// Filter: age < 30
	//   Filter: age > 20
	//     Scan: users; projection=[]
	//
	// ---After optimization---
	//
	// Filter: age < 30 AND age > 20
	//   Scan: users; selection=age < 30, age > 20; projection=[]
}
