package driver

import poly "kwil/pkg/utils/numbers/polynomial"

func (c *Connection) Plan(query string, args ...any) (plans QueryPlan, err error) {
	return c.plan(query, func(stmt *Statement) error {
		return stmt.BindMany(args)
	})
}

func (c *Connection) PlanNamed(query string, args map[string]any) (plans QueryPlan, err error) {
	return c.plan(query, func(stmt *Statement) error {
		return stmt.SetMany(args)
	})
}

func (c *Connection) plan(query string, statementSetterFn func(*Statement) error) (planForest QueryPlan, err error) {
	query = "EXPLAIN QUERY PLAN " + query
	var plans QueryPlan
	err = c.query(query, func(s *Statement) error {
		p := &QueryPlanNode{}
		p.Id = s.GetInt64("id")
		p.Parent = s.GetInt64("parent")
		p.NotUsed = s.GetInt64("notused")
		p.Detail = s.GetText("detail")

		p.Action, err = parseAction(p.Detail)
		if err != nil {
			return err
		}

		plans = append(plans, p)
		return nil
	}, statementSetterFn)
	if err != nil {
		return nil, err
	}

	plans.build()

	return plans, nil
}

type QueryPlan []*QueryPlanNode

// build builds the query plan tree from the flat list of nodes.
func (q *QueryPlan) build() {
	nodeMap := make(map[int64]*QueryPlanNode)
	var roots QueryPlan

	for _, node := range *q {
		nodeMap[node.Id] = node
	}

	for _, node := range *q {
		if node.Parent <= 0 {
			roots = append(roots, node)
		} else {
			parentNode, ok := nodeMap[node.Parent]
			if ok {
				parentNode.Children = append(parentNode.Children, node)
			}
		}
	}

	*q = roots
}

func (q *QueryPlan) Polynomial() poly.Expression {
	var expr poly.Expression
	expr = poly.NewWeight(FINAL_WEIGHT)
	for _, node := range *q {
		expr = poly.Mul(expr, node.Polynomial())
	}

	return expr
}

type QueryPlanNode struct {
	Id       int64
	Parent   int64
	NotUsed  int64
	Detail   string
	Children []*QueryPlanNode
	Action   Action
}

func (node *QueryPlanNode) AddChild(child *QueryPlanNode) {
	node.Children = append(node.Children, child)
}

func (q *QueryPlanNode) Polynomial() poly.Expression {
	expr := q.Action.Polynomial()

	if len(q.Children) == 0 {
		return expr
	}

	var innerExpr poly.Expression
	innerExpr = poly.NewWeight(1)
	for _, child := range q.Children {
		if child.Action == nil {
			continue
		}
		innerExpr = poly.Mul(innerExpr, child.Action.Polynomial())
	}

	return poly.Add(expr, innerExpr)
}
