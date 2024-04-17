package demo

import (
	"context"
	"fmt"
	"github.com/kwilteam/kwil-db/internal/engine/cost/internal/testkit"

	"github.com/kwilteam/kwil-db/internal/engine/cost/query_planner"
	sqlparser "github.com/kwilteam/kwil-db/parse/sql"
)

func ExampleDemo() {
	// enter engine
	rawSql := "SELECT state, username FROM users WHERE age = 20"
	stmt, err := sqlparser.Parse(rawSql)
	if err != nil {
		panic(err)
	}

	// load into engine
	catalog := testkit.InitMockCatalog()

	planner := query_planner.NewPlanner(catalog)
	plan := planner.ToPlan(stmt)

	//opt := optimizer.NewOptimizer()
	//plan := opt.Optimize(plan)

	ctx := NewExecutionContext()
	res := ctx.Execute(context.TODO(), plan)
	fmt.Println(res.ToCsv())

	// Output:
	// state,username
	// CA,Adam
}
