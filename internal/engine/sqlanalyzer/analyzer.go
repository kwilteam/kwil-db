package sqlanalyzer

import (
	"fmt"

	"github.com/kwilteam/kwil-db/internal/engine/sqlanalyzer/aggregate"
	"github.com/kwilteam/kwil-db/internal/engine/sqlanalyzer/clean"
	"github.com/kwilteam/kwil-db/internal/engine/sqlanalyzer/join"
	"github.com/kwilteam/kwil-db/internal/engine/sqlanalyzer/order"
	"github.com/kwilteam/kwil-db/internal/engine/types"
	sqlparser "github.com/kwilteam/kwil-db/parse/sql"
	"github.com/kwilteam/kwil-db/parse/sql/tree"
)

// acceptWrapper is a wrapper around a statement that implements the accepter interface
// it catches panics and returns them as errors
type acceptWrapper struct {
	inner tree.Walker
}

func (a *acceptWrapper) Walk(walker tree.AstWalker) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic while walking statement: %v", r)
		}
	}()

	return a.inner.Walk(walker)
}

// ApplyRules analyzes the given statement and returns the transformed statement.
// It parses it, and then traverses the AST with the given flags.
// It will alter the statement to make it conform to the given flags, or return an error if it cannot.
func ApplyRules(stmt string, flags VerifyFlag, tables []*types.Table) (*AnalyzedStatement, error) {
	cleanedTables, err := cleanTables(tables)
	if err != nil {
		return nil, fmt.Errorf("error cleaning tables: %w", err)
	}

	parsed, err := sqlparser.Parse(stmt)
	if err != nil {
		return nil, fmt.Errorf("error parsing statement: %w", err)
	}

	accept := &acceptWrapper{inner: parsed}

	clnr := clean.NewStatementCleaner()
	err = accept.Walk(clnr)
	if err != nil {
		return nil, fmt.Errorf("error cleaning statement: %w", err)
	}

	if flags&NoCartesianProduct != 0 {
		err := accept.Walk(join.NewJoinWalker())
		if err != nil {
			return nil, fmt.Errorf("error applying join rules: %w", err)
		}
	}

	if flags&GuaranteedOrder != 0 {
		err := accept.Walk(order.NewOrderWalker(cleanedTables))
		if err != nil {
			return nil, fmt.Errorf("error enforcing guaranteed order: %w", err)
		}
	}

	if flags&DeterministicAggregates != 0 {
		err := accept.Walk(aggregate.NewGroupByWalker())
		if err != nil {
			return nil, fmt.Errorf("error enforcing aggregate determinism: %w", err)
		}
	}

	mutative, err := isMutative(parsed)
	if err != nil {
		return nil, fmt.Errorf("error determining mutativity: %w", err)
	}

	generated, err := tree.SafeToSQL(parsed)
	if err != nil {
		return nil, fmt.Errorf("error generating SQL: %w", err)
	}

	return &AnalyzedStatement{
		stmt:     generated,
		mutative: mutative,
	}, nil
}

func cleanTables(tables []*types.Table) ([]*types.Table, error) {
	cleaned := make([]*types.Table, len(tables))

	for i, tbl := range tables {
		err := tbl.Clean()
		if err != nil {
			return nil, fmt.Errorf(`error cleaning table "%s": %w`, tbl.Name, err)
		}

		cleaned[i] = tbl.Copy()
	}

	return cleaned, nil
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

// AnalyzedStatement is a statement that has been analyzed by the analyzer
// As we progressively add more types of analysis (e.g. query pricing), we will add more fields to this struct
type AnalyzedStatement struct {
	stmt     string
	mutative bool
}

// Mutative returns true if the statement will mutate the database
func (a *AnalyzedStatement) Mutative() bool {
	return a.mutative
}

// Statements returns a new statement that is the result of the analysis
// It may contains changes to the original statement, depending on the flags that were passed in
func (a *AnalyzedStatement) Statement() string {
	return a.stmt
}
