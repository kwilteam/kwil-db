package sqlanalyzer

import (
	"fmt"

	"github.com/kwilteam/kwil-db/pkg/engine/sqlanalyzer/aggregate"
	"github.com/kwilteam/kwil-db/pkg/engine/sqlanalyzer/join"
	"github.com/kwilteam/kwil-db/pkg/engine/sqlanalyzer/order"
	"github.com/kwilteam/kwil-db/pkg/engine/sqlparser"
	"github.com/kwilteam/kwil-db/pkg/engine/sqlparser/tree"
	"github.com/kwilteam/kwil-db/pkg/engine/types"
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

// ApplyRules analyzes the given statement and returns the transformed statement.
// It parses it, and then traverses the AST with the given flags.
// It will alter the statement to make it conform to the given flags, or return an error if it cannot.
func ApplyRules(stmt string, flags VerifyFlag, metadata *RuleMetadata) (*AnalyzedStatement, error) {
	parsed, err := sqlparser.Parse(stmt)
	if err != nil {
		return nil, fmt.Errorf("error parsing statement: %w", err)
	}

	accept := &acceptWrapper{inner: parsed}

	if flags&NoCartesianProduct != 0 {
		err := accept.Accept(join.NewJoinWalker())
		if err != nil {
			return nil, fmt.Errorf("error applying join rules: %w", err)
		}
	}

	if flags&GuaranteedOrder != 0 {
		err := accept.Accept(order.NewOrderWalker(metadata.Tables))
		if err != nil {
			return nil, fmt.Errorf("error enforcing guaranteed order: %w", err)
		}
	}

	if flags&DeterministicAggregates != 0 {
		err := accept.Accept(aggregate.NewGroupByWalker())
		if err != nil {
			return nil, fmt.Errorf("error enforcing aggregate determinism: %w", err)
		}
	}

	mutative, err := isMutative(parsed)
	if err != nil {
		return nil, fmt.Errorf("error determining mutativity: %w", err)
	}

	generated, err := parsed.ToSQL()
	if err != nil {
		return nil, fmt.Errorf("error generating SQL: %w", err)
	}

	return &AnalyzedStatement{
		stmt:     generated,
		mutative: mutative,
	}, nil
}

// RuleMetadata contains metadata that is needed to enforce a rule
type RuleMetadata struct {
	// Tables only needs to be set if you are guaranteeing order
	Tables []*types.Table
}

type VerifyFlag uint8

const (
	// NoCartesianProduct prevents cartesian products from being generated
	NoCartesianProduct VerifyFlag = 1 << iota
	// GuaranteedOrder provides a guarantee of deterministic ordering of the results (even if it is not explicitly specified in the query)
	GuaranteedOrder
	// DeterministicAggregates enforces that aggregates are deterministic
	DeterministicAggregates
)

const (
	AllRules = NoCartesianProduct | GuaranteedOrder | DeterministicAggregates
)

type AnalyzedStatement struct {
	stmt     string
	mutative bool
}

func (a *AnalyzedStatement) Mutative() bool {
	return a.mutative
}

func (a *AnalyzedStatement) Statement() string {
	return a.stmt
}
