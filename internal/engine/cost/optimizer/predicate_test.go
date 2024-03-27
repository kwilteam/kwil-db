package optimizer

import (
	"fmt"
	"github.com/kwilteam/kwil-db/internal/engine/cost/datasource"
	dt "github.com/kwilteam/kwil-db/internal/engine/cost/datatypes"
	"github.com/kwilteam/kwil-db/internal/engine/cost/logical_plan"
)

var stubDS, _ = datasource.NewCSVDataSource("../testdata/users.csv")
var stubTable = &dt.TableRef{Table: "users"}

func ExamplePredicateRule_optimize_pushDown() {
	aop := logical_plan.NewDataFrame(logical_plan.Scan(stubTable, stubDS, nil))
	plan := aop.
		Filter(logical_plan.Eq(logical_plan.ColumnUnqualified("age"),
			logical_plan.LiteralNumeric(20))).
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
	//   Scan: users
	//
	// ---After optimization---
	//
	// Filter: age = 20
	//   Scan: users; filter=[age = 20]
}

func ExamplePredicateRule_optimize_pushDown_with_nested_filter() {
	aop := logical_plan.NewDataFrame(logical_plan.Scan(stubTable, stubDS, nil))
	plan := aop.
		Filter(logical_plan.Gt(logical_plan.ColumnUnqualified("age"),
			logical_plan.LiteralNumeric(20))).
		Filter(logical_plan.Lt(logical_plan.ColumnUnqualified("age"),
			logical_plan.LiteralNumeric(30))).
		LogicalPlan()

	fmt.Println(logical_plan.Format(plan, 0))

	r := &PredicateRule{}
	got := r.Optimize(plan)

	fmt.Printf("---After optimization---\n\n")
	fmt.Println(logical_plan.Format(got, 0))

	// Output:
	// Filter: age < 30
	//   Filter: age > 20
	//     Scan: users
	//
	// ---After optimization---
	//
	// Filter: age < 30 AND age > 20
	//   Scan: users; filter=[age < 30, age > 20]
}
