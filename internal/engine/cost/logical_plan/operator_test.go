package logical_plan_test

import (
	"fmt"

	ds "github.com/kwilteam/kwil-db/internal/engine/cost/datasource"
	dt "github.com/kwilteam/kwil-db/internal/engine/cost/datatypes"
	lp "github.com/kwilteam/kwil-db/internal/engine/cost/logical_plan"
)

var stubDS, _ = ds.NewCSVDataSource("../testdata/users.csv")
var stubTable = &dt.TableRef{Table: "users"}

func ExampleScanOp_String_no_filter() {
	op := lp.Scan(stubTable, stubDS, nil, "username", "age")
	fmt.Println(op.String())
	// Output:
	// Scan: users; projection=[username, age]
}

func ExampleScanOp_String_with_filter() {
	op := lp.Scan(stubTable, stubDS,
		[]lp.LogicalExpr{
			lp.Gt(lp.ColumnUnqualified("age"),
				lp.LiteralInt(20)),
			lp.Lt(lp.ColumnUnqualified("age"),
				lp.LiteralInt(30)),
		}, "username", "age")
	fmt.Println(op.String())
	// Output:
	// Scan: users; filter=[age > 20, age < 30]; projection=[username, age]
}

func ExampleProjectionOp_String() {
	op := lp.Projection(
		nil,
		lp.ColumnUnqualified("username"),
		lp.ColumnUnqualified("age"))
	fmt.Println(op.String())
	// Output:
	// Projection: username, age
}

func ExampleFilterOp_String() {
	op := lp.Filter(nil,
		lp.Eq(
			lp.ColumnUnqualified("age"),
			lp.LiteralInt(20)))
	fmt.Println(op.String())
	// Output:
	// Filter: age = 20
}

func ExampleAggregateOp_String() {
	op := lp.Aggregate(
		lp.Scan(stubTable, stubDS, nil),
		[]lp.LogicalExpr{lp.ColumnUnqualified("state")},
		[]lp.LogicalExpr{lp.Count(lp.ColumnUnqualified("username"))})
	fmt.Println(op.String())
	// Output:
	// Aggregate: groupBy=[state]; aggr=[COUNT(username)]
}

func ExampleAggregateOp_String_without_groupby() {
	op := lp.Aggregate(
		lp.Scan(stubTable, stubDS, nil),
		nil,
		[]lp.LogicalExpr{lp.Count(lp.ColumnUnqualified("username"))})
	fmt.Println(op.String())
	// Output:
	// Aggregate: ; aggr=[COUNT(username)]
}

func ExampleLimitOp_String_without_skip() {
	op := lp.Limit(nil, 0, 10)
	fmt.Println(op.String())
	// Output:
	// Limit: skip=0, fetch=10
}

func ExampleLimitOp_String_with_skip() {
	op := lp.Limit(nil, 5, 10)
	fmt.Println(op.String())
	// Output:
	// Limit: skip=5, fetch=10
}

func ExampleSortOp_String() {
	op := lp.Sort(nil,
		[]lp.LogicalExpr{
			lp.SortExpr(lp.ColumnUnqualified("state"), false, true),
			lp.SortExpr(lp.ColumnUnqualified("age"), true, false),
		},
	)
	fmt.Println(op.String())
	// Output:
	// Sort: state DESC NULLS FIRST, age ASC NULLS LAST
}

func ExampleLogicalPlan_Projection() {
	plan := lp.Scan(stubTable, stubDS, nil)
	plan = lp.Projection(plan,
		lp.ColumnUnqualified("username"),
		lp.ColumnUnqualified("age"))
	fmt.Println(lp.Format(plan, 0))
	// Output:
	// Projection: username, age
	//   Scan: users
}

func ExampleLogicalPlan_DataFrame() {
	aop := lp.NewDataFrame(lp.Scan(stubTable, stubDS, nil))
	plan := aop.Filter(lp.Eq(lp.ColumnUnqualified("age"), lp.LiteralInt(20))).
		Aggregate([]lp.LogicalExpr{lp.ColumnUnqualified("state")},
			[]lp.LogicalExpr{lp.Count(lp.ColumnUnqualified("username"))}).
		// the alias for aggregate result is bit weird
		Project(lp.ColumnUnqualified("state"), lp.Alias(lp.Count(lp.ColumnUnqualified("username")), "num")).
		LogicalPlan()

	fmt.Println(lp.Format(plan, 0))
	// Output:
	// Projection: state, COUNT(username) AS num
	//   Aggregate: groupBy=[state]; aggr=[COUNT(username)]
	//     Filter: age = 20
	//       Scan: users
}
