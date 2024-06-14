// package actions allows custom actions to be registered with the
// engine.
package precompiles

import (
	"context"
	"fmt"
	"strings"

	"github.com/kwilteam/kwil-db/common"
	sql "github.com/kwilteam/kwil-db/common/sql"
	"github.com/kwilteam/kwil-db/core/types"
)

// Initializer initializes a new instance of a precompile.
// It is called when a Kuneiform schema is deployed that calls
// "use <precompile> {key: "value"} as <name>". It is also called
// when the node starts up, if a database is already deployed that
// uses the precompile. The key/value pairs are passed as the
// metadata parameter. When initialize is called, the dataset is not
// yet accessible.
type Initializer func(ctx *DeploymentContext, service *common.Service, metadata map[string]string) (Instance, error)

// Instance is a named initialized instance of a precompile. It is
// returned from the precompile initialization, as specified by the
// Initializer. It will exist for the lifetime of the deployed
// dataset, and a single dataset can have multiple instances of the
// same precompile.
type Instance interface {
	// Call executes the requested method of the precompile. It is up
	// to the instance implementation to determine if a method is
	// valid, and to subsequently decode the arguments. The arguments
	// passed in as args, as well as returned, are scalar values.
	Call(scoper *ProcedureContext, app *common.App, method string, inputs []any) ([]any, error)
}

// DeploymentContext is the context for a dataset deployment
// transaction.
type DeploymentContext struct {
	Ctx    context.Context
	Schema *types.Schema
}

// ProcedureContext is the context for a procedure and action execution.
type ProcedureContext struct {
	// Ctx is the context of the current execution.
	Ctx context.Context
	// Signer is the address or public key of the caller.
	Signer []byte
	// Caller is the string identifier of the signer.
	Caller string

	// values are the variables that are available to the execution.
	values map[string]any // note: bind $args or @caller

	// DBID is the database identifier for the current scope.
	// if calling a precompile instance instead of a procedure, it
	// will be the last used DBID.
	DBID string

	// TxID is the hash of the transaction that the scope
	// was called in.
	TxID string

	// Procedure is the Procedure identifier for the current scope.
	// if calling a precompile instance instead of a Procedure, it
	// will be the last used Procedure.
	Procedure string
	// Result is the result of the most recent SQL query.
	Result *sql.ResultSet
	// Height is the block height of the current execution.
	Height int64

	// StackDepth tracks the current depth of the procedure call stack. It is
	// incremented each time a procedure calls another procedure.
	StackDepth int
	// UsedGas is the amount of gas used in the current execution.
	UsedGas uint64
}

// SetValue sets a value in the scope.
// Values are case-insensitive.
// If a value for the key already exists, it will be overwritten.
func (p *ProcedureContext) SetValue(key string, value any) {
	if p.values == nil {
		p.values = make(map[string]any)
	}
	p.values[strings.ToLower(key)] = value
}

// Values copies the values from the scope into a map. It will also
// include contextual variables, such as the caller. If a context
// variable has the same name as a scope variable, the scope variable
// will be overwritten.
func (p *ProcedureContext) Values() map[string]any {
	if p.values == nil {
		p.values = make(map[string]any)
	}

	values := make(map[string]any)
	for k, v := range p.values {
		values[strings.ToLower(k)] = v
	}

	// set environment variables
	values["@caller"] = p.Caller
	values["@txid"] = p.TxID
	values["@signer"] = p.Signer
	values["@height"] = p.Height

	return values
}

// NewScope creates a new procedure context for a child procedure.
// It will not inherit the values or last result from the parent.
// It will inherit the dbid, procedure, and stack depth from the parent.
func (p *ProcedureContext) NewScope() *ProcedureContext {
	return &ProcedureContext{
		Ctx:        p.Ctx,
		Signer:     p.Signer,
		Caller:     p.Caller,
		values:     make(map[string]any),
		DBID:       p.DBID,
		TxID:       p.TxID,
		Procedure:  p.Procedure,
		StackDepth: p.StackDepth,
		UsedGas:    p.UsedGas,
		Height:     p.Height,
	}
}

var registeredPrecompiles = make(map[string]Initializer)

func RegisteredPrecompiles() map[string]Initializer {
	return registeredPrecompiles
}

// RegisterPrecompile registers a precompile extension with the
// engine.
func RegisterPrecompile(name string, ext Initializer) error {
	name = strings.ToLower(name)
	if _, ok := registeredPrecompiles[name]; ok {
		return fmt.Errorf("precompile of same name already registered:%s ", name)
	}

	registeredPrecompiles[name] = ext
	return nil
}
