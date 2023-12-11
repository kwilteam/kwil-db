// package inregration_test contains full engine integration tests
package integration_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/kwilteam/kwil-db/internal/engine/execution"
	"github.com/kwilteam/kwil-db/internal/sql/registry"
	"github.com/kwilteam/kwil-db/internal/sql/sqlite"
)

// setup sets up the global context and registry for the tests
func setup(t *testing.T) (global *execution.GlobalContext, reg *registry.Registry, err error) {
	t.Helper()

	cleanup := func() {
		err := os.RemoveAll("./tmp")
		if err != nil {
			t.Fatal(err)
		}
	}
	// run cleanup first to ensure we don't have any leftover files
	cleanup()

	ctx := context.Background()

	reg, err = registry.NewRegistry(ctx, func(ctx context.Context, dbid string, create bool) (registry.Pool, error) {
		return sqlite.NewPool(ctx, dbid, 1, 1, true)
	}, "./tmp", registry.WithReaderWaitTimeout(time.Millisecond*100))
	if err != nil {
		return nil, nil, err
	}

	global, err = execution.NewGlobalContext(ctx, reg, map[string]execution.NamespaceInitializer{
		"math": (&mathInitializer{}).initialize,
	})
	if err != nil {
		return nil, nil, err
	}

	t.Cleanup(func() {
		err = reg.Close()
		if err != nil {
			t.Fatal(err)
		}

		cleanup()
	})

	return global, reg, nil
}

// mocks a namespace initializer
type mathInitializer struct {
	vals map[string]string
}

func (m *mathInitializer) initialize(_ context.Context, mp map[string]string) (execution.Namespace, error) {
	m.vals = mp

	_, ok := m.vals["fail"]
	if ok {
		return nil, fmt.Errorf("mock extension failed to initialize")
	}

	return &mathExt{}, nil
}

type mathExt struct{}

var _ execution.Namespace = &mathExt{}

func (m *mathExt) Call(caller *execution.ScopeContext, method string, inputs []any) ([]any, error) {
	if method != "add" {
		return nil, fmt.Errorf("unknown method: %s", method)
	}

	if len(inputs) != 2 {
		return nil, fmt.Errorf("expected 2 inputs, got %d", len(inputs))
	}

	a, ok := inputs[0].(int64)
	if !ok {
		return nil, fmt.Errorf("expected int64, got %T", inputs[0])
	}

	b, ok := inputs[1].(int64)
	if !ok {
		return nil, fmt.Errorf("expected int64, got %T", inputs[1])
	}

	return []any{a + b}, nil
}
