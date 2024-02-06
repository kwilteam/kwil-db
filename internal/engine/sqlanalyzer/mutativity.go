package sqlanalyzer

import "github.com/kwilteam/kwil-db/internal/engine/sqlanalyzer/mutative"

func isMutative(stmt Accepter) (bool, error) {
	mutativityWalker := mutative.NewMutativityWalker()

	err := stmt.Accept(mutativityWalker)
	if err != nil {
		return false, err
	}

	return mutativityWalker.Mutative, nil
}
