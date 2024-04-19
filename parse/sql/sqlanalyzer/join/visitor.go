package join

import "github.com/kwilteam/kwil-db/parse/sql/tree"

type joinWalker struct {
	tree.AstListener
}

func NewJoinWalker() tree.AstListener {
	return &joinWalker{
		AstListener: tree.NewBaseListener(),
	}
}

func (s *joinWalker) EnterJoinPredicate(j *tree.JoinPredicate) error {
	err := checkJoin(j)
	if err != nil {
		return err
	}

	return nil
}
