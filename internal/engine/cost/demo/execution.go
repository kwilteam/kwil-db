package demo

import (
	"context"

	"github.com/kwilteam/kwil-db/internal/engine/cost/datasource"
	dt "github.com/kwilteam/kwil-db/internal/engine/cost/datatypes"
	"github.com/kwilteam/kwil-db/internal/engine/cost/logical_plan"
	"github.com/kwilteam/kwil-db/internal/engine/cost/optimizer"
)

type ExecutionContext struct {
	tables map[string]logical_plan.DataFrameAPI
}

func NewExecutionContext() *ExecutionContext {
	return &ExecutionContext{}
}

func (e *ExecutionContext) Csv(table string, filepath string) *logical_plan.DataFrame {
	dataSrc, err := datasource.NewCSVDataSource(filepath)
	if err != nil {
		panic(err)
	}

	return logical_plan.NewDataFrame(
		logical_plan.ScanPlan(&dt.TableRef{Table: table}, dataSrc, nil))
}

func (e *ExecutionContext) registerBuilder(name string, builder *logical_plan.DataFrame) {
	e.tables[name] = builder
}

func (e *ExecutionContext) registerDataSource(name string, ds datasource.DataSource) {
	e.tables[name] = logical_plan.NewDataFrame(
		logical_plan.ScanPlan(&dt.TableRef{Table: name}, ds, nil))
}

func (e *ExecutionContext) registerCsv(name string, filepath string) {
	e.tables[name] = e.Csv(name, filepath)
}

func (e *ExecutionContext) Execute(ctx context.Context, plan logical_plan.LogicalPlan) *datasource.Result {
	return execute(ctx, plan)
}

//func (e *ExecutionContext) estimate(plan logical_plan.LogicalPlan) int64 {
//	return estimate(plan)
//}

func execute(ctx context.Context, plan logical_plan.LogicalPlan) *datasource.Result {
	//
	//fmt.Printf("---Original plan---\n\n")
	//fmt.Println(logical_plan.Format(plan, 0))
	//
	r := &optimizer.ProjectionRule{}
	optPlan := r.Transform(plan)
	//
	//fmt.Printf("---After optimization---\n\n")
	//fmt.Println(logical_plan.Format(optPlan, 0))
	//
	qp := optimizer.NewPlanner()
	vp := qp.ToPlan(optPlan)
	//
	//fmt.Printf("---Got virtual plan---\n\n")
	//fmt.Println(Format(vp, 0))
	//
	return vp.Execute(ctx)
}

func estimate(plan logical_plan.LogicalPlan) int64 {
	r := &optimizer.ProjectionRule{}
	optPlan := r.Transform(plan)
	qp := optimizer.NewPlanner()
	vp := qp.ToPlan(optPlan)
	return vp.Cost()
}
