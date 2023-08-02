package sqlanalyzer

import "github.com/kwilteam/kwil-db/pkg/engine/sqlanalyzer/mutative"

func IsMutative(stmt accepter) (bool, error) {
	mutativityWalker := mutative.NewMutativityWalker()

	err := stmt.Accept(mutativityWalker)
	if err != nil {
		return false, err
	}

	return mutativityWalker.Mutative, nil
}
