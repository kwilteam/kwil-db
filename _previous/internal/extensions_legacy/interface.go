package extensions

import (
	"context"

	"github.com/kwilteam/kwil-extensions/client"
	"github.com/kwilteam/kwil-extensions/types"
)

var (
	// this can be overridden for testing
	ConnectFunc Connecter = extensionConnectFunc(client.NewExtensionClient)
)

type ExtensionInitializer struct {
	Extension LegacyEngineExtension
}

// CreateInstance creates an instance of the extension with the given metadata.
func (e *ExtensionInitializer) CreateInstance(ctx context.Context, metadata map[string]string) (*Instance, error) {
	metadata, err := e.Extension.Initialize(ctx, metadata)
	if err != nil {
		return nil, err
	}

	return &Instance{
		metadata:  metadata,
		extension: e.Extension,
	}, nil
}

type ExtensionClient interface {
	CallMethod(execCtx *types.ExecutionContext, method string, args ...any) ([]any, error)
	Close() error
	Initialize(ctx context.Context, metadata map[string]string) (map[string]string, error)
	GetName(ctx context.Context) (string, error)
	ListMethods(ctx context.Context) ([]string, error)
}

type Connecter interface {
	Connect(ctx context.Context, target string, opts ...client.ClientOpt) (ExtensionClient, error)
}

type extensionConnectFunc func(ctx context.Context, target string, opts ...client.ClientOpt) (*client.ExtensionClient, error)

func (e extensionConnectFunc) Connect(ctx context.Context, target string, opts ...client.ClientOpt) (ExtensionClient, error) {
	return e(ctx, target, opts...)
}
