package join

import "github.com/kwilteam/kwil-db/parse/sql/tree"

type joinVisitor struct {
	tree.AstWalker
}

func NewJoinWalker() tree.AstWalker {
	return &joinVisitor{
		AstWalker: tree.NewBaseWalker(),
	}
}

func (s *joinVisitor) EnterJoinPredicate(j *tree.JoinPredicate) error {
	err := checkJoin(j)
	if err != nil {
		return err
	}

	return nil
}
