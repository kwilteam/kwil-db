package plantree_test

import (
	"github.com/kwilteam/kwil-db/internal/engine/cost/logical_plan"
	"testing"

	"github.com/stretchr/testify/assert"

	pt "github.com/kwilteam/kwil-db/internal/engine/cost/plantree"
)

type mockTreeNode struct {
	*pt.BaseTreeNode

	children []pt.TreeNode

	value any
}

func (n *mockTreeNode) Children() []pt.TreeNode {
	return n.children
}

func (n *mockTreeNode) Accept(v pt.TreeNodeVisitor) (bool, any) {
	return v.Visit(n)
}

func (n *mockTreeNode) TransformChildren(fn pt.TransformFunc) pt.TreeNode {
	newChildren := make([]pt.TreeNode, 0, len(n.children))
	for _, child := range n.children {
		newChildren = append(newChildren, fn(child))
	}
	return &mockTreeNode{
		BaseTreeNode: pt.NewBaseTreeNode(),
		value:        n.value,
		children:     newChildren,
	}
}

func (n *mockTreeNode) String() string {
	return logical_plan.PpList(n.children)
}

func mockLeftTree() *mockTreeNode {
	//      0
	//    /   \
	//   1     2
	//  / \   / \
	// 3   4 5   6
	//
	// preorder: 0 1 3 4 2 5 6
	// postorder: 3 4 1 5 6 2 0
	return &mockTreeNode{
		BaseTreeNode: pt.NewBaseTreeNode(),
		value:        0,
		children: []pt.TreeNode{
			&mockTreeNode{
				BaseTreeNode: pt.NewBaseTreeNode(),
				value:        1,
				children: []pt.TreeNode{
					&mockTreeNode{
						BaseTreeNode: pt.NewBaseTreeNode(),
						value:        3,
					},
					&mockTreeNode{
						BaseTreeNode: pt.NewBaseTreeNode(),
						value:        4,
					},
				},
			},
			&mockTreeNode{
				BaseTreeNode: pt.NewBaseTreeNode(),
				value:        2,
				children: []pt.TreeNode{
					&mockTreeNode{
						BaseTreeNode: pt.NewBaseTreeNode(),
						value:        5,
					},
					&mockTreeNode{
						BaseTreeNode: pt.NewBaseTreeNode(),
						value:        6,
					},
				},
			},
		},
	}
}

func mockApplyFuncPreOrderCollect(n pt.TreeNode) []any {
	collected := []any{}

	pt.PreOrderApply(n, func(n pt.TreeNode) (bool, any) {
		if v, ok := n.(*mockTreeNode); ok {
			collected = append(collected, v.value)
			return true, v.value
		}

		return true, n
	})

	return collected
}

func mockApplyFuncPostOrderCollect(n pt.TreeNode) []any {
	collected := []any{}

	pt.PostOrderApply(n, func(n pt.TreeNode) (bool, any) {
		if v, ok := n.(*mockTreeNode); ok {
			collected = append(collected, v.value)
			return true, v.value
		}
		return true, n
	})

	return collected
}

func TestOrderApply(t *testing.T) {
	node := mockLeftTree()

	assert.Equal(t, []any{0, 1, 3, 4, 2, 5, 6}, mockApplyFuncPreOrderCollect(node))
	assert.Equal(t, []any{3, 4, 1, 5, 6, 2, 0}, mockApplyFuncPostOrderCollect(node))
}

func mockTransform(node pt.TreeNode) pt.TreeNode {
	return pt.TransformPostOrder(node, func(n pt.TreeNode) pt.TreeNode {
		if v, ok := n.(*mockTreeNode); ok {
			return &mockTreeNode{
				BaseTreeNode: v.BaseTreeNode,
				value:        v.value.(int) * 2,
				children:     v.children,
			}
		}
		// otherwise, return the original node
		return n
	})
}

func TestTransform(t *testing.T) {
	node := mockLeftTree()

	originPreOrder := []any{0, 1, 3, 4, 2, 5, 6}
	originPostOrder := []any{3, 4, 1, 5, 6, 2, 0}

	transformed := mockTransform(node)
	// new tree's nodes have been transformed
	assert.Equal(t, []any{0, 2, 6, 8, 4, 10, 12}, mockApplyFuncPreOrderCollect(transformed))
	// original tree's nodes are not changed
	assert.Equal(t, originPreOrder, mockApplyFuncPreOrderCollect(node))
	assert.Equal(t, originPostOrder, mockApplyFuncPostOrderCollect(node))

	leftNode := node.children[0]
	leftTransformed := mockTransform(leftNode)
	// new left tree's nodes have been transformed
	assert.Equal(t, []any{2, 6, 8}, mockApplyFuncPreOrderCollect(leftTransformed))
	// original left tree's nodes are not changed
	assert.Equal(t, []any{1, 3, 4}, mockApplyFuncPreOrderCollect(leftNode))
	// original parent tree's nodes are not changed
	assert.Equal(t, originPreOrder, mockApplyFuncPreOrderCollect(node))
	assert.Equal(t, originPostOrder, mockApplyFuncPostOrderCollect(node))
}
