package planner3

/*
	This file handles enforcement of aggregation rules in the logical plan.
	Our aggregations rules are as follows:
		- If a SELECT uses an aggregate function and does not have a GROUP BY, then all return columns must be aggregated.
		- If a SELECT uses an aggregate function and has a GROUP BY, then all return columns must be either aggregated or in the GROUP BY.
*/

// checkAggregationsFromGroupBy takes expressions from both a GROUP BY clause and a second clause, and validates
// that all columns referenced in the second clause are either in the GROUP BY clause or are captured in an aggregate function.
// it returns all values passed in the "exprs" parameter that are aggregate functions.
// It only applies these validations on the top-level SELECT; e.g., for any nested selects / subqueries, the caller
// is responsible for calling this function again.
func checkAggregationsFromGroupBy(groupBy []LogicalExpr, exprs []LogicalExpr, ctx *SchemaContext) (aggregateFuncs []*AggregateFunctionCall, err error) {
	visitor := &columnVisitor{ctx: ctx}

}

// usedColumn represents a column that is used in a logical plan.
// It will be fully qualified with the relation it belongs to.
// It gives information as to whether a column was used within
// an aggregate or not.
type usedColumn struct {
	*Column
	// If the column is referenced inside of an aggregate function,
	// (e.g. sum(col_name)), then Aggregated will be true.
	Aggregated bool
}

// deduplicateColumns removes duplicate columns from a list of columns.
// Duplicate columns are identified by having identical parent and name.
func deduplicateColumns(cols []*usedColumn) []*usedColumn {
	seen := make(map[string]struct{})
	var deduped []*usedColumn

	for _, col := range cols {
		key := col.Parent + "." + col.Name
		if _, ok := seen[key]; ok {
			continue
		}

		seen[key] = struct{}{}
		deduped = append(deduped, col)
	}

	return deduped
}

// columnVisitor gets the columns that are used in a logical plan.
// It contains context for the relations being operated on, which is used
// to determine the columns used.
type columnVisitor struct {
	// ctx is the schema context for the logical plan.
	ctx *SchemaContext
	// err is set if an error occurs.
	// The visitor will set this field and panic
	// if an error occurs.
	err error
}

// handleErr sets the error field and panics if an error
// is not nil.
func (c *columnVisitor) handleErr(err error) {
	if err == nil {
		return
	}
	c.err = err
	panic(err)
}

/*
	####################
	# Logical Plan Ops #
	####################

	Since the purpose of implementing Logical Plan ops in this visitor
	is to detect for subquery correlation, we visit the logical plans to see
	if a query is correlated, and if so, to see if it is violating rules on
	aggregation. The caller (which will always be a subquery) will then search
	the used columns to determine if the subquery is correlated.
*/

func (c *columnVisitor) VisitNoop(p0 *Noop) any {
	return &exprColumnResult{}
}

func (c *columnVisitor) VisitTableScan(p0 *TableScanSource) any {
	// scanning a table will never reference any columns from
	// an outer context, so we can safely return nil here.
	return &exprColumnResult{}
}

func (c *columnVisitor) VisitProcedureScan(p0 *ProcedureScanSource) any {
	return projectMany2(c, append(p0.Args, p0.ContextualArgs...)...)
}

func (c *columnVisitor) VisitScanAlias(p0 *Scan) any {
	return p0.Child.Accept(c)
}

func (c *columnVisitor) VisitProject(p0 *Project) any {
	res := p0.Child.Accept(c).(*exprColumnResult)

	oldCtx := c.ctx
	ctx2 := c.ctx.Join(p0.Relation(c.ctx))
	c.ctx = ctx2

	res2 := projectMany2(c, p0.Expressions...)

	res.add(res2)

	c.ctx = oldCtx
	return res
}

func (c *columnVisitor) VisitFilter(p0 *Filter) any {
	panic("TODO: Implement")
}

func (c *columnVisitor) VisitJoin(p0 *Join) any {
	panic("TODO: Implement")
}

func (c *columnVisitor) VisitSort(p0 *Sort) any {
	panic("TODO: Implement")
}

func (c *columnVisitor) VisitLimit(p0 *Limit) any {
	panic("TODO: Implement")
}

func (c *columnVisitor) VisitDistinct(p0 *Distinct) any {
	panic("TODO: Implement")
}

func (c *columnVisitor) VisitSetOperation(p0 *SetOperation) any {
	panic("TODO: Implement")
}

func (c *columnVisitor) VisitAggregate(p0 *Aggregate) any {
	panic("TODO: Implement")
}

/*
	#######################
	# Logical Expressions #
	#######################
*/

// exprColumnResult is returned by all logical expression in the
// column visitor.
type exprColumnResult struct {
	usedCols           []*usedColumn
	usedAggregateTerms []*AggregateFunctionCall
}

// add adds another exprColumnResult to this one.
func (e *exprColumnResult) add(other *exprColumnResult) {
	e.usedCols = append(e.usedCols, other.usedCols...)
	e.usedAggregateTerms = append(e.usedAggregateTerms, other.usedAggregateTerms...)
}

func (c *columnVisitor) VisitLiteral(p0 *Literal) any {
	return &exprColumnResult{}
}

func (c *columnVisitor) VisitVariable(p0 *Variable) any {
	return &exprColumnResult{}
}

func (c *columnVisitor) VisitColumnRef(p0 *ColumnRef) any {
	col, err := c.ctx.OuterRelation.Search(p0.Parent, p0.ColumnName)
	c.handleErr(err)

	return &exprColumnResult{
		usedCols: []*usedColumn{{
			Column: col,
		}},
	}
}

func (c *columnVisitor) VisitAggregateFunctionCall(p0 *AggregateFunctionCall) any {
	res := projectMany2(c, p0.Args...)

	// set all columns as aggregated
	for _, col := range res.usedCols {
		col.Aggregated = true
	}

	return res
}

func (c *columnVisitor) VisitFunctionCall(p0 *FunctionCall) any {
	return projectMany2(c, p0.Args...)
}

func (c *columnVisitor) VisitArithmeticOp(p0 *ArithmeticOp) any {
	return projectMany2(c, p0.Left, p0.Right)
}

func (c *columnVisitor) VisitComparisonOp(p0 *ComparisonOp) any {
	return projectMany2(c, p0.Left, p0.Right)
}

func (c *columnVisitor) VisitLogicalOp(p0 *LogicalOp) any {
	return projectMany2(c, p0.Left, p0.Right)
}

func (c *columnVisitor) VisitUnaryOp(p0 *UnaryOp) any {
	return p0.Expr.Accept(c)
}

func (c *columnVisitor) VisitTypeCast(p0 *TypeCast) any {
	return p0.Expr.Accept(c)
}

func (c *columnVisitor) VisitAliasExpr(p0 *AliasExpr) any {
	return p0.Expr.Accept(c)
}

func (c *columnVisitor) VisitArrayAccess(p0 *ArrayAccess) any {
	return projectMany2(c, p0.Array, p0.Index)
}

func (c *columnVisitor) VisitArrayConstructor(p0 *ArrayConstructor) any {
	return projectMany2(c, p0.Elements...)
}

func (c *columnVisitor) VisitFieldAccess(p0 *FieldAccess) any {
	return p0.Object.Accept(c)
}

// we don't care about what a subquery does (e.g. it can do anything it wants)
// UNLESS it is correlated to our current scope.
func (c *columnVisitor) VisitSubquery(p0 *Subquery) any {
	cols := p0.Query.Accept(c).([]*usedColumn)

	return &exprColumnResult{
		usedCols: cols,
	}
}

// projectMany is a helper function that projects multiple expressions and combines the results.
// TODO: rename once I delete the old one
func projectMany2[T Accepter](c *columnVisitor, accs ...T) *exprColumnResult {
	// I use a generic here to allow using a spread operator with an interface where
	// the slices themselves are of a concrete type.

	var columns []*usedColumn
	var exprsToProject []*AggregateFunctionCall
	for _, acc := range accs {
		res := acc.Accept(c).(*exprColumnResult)

		columns = append(columns, res.usedCols...)
		exprsToProject = append(exprsToProject, res.usedAggregateTerms...)
	}

	return &exprColumnResult{
		usedCols:           columns,
		usedAggregateTerms: exprsToProject,
	}
}

// flattenRelToUsed returns a list of used columns from a relation.
func flattenRelToUsed(rel *Relation) []*usedColumn {
	var cols []*usedColumn
	for _, col := range rel.Columns {
		cols = append(cols, &usedColumn{Column: col})
	}
	return cols
}
