package memo

import (
	"bytes"
	"fmt"
	"github.com/kwilteam/kwil-db/internal/engine/cost/datatypes"
	"github.com/kwilteam/kwil-db/internal/engine/cost/logical_plan"
	"github.com/kwilteam/kwil-db/internal/engine/cost/plantree"
	"github.com/kwilteam/kwil-db/internal/engine/cost/virtual_plan"
)

type GroupRel interface {
	plantree.TreeNode

	Group() *Group
	InputGroups() []*Group
	Inputs() []GroupRel
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

func (r *LogicalRel) Inputs() []GroupRel {
	var rels []GroupRel
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

func (r *VirtualRel) Inputs() []GroupRel {
	var rels []GroupRel
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

func Format(plan GroupRel, indent int) string {
	var msg bytes.Buffer
	msg.WriteString(plan.String())
	msg.WriteString("\n")
	for _, child := range plan.Inputs() {
		msg.WriteString(Format(child, indent+2))
	}
	return msg.String()
}
