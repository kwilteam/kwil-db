package demo

import (
	"context"
	"fmt"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/internal/engine/cost/costmodel"
	"github.com/kwilteam/kwil-db/internal/engine/cost/internal/testkit"
	"github.com/kwilteam/kwil-db/internal/engine/cost/logical_plan"
	"github.com/kwilteam/kwil-db/internal/engine/cost/optimizer"
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
	// but virtual plan does not support sorting yet. (why? projection or push down issue?)
	//
	// See Example_ExecutionContext_execute instead, although it only shows how
	// a logical plan is executed.

	ctx := context.Background()

	// enter engine
	rawSql := "SELECT state, username FROM users WHERE age = 20"

	ast, err := parse.ParseSQLWithoutValidation(rawSql, &types.Schema{
		Name:   "",
		Tables: []*types.Table{testkit.MockUsersSchemaTable},
	})

	if err != nil {
		panic(err)
	}

	// load into engine
	catalog := testkit.InitMockCatalog()

	planner := query_planner.NewPlanner(catalog)
	plan := planner.ToPlan(ast)

	//opt := optimizer.NewOptimizer()
	//plan := opt.Optimize(plan)

	// ctx := NewExecutionContext()
	// res := ctx.Execute(context.TODO(), plan) // field not found "id"

	fmt.Printf("---Original plan---\n\n")
	fmt.Println(logical_plan.Format(plan, 0))
	r := &optimizer.ProjectionRule{}
	optPlan := r.Transform(plan)

	fmt.Printf("---After optimization---\n\n")
	fmt.Println(logical_plan.Format(optPlan, 0))

	relExp := costmodel.BuildRelExpr(optPlan)
	cost := costmodel.EstimateCost(relExp)
	fmt.Println(cost)

	qp := optimizer.NewPlanner() // virtual planner
	vp := qp.ToPlan(optPlan)

	//fmt.Printf("---Got virtual plan---\n\n")
	//fmt.Println(Format(vp, 0))
	res := vp.Execute(ctx)
	fmt.Println(res.ToCsv())

	cost = vp.Cost() // VProjectionOp -> VFilterOp -> VSeqScanOp <= projectionCost + (filterEqCost + (seqScanRowCost * RowCount))
	fmt.Println(cost)

	// TODO (@jchappelow): RowCount in statistics is still zero, set it

	// Output:
	// state,username
	// CA,Adam
	//
	// 30

}
