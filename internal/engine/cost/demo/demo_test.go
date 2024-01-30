package demo

import (
	"context"
	"fmt"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/internal/engine/cost/internal/testkit"
	"github.com/kwilteam/kwil-db/internal/engine/cost/query_planner"
	"github.com/kwilteam/kwil-db/parse"
)

// ExampleDemo demonstrates how a SQL is parsed, planned and executed, served as
// a simple mental model to understand the flow of the engine.
//
// For cost estimation, execution will be carried out by internal/engine, not by
// the virtual plan in this pkg.
func ExampleDemo() {
	// NOTE!!!: this will fail, as 'parser' add sorting to every AST,
	// but virutal plan does not support sorting yet.
	//
	// See Example_ExecutionContext_execute instead, although it only shows how
	// a logical plan is executed.

	// enter engine
	rawSql := "SELECT state, username FROM users WHERE age = 20"

	pr, err := parse.ParseSQL(rawSql, &types.Schema{
		Name:   "",
		Tables: []*types.Table{testkit.MockUsersSchemaTable},
	})

	if err != nil {
		panic(err)
	}

	if pr.ParseErrs.Err() != nil {
		panic(pr.ParseErrs.Err())
	}

	// load into engine
	catalog := testkit.InitMockCatalog()

	planner := query_planner.NewPlanner(catalog)
	plan := planner.ToPlan(pr.AST)

	//opt := optimizer.NewOptimizer()
	//plan := opt.Optimize(plan)

	ctx := NewExecutionContext()
	res := ctx.Execute(context.TODO(), plan)
	fmt.Println(res.ToCsv())

	// Output:
	// state,username
	// CA,Adam
}
