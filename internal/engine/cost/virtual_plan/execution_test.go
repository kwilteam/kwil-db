package virtual_plan

import (
	"fmt"

	"github.com/kwilteam/kwil-db/internal/engine/cost/logical_plan"
)

func Example_ExecutionContext_execute() {
	ctx := NewExecutionContext()
	df := ctx.csv("users", "../testdata/users.csv").
		Filter(logical_plan.Eq(logical_plan.Column("", "age"),
			logical_plan.LiteralInt(20))).
		Project(logical_plan.Column("", "state"),
			logical_plan.Column("", "username"),
		)

	res := ctx.execute(df)
	fmt.Println(res.ToCsv())
	// Output:
	// state,username
	// CA,Adam
}
