package cost_2nd

import "github.com/kwilteam/kwil-db/parse/sql/tree"

type DataFrame interface {
	// LogicalPlan returns the logical plan of the DataFrame.
	LogicalPlan() LogicalPlan

	// Project applies a projection to the DataFrame.
	Project(exprs []*tree.ResultColumnExpression) DataFrame

	// Filter applies a filter to the DataFrame.
	Filter(expr tree.Expression) DataFrame

	// Aggregate applies an aggregation to the DataFrame.
	Aggregate(groupBy []tree.Expression, aggrExpr []*tree.ExpressionFunction) DataFrame

	// Schema returns the schema of the data that will be produced by this DataFrame.
	Schema() *schema
}

type DataFrameImpl struct {
	plan LogicalPlan
}

func (df *DataFrameImpl) Project(exprs []*tree.ResultColumnExpression) DataFrame {
	return &DataFrameImpl{NewLogicalProjection(df.plan, exprs)}
}

func (df *DataFrameImpl) Filter(expr tree.Expression) DataFrame {
	return &DataFrameImpl{NewLogicalFilter(df.plan, expr)}
}

func (df *DataFrameImpl) Aggregate(groupBy []tree.Expression, aggrExpr []*tree.ExpressionFunction) DataFrame {
	return &DataFrameImpl{NewLogicalAggregate(df.plan, groupBy, aggrExpr)}
}

func (df *DataFrameImpl) Schema() *schema {
	return df.plan.Schema()
}

func (df *DataFrameImpl) LogicalPlan() LogicalPlan {
	return df.plan
}
