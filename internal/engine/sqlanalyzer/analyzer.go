package sqlanalyzer

import (
	"fmt"

	"github.com/kwilteam/kwil-db/internal/engine/sqlanalyzer/aggregate"
	"github.com/kwilteam/kwil-db/internal/engine/sqlanalyzer/clean"
	"github.com/kwilteam/kwil-db/internal/engine/sqlanalyzer/join"
	"github.com/kwilteam/kwil-db/internal/engine/sqlanalyzer/order"
	"github.com/kwilteam/kwil-db/internal/engine/sqlanalyzer/parameters"
	"github.com/kwilteam/kwil-db/internal/engine/sqlanalyzer/schema"
	"github.com/kwilteam/kwil-db/internal/engine/types"
	sqlparser "github.com/kwilteam/kwil-db/parse/sql"
	"github.com/kwilteam/kwil-db/parse/sql/tree"
)

type Accepter interface {
	Accept(walker tree.Walker) error
}

// AcceptRecoverer is a wrapper around a statement that implements the accepter interface
// it catches panics and returns them as errors
type AcceptRecoverer struct {
	Accepter
}

func NewAcceptRecoverer(a Accepter) *AcceptRecoverer {
	return &AcceptRecoverer{a}
}

func (a *AcceptRecoverer) Accept(walker tree.Walker) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic while walking statement: %v", r)
		}
	}()

	return a.Accepter.Accept(walker)
}

// ApplyRules analyzes the given statement and returns the transformed statement.
// It parses it, and then traverses the AST with the given flags.
// It will alter the statement to make it conform to the given flags, or return an error if it cannot.
// All tables will target the pgSchemaName schema.
func ApplyRules(stmt string, flags VerifyFlag, tables []*types.Table, pgSchemaName string) (*AnalyzedStatement, error) {
	cleanedTables, err := cleanTables(tables)
	if err != nil {
		return nil, fmt.Errorf("error cleaning tables: %w", err)
	}

	parsed, err := sqlparser.Parse(stmt)
	if err != nil {
		return nil, fmt.Errorf("error parsing statement: %w", err)
	}

	accept := &AcceptRecoverer{parsed}

	clnr := clean.NewStatementCleaner()
	err = accept.Accept(clnr)
	if err != nil {
		return nil, fmt.Errorf("error cleaning statement: %w", err)
	}

	schemaWalker := schema.NewSchemaWalker(pgSchemaName)
	err = accept.Accept(schemaWalker)
	if err != nil {
		return nil, fmt.Errorf("error applying schema rules: %w", err)
	}

	if flags&NoCartesianProduct != 0 {
		err := accept.Accept(join.NewJoinWalker())
		if err != nil {
			return nil, fmt.Errorf("error applying join rules: %w", err)
		}
	}

	if flags&GuaranteedOrder != 0 {
		err := accept.Accept(order.NewOrderWalker(cleanedTables))
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

	orderedParams := make([]string, 0)
	if flags&ReplaceNamedParameters != 0 {
		paramVisitor := parameters.NewParametersVisitor()
		err := accept.Accept(paramVisitor)
		if err != nil {
			return nil, fmt.Errorf("error replacing named parameters: %w", err)
		}
		orderedParams = paramVisitor.OrderedParameters
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
		Statement:      generated,
		Mutative:       mutative,
		HasTableRefs:   schemaWalker.SetCount > 0,
		ParameterOrder: orderedParams,
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
	// ReplaceNamedParameters replaces named parameters with numbered parameters
	ReplaceNamedParameters
)

const (
	AllRules = NoCartesianProduct | GuaranteedOrder | DeterministicAggregates | ReplaceNamedParameters
)

// AnalyzedStatement is a statement that has been analyzed by the analyzer
// As we progressively add more types of analysis (e.g. query pricing), we will add more fields to this struct
type AnalyzedStatement struct {
	// Statement is the rewritten SQL statement, with the correct rules applied
	Statement string
	// Mutative indicates if the statement mutates state.
	// If true, then the statement cannot run in a read-only transaction.
	Mutative bool
	// HasTableRefs indicates if the statement included tables IFF the
	// NamedParametersVisitor was run on the AST after parsing. These tables
	// would have had a schema prefixed by the walker. This can indicate if the
	// statement alone is not likely to provide type (OID) information by
	// preparing the statement with the database backend.
	HasTableRefs bool
	// ParameterOrder is a list of the parameters in the order they appear in the statement.
	// This is set if the ReplaceNamedParameters flag is set.
	// For example, if the statement is "SELECT * FROM table WHERE id = $id AND name = @caller",
	// then the parameter order would be ["$id", "@caller"]
	ParameterOrder []string
}
