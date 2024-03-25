package virtual_plan

import (
	"fmt"
	dt "github.com/kwilteam/kwil-db/internal/engine/cost/datatypes"

	"github.com/kwilteam/kwil-db/internal/engine/cost/logical_plan"
)

var stubTable = &dt.TableRef{Table: "users"}

func Example_ExecutionContext_execute() {
	ctx := NewExecutionContext()
	df := ctx.csv("users", "../testdata/users.csv").
		Filter(logical_plan.Eq(logical_plan.Column(stubTable, "age"),
			logical_plan.LiteralInt(20))).
		Project(logical_plan.Column(stubTable, "state"),
			logical_plan.Column(stubTable, "username"),
		)

	res := ctx.execute(df)
	fmt.Println(res.ToCsv())
	// Output:
	// state,username
	// CA,Adam
}
