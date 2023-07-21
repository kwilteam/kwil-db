package join

import "github.com/kwilteam/kwil-db/pkg/engine/sqlparser/tree"

type joinVisitor struct {
	tree.Walker
}

func NewJoinVisitor() tree.Walker {
	return &joinVisitor{
		Walker: tree.NewBaseWalker(),
	}
}

func (s *joinVisitor) EnterJoinPredicate(j *tree.JoinPredicate) error {
	err := checkJoin(j)
	if err != nil {
		return err
	}

	return nil
}
