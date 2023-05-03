package tree

type Function interface{}

type SimpleFunctionInvocation struct {
	FunctionName string
	Arguments    []*Expression
	All          bool
}

type AggregateFunctionInvocation struct {
	FunctionName string
	Distinct     bool
	Arguments    []*Expression
	All          bool
	Filter       *FilterClause
}

// not yet supported
// type WindowFunctionInvocation struct {}

type FilterClause struct {
	*WhereClause
}
