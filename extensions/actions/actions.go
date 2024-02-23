// package actions allows custom actions to be registered with the engine.
package actions

import (
	"context"
	"fmt"
	"strings"

	"github.com/kwilteam/kwil-db/common"
	sql "github.com/kwilteam/kwil-db/common/sql"
)

// ExtensionInitializer initializes a new instance of an extension.
// It is called when a Kuneiform schema is deployed that calls "use <extension> {key: "value"} as <name>".
// It is also called when the node starts up, if a database is already deployed that uses the extension.
// The key/value pairs are passed as the metadata parameter.
// When initialize is called, the dataset is not yet accessible.
type ExtensionInitializer func(ctx *DeploymentContext, service *common.Service, metadata map[string]string) (ExtensionNamespace, error)

// ExtensionNamespace is a named initialized instance of an extension.
// When a Kuneiform schema calls "use <extension> as <name>", a new instance is
// created for that extension, and is accessible via <name>.
// Instances exist for the lifetime of the deployed dataset, and a single
// dataset can have multiple instances of the same extension.
type ExtensionNamespace interface {
	// Call executes the requested method of the extension.
	// It is up to the extension instance implementation to determine
	// if a method is valid, and to subsequently decode the arguments.
	// The arguments passed in as args, as well as returned, are scalar values.
	Call(scoper *ProcedureContext, app *common.App, method string, inputs []any) ([]any, error)
}

// DeploymentContext is the context for a dataset deployment transaction.
type DeploymentContext struct {
	Ctx    context.Context
	Schema *common.Schema
}

// ProcedureContext is the context for a procedure execution.
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
	// if calling an extension instead of a procedure, it will be the last used DBID.
	DBID string
	// Procedure is the Procedure identifier for the current scope.
	// if calling an extension instead of a Procedure, it will be the last used Procedure.
	Procedure string
	// Result is the result of the most recent SQL query.
	Result *sql.ResultSet
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

// Values copies the values from the scope into a map. It will also include
// contextual variables, such as the caller. If a context variable has the same
// name as a scope variable, the scope variable will be overwritten.
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

	return values
}

// NewScope creates a new procedure context for a child procedure.
// It will not inherit the values or last result from the parent.
// It will inherit the dbid and procedure from the parent.
func (p *ProcedureContext) NewScope() *ProcedureContext {
	return &ProcedureContext{
		Ctx:       p.Ctx,
		Signer:    p.Signer,
		Caller:    p.Caller,
		values:    make(map[string]any),
		DBID:      p.DBID,
		Procedure: p.Procedure,
	}
}

var registeredExtensions = make(map[string]ExtensionInitializer)

func RegisteredExtensions() map[string]ExtensionInitializer {
	return registeredExtensions
}

// RegisterExtension registers an extension with the engine.
func RegisterExtension(name string, ext ExtensionInitializer) error {
	name = strings.ToLower(name)
	if _, ok := registeredExtensions[name]; ok {
		return fmt.Errorf("extension of same name already registered:%s ", name)
	}

	registeredExtensions[name] = ext
	return nil
}

// // DEPRECATED: RegisterLegacyExtension registers an extension with the engine.
// // It provides backwards compatibility with the old extension system.
// // Use RegisterExtension instead.
// func RegisterLegacyExtension(name string, ext extensions.LegacyEngineExtension) error {
// 	return RegisterExtension(name, extensions.AdaptLegacyExtension(ext))
// }
