package execution

import (
	"context"
	"strings"

	"github.com/kwilteam/kwil-db/internal/engine/types"
	sql "github.com/kwilteam/kwil-db/internal/sql"
)

// DeploymentContext is the context for a dataset deployment transaction.
type DeploymentContext struct {
	Ctx    context.Context
	Schema *types.Schema
}

// ProcedureContext is the context for a procedure execution.
type ProcedureContext struct {
	// Ctx is the context of the current execution.
	Ctx context.Context
	// Signer is the address or public key of the caller.
	Signer []byte
	// Caller is the string identifier of the signer.
	Caller string
	// globalCtx is the global context of the current execution.
	globalCtx *GlobalContext
	// values are the variables that are available to the execution.
	values map[string]any

	// DBID is the database identifier for the current scope.
	// if calling an extension instead of a procedure, it will be the last used DBID.
	DBID string
	// Procedure is the Procedure identifier for the current scope.
	// if calling an extension instead of a Procedure, it will be the last used Procedure.
	Procedure string
	// Result is the result of the most recent SQL query.
	Result *sql.ResultSet
	// Mutative is whether the execution can mutate state.
	Mutative bool
}

// SetValue sets a value in the scope.
// Values are case-insensitive.
// If a value for the key already exists, it will be overwritten.
func (p *ProcedureContext) SetValue(key string, value any) {
	p.values[strings.ToLower(key)] = value
}

// Values copies the values from the scope into a map.
// It will also include contextual variables, such as the caller.
// If a context variable has the same name as a scope variable, the scope variable will be overwritten.
func (p ProcedureContext) Values() map[string]any {
	values := make(map[string]any)
	for k, v := range p.values {
		values[strings.ToLower(k)] = v
	}

	// set environment variables
	values["@caller"] = p.Caller

	return values
}

// Dataset returns the dataset with the given identifier.
// If the dataset does not exist, it will return an error.
func (p ProcedureContext) Dataset(dbid string) (*Dataset, error) {
	dataset, ok := p.globalCtx.datasets[dbid]
	if !ok {
		return nil, types.ErrDatasetNotFound
	}

	return dataset, nil
}

// NewScope creates a new procedure context for a child procedure.
// It will not inherit the values or last result from the parent.
// It will inherit the dbid and procedure from the parent.
func (p ProcedureContext) NewScope() *ProcedureContext {
	return &ProcedureContext{
		Ctx:       p.Ctx,
		Signer:    p.Signer,
		Caller:    p.Caller,
		globalCtx: p.globalCtx,
		values:    make(map[string]any),
		DBID:      p.DBID,
		Procedure: p.Procedure,
		Mutative:  p.Mutative,
	}
}
