package join

import "github.com/kwilteam/kwil-db/pkg/engine/sqlparser/tree"

type joinVisitor struct {
	*tree.BaseVisitor
}

func NewJoinVisitor() tree.Visitor {
	return &joinVisitor{
		BaseVisitor: tree.NewBaseVisitor(),
	}
}

func (s *joinVisitor) VisitJoinPredicate(j *tree.JoinPredicate) error {
	err := checkJoin(j)
	if err != nil {
		return err
	}

	return nil
}
