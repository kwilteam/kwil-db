package join

import "github.com/kwilteam/kwil-db/parse/sql/tree"

type joinWalker struct {
	tree.AstWalker
}

func NewJoinWalker() tree.AstWalker {
	return &joinWalker{
		AstWalker: tree.NewBaseWalker(),
	}
}

func (s *joinWalker) EnterJoinPredicate(j *tree.JoinPredicate) error {
	err := checkJoin(j)
	if err != nil {
		return err
	}

	return nil
}
