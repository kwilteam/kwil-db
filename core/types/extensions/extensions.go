package extensions

import (
	"context"
)

// EngineExtension is an extension that can be loaded into the engine.
// It can be used to extend the functionality of the engine.
type EngineExtension interface {
	// Name returns the name of the extension.
	// This is used to identify the extension in the engine.
	Name() string
	// Initialize initializes the extension with the given metadata.
	// It is called each time a database is deployed that uses the extension,
	// or for each database that uses the extension when the engine starts.
	// If a database initializes an extension several times, it will be called
	// each times.
	// It should return the metadata that it wants to be returned on each
	// subsequent call from the extension.
	// If it returns an error, the database will fail to deploy.
	Initialize(ctx context.Context, metadata map[string]string) (map[string]string, error)
	// Execute executes the requested method of the extension.
	// It includes the metadata that was returned from the `Initialize` method.
	Execute(scope CallContext, metadata map[string]string, method string, args ...any) ([]any, error)
}

// CallContext is a context that Kwil gives to extensions when they are called.
// It contains information about the current execution, allows extensions to
// query and mutate the database, and allows extensions to return results.
type CallContext interface {
	// Ctx returns the context of the current execution.
	Ctx() context.Context
	// Mutative returns whether or not the current execution is allowed
	// to mutate state. Normal transactions are allowed to mutate state,
	// while view transactions are not.
	Mutative() bool
	// SetResult sets a result that will be returned to the caller.
	// If the caller is a view action, this can be used to return extension
	// data.
	SetResult(result Result) error
	// Datastore returns a datastore that can be used to query and mutate
	// the databases on the network. If the scope is not mutative, the
	// datastore will error if `Execute` is called.
	Datastore() Datastore
	// Caller returns the string identifier of the caller.
	// A caller is derived from the signer's authenticator.
	// For EVM chains, this is an 0x address.
	Caller() string
	// Signer returns the address or public key of the caller.
	// It is what is used to check the signature of the incoming transaction.
	Signer() []byte
	// DBID returns the database identifier for the current scope.
	// This is will always be the database that called the extension.
	DBID() string
	// Procedure returns the procedure identifier for the current scope.
	// This will always be the procedure that called the extension.
	Procedure() string

	// Values returns environment variables that are available to the extension.
	// These are variables such as @caller, which is the string identifier of the caller.
	// They should be added to the arguments passed to queries.
	Values() map[string]any
}

// Result is a result that is returned from a query.
// It should be used in the following way:
//
// columns := res.Columns()
//
// var res extensions.Result
//
//	for {
//		rowReturned, err := res.Next()
//		if err != nil {
//			return err
//		}
//
//		if !rowReturned {
//			break
//		}
//
//		values, err := res.Values()
//		if err != nil {
//			return err
//		}
//
//		// values[0] is the value for columns[0]
//		// values[1] is the value for columns[1]
//	}
type Result interface {
	// Columns returns the columns that are returned by the result.
	Columns() []string
	// Next returns whether or not there is another row to be read.
	// If there is, it will increment the row index, which can be used
	// to get the values for the current row.
	Next() (rowReturned bool, err error)
	// Values returns the values for the current row.
	// The values are returned in the same order as the columns.
	Values() ([]any, error)
}

// Datastore can be used to query and mutate databases.
// All statements will be parsed and made deterministic before they are executed.
type Datastore interface {
	// Query executes a query against a reader connection.
	// It will not read uncommitted data, and cannot be used to write data.
	Query(ctx context.Context, dbid string, stmt string, params map[string]any) (Result, error)

	// Execute executes a query against a writer connection.
	// It can also be used to read uncommitted data.
	// If called on a read-only context (such as in a view action), it will return an error.
	Execute(ctx context.Context, dbid string, stmt string, params map[string]any) (Result, error)
}
