package plantree

import "fmt"

type NodeFunc func(TreeNode) (bool, any)
type TransformFunc func(TreeNode) TreeNode

// Tree represents a node in a tree, which is visitable.
type Tree interface {
	Children() []TreeNode
	ShallowClone() TreeNode
	fmt.Stringer
}

type TreeNode interface {
	Tree

	// Accept visits the node and its children using the provided TreeNodeVisitor.
	Accept(TreeNodeVisitor) (keepGoing bool, res any)

	//// Apply walks through the node and its children and applies the provided NodeFunc.
	//// The point is that you can quickly apply a function to the whole tree.
	//// It's a light version of Accept.
	//Apply(NodeFunc) (bool, any)

	// TransformChildren applies the provided TransformFunc to children node,
	// returns current node (with children node transformed).
	TransformChildren(TransformFunc) TreeNode
}

//type ExprNode interface {
//	TreeNode
//
//	ExprNode()
//}

//type PlanNode interface {
//	TreeNode
//
//	PlanNode()
//}

//type BaseNode struct {
//	children []Tree
//}
//
//func (n *BaseNode) Children() []Tree {
//	return n.children
//}
//
//// Accept visits the node and its children using the provided TreeNodeVisitor.
//// It traverses the tree in DFS pre-order.
//func (n *BaseNode) Accept(v TreeNodeVisitor) (bool, any) {
//	keepGoing, res := v.PreVisit(n)
//	if !keepGoing {
//		return false, res
//	}
//
//	keepGoing, res = n.visitChildren(v)
//	if !keepGoing {
//		return false, res
//	}
//
//	return v.PostVisit(n)
//}
//
//func (n *BaseNode) visitChildren(v TreeNodeVisitor) (bool, any) {
//	for _, child := range n.children {
//		keepGoing, res := child.Accept(v)
//		if !keepGoing {
//			return false, res
//		}
//	}
//	return true, nil
//}
//
//func (n *BaseNode) TransformUp(f TransformFunc) Tree {
//	return n.transformPostOrder(f)
//}
//
//func (n *BaseNode) transformPostOrder(f TransformFunc) Tree {
//	transformed := n.TransformChildren(f)
//	return f(transformed)
//}
//
//func (n *BaseNode) TransformChildren(f TransformFunc) Tree {
//	panic("not implemented")
//}

////type RecursiveNext int8
////
////const (
////	RecursiveNextStop RecursiveNext = iota
////	RecursiveNextContinue
////	RecursiveNextSkip
////)
//
//// TreeNodeVisitor implements the visitor pattern for walking Tree recursively.
//type TreeNodeVisitor interface {
//	// PreVisit is called before visiting the children of the node.
//	PreVisit(node Tree) (bool, any)
//
//	// PostVisit is called after visiting the children of the node.
//	PostVisit(node Tree) (bool, any)
//}
//
//type BaseNodeVisitor struct{}
//
//func (v *BaseNodeVisitor) PreVisit(node Tree) (bool, any) {
//	return true, nil
//}
//
//func (v *BaseNodeVisitor) PostVisit(node Tree) (bool, any) {
//	return true, nil
//}
//
//type BaseTreeNode struct{}
//
//func (n *BaseTreeNode) Children() []TreeNode {
//	panic("not implemented")
//}
//
//func (n *BaseTreeNode) Accept(v TreeNodeVisitor) (bool, any) {
//	keepGoing, res := v.PreVisit(n)
//	if !keepGoing {
//		return false, res
//	}
//
//	keepGoing, res = v.VisitChildren(n)
//	if !keepGoing {
//		return false, res
//	}
//
//	return v.PostVisit(n)
//}

// OnionOrderVisit visits the tree in onion order, ((())) like.
func OnionOrderVisit(v TreeNodeVisitor, node TreeNode) (bool, any) {
	keepGoing, res := v.PreVisit(node)
	if !keepGoing {
		return false, res
	}

	keepGoing, res = v.VisitChildren(node)
	if !keepGoing {
		return false, res
	}

	return v.PostVisit(node)
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

type TreeNodeVisitor interface {
	Visit(TreeNode) (bool, any)
	PreVisit(TreeNode) (bool, any)
	VisitChildren(TreeNode) (bool, any)
	PostVisit(TreeNode) (bool, any)
}

//
//type BaseTreeVisitor struct{}
//
//func (v *BaseTreeVisitor) Visit(node TreeNode) (bool, any) {
//	return true, node.Accept(v)
//}
//
//func (v *BaseTreeVisitor) VisitChildren(node TreeNode) (bool, any) {
//	return O
//}
//
//func (v *BaseTreeVisitor) PreVisit(node TreeNode) (bool, any) {
//	return true, nil
//}
//
//func (v *BaseTreeVisitor) PostVisit(node TreeNode) (bool, any) {
//	return true, nil
//}

type BaseTreeNode struct{}

func (n *BaseTreeNode) String() string {
	return fmt.Sprintf("%T", n)
}

func (n *BaseTreeNode) Children() []TreeNode {
	panic("implement me")
}

func (n *BaseTreeNode) ShallowClone() TreeNode {
	nn := *n
	return &nn
}

func (n *BaseTreeNode) Accept(v TreeNodeVisitor) (bool, interface{}) {
	return v.Visit(n)
}

func (n *BaseTreeNode) Apply(fn NodeFunc) (bool, any) {
	return PreOrderApply(n, fn)
}

//
//// Transform applies the provided TransformFunc to copied node in post-order.
//// NOTE: this should be implemented by the concrete node, otherwise it won't
//// call concrete node's TransformChildren.
//func (n *BaseTreeNode) TransformUp(fn TransformFunc) TreeNode {
//	//	newChildren := n.TransformChildren(func(node TreeNode) TreeNode {
//	//		return n.TransformUp(fn)
//	//	})
//	//
//	//	return fn(newChildren)
//	panic("implement me")
//}

//// NodeTransformFunc is a function that transforms a node and its children using
//// the provided TransformFunc.
//type NodeTransformFunc func(node TreeNode, transformFunc TransformFunc) TreeNode
//
//func PostOrderTransform(node TreeNode, fn TransformFunc, nodeFn NodeTransformFunc) TreeNode {
//	newChildren := nodeFn(node, func(n TreeNode) TreeNode {
//		return PostOrderTransform(n, fn, nodeFn)
//	})
//
//	return fn(newChildren)
//}
//
//func PreOrderTransform(node TreeNode, fn TransformFunc, nodeFn NodeTransformFunc) TreeNode {
//	newNode := fn(node)
//
//	return nodeFn(newNode, func(n TreeNode) TreeNode {
//		return fn(n)
//	})
//}

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

//func (n *BaseTreeNode) TransformDown(fn TransformFunc) TreeNode {
//	//newNode := fn(n)
//	//
//	//return newNode.TransformChildren(func(node TreeNode) TreeNode {
//	//	return node.TransformDown(fn)
//	//})
//	panic("implement me")
//}

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

func (n *BaseTreeNode) TransformChildren(fn TransformFunc) TreeNode {
	panic("implement me")
}

func NewBaseTreeNode() *BaseTreeNode {
	return &BaseTreeNode{}
}
