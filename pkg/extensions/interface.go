package extensions

import (
	"context"

	"github.com/kwilteam/kwil-extensions/client"
	"github.com/kwilteam/kwil-extensions/types"
)

type ExtensionClient interface {
	CallMethod(execCtx *types.ExecutionContext, method string, args ...any) ([]any, error)
	Close() error
	GetMetadata(ctx context.Context) (map[string]string, error)
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

var (
	// this can be overridden for testing
	ConnectFunc Connecter = extensionConnectFunc(client.NewExtensionClient)
)
