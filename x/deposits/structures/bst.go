package structures

import (
	"sync"
)

// This BST was based off the implementation found here: https://flaviocopes.com/golang-data-structure-binary-search-tree/
// I could not find a license for the code, so I have not included it in the license file.

// Item the type of the binary search tree
type Item Withdrawal

type Node struct {
	key   int64
	item  Item
	left  *Node
	right *Node
}

func newNode(key int64, item Item) *Node {
	return &Node{key, item, nil, nil}
}

type BST struct {
	root *Node
	lock sync.RWMutex
}

func NewBST() *BST {
	return &BST{
		root: nil,
		lock: sync.RWMutex{},
	}
}

/*
	We will need insert, remove, min, and get
*/

func (bst *BST) Insert(key int64, item Item) *Node {
	bst.lock.Lock()
	defer bst.lock.Unlock()
	n := newNode(key, item)
	if bst.root == nil {
		bst.root = n
	} else {
		insertNode(bst.root, n)
	}

	return n
}

func insertNode(node, newNode *Node) {
	if newNode.key < node.key {
		// go left
		if node.left == nil {
			// furthest left reached
			node.left = newNode
		} else {
			// not furthest left, go another level down
			insertNode(node.left, newNode)
		}
	} else {
		// go right
		if node.right == nil {
			// furthest right reached
			node.right = newNode
		} else {
			// not furthest right, go another level down
			insertNode(node.right, newNode)
		}
	}
}

func (bst *BST) Remove(key int64) {
	bst.lock.Lock()
	defer bst.lock.Unlock()
	bst.root = remove(bst.root, key)
}

// remove removes given key given the root node
func remove(node *Node, key int64) *Node {
	if node == nil {
		// no nodes in tree
		return nil
	}
	if key < node.key {
		// there are further left nodes, go left
		node.left = remove(node.left, key)
		return node
	}
	if key > node.key {
		// there are further right nodes, go right
		node.right = remove(node.right, key)
		return node
	}
	// key == node.key, we have found the node to remove
	if node.left == nil && node.right == nil {
		// node is a leaf node, remove it
		node = nil
		return nil
	}
	if node.left == nil {
		// node has no left child but has a right child, replace it with right child
		node = node.right
		return node
	}
	if node.right == nil {
		// node has no right child but has a left child with the same key, replace it with left child
		// can this ever happen?  I don't think so, but ill leave it in for now
		node = node.left
		return node
	}

	leftmostrightside := node.right
	for {
		//find smallest value on the right side
		if leftmostrightside != nil && leftmostrightside.left != nil {
			leftmostrightside = leftmostrightside.left
		} else {
			break
		}
	}
	node.key, node.item = leftmostrightside.key, leftmostrightside.item
	node.right = remove(node.right, node.key)
	return node
}

// min returns the minimum key and value
func (bst *BST) Min() *Node {
	bst.lock.RLock()
	defer bst.lock.RUnlock()
	n := bst.root
	if n == nil {
		return nil
	}
	for {
		if n.left == nil {
			return n
		}
		n = n.left
	}
}

// Get returns the value for the given key
func (bst *BST) Get(key int64) *Node {
	bst.lock.RLock()
	defer bst.lock.RUnlock()
	return get(bst.root, key)
}

func get(node *Node, key int64) *Node {
	if node == nil {
		return nil
	}
	if key < node.key {
		// go left
		return get(node.left, key)
	}
	if key > node.key {
		// go right
		return get(node.right, key)
	}
	// key == node.key, we have found the node
	return node
}

func (n *Node) Item() Item {
	return n.item
}

func (n *Node) Key() int64 {
	return n.key
}

func (n *Node) Left() *Node {
	return n.left
}

func (n *Node) Right() *Node {
	return n.right
}
