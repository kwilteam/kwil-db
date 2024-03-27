package plantree_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	pt "github.com/kwilteam/kwil-db/internal/engine/cost/plantree"
)

type mockValueNode struct {
	*pt.BaseTreeNode

	value any
}

func (n *mockValueNode) Children() []pt.TreeNode {
	return []pt.TreeNode{}
}

func (n *mockValueNode) Accept(v pt.TreeNodeVisitor) (bool, any) {
	return v.Visit(n)
}

//func (n *mockValueNode) TransformUp(fn pt.TransformFunc) pt.TreeNode {
//	newChildren := n.TransformChildren(func(node pt.TreeNode) pt.TreeNode {
//		return node.TransformUp(fn)
//	})
//
//	return fn(newChildren)
//}
//
//func (n *mockValueNode) TransformDown(fn pt.TransformFunc) pt.TreeNode {
//	newNode := fn(n)
//
//	return newNode.TransformChildren(func(node pt.TreeNode) pt.TreeNode {
//		return node.TransformDown(fn)
//	})
//}

func (n *mockValueNode) TransformChildren(fn pt.TransformFunc) pt.TreeNode {
	return &mockValueNode{
		BaseTreeNode: pt.NewBaseTreeNode(),
		value:        n.value,
	}
}

func (n *mockValueNode) String() string {
	return fmt.Sprintf("%v", n.value)
}

type mockBinaryTreeNode struct {
	*pt.BaseTreeNode

	left  pt.TreeNode
	right pt.TreeNode
}

func (n *mockBinaryTreeNode) Children() []pt.TreeNode {
	return []pt.TreeNode{n.left, n.right}
}

func (n *mockBinaryTreeNode) Accept(v pt.TreeNodeVisitor) (bool, any) {
	return v.Visit(n)
}

//func (n *mockBinaryTreeNode) TransformUp(fn pt.TransformFunc) pt.TreeNode {
//	newChildren := n.TransformChildren(func(node pt.TreeNode) pt.TreeNode {
//		return node.TransformUp(fn)
//	})
//
//	return fn(newChildren)
//}
//
//func (n *mockBinaryTreeNode) TransformDown(fn pt.TransformFunc) pt.TreeNode {
//	newNode := fn(n)
//
//	return newNode.TransformChildren(func(node pt.TreeNode) pt.TreeNode {
//		return node.TransformDown(fn)
//	})
//}

func (n *mockBinaryTreeNode) TransformChildren(fn pt.TransformFunc) pt.TreeNode {
	return &mockBinaryTreeNode{
		BaseTreeNode: pt.NewBaseTreeNode(),

		left:  fn(n.left),
		right: fn(n.right),
	}
}

func (n *mockBinaryTreeNode) String() string {
	return fmt.Sprintf("(%v, %v)", n.left, n.right)
}

func mockLeftTree() *mockBinaryTreeNode {
	//    /\
	//   /\ 4
	//  /\ 3
	// 1  2
	return &mockBinaryTreeNode{
		BaseTreeNode: pt.NewBaseTreeNode(),
		left: &mockBinaryTreeNode{
			BaseTreeNode: pt.NewBaseTreeNode(),
			left: &mockBinaryTreeNode{
				BaseTreeNode: pt.NewBaseTreeNode(),
				left: &mockValueNode{
					BaseTreeNode: pt.NewBaseTreeNode(),
					value:        1,
				},
				right: &mockValueNode{
					BaseTreeNode: pt.NewBaseTreeNode(),
					value:        2,
				},
			},
			right: &mockValueNode{
				BaseTreeNode: pt.NewBaseTreeNode(),
				value:        3,
			},
		},
		right: &mockValueNode{
			BaseTreeNode: pt.NewBaseTreeNode(),
			value:        4,
		},
	}
}

func mockRightTree() *mockBinaryTreeNode {
	//   /\
	//  4 /\
	//   3 /\
	//    1  2
	return &mockBinaryTreeNode{
		BaseTreeNode: pt.NewBaseTreeNode(),
		left: &mockValueNode{
			BaseTreeNode: pt.NewBaseTreeNode(),
			value:        4,
		},
		right: &mockBinaryTreeNode{
			BaseTreeNode: pt.NewBaseTreeNode(),
			left: &mockValueNode{
				BaseTreeNode: pt.NewBaseTreeNode(),
				value:        3,
			},
			right: &mockBinaryTreeNode{
				BaseTreeNode: pt.NewBaseTreeNode(),
				left: &mockValueNode{
					BaseTreeNode: pt.NewBaseTreeNode(),
					value:        1,
				},
				right: &mockValueNode{
					BaseTreeNode: pt.NewBaseTreeNode(),
					value:        2,
				},
			},
		},
	}
}

func mockApplyFuncPreOrderCollect(n pt.TreeNode) []any {
	collected := []any{}

	pt.PreOrderApply(n, func(n pt.TreeNode) (bool, any) {
		if v, ok := n.(*mockValueNode); ok {
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
		if v, ok := n.(*mockValueNode); ok {
			collected = append(collected, v.value)
			return true, v.value
		}
		return true, n
	})

	return collected
}

func TestOrderApply_left_tree(t *testing.T) {
	node := mockLeftTree()

	assert.Equal(t, []any{1, 2, 3, 4}, mockApplyFuncPreOrderCollect(node))
	assert.Equal(t, []any{1, 2, 3, 4}, mockApplyFuncPostOrderCollect(node))
}

func TestOrderApply_right_tree(t *testing.T) {
	node := mockRightTree()

	assert.Equal(t, []any{4, 3, 1, 2}, mockApplyFuncPreOrderCollect(node))
	assert.Equal(t, []any{4, 3, 1, 2}, mockApplyFuncPostOrderCollect(node))
}

func mockTransform(node pt.TreeNode) pt.TreeNode {
	return pt.TransformPostOrder(node, func(n pt.TreeNode) pt.TreeNode {
		if v, ok := n.(*mockValueNode); ok {
			return &mockValueNode{
				value: v.value.(int) * 2,
			}
		}
		// otherwise, return the original node
		return n
	})
}

func TestTransform_left_tree(t *testing.T) {
	node := mockLeftTree()

	transformed := mockTransform(node)
	// new tree's nodes have been transformed
	assert.Equal(t, []any{2, 4, 6, 8}, mockApplyFuncPreOrderCollect(transformed))
	// original tree's nodes are not changed
	assert.Equal(t, []any{1, 2, 3, 4}, mockApplyFuncPreOrderCollect(node))

	leftNode := node.left
	leftTransformed := mockTransform(leftNode)
	// new left tree's nodes have been transformed
	assert.Equal(t, []any{2, 4, 6}, mockApplyFuncPreOrderCollect(leftTransformed))
	// original left tree's nodes are not changed
	assert.Equal(t, []any{1, 2, 3}, mockApplyFuncPreOrderCollect(leftNode))
	// original parent tree's nodes are not changed
	assert.Equal(t, []any{1, 2, 3, 4}, mockApplyFuncPreOrderCollect(node))
}

func mockNodeTransformFunc(node pt.TreeNode, transformFunc pt.TransformFunc) pt.TreeNode {
	switch t := node.(type) {
	case *mockValueNode:
		return &mockValueNode{
			BaseTreeNode: pt.NewBaseTreeNode(),
			value:        t.value,
		}
	case *mockBinaryTreeNode:
		return &mockBinaryTreeNode{
			BaseTreeNode: pt.NewBaseTreeNode(),
			left:         transformFunc(t.left),
			right:        transformFunc(t.right),
		}
	default:
		panic("unknown node type")
	}
}

func TestTransform_left_tree_using_fn(t *testing.T) {

	node := mockLeftTree()

	transformed := mockTransform(node)

	// new tree's nodes have been transformed
	assert.Equal(t, []any{2, 4, 6, 8}, mockApplyFuncPreOrderCollect(transformed))
	// original tree's nodes are not changed
	assert.Equal(t, []any{1, 2, 3, 4}, mockApplyFuncPreOrderCollect(node))

	leftNode := node.left
	leftTransformed := mockTransform(leftNode)

	// new left tree's nodes have been transformed
	assert.Equal(t, []any{2, 4, 6}, mockApplyFuncPreOrderCollect(leftTransformed))
	// original left tree's nodes are not changed
	assert.Equal(t, []any{1, 2, 3}, mockApplyFuncPreOrderCollect(leftNode))
	// original parent tree's nodes are not changed
	assert.Equal(t, []any{1, 2, 3, 4}, mockApplyFuncPreOrderCollect(node))
}

type cloneB struct {
	a int
}

type cloneA struct {
	b *cloneB
}

func (c *cloneA) Clone1() *cloneA {
	bb := *c.b
	return &cloneA{b: &bb}
}

func (c *cloneA) Clone2() *cloneA {
	cc := *c
	return &cc
}

func TestClone(t *testing.T) {
	a := &cloneA{b: &cloneB{a: 1}}
	b := a.Clone1()
	assert.Equal(t, 1, b.b.a)

	c := a.Clone2()
	assert.Equal(t, 1, c.b.a)

	fmt.Printf("a: %p, a.b: %p\n", a, a.b)
	fmt.Printf("b: %p, b.b: %p\n", b, b.b)
	fmt.Printf("c: %p, c.b: %p\n", c, c.b)
}

type baseI interface {
	foo()
	bar()
}

type baseA struct {
}

func (a *baseA) foo() {
	fmt.Println("baseA foo")
	a.bar()
}

func (a *baseA) bar() {
	fmt.Println("baseA bar")
}

//type baseB struct {
//}
//
//func (a *baseB) foo() {}
//
//func (a *baseB) bar() {
// 	switch a.(type) {
//
//	}
//}

type derivedA struct {
	*baseA
}

//func (a *derivedA) foo() {
//	fmt.Println("derivedA foo")
//	//a.baseA.foo()
//}

func (a *derivedA) bar() {
	fmt.Println("derivedA bar")
}

type derivedB struct {
	*baseA
}

//func (a *derivedB) foo() {
//
//}

func (a *derivedB) bar() {
	fmt.Println("derivedB bar")
}

func TestInheritance(t *testing.T) {
	a := &derivedA{&baseA{}}
	a.foo()
}
