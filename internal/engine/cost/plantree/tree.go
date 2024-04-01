package plantree

import "fmt"

type NodeFunc func(TreeNode) (bool, any)
type TransformFunc func(TreeNode) TreeNode

// Tree represents a node in a tree, which is visitable.
type Tree interface {
	Children() []TreeNode
	fmt.Stringer
}

type TreeNode interface {
	Tree

	// Accept visits the node using the provided TreeNodeVisitor.
	Accept(TreeNodeVisitor) any

	// TransformChildren applies the provided TransformFunc to children node,
	// returns current node (with children node transformed).
	TransformChildren(TransformFunc) TreeNode
}

type ExprNode interface {
	TreeNode

	ExprNode()
}

type PlanNode interface {
	TreeNode

	PlanNode()
}

func ApplyNodeFuncToChildren(node TreeNode, fn NodeFunc) (bool, any) {
	for _, child := range node.Children() {
		keepGoing, res := fn(child)
		if !keepGoing {
			return false, res
		}
	}
	return true, nil
}

// PreOrderApply walks through the node and its children and applies the
// provided NodeFunc, in pre-order.
// The point is that you can quickly apply a function to the whole tree.
// It's a light version of TreeNodeVisitor.
func PreOrderApply(node TreeNode, fn NodeFunc) (bool, any) {
	keepGoing, res := fn(node)
	if !keepGoing {
		return false, res
	}

	return ApplyNodeFuncToChildren(node,
		func(node TreeNode) (bool, any) {
			return PreOrderApply(node, fn)
		})
}

// PostOrderApply walks through the node and its children and applies the
// provided NodeFunc, in post-order.
// The point is that you can quickly apply a function to the whole tree.
// It's a light version of TreeNodeVisitor.
// PostOrder means all children are visited before the node itself.
func PostOrderApply(node TreeNode, fn NodeFunc) (bool, any) {
	keepGoing, res := ApplyNodeFuncToChildren(node,
		func(node TreeNode) (bool, any) {
			return PostOrderApply(node, fn)
		})
	if !keepGoing {
		return false, res
	}

	return fn(node)
}

// TransformPostOrder applies the provided TransformFunc to copied node in post-order.
// and apply it to its children by calling TransformChildren, returns transformed node.
// It traverses the tree in DFS post-order.
// NOTE: Since Go using composition instead of inheritance, we can't define
// a default implementation for TransformPostOrder in the BaseTreeNode.
// So use this function when you want to apply TransformFunc to a node in post-order.
func TransformPostOrder(node TreeNode, fn TransformFunc) TreeNode {

	cfn := func(n TreeNode) TreeNode {
		return TransformPostOrder(n, fn)
	}

	//newChildren := node.TransformChildren(func(n TreeNode) TreeNode {
	//	return TransformPostOrder(n, fn)
	//})

	newNode := node.TransformChildren(cfn)
	return fn(newNode)
}

// TransformPreOrder applies the provided TransformFunc to copied node in pre-order.
// and apply it to its children by calling TransformChildren, returns transformed node.
// It traverses the tree in DFS pre-order.
// NOTE: Since Go using composition instead of inheritance, we can't define
// a default implementation for TransformPreOrder in the BaseTreeNode.
// So use this function when you want to apply TransformFunc to a node in pre-order.
func TransformPreOrder(node TreeNode, fn TransformFunc) TreeNode {
	newNode := fn(node)

	return newNode.TransformChildren(func(n TreeNode) TreeNode {
		return TransformPreOrder(n, fn)
	})
}

type BaseTreeNode struct{}

func (n *BaseTreeNode) String() string {
	return fmt.Sprintf("%T", n)
}

func (n *BaseTreeNode) Children() []TreeNode {
	panic("implement me")
}

func (n *BaseTreeNode) Accept(v TreeNodeVisitor) any {
	return n.Accept(v)
}

func (n *BaseTreeNode) TransformChildren(fn TransformFunc) TreeNode {
	panic("implement me")
}

func NewBaseTreeNode() *BaseTreeNode {
	return &BaseTreeNode{}
}

type TreeNodeVisitor func(TreeNode) any
