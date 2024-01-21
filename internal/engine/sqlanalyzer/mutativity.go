package sqlanalyzer

import "github.com/kwilteam/kwil-db/internal/engine/sqlanalyzer/mutative"

func isMutative(stmt accepter) (bool, error) {
	mutativityWalker := mutative.NewMutativityWalker()

	err := stmt.Walk(mutativityWalker)
	if err != nil {
		return false, err
	}

	return mutativityWalker.Mutative, nil
}
