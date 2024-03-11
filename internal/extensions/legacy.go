package extensions

import (
	"context"

	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/extensions/precompiles"
)

// LegacyEngineExtension is an extension that can be loaded into the engine.
// It can be used to extend the functionality of the engine.
type LegacyEngineExtension interface {
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
	Execute(scope *precompiles.ProcedureContext, metadata map[string]string, method string, args ...any) ([]any, error)
}

// AdapterFunc is a function that adapts a LegacyEngineExtension to an InitializeFunc.
func AdaptLegacyExtension(ext LegacyEngineExtension) precompiles.Initializer {
	return func(ctx *precompiles.DeploymentContext, service *common.Service, metadata map[string]string) (precompiles.Instance, error) {

		m, err := ext.Initialize(ctx.Ctx, metadata)
		if err != nil {
			return nil, err
		}

		return &legacyExtensionAdapter{
			ext:      ext,
			metadata: m,
		}, nil
	}
}

// legacyExtensionAdapter adapts a LegacyEngineExtension to an EngineExtension.
type legacyExtensionAdapter struct {
	ext      LegacyEngineExtension
	metadata map[string]string
}

var _ precompiles.Instance = (*legacyExtensionAdapter)(nil)

func (l *legacyExtensionAdapter) Call(scope *precompiles.ProcedureContext, app *common.App, method string, args []any) ([]any, error) {
	return l.ext.Execute(scope, l.metadata, method, args...)
}
