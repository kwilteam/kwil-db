package structures

import (
	"sync"
)

/*
	A withdrawal tracker is a struct that include both a bst and a hash map.
	The bst stores pending withdrawals based on their expiration height.
	The hash map maps a nonce to a height.
*/

type WithdrawalTracker struct {
	bst *BST
	m   map[string]*Node
	mu  sync.Mutex // only used for garbage collection
}

func NewWithdrawalTracker() *WithdrawalTracker {
	return &WithdrawalTracker{
		bst: NewBST(),
		m:   make(map[string]*Node),
		mu:  sync.Mutex{},
	}
}

// Withdrawal will be used as the Item type for the BST
type Withdrawal interface {
	Expiration() int64
	Nonce() string
	Amount() string
	Spent() string
	Wallet() string
}

// Insert will insert a withdrawal into the BST and the hash map
func (wt *WithdrawalTracker) Insert(w Withdrawal) {
	n := wt.bst.Insert(w.Expiration(), w)
	wt.m[w.Nonce()] = n
}

// Remove will remove a withdrawal from the BST and the hash map
func (wt *WithdrawalTracker) RemoveByNonce(nonce string) {
	n := wt.m[nonce]
	if n == nil {
		return
	}
	wt.bst.Remove(n.Key())
	delete(wt.m, nonce)
}

// Poll will return all nodes with an expiration height less than or equal to the given height
func (wt *WithdrawalTracker) PopExpired(h int64) []*Node {
	var nodes []*Node
	for {
		min := wt.bst.Min()
		if min == nil {
			break
		}
		if min.Key() > h {
			break
		}
		nodes = append(nodes, min)
		wt.bst.Remove(min.Key())
		delete(wt.m, min.Item().Nonce())
	}
	return nodes
}

// Get will return the node with the given nonce
func (wt *WithdrawalTracker) GetByNonce(nonce string) *Node {
	return wt.m[nonce]
}

// RunGC will remake the hash map
func (wt *WithdrawalTracker) RunGC() {
	wt.mu.Lock()
	defer wt.mu.Unlock()
	nm := make(map[string]*Node)

	// loop through wt.m and add values to nm
	for k, v := range wt.m {
		nm[k] = v
	}

	wt.m = nm
}
