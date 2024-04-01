package costmodel

import "github.com/kwilteam/kwil-db/internal/engine/cost/plantree"

// CostEstimator implements the visitor pattern.
type CostEstimator struct {
}

func (c *CostEstimator) Visit(node plantree.TreeNode) (bool, any) {
	//TODO implement me
	panic("implement me")
}

func (c *CostEstimator) PreVisit(node plantree.TreeNode) (bool, any) {
	//TODO implement me
	panic("implement me")
}

func (c *CostEstimator) VisitChildren(node plantree.TreeNode) (bool, any) {
	//TODO implement me
	panic("implement me")
}

func (c *CostEstimator) PostVisit(node plantree.TreeNode) (bool, any) {
	//TODO implement me
	panic("implement me")
}

var _ plantree.TreeNodeVisitor = (*CostEstimator)(nil)
