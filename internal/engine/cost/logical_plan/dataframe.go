package logical_plan

import "github.com/kwilteam/kwil-db/internal/engine/cost/datatypes"

type DataFrameAPI interface {
	// Project applies a projection
	Project(expr ...LogicalExpr) DataFrameAPI

	// Filter applies a filter
	Filter(expr LogicalExpr) DataFrameAPI

	// Aggregate applies an aggregation
	Aggregate(groupBy []LogicalExpr, aggregateExpr []LogicalExpr) DataFrameAPI

	// Schema returns the schema of the data that will be produced by this DataFrameAPI.
	Schema() *datatypes.Schema

	// LogicalPlan returns the logical plan
	LogicalPlan() LogicalPlan
}

// DataFrame adapts a LogicalPlan with methods to project, filter, and apply
// aggregation with a chained syntax.
//
// NOTE: You can use the package level functions directly instead: Projection,
// Filter, Aggregate.
type DataFrame struct {
	plan LogicalPlan
}

func (df *DataFrame) Project(exprs ...LogicalExpr) DataFrameAPI {
	return &DataFrame{Projection(df.plan, exprs...)}
}

func (df *DataFrame) Filter(expr LogicalExpr) DataFrameAPI {
	return &DataFrame{Filter(df.plan, expr)}
}

func (df *DataFrame) Aggregate(groupBy []LogicalExpr, aggregateExpr []LogicalExpr) DataFrameAPI {
	return &DataFrame{Aggregate(df.plan, groupBy, aggregateExpr)}
}

func (df *DataFrame) Schema() *datatypes.Schema {
	return df.plan.Schema()
}

func (df *DataFrame) LogicalPlan() LogicalPlan {
	return df.plan
}

func NewDataFrame(plan LogicalPlan) *DataFrame {
	return &DataFrame{plan: plan}
}
