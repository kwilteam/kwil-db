package interpreter

import (
	"context"

	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/extensions/precompiles"
)

// TODO: the type definitions in this file should be moved to the extensions package.
// I am keeping them here for now until we merge this with nodev2, as to avoid merge conflicts.
// Initializer initializes a new instance of a precompile.
// It is called when a Kuneiform schema is deployed that calls
// "use <precompile> {key: "value"} as <name>". It is also called
// when the node starts up, if a database is already deployed that
// uses the precompile. The key/value pairs are passed as the
// metadata parameter. When initialize is called, the dataset is not
// yet accessible.
type Initializer func(ctx context.Context, service *common.Service, metadata map[string]string) (Instance, error)

// Instance is a named initialized instance of a precompile. It is
// returned from the precompile initialization, as specified by the
// Initializer. It will exist for the lifetime of the deployed
// dataset, and a single dataset can have multiple instances of the
// same precompile.
type Instance interface {
	OnDeploy(scoper *precompiles.ProcedureContext, app *common.App, metadata map[string]string) error
	// Call executes the requested method of the precompile. It is up
	// to the instance implementation to determine if a method is
	// valid, and to subsequently decode the arguments. The arguments
	// passed in as args, as well as returned, are scalar values.
	Call(scoper *precompiles.ProcedureContext, app *common.App, method string, inputs []any) ([]any, error)
}

/*
two cases:

- node is running, extension gets used

1. extension gets used
2. initialize gets called
3. OnDeploy gets called

- extension has already been used, node restarts
1. Node starts
2. initialize gets called
*/
