package demo

import (
	"context"
	"fmt"

	dt "github.com/kwilteam/kwil-db/internal/engine/cost/datatypes"
	"github.com/kwilteam/kwil-db/internal/engine/cost/logical_plan"
)

var stubTable = &dt.TableRef{Table: "users"}

// Example_ExecutionContext_execute demonstrates how to use the ExecutionContext
// to execute a logical plan.
func Example_ExecutionContext_execute() {
	ctx := NewExecutionContext()
	df := logical_plan.DataFrameAPI(ctx.Csv("users", "../testdata/users.csv"))
	// SELECT state, username FROM users WHERE age = 20;
	df = df.Filter(logical_plan.Eq(logical_plan.Column(stubTable, "age"),
		logical_plan.LiteralNumeric(20)))
	df = df.Project(logical_plan.Column(stubTable, "state"),
		logical_plan.Column(stubTable, "username"),
	)

	res := ctx.Execute(context.TODO(), df.LogicalPlan())
	fmt.Println(res.ToCsv())

	// Output:
	// state,username
	// CA,Adam
}
