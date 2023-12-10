package execution

import (
	"context"
	"errors"

	"github.com/kwilteam/kwil-db/core/types/extensions"
	"github.com/kwilteam/kwil-db/internal/engine/sqlanalyzer"
	sql "github.com/kwilteam/kwil-db/internal/sql"
)

// executionContext is the context for an execution.
// It is scoped to the lifetime of a single execution.
type executionContext struct {
	Ctx context.Context
	// mutative is whether the execution can mutate state.
	mutative bool

	// signer is the address or public key of the wallet that signed the transaction.
	signer []byte
	// caller is the string identifier of the calling signer.
	// it is derived from the signer's address or public key.
	caller string

	// FinalResult is the most recent SQL query result.
	FinalResult *sql.ResultSet

	// global is the global context.
	global *GlobalContext
}

// NewScope creates a new scope context.
// It will include the execution context.
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

	// dbid is the database identifier for the current scope.
	// if calling an extension instead of a procedure, it will be the last used dbid.
	dbid string
	// procedure is the procedure identifier for the current scope.
	// if calling an extension instead of a procedure, it will be the last used procedure.
	procedure string

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

// SetResult sets the result of the most recent SQL query.
func (s *ScopeContext) SetResult(result *sql.ResultSet) {
	s.execution.FinalResult = result
}

// Caller returns the caller identifier.
func (s *ScopeContext) Caller() string {
	return s.execution.caller
}

// Signer returns the address or public key of the wallet that signed the transaction.
// It returns a copy of the signer, and nil if the execution is not signed.
func (s *ScopeContext) Signer() []byte {
	var signer []byte
	if s.execution.signer == nil {
		return nil
	}
	signer = make([]byte, len(s.execution.signer))
	copy(signer, s.execution.signer)
	return signer
}

// Values copies the values from the scope into a map.
func (s *ScopeContext) Values() map[string]any {
	values := make(map[string]any)
	for k, v := range s.values {
		values[k] = v
	}

	// set environment variables
	values["@caller"] = s.Caller()

	return values
}

// Mutative returns whether the execution can mutate state.
func (s *ScopeContext) Mutative() bool {
	return s.execution.mutative
}

// Ctx returns the running context.
func (s *ScopeContext) Ctx() context.Context {
	return s.execution.Ctx
}

// New creates a new scope context.
// It will inherit the execution context from the parent.
func (s *ScopeContext) NewScope() *ScopeContext {
	return &ScopeContext{
		values:    make(map[string]any),
		execution: s.execution,
	}
}

// DBID returns the database identifier for the current scope.
func (s *ScopeContext) DBID() string {
	return s.dbid
}

// Procedure returns the procedure identifier for the current scope.
func (s *ScopeContext) Procedure() string {
	return s.procedure
}

// ExtensionScoper is a scope context that implements the extensions.ScopeContext interface.
type ExtensionScoper struct {
	*ScopeContext
}

var _ extensions.CallContext = (*ExtensionScoper)(nil)

// NewScope creates a new scope context.
func (e *ExtensionScoper) NewScope() extensions.CallContext {
	return &ExtensionScoper{
		ScopeContext: e.ScopeContext.NewScope(),
	}
}

// Query executes a query against a reader connection
func (e *ExtensionScoper) Datastore() extensions.Datastore {
	return &extensionDatastore{
		scope: e.ScopeContext,
	}
}

// SetResult sets the result of the most recent SQL query.
func (e *ExtensionScoper) SetResult(result extensions.Result) error {
	res := &sql.ResultSet{
		ReturnedColumns: result.Columns(),
	}

	for {
		rowReturned, err := result.Next()
		if err != nil {
			return err
		}

		if !rowReturned {
			break
		}

		values, err := result.Values()
		if err != nil {
			return err
		}

		res.Rows = append(res.Rows, values)
	}

	e.ScopeContext.SetResult(res)
	return nil
}

var ErrReadOnly = errors.New("cannot write to  datastore: context is read-only")

type extensionDatastore struct {
	scope *ScopeContext
}

var _ extensions.Datastore = (*extensionDatastore)(nil)

func (e *extensionDatastore) Query(ctx context.Context, dbid string, stmt string, params map[string]any) (extensions.Result, error) {
	var parsedStmt *sqlanalyzer.AnalyzedStatement
	dataset, ok := e.scope.execution.global.datasets[dbid]
	if !ok {
		return nil, errors.New("unknown dataset")
	}

	var err error
	if e.scope.Mutative() {
		parsedStmt, err = sqlanalyzer.ApplyRules(stmt, sqlanalyzer.AllRules, dataset.schema.Tables)
		if err != nil {
			return nil, err
		}
	} else {
		parsedStmt, err := sqlanalyzer.ApplyRules(stmt, sqlanalyzer.NoCartesianProduct, dataset.schema.Tables)
		if err != nil {
			return nil, err
		}
		if parsedStmt.Mutative() {
			return nil, ErrReadOnly
		}
	}

	return e.scope.execution.global.datastore.Query(ctx, dbid, parsedStmt.Statement(), params)
}

func (e *extensionDatastore) Execute(ctx context.Context, dbid string, stmt string, params map[string]any) (extensions.Result, error) {
	if !e.scope.Mutative() {
		return nil, ErrReadOnly
	}

	dataset, ok := e.scope.execution.global.datasets[dbid]
	if !ok {
		return nil, errors.New("unknown dataset")
	}

	parsedStmt, err := sqlanalyzer.ApplyRules(stmt, sqlanalyzer.AllRules, dataset.schema.Tables)
	if err != nil {
		return nil, err
	}

	return e.scope.execution.global.datastore.Execute(ctx, dbid, parsedStmt.Statement(), params)
}
