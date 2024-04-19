package sqlanalyzer

import (
	"github.com/kwilteam/kwil-db/parse/sql/sqlanalyzer/mutative"
	"github.com/kwilteam/kwil-db/parse/sql/tree"
)

func IsMutative(stmt tree.AstWalker) (bool, error) {
	mutativityWalker := mutative.NewMutativityWalker()

	err := stmt.Walk(mutativityWalker)
	if err != nil {
		return false, err
	}

	return mutativityWalker.Mutative, nil
}
