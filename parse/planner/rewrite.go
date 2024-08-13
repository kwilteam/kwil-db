package planner

import (
	"fmt"
)

// RewriteConfig is a configuration for the rewriter.
type RewriteConfig struct {
	// ExprCallback is the function that will be called on each expression
	ExprCallback func(LogicalExpr) (LogicalExpr, error)
	// PlanCallback is the function that will be called on each plan
	PlanCallback func(LogicalPlan) (LogicalPlan, error)
	// ScanSourceCallback is the function that will be called on each scan source
	ScanSourceCallback func(ScanSource) (ScanSource, error)
	// If true, the callback will be called before visiting children,
	// and any expression acting on a plan will be called before visiting the plan.
	// If false, the callback will be called after visiting children,
	// and any expression acting on a plan will be called after visiting the plan.
	// If order doesn't matter, it is recommended to set this to false, since
	// setting it to true can lead to infinite loops.
	CallbackBeforeVisit bool
	// PostOrderVisit determines the order in which fields are visited.
	// If visitng in post order, then children are visited first, then the parent.
	// For example, for a Project node, if PostOrderVisit is true, then the child
	// is visited first, then the expressions.
	// If PostOrderVisit is false, then the expressions are visited first, then the child.
	PostOrderVisit bool
}

// Rewrite rewrites a logical plan using the given configuration.
// It returns the rewritten plan, but it also modifies the original plan in place.
// The returned plan should always be used.
func Rewrite(node LogicalNode, cfg *RewriteConfig) (lp LogicalNode, err error) {
	defer func() {
		if r := recover(); r != nil {
			err2, ok := r.(error)
			if !ok {
				err = fmt.Errorf("%v", r)
			} else {
				err = err2
			}
		}
	}()

	v := &rewriteVisitor{
		exprCallback:       cfg.ExprCallback,
		planCallback:       cfg.PlanCallback,
		scanSourceCallback: cfg.ScanSourceCallback,
		callbackFuncFirst:  cfg.CallbackBeforeVisit,
		postOrderVisit:     cfg.PostOrderVisit,
	}
	if v.exprCallback == nil {
		v.exprCallback = func(e LogicalExpr) (LogicalExpr, error) {
			return e, nil
		}
	}
	if v.planCallback == nil {
		v.planCallback = func(p LogicalPlan) (LogicalPlan, error) {
			return p, nil
		}
	}
	if v.scanSourceCallback == nil {
		v.scanSourceCallback = func(s ScanSource) (ScanSource, error) {
			return s, nil
		}
	}

	return node.Accept(v).(LogicalNode), nil
}

// rewriteVisitor is a visitor that can be used to rewrite a logical plan.
// For all visits, functions can be given to determine the order in which
// fields are visited in. They should be passed in post-order (e.g. in the same
// order as they would need to be visited to evaluate the relation).
type rewriteVisitor struct {
	// exprCallback is the function that will be called on each expression
	exprCallback func(LogicalExpr) (LogicalExpr, error)
	// planCallback is the function that will be called on each plan
	planCallback func(LogicalPlan) (LogicalPlan, error)
	// scanSourceCallback is the function that will be called on each scan source
	scanSourceCallback func(ScanSource) (ScanSource, error)
	// if true, the callback will be called before visiting children
	// if false, the callback will be called after visiting children
	callbackFuncFirst bool
	// if true, then fields are visited in post order
	// if false, then fields are visited in pre order
	postOrderVisit bool
}

func (r *rewriteVisitor) VisitTableScanSource(p0 *TableScanSource) any {
	return r.scanSource(p0, func() {})
}

func (r *rewriteVisitor) slice(v any) {
	if v == nil {
		return
	}
	switch v := v.(type) {
	case []LogicalPlan:
		for i := range v {
			v[i] = v[i].Accept(r).(LogicalPlan)
		}
	case []LogicalExpr:
		for i := range v {
			v[i] = v[i].Accept(r).(LogicalExpr)
		}
	}
}

func (r *rewriteVisitor) VisitProcedureScanSource(p0 *ProcedureScanSource) any {
	return r.scanSource(p0, func() {
		r.slice(p0.ContextualArgs)
		r.slice(p0.Args)
	})
}

func (r *rewriteVisitor) VisitSubquery(p0 *Subquery) any {
	return r.scanSource(p0, r.subqueryFuncs(p0)...)
}

// defining this separately since we use it in several places
func (r *rewriteVisitor) subqueryFuncs(p0 *Subquery) []func() {
	return []func(){func() {
		p0.Plan.Plan = p0.Plan.Plan.Accept(r).(LogicalPlan)
	}} // TODO: once we switch the Correlated column from column refs to exprs, we should visit the expressions here
}

func (r *rewriteVisitor) VisitEmptyScan(p0 *EmptyScan) any {
	return r.plan(p0, func() {})
}

func (r *rewriteVisitor) VisitScan(p0 *Scan) any {
	return r.plan(p0, func() {
		p0.Source = p0.Source.Accept(r).(ScanSource)
	})
}

func (r *rewriteVisitor) VisitProject(p0 *Project) any {
	return r.plan(p0,
		func() { p0.Child = p0.Child.Accept(r).(LogicalPlan) },
		func() { r.slice(p0.Expressions) },
	)
}

func (r *rewriteVisitor) VisitFilter(p0 *Filter) any {
	return r.plan(p0,
		func() { p0.Child = p0.Child.Accept(r).(LogicalPlan) },
		func() { p0.Condition = p0.Condition.Accept(r).(LogicalExpr) },
	)
}

func (r *rewriteVisitor) VisitJoin(p0 *Join) any {
	return r.plan(p0,
		func() { p0.Left = p0.Left.Accept(r).(LogicalPlan) },
		func() { p0.Right = p0.Right.Accept(r).(LogicalPlan) },
		func() { p0.Condition = p0.Condition.Accept(r).(LogicalExpr) },
	)
}

func (r *rewriteVisitor) VisitSort(p0 *Sort) any {
	return r.plan(p0,
		func() { p0.Child = p0.Child.Accept(r).(LogicalPlan) },
		func() {
			for _, sort := range p0.SortExpressions {
				sort.Expr = sort.Expr.Accept(r).(LogicalExpr)
			}
		},
	)
}

func (r *rewriteVisitor) VisitLimit(p0 *Limit) any {
	return r.plan(p0, func() {
		p0.Child = p0.Child.Accept(r).(LogicalPlan)
		p0.Limit = p0.Limit.Accept(r).(LogicalExpr)
		p0.Offset = p0.Offset.Accept(r).(LogicalExpr)
	})
}

func (r *rewriteVisitor) VisitDistinct(p0 *Distinct) any {
	return r.plan(p0,
		func() { p0.Child = p0.Child.Accept(r).(LogicalPlan) },
	)
}

func (r *rewriteVisitor) VisitSetOperation(p0 *SetOperation) any {
	return r.plan(p0,
		func() { p0.Left = p0.Left.Accept(r).(LogicalPlan) },
		func() { p0.Right = p0.Right.Accept(r).(LogicalPlan) },
	)
}

func (r *rewriteVisitor) VisitAggregate(p0 *Aggregate) any {
	return r.plan(p0,
		func() { p0.Child = p0.Child.Accept(r).(LogicalPlan) },
		func() { r.slice(p0.GroupingExpressions) },
		func() { r.slice(p0.AggregateExpressions) },
	)
}

func (r *rewriteVisitor) VisitSubplan(p0 *Subplan) any {
	return r.plan(p0,
		func() { p0.Plan = p0.Plan.Accept(r).(LogicalPlan) },
	)
}

func (r *rewriteVisitor) VisitLiteral(p0 *Literal) any {
	return r.expr(p0, func() {})
}

func (r *rewriteVisitor) VisitVariable(p0 *Variable) any {
	return r.expr(p0, func() {})
}

func (r *rewriteVisitor) VisitColumnRef(p0 *ColumnRef) any {
	return r.expr(p0, func() {})
}

func (r *rewriteVisitor) VisitAggregateFunctionCall(p0 *AggregateFunctionCall) any {
	return r.expr(p0,
		func() { r.slice(p0.Args) },
	)
}

func (r *rewriteVisitor) VisitScalarFunctionCall(p0 *ScalarFunctionCall) any {
	return r.expr(p0,
		func() { r.slice(p0.Args) },
	)
}

func (r *rewriteVisitor) VisitProcedureCall(p0 *ProcedureCall) any {
	return r.expr(p0,
		func() { r.slice(p0.ContextArgs) },
		func() { r.slice(p0.Args) },
	)
}

func (r *rewriteVisitor) VisitArithmeticOp(p0 *ArithmeticOp) any {
	return r.expr(p0,
		func() { p0.Left = p0.Left.Accept(r).(LogicalExpr) },
		func() { p0.Right = p0.Right.Accept(r).(LogicalExpr) },
	)
}

func (r *rewriteVisitor) VisitComparisonOp(p0 *ComparisonOp) any {
	return r.expr(p0,
		func() { p0.Left = p0.Left.Accept(r).(LogicalExpr) },
		func() { p0.Right = p0.Right.Accept(r).(LogicalExpr) },
	)
}

func (r *rewriteVisitor) VisitLogicalOp(p0 *LogicalOp) any {
	return r.expr(p0,
		func() { p0.Left = p0.Left.Accept(r).(LogicalExpr) },
		func() { p0.Right = p0.Right.Accept(r).(LogicalExpr) },
	)
}

func (r *rewriteVisitor) VisitUnaryOp(p0 *UnaryOp) any {
	return r.expr(p0,
		func() { p0.Expr = p0.Expr.Accept(r).(LogicalExpr) },
	)
}

func (r *rewriteVisitor) VisitTypeCast(p0 *TypeCast) any {
	return r.expr(p0,
		func() { p0.Expr = p0.Expr.Accept(r).(LogicalExpr) },
	)
}

func (r *rewriteVisitor) VisitAliasExpr(p0 *AliasExpr) any {
	return r.expr(p0,
		func() { p0.Expr = p0.Expr.Accept(r).(LogicalExpr) },
	)
}

func (r *rewriteVisitor) VisitArrayAccess(p0 *ArrayAccess) any {
	return r.expr(p0,
		func() { p0.Array = p0.Array.Accept(r).(LogicalExpr) },
		func() { p0.Index = p0.Index.Accept(r).(LogicalExpr) },
	)
}

func (r *rewriteVisitor) VisitArrayConstructor(p0 *ArrayConstructor) any {
	return r.expr(p0,
		func() { r.slice(p0.Elements) },
	)
}

func (r *rewriteVisitor) VisitFieldAccess(p0 *FieldAccess) any {
	return r.expr(p0,
		func() { p0.Object = p0.Object.Accept(r).(LogicalExpr) },
	)
}

func (r *rewriteVisitor) VisitSubqueryExpr(p0 *SubqueryExpr) any {
	return r.expr(p0,
		func() {},
	)
}

func (r *rewriteVisitor) VisitCollate(p0 *Collate) any {
	return r.expr(p0,
		func() { p0.Expr = p0.Expr.Accept(r).(LogicalExpr) },
	)
}

func (r *rewriteVisitor) VisitIsIn(p0 *IsIn) any {
	return r.expr(p0,
		func() { p0.Left = p0.Left.Accept(r).(LogicalExpr) },
		func() {
			if p0.Subquery != nil {
				for _, fn := range r.subqueryFuncs(p0.Subquery.Query) {
					fn()
				}
			} else {
				r.slice(p0.Expressions)
			}
		},
	)
}

func (r *rewriteVisitor) VisitCase(p0 *Case) any {
	return r.expr(p0,
		func() { p0.Value = p0.Value.Accept(r).(LogicalExpr) },
		func() {
			for _, whenThen := range p0.WhenClauses {
				whenThen[0] = whenThen[0].Accept(r).(LogicalExpr)
				whenThen[1] = whenThen[1].Accept(r).(LogicalExpr)
			}
		},
		func() { p0.Else = p0.Else.Accept(r).(LogicalExpr) },
	)
}

func (r *rewriteVisitor) VisitExprRef(p0 *ExprRef) any {
	return r.expr(p0,
		func() {
			p0.Identified.Expr = p0.Identified.Expr.Accept(r).(LogicalExpr)
		},
	)
}

func (r *rewriteVisitor) VisitIdentifiedExpr(p0 *IdentifiedExpr) any {
	return r.expr(p0,
		func() { p0.Expr = p0.Expr.Accept(r).(LogicalExpr) },
	)
}

func (r *rewriteVisitor) VisitReturn(p0 *Return) any {
	return r.plan(p0, func() {
		p0.Child = p0.Child.Accept(r).(LogicalPlan)
	})
}

func (r *rewriteVisitor) VisitCartesianProduct(p0 *CartesianProduct) any {
	return r.plan(p0,
		func() { p0.Left = p0.Left.Accept(r).(LogicalPlan) },
		func() { p0.Right = p0.Right.Accept(r).(LogicalPlan) },
	)
}

func (r *rewriteVisitor) VisitUpdate(p0 *Update) any {
	return r.plan(p0, func() {
		p0.Child = p0.Child.Accept(r).(LogicalPlan)
	})
}

func (r *rewriteVisitor) VisitDelete(p0 *Delete) any {
	return r.plan(p0, func() {
		p0.Child = p0.Child.Accept(r).(LogicalPlan)
	})
}

func (r *rewriteVisitor) VisitInsert(p0 *Insert) any {
	return r.plan(p0,
		func() {
			for _, row := range p0.Values {
				r.slice(row)
			}
		},
		func() {
			if p0.ConflictResolution != nil {
				doUpdate, ok := p0.ConflictResolution.(*ConflictUpdate)
				if ok {
					doUpdate.ConflictFilter = doUpdate.ConflictFilter.Accept(r).(LogicalExpr)
				}
			}
		},
	)
}

// execFields executes the given fields in the correct order.
func (r *rewriteVisitor) execFields(fields []func()) {
	if r.postOrderVisit {
		for i := len(fields) - 1; i >= 0; i-- {
			fields[i]()
		}
	} else {
		for _, f := range fields {
			f()
		}
	}
}

// expr is a helper method for traversing expressions.
func (r *rewriteVisitor) expr(node LogicalExpr, fn ...func()) LogicalExpr {
	return rewriteInOrder(r, r.exprCallback, fn, node)
}

// plan is a helper method for traversing plans.
// It takes a list of functions which should be passed in their pre-order order.
func (r *rewriteVisitor) plan(node LogicalPlan, fn ...func()) LogicalPlan {
	return rewriteInOrder(r, r.planCallback, fn, node)
}

// scanSource is a helper method for traversing scan sources.
func (r *rewriteVisitor) scanSource(node ScanSource, fn ...func()) ScanSource {
	return rewriteInOrder(r, r.scanSourceCallback, fn, node)
}

// rewriteInOrder is a generic function for executing a rewrite based on a certain order.
func rewriteInOrder[T Traversable](r *rewriteVisitor, callback func(T) (T, error), fields []func(), node T) T {
	if r.callbackFuncFirst {
		res, err := callback(node)
		if err != nil {
			panic(err)
		}

		r.execFields(fields)
		return res
	}

	r.execFields(fields)
	res, err := callback(node)
	if err != nil {
		panic(err)
	}

	return res
}
