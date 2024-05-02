package sqlanalyzer

import (
	"github.com/kwilteam/kwil-db/parse/sql/sqlanalyzer/mutative"
	"github.com/kwilteam/kwil-db/parse/sql/tree"
)

// IsMutative returns true if the statement is mutative, false otherwise.
// It doesn't need an error listener since it doesn't actually perform any syntax/
// semantic analysis, it simply checks if there is an INSERT, UPDATE, or DELETE.
func IsMutative(stmt tree.AstWalker) (bool, error) {
	mutativityWalker := mutative.NewMutativityWalker()

	err := stmt.Walk(mutativityWalker)
	if err != nil {
		return false, err
	}

	return mutativityWalker.Mutative, nil
}
