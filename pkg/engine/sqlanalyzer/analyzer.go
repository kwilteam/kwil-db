package sqlanalyzer

import (
	"fmt"

	"github.com/kwilteam/kwil-db/pkg/engine/sqlanalyzer/aggregate"
	"github.com/kwilteam/kwil-db/pkg/engine/sqlanalyzer/join"
	"github.com/kwilteam/kwil-db/pkg/engine/sqlparser/tree"
)

type accepter interface {
	Accept(walker tree.Walker) error
}

// acceptWrapper is a wrapper around a statement that implements the accepter interface
// it catches panics and returns them as errors
type acceptWrapper struct {
	inner accepter
}

func (a *acceptWrapper) Accept(walker tree.Walker) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic while walking statement: %v", r)
		}
	}()

	return a.inner.Accept(walker)
}

// ApplyRules analyzes the given statement and returns the statement.
// NOTE: this can change the statement, so it is recommended to clone the statement before analyzing it
// if you want to keep the original statement.
func ApplyRules(stmt accepter, flags VerifyFlag) (accepter, error) {
	accept := &acceptWrapper{inner: stmt}

	if flags&NoCartesianProduct != 0 {
		err := accept.Accept(join.NewJoinWalker())
		if err != nil {
			return nil, fmt.Errorf("error applying join rules: %w", err)
		}
	}

	if flags&DeterministicAggregates != 0 {
		err := accept.Accept(aggregate.NewGroupByWalker())
		if err != nil {
			return nil, fmt.Errorf("error enforcing aggregate determinism: %w", err)
		}
	}
}

type VerifyFlag uint8

const (
	NoCartesianProduct VerifyFlag = 1 << iota
	GuaranteedOrder
	DeterministicAggregates
)

type Statement struct {
	VerifiedFor VerifyFlag

	SQL string
}
