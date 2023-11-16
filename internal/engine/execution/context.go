package execution

import (
	"context"

	sql "github.com/kwilteam/kwil-db/internal/sql"
)

// Scoper is an interface that can create a new scope from a parent scope.
type Scoper interface {
	// NewScope creates a new scope context.
	NewScope() *ScopeContext
}

// executionContext is the context for an execution.
// It is scoped to the lifetime of a single execution.
type executionContext struct {
	Ctx context.Context

	// Caller is the user identifier of the sender of the transaction.
	Caller []byte

	// PublicKey is the public key of the sender of the transaction.
	// TODO: this will get removed once we have make the auth updates
	PublicKey []byte

	// FinalResult is the most recent SQL query result.
	FinalResult *sql.ResultSet

	// Mutative indicates whether the execution can mutate state.
	Mutative bool
}

var _ Scoper = (*executionContext)(nil)

// NewScope creates a new scope context.
// It will include the execution context.
// It implements the Scoper interface.
func (e *executionContext) NewScope() *ScopeContext {
	return &ScopeContext{
		values:    make(map[string]any),
		execution: e,
	}
}

// ScopeContext is the context that encapsulates a bounded set of variables.
// These variables are only accessible within the scope.
type ScopeContext struct {
	// values are the variables that are available to the execution.
	values map[string]any

	// execution is the context for the entire execution lifetime across all scopes.
	execution *executionContext
}

// Set sets a value in the scope.
func (s *ScopeContext) Set(key string, value any) {
	s.values[key] = value
}

// Get gets a value from the scope.
func (s *ScopeContext) Get(key string) (any, bool) {
	value, ok := s.values[key]
	return value, ok
}

// Signer returns the public key of the sender of the transaction.
func (s *ScopeContext) Signer() []byte {
	return s.execution.Caller
}

// SetResult sets the result of the most recent SQL query.
func (s *ScopeContext) SetResult(result *sql.ResultSet) {
	s.execution.FinalResult = result
}

// Values copies the values from the scope into a map.
func (s *ScopeContext) Values() map[string]any {
	values := make(map[string]any)
	for k, v := range s.values {
		values[k] = v
	}

	// set environment variables
	values["@caller"] = s.execution.Caller

	return values
}

// Mutative returns whether the execution can mutate state.
func (s *ScopeContext) Mutative() bool {
	return s.execution.Mutative
}

// Ctx returns the running context.
func (s *ScopeContext) Ctx() context.Context {
	return s.execution.Ctx
}

var _ Scoper = (*ScopeContext)(nil)

// New creates a new scope context.
// It will inherit the execution context from the parent.
func (s *ScopeContext) NewScope() *ScopeContext {
	return &ScopeContext{
		values:    make(map[string]any),
		execution: s.execution,
	}
}
