//go:build precompiles_math || ext_test

package extensions_test

import (
	"context"
	"errors"
	"math/big"
	"testing"

	actions "github.com/kwilteam/kwil-db/extensions/precompiles"
	extensions "github.com/kwilteam/kwil-db/internal/extensions"
	"github.com/kwilteam/kwil-extensions/client"
	"github.com/kwilteam/kwil-extensions/types"
	"github.com/stretchr/testify/assert"
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

func (m *mockClient) Initialize(ctx context.Context, metadata map[string]string) (map[string]string, error) {
	return metadata, nil
}

func Test_LocalExtension(t *testing.T) {
	ctx := context.Background()
	callCtx := &actions.ProcedureContext{
		Ctx: ctx,
	}

	metadata := map[string]string{
		"round": "down",
	}
	incorrectMetadata := map[string]string{
		"roundoff": "down",
	}

	initializer := &extensions.ExtensionInitializer{
		Extension: &legacyAdapter{
			init: actions.InitializeMath,
		},
	}

	// Create instance with correct metadata
	instance1, err := initializer.CreateInstance(ctx, metadata)
	assert.NoError(t, err)

	result, err := instance1.Execute(callCtx, "add", 1, 2)
	assert.NoError(t, err)
	assert.Equal(t, int(3), result[0])

	result, err = instance1.Execute(callCtx, "add", 1.2, 2.3)
	assert.Error(t, err)

	result, err = instance1.Execute(callCtx, "divide", 1, 2)
	assert.NoError(t, err)
	assert.Equal(t, big.NewInt(0), result[0]) // 1/2 rounded down to 0

	// Create instance with incorrect metadata, uses defaults
	instance2, err := initializer.CreateInstance(ctx, incorrectMetadata)
	assert.NoError(t, err)
	updatedMetadata := instance2.Metadata()
	assert.Equal(t, updatedMetadata["round"], "up")

	result, err = instance2.Execute(callCtx, "divide", 1, 2)
	assert.NoError(t, err)
	assert.Equal(t, big.NewInt(1), result[0]) // 1/2 rounded up -> 1
}

// this only works specifically for this test, and should not be used generally.
// legacy extensions are forwards compatible, but current extensions are not backwards compatible,
// and the extension tested here is not a legacy extension.
type legacyAdapter struct {
	init        actions.Initializer
	ext         actions.Instance
	initialized bool
}

func (l *legacyAdapter) Execute(scope *actions.ProcedureContext, metadata map[string]string, method string, args ...any) ([]any, error) {
	if !l.initialized {
		return nil, errors.New("not initialized")
	}

	return l.ext.Call(scope, nil, method, args)
}

func (l *legacyAdapter) Initialize(ctx context.Context, metadata map[string]string) (map[string]string, error) {
	ext, err := l.init(&actions.DeploymentContext{
		Ctx: ctx,
	}, nil, metadata)
	if err != nil {
		return nil, err
	}

	l.ext = ext
	l.initialized = true

	return metadata, nil
}

func Test_RemoteExtension(t *testing.T) {
	ctx := context.Background()
	ext := extensions.New("local:8080")
	callCtx := &actions.ProcedureContext{
		Ctx: ctx,
	}

	err := ext.Connect(ctx)
	if err != nil {
		t.Fatal(err)
	}
	initializer := &extensions.ExtensionInitializer{
		Extension: ext,
	}
	instance, err := initializer.CreateInstance(ctx, map[string]string{
		"token_address":  "0x12345",
		"wallet_address": "0xabcd",
	})
	if err != nil {
		t.Fatal(err)
	}

	results, err := instance.Execute(callCtx, "method1", "0x12345")
	if err != nil {
		t.Fatal(err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
}
