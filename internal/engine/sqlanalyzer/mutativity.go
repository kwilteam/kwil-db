package sqlanalyzer

import (
	"github.com/kwilteam/kwil-db/internal/engine/sqlanalyzer/mutative"
	"github.com/kwilteam/kwil-db/parse/sql/tree"
)

func isMutative(stmt tree.Walker) (bool, error) {
	mutativityWalker := mutative.NewMutativityWalker()

	err := stmt.Walk(mutativityWalker)
	if err != nil {
		return false, err
	}

	return mutativityWalker.Mutative, nil
}
