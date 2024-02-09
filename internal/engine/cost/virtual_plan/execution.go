package virtual_plan

import (
	"github.com/kwilteam/kwil-db/internal/engine/cost/datasource"
	"github.com/kwilteam/kwil-db/internal/engine/cost/logical_plan"
	"github.com/kwilteam/kwil-db/internal/engine/cost/optimizer"
)

type ExecutionContext struct {
	tables map[string]logical_plan.DataFrameAPI
}

func NewExecutionContext() *ExecutionContext {
	return &ExecutionContext{}
}

func (e *ExecutionContext) csv(table string, filepath string) *logical_plan.DataFrame {
	datasource, err := datasource.NewCSVDataSource(filepath)
	if err != nil {
		panic(err)
	}
	return logical_plan.NewDataFrame(logical_plan.Scan(table, datasource, nil))
}

func (e *ExecutionContext) registerBuilder(name string, builder *logical_plan.DataFrame) {
	e.tables[name] = builder
}

func (e *ExecutionContext) registerDataSource(name string, ds datasource.DataSource) {
	e.tables[name] = logical_plan.NewDataFrame(logical_plan.Scan(name, ds, nil))
}

func (e *ExecutionContext) registerCsv(name string, filepath string) {
	e.tables[name] = e.csv(name, filepath)
}

func (e *ExecutionContext) execute(df logical_plan.DataFrameAPI) *datasource.Result {
	return execute(df.LogicalPlan())
}

func execute(plan logical_plan.LogicalPlan) *datasource.Result {
	//
	//fmt.Printf("---Original plan---\n\n")
	//fmt.Println(logical_plan.Format(plan, 0))
	//
	r := &optimizer.ProjectionRule{}
	optPlan := r.Optimize(plan)
	//
	//fmt.Printf("---After optimization---\n\n")
	//fmt.Println(logical_plan.Format(optPlan, 0))
	//
	qp := NewQueryPlanner()
	vp := qp.CreateVirtualPlan(optPlan)
	//
	//fmt.Printf("---Got virtual plan---\n\n")
	//fmt.Println(Format(vp, 0))
	//
	return vp.Execute()
}
