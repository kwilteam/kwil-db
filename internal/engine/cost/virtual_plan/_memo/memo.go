package memo

import (
	"github.com/kwilteam/kwil-db/internal/engine/cost/logical_plan"
	"github.com/kwilteam/kwil-db/internal/engine/cost/virtual_plan"
)

type Expression struct {
	expr      string
	operation string
	cost      int
}

// Group holds a set of logically equivalent logical/virtual plans.
type Group struct {
	id      int
	bestIdx int

	best virtual_plan.VirtualPlan

	logical []GroupExpression
	virtual []GroupExpression
}

//func (g *Group) String() string {
//	return fmt.Sprintf("Group %d\n  Logical: %s\n  Virtual: %v",
//		g.id, g.logical, g.virtual)
//}

func NewGroup(id int) *Group {
	return &Group{id: id}
}

func (g *Group) addRelExpr(rel GroupExpression) {
	switch t := rel.(type) {
	case *LogicalRel:
		g.logical = append(g.logical, t)
	case *VirtualRel:
		g.virtual = append(g.virtual, t)
	}
}

//var _ PlanNode = (*Group)(nil)

// Memo holds the state of the memoization.
// A plan like:
// a - b - c - e
// . . .  \d - f
// will be represented as:
// root: G5
//
// G5: a -> G4
// G4: b -> [G3, G1]
// G1: d -> G0
// G0: f
// G3: c -> G0
// G2: e

type Memo struct {
	root *Group

	nextGroupId int

	groups []*Group

	//groups map[plantree.TreeNode]Group
}

func NewMemo() *Memo {
	return &Memo{
		groups: []*Group{},
	}
}

func (m *Memo) newGroup() *Group {
	g := NewGroup(m.nextGroupId)
	m.nextGroupId++
	return g
}

// Init initializes the memo, and returns the root group.
func (m *Memo) Init(plan logical_plan.LogicalPlan) *Group {
	// todo:
	// 1. implement PlanNode for all LogicalPlan/VirtualPlan/GroupPlan
	// 2. build a tree of GroupPlan

	return m.addPlanToGroup(plan, nil)
}

//// transformPlanToRel transforms a logical plan to a Rel.
//// It does the same as addPlanToGroup, but using TranformPostOrder.
//// NOTE: not tested.
//func (m *Memo) transformPlanToRel(plan logical_plan.LogicalPlan) Rel {
//	rel := plantree.TransformPostOrder(plan, func(n plantree.TreeNode) plantree.TreeNode {
//		target := m.newGroup()
//		m.groups = append(m.groups, target)
//
//		relExpr := Rel{
//			TreeNode: plan,
//			group:    target,
//			inputs:   plan.Inputs(),
//		}
//
//		target.addRelExpr(relExpr)
//
//		return relExpr
//	})
//
//	return rel
//}

// addToGroup recursively adds a plan to a group.
func (m *Memo) addPlanToGroup(src logical_plan.LogicalPlan, target *Group) *Group {
	inputs := make([]*Group, len(src.Inputs()))
	for i, child := range src.Inputs() {
		ng := m.addPlanToGroup(child, nil)
		inputs[i] = ng
	}

	if target == nil {
		target = m.newGroup()
		m.groups = append(m.groups, target)
	}

	rel := NewLogicalRel(src, target, inputs)

	target.addRelExpr(rel)

	return target
}

// explore explores the memo group by expanding it.
func (m *Memo) explore() {

}
