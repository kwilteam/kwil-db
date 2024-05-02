package join

import (
	"github.com/kwilteam/kwil-db/parse/sql/tree"
	parseTypes "github.com/kwilteam/kwil-db/parse/types"
)

type joinWalker struct {
	tree.AstListener
	errs parseTypes.NativeErrorListener
}

func NewJoinWalker(errLis parseTypes.NativeErrorListener) tree.AstListener {
	return &joinWalker{
		AstListener: tree.NewBaseListener(),
		errs:        errLis,
	}
}

func (s *joinWalker) EnterJoinPredicate(j *tree.JoinPredicate) error {
	err := checkJoin(j)
	if err != nil {
		s.errs.NodeErr(j.GetNode(), parseTypes.ParseErrorTypeSemantic, err.Error())
	}

	return nil
}
