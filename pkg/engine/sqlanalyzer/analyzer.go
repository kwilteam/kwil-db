package sqlanalyzer

import "github.com/kwilteam/kwil-db/pkg/engine/sqlparser/tree"

type Analyzer struct{}

func Analyze(stmt tree.Ast, flags VerifyFlag) (*Statement, error) {
	panic("not implemented")
	return &Statement{}, nil
}

type VerifyFlag uint8

const (
	NoCarterianProduct VerifyFlag = 1 << iota
	GuaranteedOrder
	Deterministic
)

type Statement struct {
	VerifiedFor VerifyFlag

	SQL string
}
