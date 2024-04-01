package demo

import (
	"context"
	"fmt"
	dt "github.com/kwilteam/kwil-db/internal/engine/cost/datatypes"
	"github.com/kwilteam/kwil-db/internal/engine/cost/logical_plan"
)

var stubTable = &dt.TableRef{Table: "users"}

func Example_ExecutionContext_execute() {
	ctx := NewExecutionContext()
	df := ctx.Csv("users", "../testdata/users.csv").
		Filter(logical_plan.Eq(logical_plan.Column(stubTable, "age"),
			logical_plan.LiteralNumeric(20))).
		Project(logical_plan.Column(stubTable, "state"),
			logical_plan.Column(stubTable, "username"),
		)

	res := ctx.Execute(context.TODO(), df.LogicalPlan())
	fmt.Println(res.ToCsv())

	// Output:
	// state,username
	// CA,Adam
}
