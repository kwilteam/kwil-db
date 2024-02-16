package sqlanalyzer

import (
	"github.com/kwilteam/kwil-db/internal/engine/sqlanalyzer/mutative"
	"github.com/kwilteam/kwil-db/parse/sql/tree"
)

func isMutative(stmt tree.Accepter) (bool, error) {
	mutativityWalker := mutative.NewMutativityWalker()

	err := stmt.Accept(mutativityWalker)
	if err != nil {
		return false, err
	}

	return mutativityWalker.Mutative, nil
}
