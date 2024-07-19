package memo

import (
	"bytes"
	"fmt"

	"github.com/kwilteam/kwil-db/internal/engine/cost/datatypes"
	"github.com/kwilteam/kwil-db/internal/engine/cost/logical_plan"
	"github.com/kwilteam/kwil-db/internal/engine/cost/plantree"
	"github.com/kwilteam/kwil-db/internal/engine/cost/virtual_plan"
)

//// GroupExpr is used to store all the logically equivalent expressions which
//// have the same root operator. Different from a normal expression, the
//// Children of a Group expression are expression Groups, not expressions.
//// Another property of Group expression is that the child Group references will
//// never be changed once the Group expression is created.
//type GroupExpr struct {
//	ExprNode plannercore.LogicalPlan
//	Children []*Group
//	Group    *Group
//
//	// ExploreMark is uses to mark whether this GroupExpr has been fully
//	// explored by a transformation rule batch in a certain round.
//	ExploreMark
//
//	selfFingerprint string
//	// appliedRuleSet saves transformation rules which have been applied to this
//	// GroupExpr, and will not be applied again. Use `uint64` which should be the
//	// id of a Transformation instead of `Transformation` itself to avoid import cycle.
//	appliedRuleSet map[uint64]struct{}
//}

// GroupExpression is an interface for a relation expression(from Volcano/Cascades model).
// A GroupExpression is to store all the logically equivalent expressions which
// have the same root operator. Every child of a GroupExpression is a Group.
type GroupExpression interface {
	plantree.TreeNode

	Group() *Group
	InputGroups() []*Group
	Inputs() []GroupExpression
	Statistics() *datatypes.Statistics
	SetStatistics(stat *datatypes.Statistics)
	Cost() int64
}

//type baseRel struct {
//	plantree.TreeNode
//
//	stat   *datatypes.Statistics // intermediate statistics
//	group  *Group
//	cost   int64
//	inputs []*Group
//}

// LogicalRel is a wrapper of a logical plan,it's used in memo.
type LogicalRel struct {
	plantree.TreeNode

	stat   *datatypes.Statistics // intermediate statistics
	group  *Group
	cost   int64
	inputs []*Group

	plan logical_plan.LogicalPlan
}

func NewLogicalRel(plan logical_plan.LogicalPlan,
	group *Group,
	inputs []*Group) *LogicalRel {
	return &LogicalRel{
		TreeNode: plantree.NewBaseTreeNode(),
		plan:     plan,
		group:    group,
		inputs:   inputs,
		cost:     0,
	}
}

func (r *LogicalRel) String() string {
	inputGroupId := make([]int, len(r.inputs))
	for i, input := range r.inputs {
		inputGroupId[i] = input.id
	}

	return fmt.Sprintf("Group: %d\n  Plans: %T\n  Inputs: %v",
		r.group.id, r.plan, inputGroupId)
}

func (r *LogicalRel) Group() *Group {
	return r.group
}

func (r *LogicalRel) InputGroups() []*Group {
	return r.inputs
}

func (r *LogicalRel) Inputs() []GroupExpression {
	var rels []GroupExpression
	for _, input := range r.inputs {
		rels = append(rels, input.logical...)
		rels = append(rels, input.virtual...)
	}

	return rels
}

func (r *LogicalRel) Statistics() *datatypes.Statistics {
	return r.stat
}

func (r *LogicalRel) SetStatistics(stat *datatypes.Statistics) {
	r.stat = stat
}

func (r *LogicalRel) Cost() int64 {
	return r.cost
}

type VirtualRel struct {
	plantree.TreeNode

	stat   *datatypes.Statistics // intermediate statistics
	group  *Group
	cost   int64
	inputs []*Group

	plan virtual_plan.VirtualPlan
}

func NewVirtualRel(plan virtual_plan.VirtualPlan) *VirtualRel {
	return &VirtualRel{
		TreeNode: plantree.NewBaseTreeNode(),
		plan:     plan,
	}
}

func (r *VirtualRel) String() string {
	inputGroupId := make([]int, len(r.inputs))
	for i, input := range r.inputs {
		inputGroupId[i] = input.id
	}

	return fmt.Sprintf("Group: %d\n  Plans: %T\n  Inputs: %v",
		r.group.id, r.plan, inputGroupId)
}

func (r *VirtualRel) Group() *Group {
	return r.group
}

func (r *VirtualRel) InputGroups() []*Group {
	return r.inputs
}

func (r *VirtualRel) Inputs() []GroupExpression {
	var rels []GroupExpression
	for _, input := range r.inputs {
		rels = append(rels, input.logical...)
		rels = append(rels, input.virtual...)
	}

	return rels
}

func (r *VirtualRel) Statistics() *datatypes.Statistics {
	return r.stat
}

func (r *VirtualRel) SetStatistics(stat *datatypes.Statistics) {
	r.stat = stat
}

func (r *VirtualRel) Cost() int64 {
	return r.cost
}

func Format(plan GroupExpression, indent int) string {
	var msg bytes.Buffer
	msg.WriteString(plan.String())
	msg.WriteString("\n")
	for _, child := range plan.Inputs() {
		msg.WriteString(Format(child, indent+2))
	}
	return msg.String()
}
