package extensions_test

import (
	"context"

	"github.com/kwilteam/kwil-db/pkg/extensions"
	"github.com/kwilteam/kwil-extensions/client"
	"github.com/kwilteam/kwil-extensions/types"
)

func init() {
	extensions.ConnectFunc = connecterFunc(mockConnect)
}

// this is used to inject a mock connection function for testing
func mockConnect(ctx context.Context, target string, opts ...client.ClientOpt) (extensions.ExtensionClient, error) {
	return &mockClient{}, nil
}

type connecterFunc func(ctx context.Context, target string, opts ...client.ClientOpt) (extensions.ExtensionClient, error)

func (m connecterFunc) Connect(ctx context.Context, target string, opts ...client.ClientOpt) (extensions.ExtensionClient, error) {
	return &mockClient{}, nil
}

// mockClient implements the ExtensionClient interface
type mockClient struct{}

func (m *mockClient) GetName(ctx context.Context) (string, error) {
	return "mock", nil
}

func (m *mockClient) CallMethod(ctx *types.ExecutionContext, method string, args ...any) ([]any, error) {
	return []any{"val1", 2}, nil
}

func (m *mockClient) Close() error {
	return nil
}

func (m *mockClient) ListMethods(ctx context.Context) ([]string, error) {
	return []string{"method1", "method2"}, nil
}

func (m *mockClient) GetMetadata(ctx context.Context) (map[string]string, error) {
	return map[string]string{
		"token_address":  "0x1234", // not required
		"wallet_address": "",       // required
	}, nil
}
